package service

import (
"context"
"fmt"

"github.com/IzuCas/flagflash/internal/domain/entity"
"github.com/IzuCas/flagflash/internal/domain/repository"
"github.com/google/uuid"
)

// RolloutService handles progressive rollout operations
type RolloutService struct {
	rolloutRepo repository.RolloutPlanRepository
	historyRepo repository.RolloutHistoryRepository
	flagRepo    repository.FeatureFlagRepository
	webhookSvc  *WebhookService
}

// NewRolloutService creates a new rollout service
func NewRolloutService(
rolloutRepo repository.RolloutPlanRepository,
historyRepo repository.RolloutHistoryRepository,
flagRepo repository.FeatureFlagRepository,
webhookSvc *WebhookService,
) *RolloutService {
	return &RolloutService{
		rolloutRepo: rolloutRepo,
		historyRepo: historyRepo,
		flagRepo:    flagRepo,
		webhookSvc:  webhookSvc,
	}
}

// Create creates a new rollout plan
func (s *RolloutService) Create(
ctx context.Context,
flagID uuid.UUID,
name string,
targetPercentage, incrementPercentage, intervalMinutes int,
createdBy *uuid.UUID,
) (*entity.RolloutPlan, error) {
	// Check if flag exists
	_, err := s.flagRepo.GetByID(ctx, flagID)
	if err != nil {
		return nil, fmt.Errorf("flag not found: %w", err)
	}

	// Check for existing active rollout
	existing, _ := s.rolloutRepo.GetActiveByFlag(ctx, flagID)
	if existing != nil {
		return nil, fmt.Errorf("flag already has an active rollout plan")
	}

	rollout := entity.NewRolloutPlan(
flagID,
name,
targetPercentage,
incrementPercentage,
intervalMinutes,
createdBy,
)

	if err := s.rolloutRepo.Create(ctx, rollout); err != nil {
		return nil, err
	}

	return rollout, nil
}

// GetByID gets a rollout plan by ID
func (s *RolloutService) GetByID(ctx context.Context, id uuid.UUID) (*entity.RolloutPlan, error) {
	return s.rolloutRepo.GetByID(ctx, id)
}

// GetActiveByFlag gets the active rollout for a flag
func (s *RolloutService) GetActiveByFlag(ctx context.Context, flagID uuid.UUID) (*entity.RolloutPlan, error) {
	return s.rolloutRepo.GetActiveByFlag(ctx, flagID)
}

// ListByFlag lists all rollout plans for a flag
func (s *RolloutService) ListByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.RolloutPlan, error) {
	return s.rolloutRepo.ListByFlag(ctx, flagID)
}

// ListActive lists all active rollout plans
func (s *RolloutService) ListActive(ctx context.Context) ([]*entity.RolloutPlan, error) {
	return s.rolloutRepo.ListActive(ctx)
}

// Start starts a rollout plan
func (s *RolloutService) Start(ctx context.Context, id uuid.UUID) error {
	rollout, err := s.rolloutRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rollout.Status != entity.RolloutStatusDraft && rollout.Status != entity.RolloutStatusPaused {
		return fmt.Errorf("rollout cannot be started from status: %s", rollout.Status)
	}

	rollout.Start()

	return s.rolloutRepo.Update(ctx, rollout)
}

// Pause pauses a rollout
func (s *RolloutService) Pause(ctx context.Context, id uuid.UUID) error {
	rollout, err := s.rolloutRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rollout.Status != entity.RolloutStatusActive {
		return fmt.Errorf("only active rollouts can be paused")
	}

	previousPct := rollout.CurrentPercentage
	rollout.Pause()

	if err := s.rolloutRepo.Update(ctx, rollout); err != nil {
		return err
	}

	// Record history
	history := entity.NewRolloutHistory(
rollout.ID,
previousPct,
rollout.CurrentPercentage,
entity.RolloutActionPause,
"Rollout paused",
nil,
)
	return s.historyRepo.Create(ctx, history)
}

