package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// FeatureFlagRepository defines the interface for feature flag persistence
type FeatureFlagRepository interface {
	Create(ctx context.Context, flag *entity.FeatureFlag) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.FeatureFlag, error)
	GetByKey(ctx context.Context, environmentID uuid.UUID, key string) (*entity.FeatureFlag, error)
	GetByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]*entity.FeatureFlag, error)
	GetByEnvironmentIDPaginated(ctx context.Context, environmentID uuid.UUID, offset, limit int) ([]*entity.FeatureFlag, int, error)
	ListByEnvironment(ctx context.Context, environmentID uuid.UUID, includeDeleted bool) ([]*entity.FeatureFlag, error)
	ListByEnvironmentWithPagination(ctx context.Context, environmentID uuid.UUID, limit, offset int, search string) ([]*entity.FeatureFlag, int, error)
	GetFlagWithTenant(ctx context.Context, id uuid.UUID) (*entity.FeatureFlag, uuid.UUID, error)
	Update(ctx context.Context, flag *entity.FeatureFlag) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementVersion(ctx context.Context, id uuid.UUID) error
	CopyFlags(ctx context.Context, sourceEnvID, targetEnvID uuid.UUID) error
}
