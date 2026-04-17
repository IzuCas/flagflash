package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// ExperimentService handles experiment operations
type ExperimentService struct {
	experimentRepo repository.ExperimentRepository
	variantRepo    repository.ExperimentVariantRepository
	metricRepo     repository.ExperimentMetricRepository
	resultRepo     repository.ExperimentResultRepository
	flagRepo       repository.FeatureFlagRepository
	auditRepo      repository.AuditLogRepository
	webhookSvc     *WebhookService
}

// NewExperimentService creates a new experiment service
func NewExperimentService(
	experimentRepo repository.ExperimentRepository,
	variantRepo repository.ExperimentVariantRepository,
	metricRepo repository.ExperimentMetricRepository,
	resultRepo repository.ExperimentResultRepository,
	flagRepo repository.FeatureFlagRepository,
	auditRepo repository.AuditLogRepository,
	webhookSvc *WebhookService,
) *ExperimentService {
	return &ExperimentService{
		experimentRepo: experimentRepo,
		variantRepo:    variantRepo,
		metricRepo:     metricRepo,
		resultRepo:     resultRepo,
		flagRepo:       flagRepo,
		auditRepo:      auditRepo,
		webhookSvc:     webhookSvc,
	}
}

// Create creates a new experiment
func (s *ExperimentService) Create(
	ctx context.Context,
	tenantID, environmentID, flagID uuid.UUID,
	name, description, hypothesis string,
	createdBy *uuid.UUID,
) (*entity.Experiment, error) {
	// Check if flag exists
	flag, err := s.flagRepo.GetByID(ctx, flagID)
	if err != nil {
		return nil, fmt.Errorf("flag not found: %w", err)
	}

	// Check if flag already has an active experiment
	existing, _ := s.experimentRepo.GetActiveByFlag(ctx, flagID)
	if existing != nil {
		return nil, fmt.Errorf("flag already has an active experiment")
	}

	experiment := entity.NewExperiment(tenantID, environmentID, flagID, name, description, hypothesis, createdBy)

	if err := s.experimentRepo.Create(ctx, experiment); err != nil {
		return nil, fmt.Errorf("failed to create experiment: %w", err)
	}

	// Create default variants (control and treatment)
	controlVariant := entity.NewExperimentVariant(
		experiment.ID,
		"Control",
		"Original flag value",
		flag.DefaultValue,
		50,
		true,
	)

	treatmentVariant := entity.NewExperimentVariant(
		experiment.ID,
		"Treatment",
		"Modified flag value",
		flag.DefaultValue, // Will be updated by user
		50,
		false,
	)

	if err := s.variantRepo.CreateBatch(ctx, []*entity.ExperimentVariant{controlVariant, treatmentVariant}); err != nil {
		return nil, fmt.Errorf("failed to create variants: %w", err)
	}

	return experiment, nil
}

// GetByID gets an experiment by ID with all details
func (s *ExperimentService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Experiment, error) {
	return s.experimentRepo.GetByIDWithDetails(ctx, id)
}

// ListByTenant lists all experiments for a tenant
func (s *ExperimentService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Experiment, error) {
	return s.experimentRepo.ListByTenant(ctx, tenantID)
}

// ListByEnvironment lists experiments for an environment
func (s *ExperimentService) ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.Experiment, error) {
	return s.experimentRepo.ListByEnvironment(ctx, environmentID)
}

// Start starts an experiment
func (s *ExperimentService) Start(ctx context.Context, id uuid.UUID) error {
	experiment, err := s.experimentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if experiment.Status != entity.ExperimentStatusDraft && experiment.Status != entity.ExperimentStatusPaused {
		return fmt.Errorf("experiment cannot be started from status: %s", experiment.Status)
	}

	// Validate variants exist and weights sum to 100
	variants, err := s.variantRepo.ListByExperiment(ctx, id)
	if err != nil {
		return err
	}

	if len(variants) < 2 {
		return fmt.Errorf("experiment must have at least 2 variants")
	}

	totalWeight := 0
	for _, v := range variants {
		totalWeight += v.Weight
	}
	if totalWeight != 100 {
		return fmt.Errorf("variant weights must sum to 100, got %d", totalWeight)
	}

	experiment.Start()

	if err := s.experimentRepo.Update(ctx, experiment); err != nil {
		return err
	}

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, experiment.TenantID, entity.WebhookEventExperimentStarted, experiment)
	}

	return nil
}

// Pause pauses an experiment
func (s *ExperimentService) Pause(ctx context.Context, id uuid.UUID) error {
	experiment, err := s.experimentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if experiment.Status != entity.ExperimentStatusRunning {
		return fmt.Errorf("only running experiments can be paused")
	}

	experiment.Pause()
	return s.experimentRepo.Update(ctx, experiment)
}