// Resume resumes a paused rollout
func (s *RolloutService) Resume(ctx context.Context, id uuid.UUID) error {
	rollout, err := s.rolloutRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rollout.Status != entity.RolloutStatusPaused {
		return fmt.Errorf("only paused rollouts can be resumed")
	}

	previousPct := rollout.CurrentPercentage
	rollout.Resume()

	if err := s.rolloutRepo.Update(ctx, rollout); err != nil {
		return err
	}

	// Record history
	history := entity.NewRolloutHistory(
rollout.ID,
previousPct,
rollout.CurrentPercentage,
entity.RolloutActionResume,
"Rollout resumed",
nil,
)
	return s.historyRepo.Create(ctx, history)
}

// Increment processes automatic or manual increment
func (s *RolloutService) Increment(ctx context.Context, id uuid.UUID) error {
	rollout, err := s.rolloutRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rollout.Status != entity.RolloutStatusActive {
		return fmt.Errorf("cannot increment rollout with status: %s", rollout.Status)
	}

	previousPercentage := rollout.CurrentPercentage
	continued := rollout.Increment()

	if err := s.rolloutRepo.Update(ctx, rollout); err != nil {
		return err
	}

	// Record history
	history := entity.NewRolloutHistory(
rollout.ID,
previousPercentage,
rollout.CurrentPercentage,
entity.RolloutActionIncrement,
fmt.Sprintf("Incremented from %d%% to %d%%", previousPercentage, rollout.CurrentPercentage),
nil,
)
	s.historyRepo.Create(ctx, history)

	// Check if completed
	if !continued && rollout.Status == entity.RolloutStatusCompleted {
		if s.webhookSvc != nil {
			s.webhookSvc.TriggerEvent(ctx, uuid.Nil, entity.WebhookEventRolloutComplete, rollout)
		}
	}

	return nil
}

// Rollback rolls back the rollout
func (s *RolloutService) Rollback(ctx context.Context, id uuid.UUID, reason string) error {
	rollout, err := s.rolloutRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rollout.Status == entity.RolloutStatusCompleted || rollout.Status == entity.RolloutStatusFailed {
		return fmt.Errorf("cannot rollback finished rollout")
	}

	previousPercentage := rollout.CurrentPercentage
	rollout.Rollback(reason)

	if err := s.rolloutRepo.Update(ctx, rollout); err != nil {
		return err
	}

	// Record history
	history := entity.NewRolloutHistory(
rollout.ID,
previousPercentage,
rollout.CurrentPercentage,
entity.RolloutActionRollback,
reason,
nil,
)
	s.historyRepo.Create(ctx, history)

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, uuid.Nil, entity.WebhookEventRolloutRollback, rollout)
	}

	return nil
}

// Complete completes a rollout immediately
func (s *RolloutService) Complete(ctx context.Context, id uuid.UUID) error {
	rollout, err := s.rolloutRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	previousPercentage := rollout.CurrentPercentage
	rollout.Complete()

	if err := s.rolloutRepo.Update(ctx, rollout); err != nil {
		return err
	}

	// Record history
	history := entity.NewRolloutHistory(
rollout.ID,
previousPercentage,
rollout.CurrentPercentage,
entity.RolloutActionComplete,
"Rollout completed",
nil,
)
	s.historyRepo.Create(ctx, history)

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, uuid.Nil, entity.WebhookEventRolloutComplete, rollout)
	}

	return nil
}

// Delete deletes a rollout plan
func (s *RolloutService) Delete(ctx context.Context, id uuid.UUID) error {
	rollout, err := s.rolloutRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rollout.Status == entity.RolloutStatusActive {
		return fmt.Errorf("cannot delete an active rollout")
	}

	// Delete history first
	s.historyRepo.DeleteByPlan(ctx, id)

	return s.rolloutRepo.Delete(ctx, id)
}

// GetHistory gets the history of a rollout
func (s *RolloutService) GetHistory(ctx context.Context, rolloutID uuid.UUID) ([]*entity.RolloutHistory, error) {
	return s.historyRepo.ListByPlan(ctx, rolloutID)
}

// ProcessScheduledIncrements processes rollouts that need automatic increment
func (s *RolloutService) ProcessScheduledIncrements(ctx context.Context) error {
	rollouts, err := s.rolloutRepo.ListNeedingIncrement(ctx)
	if err != nil {
		return err
	}

	for _, rollout := range rollouts {
		if rollout.NeedsIncrement() {
			s.Increment(ctx, rollout.ID)
		}
	}

	return nil
}
