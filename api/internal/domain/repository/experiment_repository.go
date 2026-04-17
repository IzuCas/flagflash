package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// ExperimentRepository defines the interface for experiment persistence
type ExperimentRepository interface {
	Create(ctx context.Context, experiment *entity.Experiment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Experiment, error)
	GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.Experiment, error)
	GetByFlag(ctx context.Context, flagID uuid.UUID) (*entity.Experiment, error)
	GetActiveByFlag(ctx context.Context, flagID uuid.UUID) (*entity.Experiment, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Experiment, error)
	ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.Experiment, error)
	ListByStatus(ctx context.Context, tenantID uuid.UUID, status entity.ExperimentStatus) ([]*entity.Experiment, error)
	ListRunning(ctx context.Context) ([]*entity.Experiment, error)
	Update(ctx context.Context, experiment *entity.Experiment) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ExperimentVariantRepository defines the interface for experiment variant persistence
type ExperimentVariantRepository interface {
	Create(ctx context.Context, variant *entity.ExperimentVariant) error
	CreateBatch(ctx context.Context, variants []*entity.ExperimentVariant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ExperimentVariant, error)
	ListByExperiment(ctx context.Context, experimentID uuid.UUID) ([]*entity.ExperimentVariant, error)
	GetControlVariant(ctx context.Context, experimentID uuid.UUID) (*entity.ExperimentVariant, error)
	Update(ctx context.Context, variant *entity.ExperimentVariant) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByExperiment(ctx context.Context, experimentID uuid.UUID) error
}

// ExperimentMetricRepository defines the interface for experiment metric persistence
type ExperimentMetricRepository interface {
	Create(ctx context.Context, metric *entity.ExperimentMetric) error
	CreateBatch(ctx context.Context, metrics []*entity.ExperimentMetric) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ExperimentMetric, error)
	ListByExperiment(ctx context.Context, experimentID uuid.UUID) ([]*entity.ExperimentMetric, error)
	GetPrimaryMetric(ctx context.Context, experimentID uuid.UUID) (*entity.ExperimentMetric, error)
	Update(ctx context.Context, metric *entity.ExperimentMetric) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByExperiment(ctx context.Context, experimentID uuid.UUID) error
}

// ExperimentResultRepository defines the interface for experiment result persistence
type ExperimentResultRepository interface {
	Create(ctx context.Context, result *entity.ExperimentResult) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ExperimentResult, error)
	GetByVariantAndMetric(ctx context.Context, variantID, metricID uuid.UUID) (*entity.ExperimentResult, error)
	ListByExperiment(ctx context.Context, experimentID uuid.UUID) ([]*entity.ExperimentResult, error)
	ListByVariant(ctx context.Context, variantID uuid.UUID) ([]*entity.ExperimentResult, error)
	Update(ctx context.Context, result *entity.ExperimentResult) error
	Upsert(ctx context.Context, result *entity.ExperimentResult) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByExperiment(ctx context.Context, experimentID uuid.UUID) error
}
