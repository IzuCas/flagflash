package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// WebhookRepository defines the interface for webhook persistence
type WebhookRepository interface {
	Create(ctx context.Context, webhook *entity.Webhook) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Webhook, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Webhook, error)
	ListEnabledByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Webhook, error)
	ListByEvent(ctx context.Context, tenantID uuid.UUID, event entity.WebhookEvent) ([]*entity.Webhook, error)
	Update(ctx context.Context, webhook *entity.Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// WebhookDeliveryRepository defines the interface for webhook delivery persistence
type WebhookDeliveryRepository interface {
	Create(ctx context.Context, delivery *entity.WebhookDelivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.WebhookDelivery, error)
	ListByWebhook(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]*entity.WebhookDelivery, int, error)
	ListPending(ctx context.Context) ([]*entity.WebhookDelivery, error)
	ListRetrying(ctx context.Context) ([]*entity.WebhookDelivery, error)
	Update(ctx context.Context, delivery *entity.WebhookDelivery) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOlderThan(ctx context.Context, days int) (int, error)
}
