package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// WebhookRepo implements repository.WebhookRepository
type WebhookRepo struct {
	db *DB
}

// NewWebhookRepo creates a new webhook repository
func NewWebhookRepo(db *DB) repository.WebhookRepository {
	return &WebhookRepo{db: db}
}

// Create creates a new webhook
func (r *WebhookRepo) Create(ctx context.Context, webhook *entity.Webhook) error {
	headers, err := json.Marshal(webhook.Headers)
	if err != nil {
		headers = []byte("{}")
	}

	query := `
		INSERT INTO webhooks (id, tenant_id, name, url, secret, events, headers, enabled, retry_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		webhook.ID, webhook.TenantID, webhook.Name, webhook.URL, webhook.Secret,
		pq.Array(webhook.Events), headers, webhook.Enabled, webhook.RetryCount,
		webhook.CreatedAt, webhook.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	return nil
}

// Update updates a webhook
func (r *WebhookRepo) Update(ctx context.Context, webhook *entity.Webhook) error {
	headers, err := json.Marshal(webhook.Headers)
	if err != nil {
		headers = []byte("{}")
	}

	webhook.UpdatedAt = time.Now()

	query := `
		UPDATE webhooks
		SET name = $1, url = $2, secret = $3, events = $4, headers = $5, enabled = $6, retry_count = $7, updated_at = $8
		WHERE id = $9
	`

	result, err := r.db.ExecContext(ctx, query,
		webhook.Name, webhook.URL, webhook.Secret, pq.Array(webhook.Events),
		headers, webhook.Enabled, webhook.RetryCount, webhook.UpdatedAt, webhook.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

// Delete deletes a webhook
func (r *WebhookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

// GetByID retrieves a webhook by ID
func (r *WebhookRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, events, headers, enabled, retry_count, created_at, updated_at
		FROM webhooks
		WHERE id = $1
	`

	return r.scanWebhook(r.db.QueryRowContext(ctx, query, id))
}

// ListByTenant lists webhooks for a tenant
func (r *WebhookRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, events, headers, enabled, retry_count, created_at, updated_at
		FROM webhooks
		WHERE tenant_id = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	return r.scanWebhooks(rows)
}

// ListEnabledByTenant lists enabled webhooks for a tenant
func (r *WebhookRepo) ListEnabledByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, events, headers, enabled, retry_count, created_at, updated_at
		FROM webhooks
		WHERE tenant_id = $1 AND enabled = true
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	return r.scanWebhooks(rows)
}

