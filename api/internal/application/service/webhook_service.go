package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// WebhookService handles webhook operations
type WebhookService struct {
	webhookRepo  repository.WebhookRepository
	deliveryRepo repository.WebhookDeliveryRepository
	httpClient   *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	webhookRepo repository.WebhookRepository,
	deliveryRepo repository.WebhookDeliveryRepository,
) *WebhookService {
	return &WebhookService{
		webhookRepo:  webhookRepo,
		deliveryRepo: deliveryRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Create creates a new webhook
func (s *WebhookService) Create(ctx context.Context, webhook *entity.Webhook) error {
	return s.webhookRepo.Create(ctx, webhook)
}

// GetByID gets a webhook by ID
func (s *WebhookService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Webhook, error) {
	return s.webhookRepo.GetByID(ctx, id)
}

// ListByTenant lists all webhooks for a tenant
func (s *WebhookService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Webhook, error) {
	return s.webhookRepo.ListByTenant(ctx, tenantID)
}

// Update updates a webhook
func (s *WebhookService) Update(ctx context.Context, webhook *entity.Webhook) error {
	return s.webhookRepo.Update(ctx, webhook)
}

// Delete deletes a webhook
func (s *WebhookService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.webhookRepo.Delete(ctx, id)
}

// TriggerEvent triggers webhooks for a specific event
func (s *WebhookService) TriggerEvent(ctx context.Context, tenantID uuid.UUID, event entity.WebhookEvent, data interface{}) error {
	// Get all webhooks subscribed to this event
	webhooks, err := s.webhookRepo.ListByEvent(ctx, tenantID, event)
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	if len(webhooks) == 0 {
		return nil
	}

	// Create payload
	payload, err := entity.NewWebhookPayload(event, tenantID, data)
	if err != nil {
		return fmt.Errorf("failed to create payload: %w", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Deliver to each webhook asynchronously
	for _, webhook := range webhooks {
		go s.deliver(context.Background(), webhook, event, payloadJSON)
	}

	return nil
}

func (s *WebhookService) deliver(ctx context.Context, webhook *entity.Webhook, event entity.WebhookEvent, payload []byte) {
	delivery := entity.NewWebhookDelivery(webhook.ID, event, payload)

	if err := s.deliveryRepo.Create(ctx, delivery); err != nil {
		log.Printf("Failed to create webhook delivery record: %v", err)
		return
	}

	s.executeDelivery(ctx, webhook, delivery)
}

func (s *WebhookService) executeDelivery(ctx context.Context, webhook *entity.Webhook, delivery *entity.WebhookDelivery) {
	startTime := time.Now()

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(delivery.Payload))
	if err != nil {
		delivery.MarkFailed(nil, "", fmt.Sprintf("failed to create request: %v", err), 0, webhook.RetryCount)
		s.deliveryRepo.Update(ctx, delivery)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FlagFlash-Webhook/1.0")
	req.Header.Set("X-Webhook-ID", webhook.ID.String())
	req.Header.Set("X-Webhook-Event", string(delivery.EventType))
	req.Header.Set("X-Delivery-ID", delivery.ID.String())

	// Add signature if secret is set
	if webhook.Secret != "" {
		signature := s.generateSignature(delivery.Payload, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Add custom headers
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Execute request
	client := &http.Client{
		Timeout: time.Duration(webhook.TimeoutSeconds) * time.Second,
	}

	resp, err := client.Do(req)
	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		delivery.MarkFailed(nil, "", fmt.Sprintf("request failed: %v", err), durationMs, webhook.RetryCount)
		s.deliveryRepo.Update(ctx, delivery)
		return
	}
	defer resp.Body.Close()

	// Read response body (limit to 64KB)
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))

	// Parse response headers
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.MarkSuccess(resp.StatusCode, string(body), headers, durationMs)
	} else {
		delivery.MarkFailed(&resp.StatusCode, string(body), fmt.Sprintf("HTTP %d", resp.StatusCode), durationMs, webhook.RetryCount)
	}

	s.deliveryRepo.Update(ctx, delivery)
}

func (s *WebhookService) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// ProcessRetries processes webhook deliveries that need to be retried
func (s *WebhookService) ProcessRetries(ctx context.Context) error {
	deliveries, err := s.deliveryRepo.ListRetrying(ctx)
	if err != nil {
		return fmt.Errorf("failed to list retrying deliveries: %w", err)
	}

	for _, delivery := range deliveries {
		if delivery.NextRetryAt == nil || time.Now().Before(*delivery.NextRetryAt) {
			continue
		}

		webhook, err := s.webhookRepo.GetByID(ctx, delivery.WebhookID)
		if err != nil || webhook == nil || !webhook.Enabled {
			delivery.Status = entity.WebhookDeliveryStatusFailed
			delivery.ErrorMessage = "webhook not found or disabled"
			s.deliveryRepo.Update(ctx, delivery)
			continue
		}

		go s.executeDelivery(context.Background(), webhook, delivery)
	}

	return nil
}

// GetDeliveries gets delivery history for a webhook
func (s *WebhookService) GetDeliveries(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*entity.WebhookDelivery, int, error) {
	return s.deliveryRepo.ListByWebhook(ctx, webhookID, limit, offset)
}

// RetryDelivery manually retries a failed delivery
func (s *WebhookService) RetryDelivery(ctx context.Context, deliveryID uuid.UUID) error {
	delivery, err := s.deliveryRepo.GetByID(ctx, deliveryID)
	if err != nil {
		return err
	}

	if delivery.Status == entity.WebhookDeliveryStatusSuccess {
		return fmt.Errorf("delivery was already successful")
	}

	webhook, err := s.webhookRepo.GetByID(ctx, delivery.WebhookID)
	if err != nil {
		return err
	}

	// Reset delivery status
	delivery.Status = entity.WebhookDeliveryStatusPending
	delivery.Attempt++
	delivery.NextRetryAt = nil

	go s.executeDelivery(context.Background(), webhook, delivery)

	return nil
}

// CleanupOldDeliveries removes old delivery records
func (s *WebhookService) CleanupOldDeliveries(ctx context.Context, days int) (int, error) {
	return s.deliveryRepo.DeleteOlderThan(ctx, days)
}

// TestWebhook sends a test event to verify webhook configuration
func (s *WebhookService) TestWebhook(ctx context.Context, webhook *entity.Webhook) (*entity.WebhookDelivery, error) {
	testPayload := map[string]interface{}{
		"event":     "test",
		"timestamp": time.Now().Format(time.RFC3339),
		"tenant_id": webhook.TenantID,
		"message":   "This is a test webhook delivery from FlagFlash",
	}

	payloadJSON, err := json.Marshal(testPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal test payload: %w", err)
	}

	delivery := entity.NewWebhookDelivery(webhook.ID, "test", payloadJSON)

	if err := s.deliveryRepo.Create(ctx, delivery); err != nil {
		return nil, fmt.Errorf("failed to create delivery record: %w", err)
	}

	// Execute synchronously for test
	s.executeDelivery(ctx, webhook, delivery)

	// Refresh delivery status
	return s.deliveryRepo.GetByID(ctx, delivery.ID)
}
