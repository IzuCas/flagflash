package repository

import (
	"context"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"

	"github.com/google/uuid"
)

// EvaluationEventRepository defines the interface for evaluation event operations
type EvaluationEventRepository interface {
	// Create stores a new evaluation event
	Create(ctx context.Context, event *entity.EvaluationEvent) error

	// CreateBatch stores multiple evaluation events
	CreateBatch(ctx context.Context, events []*entity.EvaluationEvent) error

	// GetByTenant returns evaluation events for a tenant
	GetByTenant(ctx context.Context, tenantID uuid.UUID, filters EvaluationFilters) ([]*entity.EvaluationEvent, int, error)

	// GetByFlag returns evaluation events for a specific flag
	GetByFlag(ctx context.Context, flagID uuid.UUID, filters EvaluationFilters) ([]*entity.EvaluationEvent, int, error)

	// GetSummary returns aggregated metrics
	GetSummary(ctx context.Context, tenantID uuid.UUID, filters MetricsFilters) (*entity.UsageMetrics, error)

	// GetSummaryByEnvironment returns aggregated metrics per environment
	GetSummaryByEnvironment(ctx context.Context, tenantID uuid.UUID, filters MetricsFilters) ([]entity.EnvironmentMetrics, error)

	// GetSummaryByFlag returns aggregated metrics per flag
	GetSummaryByFlag(ctx context.Context, tenantID uuid.UUID, filters MetricsFilters) ([]entity.FlagMetrics, error)

	// GetTimeline returns time-series data
	GetTimeline(ctx context.Context, tenantID uuid.UUID, filters MetricsFilters) ([]entity.TimelinePoint, error)

	// DeleteOlderThan removes events older than a specific date
	DeleteOlderThan(ctx context.Context, tenantID uuid.UUID, before time.Time) (int64, error)
}

// EvaluationFilters represents filters for querying evaluation events
type EvaluationFilters struct {
	EnvironmentID *uuid.UUID
	FlagID        *uuid.UUID
	FlagKey       *string
	UserID        *string
	StartDate     *time.Time
	EndDate       *time.Time
	Limit         int
	Offset        int
}

// MetricsFilters represents filters for aggregated metrics
type MetricsFilters struct {
	EnvironmentID *uuid.UUID
	FlagID        *uuid.UUID
	StartDate     time.Time
	EndDate       time.Time
	Granularity   string // hour, day, week, month
}