// ListByEvent lists webhooks subscribed to an event
func (r *WebhookRepo) ListByEvent(ctx context.Context, tenantID uuid.UUID, eventType entity.WebhookEvent) ([]*entity.Webhook, error) {
	query := `
		SELECT id, tenant_id, name, url, secret, events, headers, enabled, retry_count, created_at, updated_at
		FROM webhooks
		WHERE tenant_id = $1 AND enabled = true AND $2 = ANY(events)
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, string(eventType))
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks by event: %w", err)
	}
	defer rows.Close()

	return r.scanWebhooks(rows)
}

func (r *WebhookRepo) scanWebhook(row *sql.Row) (*entity.Webhook, error) {
	var webhook entity.Webhook
	var secret sql.NullString
	var events pq.StringArray
	var headers []byte

	err := row.Scan(
		&webhook.ID, &webhook.TenantID, &webhook.Name, &webhook.URL, &secret,
		&events, &headers, &webhook.Enabled, &webhook.RetryCount,
		&webhook.CreatedAt, &webhook.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("webhook not found")
		}
		return nil, fmt.Errorf("failed to scan webhook: %w", err)
	}

	if secret.Valid {
		webhook.Secret = secret.String
	}

	webhook.Events = make([]entity.WebhookEvent, len(events))
	for i, e := range events {
		webhook.Events[i] = entity.WebhookEvent(e)
	}

	json.Unmarshal(headers, &webhook.Headers)

	return &webhook, nil
}

func (r *WebhookRepo) scanWebhooks(rows *sql.Rows) ([]*entity.Webhook, error) {
	webhooks := make([]*entity.Webhook, 0)

	for rows.Next() {
		var webhook entity.Webhook
		var secret sql.NullString
		var events pq.StringArray
		var headers []byte

		err := rows.Scan(
			&webhook.ID, &webhook.TenantID, &webhook.Name, &webhook.URL, &secret,
			&events, &headers, &webhook.Enabled, &webhook.RetryCount,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}

		if secret.Valid {
			webhook.Secret = secret.String
		}

		webhook.Events = make([]entity.WebhookEvent, len(events))
		for i, e := range events {
			webhook.Events[i] = entity.WebhookEvent(e)
		}

		json.Unmarshal(headers, &webhook.Headers)

		webhooks = append(webhooks, &webhook)
	}

	return webhooks, nil
}

// WebhookDeliveryRepo implements repository.WebhookDeliveryRepository
type WebhookDeliveryRepo struct {
	db *DB
}

// NewWebhookDeliveryRepo creates a new webhook delivery repository
func NewWebhookDeliveryRepo(db *DB) repository.WebhookDeliveryRepository {
	return &WebhookDeliveryRepo{db: db}
}

// Create creates a new webhook delivery
func (r *WebhookDeliveryRepo) Create(ctx context.Context, delivery *entity.WebhookDelivery) error {
	var responseHeadersJSON []byte
	if delivery.ResponseHeaders != nil {
		responseHeadersJSON, _ = json.Marshal(delivery.ResponseHeaders)
	}

	query := `
		INSERT INTO webhook_deliveries (id, webhook_id, event_type, payload, response_status, response_body, response_headers, duration_ms, attempt, status, error_message, next_retry_at, delivered_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.ExecContext(ctx, query,
		delivery.ID, delivery.WebhookID, delivery.EventType, delivery.Payload,
		delivery.ResponseStatus, delivery.ResponseBody, responseHeadersJSON,
		delivery.DurationMs, delivery.Attempt, delivery.Status, delivery.ErrorMessage,
		delivery.NextRetryAt, delivery.DeliveredAt, delivery.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create webhook delivery: %w", err)
	}

	return nil
}

// Update updates a webhook delivery
func (r *WebhookDeliveryRepo) Update(ctx context.Context, delivery *entity.WebhookDelivery) error {
	var responseHeadersJSON []byte
	if delivery.ResponseHeaders != nil {
		responseHeadersJSON, _ = json.Marshal(delivery.ResponseHeaders)
	}

	query := `
		UPDATE webhook_deliveries
		SET response_status = $1, response_body = $2, response_headers = $3, duration_ms = $4, attempt = $5, status = $6, error_message = $7, next_retry_at = $8, delivered_at = $9
		WHERE id = $10
	`

	result, err := r.db.ExecContext(ctx, query,
		delivery.ResponseStatus, delivery.ResponseBody, responseHeadersJSON,
		delivery.DurationMs, delivery.Attempt, delivery.Status, delivery.ErrorMessage,
		delivery.NextRetryAt, delivery.DeliveredAt, delivery.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update webhook delivery: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("webhook delivery not found")
	}

	return nil
}

// GetByID retrieves a webhook delivery by ID
func (r *WebhookDeliveryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event_type, payload, response_status, response_body, response_headers, duration_ms, attempt, status, error_message, next_retry_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE id = $1
	`

	var delivery entity.WebhookDelivery
	var payload, responseHeadersJSON []byte
	var responseStatus sql.NullInt32
	var responseBody, errorMessage sql.NullString
	var nextRetryAt, deliveredAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&delivery.ID, &delivery.WebhookID, &delivery.EventType, &payload,
		&responseStatus, &responseBody, &responseHeadersJSON, &delivery.DurationMs,
		&delivery.Attempt, &delivery.Status, &errorMessage, &nextRetryAt, &deliveredAt, &delivery.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("webhook delivery not found")
		}
		return nil, fmt.Errorf("failed to get webhook delivery: %w", err)
	}

	if len(payload) > 0 {
		delivery.Payload = payload
	}
	if responseStatus.Valid {
		status := int(responseStatus.Int32)
		delivery.ResponseStatus = &status
	}
	if responseBody.Valid {
		delivery.ResponseBody = responseBody.String
	}
	if len(responseHeadersJSON) > 0 {
		json.Unmarshal(responseHeadersJSON, &delivery.ResponseHeaders)
	}
	if errorMessage.Valid {
		delivery.ErrorMessage = errorMessage.String
	}
	if nextRetryAt.Valid {
		delivery.NextRetryAt = &nextRetryAt.Time
	}
	if deliveredAt.Valid {
		delivery.DeliveredAt = &deliveredAt.Time
	}

	return &delivery, nil
}

// ListByWebhook lists deliveries for a webhook with pagination
func (r *WebhookDeliveryRepo) ListByWebhook(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*entity.WebhookDelivery, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM webhook_deliveries WHERE webhook_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, webhookID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count webhook deliveries: %w", err)
	}

	query := `
		SELECT id, webhook_id, event_type, payload, response_status, response_body, response_headers, duration_ms, attempt, status, error_message, next_retry_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE webhook_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, webhookID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list webhook deliveries: %w", err)
	}
	defer rows.Close()

	deliveries, err := r.scanDeliveries(rows)
	if err != nil {
		return nil, 0, err
	}
	return deliveries, total, nil
}