// Resume resumes a paused experiment
func (s *ExperimentService) Resume(ctx context.Context, id uuid.UUID) error {
	experiment, err := s.experimentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if experiment.Status != entity.ExperimentStatusPaused {
		return fmt.Errorf("only paused experiments can be resumed")
	}

	experiment.Resume()
	return s.experimentRepo.Update(ctx, experiment)
}

// Complete completes an experiment with a winner
func (s *ExperimentService) Complete(ctx context.Context, id uuid.UUID, winnerVariantName string) error {
	experiment, err := s.experimentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if experiment.Status != entity.ExperimentStatusRunning && experiment.Status != entity.ExperimentStatusPaused {
		return fmt.Errorf("experiment cannot be completed from status: %s", experiment.Status)
	}

	// Calculate statistical significance (simplified)
	significance := s.calculateSignificance(ctx, experiment)

	experiment.Complete(winnerVariantName, significance)

	if err := s.experimentRepo.Update(ctx, experiment); err != nil {
		return err
	}

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, experiment.TenantID, entity.WebhookEventExperimentEnded, experiment)
	}

	return nil
}

// Cancel cancels an experiment
func (s *ExperimentService) Cancel(ctx context.Context, id uuid.UUID) error {
	experiment, err := s.experimentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if experiment.Status == entity.ExperimentStatusCompleted || experiment.Status == entity.ExperimentStatusCancelled {
		return fmt.Errorf("experiment is already finished")
	}

	experiment.Cancel()
	return s.experimentRepo.Update(ctx, experiment)
}

// Delete deletes an experiment
func (s *ExperimentService) Delete(ctx context.Context, id uuid.UUID) error {
	experiment, err := s.experimentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if experiment.IsRunning() {
		return fmt.Errorf("cannot delete a running experiment")
	}

	// Delete related data
	s.resultRepo.DeleteByExperiment(ctx, id)
	s.metricRepo.DeleteByExperiment(ctx, id)
	s.variantRepo.DeleteByExperiment(ctx, id)

	return s.experimentRepo.Delete(ctx, id)
}

// AddVariant adds a variant to an experiment
func (s *ExperimentService) AddVariant(ctx context.Context, variant *entity.ExperimentVariant) error {
	experiment, err := s.experimentRepo.GetByID(ctx, variant.ExperimentID)
	if err != nil {
		return err
	}

	if experiment.Status != entity.ExperimentStatusDraft {
		return fmt.Errorf("can only add variants to draft experiments")
	}

	return s.variantRepo.Create(ctx, variant)
}

// UpdateVariant updates a variant
func (s *ExperimentService) UpdateVariant(ctx context.Context, variant *entity.ExperimentVariant) error {
	return s.variantRepo.Update(ctx, variant)
}

// DeleteVariant deletes a variant
func (s *ExperimentService) DeleteVariant(ctx context.Context, id uuid.UUID) error {
	variant, err := s.variantRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	experiment, err := s.experimentRepo.GetByID(ctx, variant.ExperimentID)
	if err != nil {
		return err
	}

	if experiment.Status != entity.ExperimentStatusDraft {
		return fmt.Errorf("can only delete variants from draft experiments")
	}

	if variant.IsControl {
		return fmt.Errorf("cannot delete control variant")
	}

	return s.variantRepo.Delete(ctx, id)
}

// AddMetric adds a metric to an experiment
func (s *ExperimentService) AddMetric(ctx context.Context, metric *entity.ExperimentMetric) error {
	return s.metricRepo.Create(ctx, metric)
}

// ListMetrics lists metrics for an experiment
func (s *ExperimentService) ListMetrics(ctx context.Context, experimentID uuid.UUID) ([]*entity.ExperimentMetric, error) {
	return s.metricRepo.ListByExperiment(ctx, experimentID)
}

// GetResults gets results for an experiment
func (s *ExperimentService) GetResults(ctx context.Context, experimentID uuid.UUID) ([]*entity.ExperimentResult, error) {
	return s.resultRepo.ListByExperiment(ctx, experimentID)
}

