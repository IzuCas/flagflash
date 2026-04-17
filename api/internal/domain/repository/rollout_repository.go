package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// RolloutPlanRepository defines the interface for rollout plan persistence
type RolloutPlanRepository interface {
	Create(ctx context.Context, plan *entity.RolloutPlan) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.RolloutPlan, error)
	GetByFlag(ctx context.Context, flagID uuid.UUID) (*entity.RolloutPlan, error)
	GetActiveByFlag(ctx context.Context, flagID uuid.UUID) (*entity.RolloutPlan, error)
	ListByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.RolloutPlan, error)
	ListActive(ctx context.Context) ([]*entity.RolloutPlan, error)
	ListNeedingIncrement(ctx context.Context) ([]*entity.RolloutPlan, error)
	Update(ctx context.Context, plan *entity.RolloutPlan) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// RolloutHistoryRepository defines the interface for rollout history persistence
type RolloutHistoryRepository interface {
	Create(ctx context.Context, history *entity.RolloutHistory) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.RolloutHistory, error)
	ListByPlan(ctx context.Context, planID uuid.UUID) ([]*entity.RolloutHistory, error)
	ListByPlanPaginated(ctx context.Context, planID uuid.UUID, limit, offset int) ([]*entity.RolloutHistory, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByPlan(ctx context.Context, planID uuid.UUID) error
}