// ListPending lists all pending deliveries
func (r *WebhookDeliveryRepo) ListPending(ctx context.Context) ([]*entity.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event_type, payload, response_status, response_body, response_headers, duration_ms, attempt, status, error_message, next_retry_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE status = 'pending'
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending webhook deliveries: %w", err)
	}
	defer rows.Close()

	return r.scanDeliveries(rows)
}

// ListRetrying lists all deliveries that need retry
func (r *WebhookDeliveryRepo) ListRetrying(ctx context.Context) ([]*entity.WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, event_type, payload, response_status, response_body, response_headers, duration_ms, attempt, status, error_message, next_retry_at, delivered_at, created_at
		FROM webhook_deliveries
		WHERE status = 'retrying' AND next_retry_at <= NOW()
		ORDER BY next_retry_at
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list retrying webhook deliveries: %w", err)
	}
	defer rows.Close()

	return r.scanDeliveries(rows)
}

// Delete deletes a webhook delivery
func (r *WebhookDeliveryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhook_deliveries WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete webhook delivery: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("webhook delivery not found")
	}

	return nil
}

// DeleteOlderThan deletes deliveries older than specified days
func (r *WebhookDeliveryRepo) DeleteOlderThan(ctx context.Context, days int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	query := `DELETE FROM webhook_deliveries WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old webhook deliveries: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

func (r *WebhookDeliveryRepo) scanDeliveries(rows *sql.Rows) ([]*entity.WebhookDelivery, error) {
	deliveries := make([]*entity.WebhookDelivery, 0)

	for rows.Next() {
		var delivery entity.WebhookDelivery
		var payload, responseHeadersJSON []byte
		var responseStatus sql.NullInt32
		var responseBody, errorMessage sql.NullString
		var nextRetryAt, deliveredAt sql.NullTime

		err := rows.Scan(
			&delivery.ID, &delivery.WebhookID, &delivery.EventType, &payload,
			&responseStatus, &responseBody, &responseHeadersJSON, &delivery.DurationMs,
			&delivery.Attempt, &delivery.Status, &errorMessage, &nextRetryAt, &deliveredAt, &delivery.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook delivery: %w", err)
		}

		if len(payload) > 0 {
			delivery.Payload = payload
		}
		if responseStatus.Valid {
			status := int(responseStatus.Int32)
			delivery.ResponseStatus = &status
		}
		if responseBody.Valid {
			delivery.ResponseBody = responseBody.String
		}
		if len(responseHeadersJSON) > 0 {
			json.Unmarshal(responseHeadersJSON, &delivery.ResponseHeaders)
		}
		if errorMessage.Valid {
			delivery.ErrorMessage = errorMessage.String
		}
		if nextRetryAt.Valid {
			delivery.NextRetryAt = &nextRetryAt.Time
		}
		if deliveredAt.Valid {
			delivery.DeliveredAt = &deliveredAt.Time
		}

		deliveries = append(deliveries, &delivery)
	}

	return deliveries, nil
}
