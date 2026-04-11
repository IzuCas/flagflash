package service

import (
	"context"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"

	"github.com/google/uuid"
)

// UsageMetricsService handles usage metrics business logic
type UsageMetricsService struct {
	evalRepo repository.EvaluationEventRepository
}

// NewUsageMetricsService creates a new usage metrics service
func NewUsageMetricsService(evalRepo repository.EvaluationEventRepository) *UsageMetricsService {
	return &UsageMetricsService{evalRepo: evalRepo}
}

// RecordEvaluation records a single flag evaluation event
func (s *UsageMetricsService) RecordEvaluation(ctx context.Context, event *entity.EvaluationEvent) error {
	return s.evalRepo.Create(ctx, event)
}

// RecordEvaluationBatch records multiple flag evaluation events
func (s *UsageMetricsService) RecordEvaluationBatch(ctx context.Context, events []*entity.EvaluationEvent) error {
	return s.evalRepo.CreateBatch(ctx, events)
}

// GetMetricsSummary returns aggregated metrics for a tenant
func (s *UsageMetricsService) GetMetricsSummary(ctx context.Context, tenantID uuid.UUID, environmentID *uuid.UUID, flagID *uuid.UUID, startDate, endDate time.Time, granularity string) (*entity.UsageMetrics, error) {
	filters := repository.MetricsFilters{
		EnvironmentID: environmentID,
		FlagID:        flagID,
		StartDate:     startDate,
		EndDate:       endDate,
		Granularity:   granularity,
	}

	metrics, err := s.evalRepo.GetSummary(ctx, tenantID, filters)
	if err != nil {
		return nil, err
	}

	// Get timeline data
	timeline, err := s.evalRepo.GetTimeline(ctx, tenantID, filters)
	if err != nil {
		return nil, err
	}
	metrics.Timeline = timeline

	// Get by environment breakdown
	byEnv, err := s.evalRepo.GetSummaryByEnvironment(ctx, tenantID, filters)
	if err != nil {
		return nil, err
	}
	metrics.ByEnvironment = byEnv

	// Get by flag breakdown
	byFlag, err := s.evalRepo.GetSummaryByFlag(ctx, tenantID, filters)
	if err != nil {
		return nil, err
	}
	metrics.ByFlag = byFlag

	return metrics, nil
}

// GetEnvironmentMetrics returns metrics broken down by environment
func (s *UsageMetricsService) GetEnvironmentMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]entity.EnvironmentMetrics, error) {
	filters := repository.MetricsFilters{
		StartDate: startDate,
		EndDate:   endDate,
	}
	return s.evalRepo.GetSummaryByEnvironment(ctx, tenantID, filters)
}

// GetFlagMetrics returns metrics broken down by flag
func (s *UsageMetricsService) GetFlagMetrics(ctx context.Context, tenantID uuid.UUID, environmentID *uuid.UUID, startDate, endDate time.Time) ([]entity.FlagMetrics, error) {
	filters := repository.MetricsFilters{
		EnvironmentID: environmentID,
		StartDate:     startDate,
		EndDate:       endDate,
	}
	return s.evalRepo.GetSummaryByFlag(ctx, tenantID, filters)
}

// GetTimeline returns time-series data for evaluations
func (s *UsageMetricsService) GetTimeline(ctx context.Context, tenantID uuid.UUID, environmentID *uuid.UUID, flagID *uuid.UUID, startDate, endDate time.Time, granularity string) ([]entity.TimelinePoint, error) {
	filters := repository.MetricsFilters{
		EnvironmentID: environmentID,
		FlagID:        flagID,
		StartDate:     startDate,
		EndDate:       endDate,
		Granularity:   granularity,
	}
	return s.evalRepo.GetTimeline(ctx, tenantID, filters)
}

// GetRecentEvaluations returns recent evaluation events
func (s *UsageMetricsService) GetRecentEvaluations(ctx context.Context, tenantID uuid.UUID, environmentID *uuid.UUID, flagID *uuid.UUID, limit, offset int) ([]*entity.EvaluationEvent, int, error) {
	filters := repository.EvaluationFilters{
		EnvironmentID: environmentID,
		FlagID:        flagID,
		Limit:         limit,
		Offset:        offset,
	}
	return s.evalRepo.GetByTenant(ctx, tenantID, filters)
}

// CleanupOldEvents removes evaluation events older than the specified retention period
func (s *UsageMetricsService) CleanupOldEvents(ctx context.Context, tenantID uuid.UUID, retentionDays int) (int64, error) {
	before := time.Now().AddDate(0, 0, -retentionDays)
	return s.evalRepo.DeleteOlderThan(ctx, tenantID, before)
}