// RecordExposure records that a user was exposed to a variant
func (s *ExperimentService) RecordExposure(ctx context.Context, experimentID uuid.UUID, userID string) (*entity.ExperimentVariant, error) {
	experiment, err := s.experimentRepo.GetByID(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	if !experiment.IsRunning() {
		return nil, fmt.Errorf("experiment is not running")
	}

	// Get variants
	variants, err := s.variantRepo.ListByExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	// Assign variant based on user ID hash and weights
	variant := s.assignVariant(userID, variants)

	// Increment sample size
	experiment.IncrementSampleSize(1)
	s.experimentRepo.Update(ctx, experiment)

	return variant, nil
}

// GetVariantForUser gets the variant assigned to a user for an experiment
func (s *ExperimentService) GetVariantForUser(ctx context.Context, experimentID uuid.UUID, userID string) (*entity.ExperimentVariant, error) {
	variants, err := s.variantRepo.ListByExperiment(ctx, experimentID)
	if err != nil {
		return nil, err
	}

	return s.assignVariant(userID, variants), nil
}

// assignVariant assigns a variant based on weighted random selection
func (s *ExperimentService) assignVariant(userID string, variants []*entity.ExperimentVariant) *entity.ExperimentVariant {
	// Use hash of user ID for consistent assignment
	hash := hashString(userID)
	bucket := hash % 100

	var cumulative int
	for _, v := range variants {
		cumulative += v.Weight
		if bucket < cumulative {
			return v
		}
	}

	// Fallback to first variant
	if len(variants) > 0 {
		return variants[0]
	}
	return nil
}

func hashString(s string) int {
	hash := 0
	for _, c := range s {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// calculateSignificance calculates statistical significance (simplified)
func (s *ExperimentService) calculateSignificance(ctx context.Context, experiment *entity.Experiment) float64 {
	// This is a simplified calculation
	// In production, use proper statistical methods
	results, _ := s.resultRepo.ListByExperiment(ctx, experiment.ID)
	if len(results) < 2 {
		return 0
	}

	// Simple confidence based on sample size
	totalSamples := 0
	for _, r := range results {
		totalSamples += r.SampleCount
	}

	if totalSamples < 100 {
		return 0.5
	} else if totalSamples < 1000 {
		return 0.75
	} else if totalSamples < 10000 {
		return 0.90
	}
	return 0.95
}

// RecordConversion records a conversion event for a metric
func (s *ExperimentService) RecordConversion(ctx context.Context, experimentID, variantID, metricID uuid.UUID, value float64) error {
	result, err := s.resultRepo.GetByVariantAndMetric(ctx, variantID, metricID)
	if err != nil {
		// Create new result
		result = entity.NewExperimentResult(experimentID, variantID, metricID)
		if err := s.resultRepo.Create(ctx, result); err != nil {
			return err
		}
	}

	result.AddSample(value, true)
	return s.resultRepo.Update(ctx, result)
}

// CreateExperimentInput holds input for creating an experiment
type CreateExperimentInput struct {
	TenantID      uuid.UUID            `json:"tenant_id"`
	EnvironmentID uuid.UUID            `json:"environment_id"`
	FlagID        uuid.UUID            `json:"flag_id"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	Hypothesis    string               `json:"hypothesis"`
	Variants      []CreateVariantInput `json:"variants"`
	Metrics       []CreateMetricInput  `json:"metrics"`
}

// CreateVariantInput holds input for creating a variant
type CreateVariantInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Value       json.RawMessage `json:"value"`
	Weight      int             `json:"weight"`
	IsControl   bool            `json:"is_control"`
}

// CreateMetricInput holds input for creating a metric
type CreateMetricInput struct {
	Name          string               `json:"name"`
	MetricType    entity.MetricType    `json:"metric_type"`
	IsPrimary     bool                 `json:"is_primary"`
	GoalDirection entity.GoalDirection `json:"goal_direction"`
}

// CreateFull creates an experiment with variants and metrics
func (s *ExperimentService) CreateFull(ctx context.Context, input CreateExperimentInput, createdBy *uuid.UUID) (*entity.Experiment, error) {
	experiment := entity.NewExperiment(
		input.TenantID,
		input.EnvironmentID,
		input.FlagID,
		input.Name,
		input.Description,
		input.Hypothesis,
		createdBy,
	)

	if err := s.experimentRepo.Create(ctx, experiment); err != nil {
		return nil, err
	}

	// Create variants
	for _, v := range input.Variants {
		variant := entity.NewExperimentVariant(
			experiment.ID,
			v.Name,
			v.Description,
			v.Value,
			v.Weight,
			v.IsControl,
		)
		if err := s.variantRepo.Create(ctx, variant); err != nil {
			return nil, err
		}
	}

	// Create metrics
	for _, m := range input.Metrics {
		metric := entity.NewExperimentMetric(
			experiment.ID,
			m.Name,
			m.MetricType,
			m.IsPrimary,
			m.GoalDirection,
		)
		if err := s.metricRepo.Create(ctx, metric); err != nil {
			return nil, err
		}
	}

	return s.experimentRepo.GetByIDWithDetails(ctx, experiment.ID)
}

func init() {
	rand.Seed(42) // For consistent variant assignment in tests
}
