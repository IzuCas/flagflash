package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// SchedulerService handles scheduled flag operations
type SchedulerService struct {
	flagRepo           repository.FeatureFlagRepository
	auditRepo          repository.AuditLogRepository
	rolloutRepo        repository.RolloutPlanRepository
	rolloutHistoryRepo repository.RolloutHistoryRepository
	emergencyRepo      repository.EmergencyControlRepository
	pendingChangeRepo  repository.PendingChangeRepository
	webhookSvc         *WebhookService
	notificationSvc    *NotificationService
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(
	flagRepo repository.FeatureFlagRepository,
	auditRepo repository.AuditLogRepository,
	rolloutRepo repository.RolloutPlanRepository,
	rolloutHistoryRepo repository.RolloutHistoryRepository,
	emergencyRepo repository.EmergencyControlRepository,
	pendingChangeRepo repository.PendingChangeRepository,
	webhookSvc *WebhookService,
	notificationSvc *NotificationService,
) *SchedulerService {
	return &SchedulerService{
		flagRepo:           flagRepo,
		auditRepo:          auditRepo,
		rolloutRepo:        rolloutRepo,
		rolloutHistoryRepo: rolloutHistoryRepo,
		emergencyRepo:      emergencyRepo,
		pendingChangeRepo:  pendingChangeRepo,
		webhookSvc:         webhookSvc,
		notificationSvc:    notificationSvc,
	}
}

// ProcessScheduledFlags processes flags that need to be enabled/disabled based on schedule
func (s *SchedulerService) ProcessScheduledFlags(ctx context.Context) error {
	now := time.Now()

	// Get all environments and process their flags
	// This is a simplified version - in production, use a more efficient query

	// Process enable schedules
	enableCount, err := s.processEnableSchedules(ctx, now)
	if err != nil {
		log.Printf("Error processing enable schedules: %v", err)
	}

	// Process disable schedules
	disableCount, err := s.processDisableSchedules(ctx, now)
	if err != nil {
		log.Printf("Error processing disable schedules: %v", err)
	}

	if enableCount > 0 || disableCount > 0 {
		log.Printf("Scheduler: enabled %d flags, disabled %d flags", enableCount, disableCount)
	}

	return nil
}

func (s *SchedulerService) processEnableSchedules(ctx context.Context, now time.Time) (int, error) {
	// In a real implementation, this would query flags with scheduled_enable_at <= now
	// For now, we'll use a simple approach
	return 0, nil
}

func (s *SchedulerService) processDisableSchedules(ctx context.Context, now time.Time) (int, error) {
	// In a real implementation, this would query flags with scheduled_disable_at <= now
	return 0, nil
}

// ProcessRollouts processes active rollout plans
func (s *SchedulerService) ProcessRollouts(ctx context.Context) error {
	plans, err := s.rolloutRepo.ListNeedingIncrement(ctx)
	if err != nil {
		return fmt.Errorf("failed to list rollout plans: %w", err)
	}

	for _, plan := range plans {
		if err := s.processRollout(ctx, plan); err != nil {
			log.Printf("Error processing rollout %s: %v", plan.ID, err)
		}
	}

	return nil
}

func (s *SchedulerService) processRollout(ctx context.Context, plan *entity.RolloutPlan) error {
	if !plan.NeedsIncrement() {
		return nil
	}

	fromPct := plan.CurrentPercentage

	// TODO: Check metrics for auto-rollback
	// If metrics exceed thresholds, rollback instead of increment

	if plan.Increment() {
		// Record history
		history := entity.NewRolloutHistory(
			plan.ID,
			fromPct,
			plan.CurrentPercentage,
			entity.RolloutActionIncrement,
			"Automatic increment",
			nil,
		)

		if err := s.rolloutHistoryRepo.Create(ctx, history); err != nil {
			log.Printf("Failed to create rollout history: %v", err)
		}

		if err := s.rolloutRepo.Update(ctx, plan); err != nil {
			return fmt.Errorf("failed to update rollout plan: %w", err)
		}

		// TODO: Update the actual flag's targeting percentage

		log.Printf("Rollout %s: incremented from %d%% to %d%%", plan.Name, fromPct, plan.CurrentPercentage)
	}

	return nil
}

// ProcessExpiredEmergencyControls disables expired emergency controls
func (s *SchedulerService) ProcessExpiredEmergencyControls(ctx context.Context) error {
	count, err := s.emergencyRepo.DisableExpired(ctx)
	if err != nil {
		return fmt.Errorf("failed to disable expired emergency controls: %w", err)
	}

	if count > 0 {
		log.Printf("Disabled %d expired emergency controls", count)
	}

	return nil
}

// ProcessExpiredPendingChanges expires old pending changes
func (s *SchedulerService) ProcessExpiredPendingChanges(ctx context.Context) error {
	count, err := s.pendingChangeRepo.ExpireOld(ctx)
	if err != nil {
		return fmt.Errorf("failed to expire pending changes: %w", err)
	}

	if count > 0 {
		log.Printf("Expired %d pending changes", count)
	}

	return nil
}

// ScheduleEnable schedules a flag to be enabled at a specific time
func (s *SchedulerService) ScheduleEnable(ctx context.Context, flagID uuid.UUID, enableAt time.Time, timezone string, actorID string) error {
	flag, err := s.flagRepo.GetByID(ctx, flagID)
	if err != nil {
		return fmt.Errorf("failed to get flag: %w", err)
	}

	flag.ScheduledEnableAt = &enableAt
	flag.ScheduleTimezone = timezone
	flag.UpdatedAt = time.Now()

	if err := s.flagRepo.Update(ctx, flag); err != nil {
		return fmt.Errorf("failed to update flag: %w", err)
	}

	return nil
}

// ScheduleDisable schedules a flag to be disabled at a specific time
func (s *SchedulerService) ScheduleDisable(ctx context.Context, flagID uuid.UUID, disableAt time.Time, timezone string, actorID string) error {
	flag, err := s.flagRepo.GetByID(ctx, flagID)
	if err != nil {
		return fmt.Errorf("failed to get flag: %w", err)
	}

	flag.ScheduledDisableAt = &disableAt
	flag.ScheduleTimezone = timezone
	flag.UpdatedAt = time.Now()

	if err := s.flagRepo.Update(ctx, flag); err != nil {
		return fmt.Errorf("failed to update flag: %w", err)
	}

	return nil
}

// CancelSchedule cancels any scheduled enable/disable
func (s *SchedulerService) CancelSchedule(ctx context.Context, flagID uuid.UUID, actorID string) error {
	flag, err := s.flagRepo.GetByID(ctx, flagID)
	if err != nil {
		return fmt.Errorf("failed to get flag: %w", err)
	}

	flag.ScheduledEnableAt = nil
	flag.ScheduledDisableAt = nil
	flag.UpdatedAt = time.Now()

	if err := s.flagRepo.Update(ctx, flag); err != nil {
		return fmt.Errorf("failed to update flag: %w", err)
	}

	return nil
}

// RunSchedulerTick runs all scheduler tasks
func (s *SchedulerService) RunSchedulerTick(ctx context.Context) error {
	// Process scheduled flags
	if err := s.ProcessScheduledFlags(ctx); err != nil {
		log.Printf("Error in ProcessScheduledFlags: %v", err)
	}

	// Process rollouts
	if err := s.ProcessRollouts(ctx); err != nil {
		log.Printf("Error in ProcessRollouts: %v", err)
	}

	// Process expired emergency controls
	if err := s.ProcessExpiredEmergencyControls(ctx); err != nil {
		log.Printf("Error in ProcessExpiredEmergencyControls: %v", err)
	}

	// Process expired pending changes
	if err := s.ProcessExpiredPendingChanges(ctx); err != nil {
		log.Printf("Error in ProcessExpiredPendingChanges: %v", err)
	}

	return nil
}
