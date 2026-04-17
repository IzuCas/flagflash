package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	Create(ctx context.Context, notification *entity.Notification) error
	CreateBatch(ctx context.Context, notifications []*entity.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error)
	ListUnread(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOlderThan(ctx context.Context, days int) (int, error)
}
