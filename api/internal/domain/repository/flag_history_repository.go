package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// FlagHistoryRepository defines the interface for flag history persistence
type FlagHistoryRepository interface {
	Create(ctx context.Context, history *entity.FlagHistory) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.FlagHistory, error)
	ListByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.FlagHistory, error)
	ListByFlagPaginated(ctx context.Context, flagID uuid.UUID, limit, offset int) ([]*entity.FlagHistory, int, error)
	GetLatestByFlag(ctx context.Context, flagID uuid.UUID) (*entity.FlagHistory, error)
	GetByVersion(ctx context.Context, flagID uuid.UUID, version int) (*entity.FlagHistory, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByFlag(ctx context.Context, flagID uuid.UUID) error
	DeleteOlderThan(ctx context.Context, days int) (int, error)
}
