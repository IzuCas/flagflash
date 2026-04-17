package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WebhookEvent defines the types of events that can trigger webhooks
type WebhookEvent string

const (
	WebhookEventFlagCreated       WebhookEvent = "flag.created"
	WebhookEventFlagUpdated       WebhookEvent = "flag.updated"
	WebhookEventFlagDeleted       WebhookEvent = "flag.deleted"
	WebhookEventFlagEnabled       WebhookEvent = "flag.enabled"
	WebhookEventFlagDisabled      WebhookEvent = "flag.disabled"
	WebhookEventRuleCreated       WebhookEvent = "rule.created"
	WebhookEventRuleUpdated       WebhookEvent = "rule.updated"
	WebhookEventRuleDeleted       WebhookEvent = "rule.deleted"
	WebhookEventApprovalRequested WebhookEvent = "approval.requested"
	WebhookEventApprovalDecided   WebhookEvent = "approval.decided"
	WebhookEventExperimentStarted WebhookEvent = "experiment.started"
	WebhookEventExperimentEnded   WebhookEvent = "experiment.ended"
	WebhookEventRolloutProgress   WebhookEvent = "rollout.progress"
	WebhookEventRolloutComplete   WebhookEvent = "rollout.complete"
	WebhookEventRolloutRollback   WebhookEvent = "rollout.rollback"
	WebhookEventEmergencyEnabled  WebhookEvent = "emergency.enabled"
	WebhookEventEmergencyDisabled WebhookEvent = "emergency.disabled"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID             uuid.UUID         `json:"id"`
	TenantID       uuid.UUID         `json:"tenant_id"`
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	Secret         string            `json:"secret,omitempty"`
	Events         []WebhookEvent    `json:"events"`
	Headers        map[string]string `json:"headers,omitempty"`
	Enabled        bool              `json:"enabled"`
	RetryCount     int               `json:"retry_count"`
	TimeoutSeconds int               `json:"timeout_seconds"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// NewWebhook creates a new webhook
func NewWebhook(tenantID uuid.UUID, name, url, secret string, events []WebhookEvent) *Webhook {
	now := time.Now()
	return &Webhook{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Name:           name,
		URL:            url,
		Secret:         secret,
		Events:         events,
		Headers:        map[string]string{},
		Enabled:        true,
		RetryCount:     3,
		TimeoutSeconds: 30,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// HasEvent checks if the webhook is subscribed to an event
func (w *Webhook) HasEvent(event WebhookEvent) bool {
	for _, e := range w.Events {
		if e == event {
			return true
		}
	}
	return false
}

// Enable enables the webhook
func (w *Webhook) Enable() {
	w.Enabled = true
	w.UpdatedAt = time.Now()
}

// Disable disables the webhook
func (w *Webhook) Disable() {
	w.Enabled = false
	w.UpdatedAt = time.Now()
}

// Update updates webhook details
func (w *Webhook) Update(name, url, secret string, events []WebhookEvent, headers map[string]string) {
	if name != "" {
		w.Name = name
	}
	if url != "" {
		w.URL = url
	}
	if secret != "" {
		w.Secret = secret
	}
	if events != nil {
		w.Events = events
	}
	if headers != nil {
		w.Headers = headers
	}
	w.UpdatedAt = time.Now()
}

// WebhookDeliveryStatus defines the status of a delivery attempt
type WebhookDeliveryStatus string

const (
	WebhookDeliveryStatusPending  WebhookDeliveryStatus = "pending"
	WebhookDeliveryStatusSuccess  WebhookDeliveryStatus = "success"
	WebhookDeliveryStatusFailed   WebhookDeliveryStatus = "failed"
	WebhookDeliveryStatusRetrying WebhookDeliveryStatus = "retrying"
)

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID              uuid.UUID             `json:"id"`
	WebhookID       uuid.UUID             `json:"webhook_id"`
	EventType       WebhookEvent          `json:"event_type"`
	Payload         json.RawMessage       `json:"payload"`
	ResponseStatus  *int                  `json:"response_status,omitempty"`
	ResponseBody    string                `json:"response_body,omitempty"`
	ResponseHeaders map[string]string     `json:"response_headers,omitempty"`
	DurationMs      int                   `json:"duration_ms"`
	Attempt         int                   `json:"attempt"`
	Status          WebhookDeliveryStatus `json:"status"`
	ErrorMessage    string                `json:"error_message,omitempty"`
	NextRetryAt     *time.Time            `json:"next_retry_at,omitempty"`
	DeliveredAt     *time.Time            `json:"delivered_at,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
}

// NewWebhookDelivery creates a new webhook delivery
func NewWebhookDelivery(webhookID uuid.UUID, eventType WebhookEvent, payload json.RawMessage) *WebhookDelivery {
	return &WebhookDelivery{
		ID:        uuid.New(),
		WebhookID: webhookID,
		EventType: eventType,
		Payload:   payload,
		Attempt:   1,
		Status:    WebhookDeliveryStatusPending,
		CreatedAt: time.Now(),
	}
}

// MarkSuccess marks the delivery as successful
func (d *WebhookDelivery) MarkSuccess(status int, body string, headers map[string]string, durationMs int) {
	now := time.Now()
	d.ResponseStatus = &status
	d.ResponseBody = body
	d.ResponseHeaders = headers
	d.DurationMs = durationMs
	d.Status = WebhookDeliveryStatusSuccess
	d.DeliveredAt = &now
}

// MarkFailed marks the delivery as failed
func (d *WebhookDelivery) MarkFailed(status *int, body, errorMsg string, durationMs int, maxRetries int) {
	d.ResponseStatus = status
	d.ResponseBody = body
	d.ErrorMessage = errorMsg
	d.DurationMs = durationMs

	if d.Attempt < maxRetries {
		d.Status = WebhookDeliveryStatusRetrying
		// Exponential backoff: 1min, 5min, 30min
		backoffs := []time.Duration{1 * time.Minute, 5 * time.Minute, 30 * time.Minute}
		backoff := backoffs[min(d.Attempt-1, len(backoffs)-1)]
		nextRetry := time.Now().Add(backoff)
		d.NextRetryAt = &nextRetry
		d.Attempt++
	} else {
		d.Status = WebhookDeliveryStatusFailed
	}
}

// WebhookPayload represents the payload sent to webhooks
type WebhookPayload struct {
	Event     WebhookEvent    `json:"event"`
	Timestamp time.Time       `json:"timestamp"`
	TenantID  uuid.UUID       `json:"tenant_id"`
	Data      json.RawMessage `json:"data"`
}

// NewWebhookPayload creates a new webhook payload
func NewWebhookPayload(event WebhookEvent, tenantID uuid.UUID, data interface{}) (*WebhookPayload, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &WebhookPayload{
		Event:     event,
		Timestamp: time.Now(),
		TenantID:  tenantID,
		Data:      dataJSON,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
