package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// FlagCache defines the interface for flag caching
type FlagCache interface {
	GetFlags(ctx context.Context, environmentID uuid.UUID) ([]*entity.FeatureFlag, error)
	SetFlags(ctx context.Context, environmentID uuid.UUID, flags []*entity.FeatureFlag) error
	InvalidateFlags(ctx context.Context, environmentID uuid.UUID) error
	GetFlag(ctx context.Context, environmentID uuid.UUID, key string) (*entity.FeatureFlag, error)
	SetFlag(ctx context.Context, environmentID uuid.UUID, flag *entity.FeatureFlag) error
}

// FlagPublisher defines the interface for publishing flag updates
type FlagPublisher interface {
	PublishFlagUpdate(ctx context.Context, environmentID uuid.UUID, flag *entity.FeatureFlag) error
	PublishFlagDelete(ctx context.Context, environmentID uuid.UUID, flagKey string) error
}

// FeatureFlagService handles feature flag business logic
type FeatureFlagService struct {
	flagRepo      repository.FeatureFlagRepository
	targetingRepo repository.TargetingRuleRepository
	envRepo       repository.EnvironmentRepository
	auditRepo     repository.AuditLogRepository
	historyRepo   repository.FlagHistoryRepository
	cache         FlagCache
	publisher     FlagPublisher
}

// NewFeatureFlagService creates a new feature flag service
func NewFeatureFlagService(
	flagRepo repository.FeatureFlagRepository,
	targetingRepo repository.TargetingRuleRepository,
	envRepo repository.EnvironmentRepository,
	auditRepo repository.AuditLogRepository,
	historyRepo repository.FlagHistoryRepository,
	cache FlagCache,
	publisher FlagPublisher,
) *FeatureFlagService {
	return &FeatureFlagService{
		flagRepo:      flagRepo,
		targetingRepo: targetingRepo,
		envRepo:       envRepo,
		auditRepo:     auditRepo,
		historyRepo:   historyRepo,
		cache:         cache,
		publisher:     publisher,
	}
}

// Create creates a new feature flag
func (s *FeatureFlagService) Create(ctx context.Context, environmentID uuid.UUID, key, name, description string, flagType entity.FlagType, defaultValue json.RawMessage, tags []string, actorID string) (*entity.FeatureFlag, error) {
	// Validate environment exists
	env, tenantID, err := s.envRepo.GetByIDWithTenant(ctx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("environment not found: %w", err)
	}
	if env == nil {
		return nil, fmt.Errorf("environment not found")
	}

	// Check if key already exists
	existing, _ := s.flagRepo.GetByKey(ctx, environmentID, key)
	if existing != nil {
		return nil, fmt.Errorf("flag with key '%s' already exists in this environment", key)
	}

	flag := entity.NewFeatureFlag(environmentID, key, name, description, flagType, defaultValue)
	if tags != nil {
		flag.Tags = tags
	}

	if err := s.flagRepo.Create(ctx, flag); err != nil {
		return nil, fmt.Errorf("failed to create feature flag: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		s.cache.InvalidateFlags(ctx, environmentID)
	}

	// Publish update
	if s.publisher != nil {
		s.publisher.PublishFlagUpdate(ctx, environmentID, flag)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenantID,
		entity.EntityTypeFeatureFlag,
		flag.ID,
		entity.AuditActionCreate,
		actorID,
		entity.ActorTypeUser,
		nil,
		flag,
		map[string]interface{}{"environment": env.Name},
	)
	s.auditRepo.Create(ctx, auditLog)

	// Record in flag history
	s.recordFlagHistory(ctx, flag, nil, entity.FlagChangeTypeCreated, actorID)

	return flag, nil
}

// GetByID retrieves a feature flag by ID
func (s *FeatureFlagService) GetByID(ctx context.Context, id uuid.UUID) (*entity.FeatureFlag, error) {
	flag, err := s.flagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("feature flag not found: %w", err)
	}
	return flag, nil
}

// GetByKey retrieves a feature flag by environment ID and key
func (s *FeatureFlagService) GetByKey(ctx context.Context, environmentID uuid.UUID, key string) (*entity.FeatureFlag, error) {
	// Try cache first
	if s.cache != nil {
		flag, err := s.cache.GetFlag(ctx, environmentID, key)
		if err == nil && flag != nil {
			return flag, nil
		}
	}

	flag, err := s.flagRepo.GetByKey(ctx, environmentID, key)
	if err != nil {
		return nil, fmt.Errorf("feature flag not found: %w", err)
	}

	// Cache the flag
	if s.cache != nil {
		s.cache.SetFlag(ctx, environmentID, flag)
	}

	return flag, nil
}

// Update updates a feature flag
func (s *FeatureFlagService) Update(ctx context.Context, id uuid.UUID, name, description string, defaultValue json.RawMessage, tags []string, actorID string) (*entity.FeatureFlag, error) {
	flag, tenantID, err := s.flagRepo.GetFlagWithTenant(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("feature flag not found: %w", err)
	}

	oldFlag := *flag
	flag.Update(name, description, defaultValue, tags)

	if err := s.flagRepo.Update(ctx, flag); err != nil {
		return nil, fmt.Errorf("failed to update feature flag: %w", err)
	}

	// Invalidate cache and publish update
	if s.cache != nil {
		s.cache.InvalidateFlags(ctx, flag.EnvironmentID)
	}
	if s.publisher != nil {
		s.publisher.PublishFlagUpdate(ctx, flag.EnvironmentID, flag)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenantID,
		entity.EntityTypeFeatureFlag,
		flag.ID,
		entity.AuditActionUpdate,
		actorID,
		entity.ActorTypeUser,
		oldFlag,
		flag,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	// Record in flag history
	s.recordFlagHistory(ctx, flag, &oldFlag, entity.FlagChangeTypeUpdated, actorID)

	return flag, nil
}

// Toggle toggles a feature flag
func (s *FeatureFlagService) Toggle(ctx context.Context, id uuid.UUID, actorID string) (*entity.FeatureFlag, error) {
	flag, tenantID, err := s.flagRepo.GetFlagWithTenant(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("feature flag not found: %w", err)
	}

	oldFlag := *flag
	flag.Toggle()

	if err := s.flagRepo.Update(ctx, flag); err != nil {
		return nil, fmt.Errorf("failed to toggle feature flag: %w", err)
	}

	// Invalidate cache and publish update
	if s.cache != nil {
		s.cache.InvalidateFlags(ctx, flag.EnvironmentID)
	}
	if s.publisher != nil {
		s.publisher.PublishFlagUpdate(ctx, flag.EnvironmentID, flag)
	}

	// Create audit log
	action := entity.AuditActionEnable
	if !flag.Enabled {
		action = entity.AuditActionDisable
	}

	auditLog := entity.NewAuditLog(
		tenantID,
		entity.EntityTypeFeatureFlag,
		flag.ID,
		action,
		actorID,
		entity.ActorTypeUser,
		oldFlag,
		flag,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	// Record in flag history
	changeType := entity.FlagChangeTypeEnabled
	if !flag.Enabled {
		changeType = entity.FlagChangeTypeDisabled
	}
	s.recordFlagHistory(ctx, flag, &oldFlag, changeType, actorID)

	return flag, nil
}

// Delete soft deletes a feature flag
func (s *FeatureFlagService) Delete(ctx context.Context, id uuid.UUID, actorID string) error {
	flag, tenantID, err := s.flagRepo.GetFlagWithTenant(ctx, id)
	if err != nil {
		return fmt.Errorf("feature flag not found: %w", err)
	}

	if err := s.flagRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete feature flag: %w", err)
	}

	// Invalidate cache and publish delete
	if s.cache != nil {
		s.cache.InvalidateFlags(ctx, flag.EnvironmentID)
	}
	if s.publisher != nil {
		s.publisher.PublishFlagDelete(ctx, flag.EnvironmentID, flag.Key)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenantID,
		entity.EntityTypeFeatureFlag,
		flag.ID,
		entity.AuditActionDelete,
		actorID,
		entity.ActorTypeUser,
		flag,
		nil,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	// Record in flag history
	s.recordFlagHistory(ctx, flag, flag, entity.FlagChangeTypeDeleted, actorID)

	return nil
}

// ListByEnvironment lists all feature flags for an environment
func (s *FeatureFlagService) ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.FeatureFlag, error) {
	// Try cache first
	if s.cache != nil {
		flags, err := s.cache.GetFlags(ctx, environmentID)
		if err == nil && flags != nil {
			return flags, nil
		}
	}

	flags, err := s.flagRepo.ListByEnvironment(ctx, environmentID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list feature flags: %w", err)
	}

	// Cache the flags
	if s.cache != nil {
		s.cache.SetFlags(ctx, environmentID, flags)
	}

	return flags, nil
}

// ListByEnvironmentWithPagination lists feature flags with pagination
func (s *FeatureFlagService) ListByEnvironmentWithPagination(ctx context.Context, environmentID uuid.UUID, limit, offset int, search string) ([]*entity.FeatureFlag, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.flagRepo.ListByEnvironmentWithPagination(ctx, environmentID, limit, offset, search)
}

// CopyFlags copies flags from one environment to another
func (s *FeatureFlagService) CopyFlags(ctx context.Context, sourceEnvID, targetEnvID uuid.UUID, actorID string) error {
	// Validate both environments exist
	_, err := s.envRepo.GetByID(ctx, sourceEnvID)
	if err != nil {
		return fmt.Errorf("source environment not found: %w", err)
	}

	_, err = s.envRepo.GetByID(ctx, targetEnvID)
	if err != nil {
		return fmt.Errorf("target environment not found: %w", err)
	}

	if err := s.flagRepo.CopyFlags(ctx, sourceEnvID, targetEnvID); err != nil {
		return fmt.Errorf("failed to copy flags: %w", err)
	}

	// Invalidate target cache
	if s.cache != nil {
		s.cache.InvalidateFlags(ctx, targetEnvID)
	}

	return nil
}

// GetHistory retrieves the history of changes for a feature flag
func (s *FeatureFlagService) GetHistory(ctx context.Context, flagID uuid.UUID, limit int) ([]*entity.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.auditRepo.GetByEntity(ctx, entity.EntityTypeFeatureFlag, flagID, limit)
}

// FlagWithRules represents a flag with its targeting rules
type FlagWithRules struct {
	*entity.FeatureFlag
	TargetingRules []*entity.TargetingRule `json:"targeting_rules"`
}

// GetByIDWithRules retrieves a flag with its targeting rules
func (s *FeatureFlagService) GetByIDWithRules(ctx context.Context, id uuid.UUID) (*FlagWithRules, error) {
	flag, err := s.flagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("feature flag not found: %w", err)
	}

	rules, err := s.targetingRepo.ListByFlag(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get targeting rules: %w", err)
	}

	return &FlagWithRules{
		FeatureFlag:    flag,
		TargetingRules: rules,
	}, nil
}

// CreateTargetingRule creates a new targeting rule
func (s *FeatureFlagService) CreateTargetingRule(ctx context.Context, rule *entity.TargetingRule) error {
	// Verify flag exists
	_, err := s.flagRepo.GetByID(ctx, rule.FeatureFlagID)
	if err != nil {
		return fmt.Errorf("feature flag not found: %w", err)
	}

	if err := s.targetingRepo.Create(ctx, rule); err != nil {
		return fmt.Errorf("failed to create targeting rule: %w", err)
	}

	return nil
}

// GetTargetingRule retrieves a targeting rule by ID
func (s *FeatureFlagService) GetTargetingRule(ctx context.Context, id uuid.UUID) (*entity.TargetingRule, error) {
	rule, err := s.targetingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("targeting rule not found: %w", err)
	}
	return rule, nil
}

// ListTargetingRules lists all targeting rules for a flag
func (s *FeatureFlagService) ListTargetingRules(ctx context.Context, flagID uuid.UUID) ([]*entity.TargetingRule, error) {
	return s.targetingRepo.ListByFlag(ctx, flagID)
}

// UpdateTargetingRule updates a targeting rule
func (s *FeatureFlagService) UpdateTargetingRule(ctx context.Context, rule *entity.TargetingRule) error {
	if err := s.targetingRepo.Update(ctx, rule); err != nil {
		return fmt.Errorf("failed to update targeting rule: %w", err)
	}
	return nil
}

// DeleteTargetingRule deletes a targeting rule
func (s *FeatureFlagService) DeleteTargetingRule(ctx context.Context, id uuid.UUID) error {
	if err := s.targetingRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete targeting rule: %w", err)
	}
	return nil
}

// ReorderTargetingRules reorders targeting rules by priority
func (s *FeatureFlagService) ReorderTargetingRules(ctx context.Context, flagID uuid.UUID, ruleIDs []uuid.UUID) error {
	for i, ruleID := range ruleIDs {
		rule, err := s.targetingRepo.GetByID(ctx, ruleID)
		if err != nil {
			return fmt.Errorf("targeting rule not found: %w", err)
		}
		if rule.FeatureFlagID != flagID {
			return fmt.Errorf("targeting rule does not belong to this flag")
		}
		rule.Priority = i
		if err := s.targetingRepo.Update(ctx, rule); err != nil {
			return fmt.Errorf("failed to update targeting rule priority: %w", err)
		}
	}
	return nil
}

// recordFlagHistory records a flag change in history
func (s *FeatureFlagService) recordFlagHistory(ctx context.Context, flag *entity.FeatureFlag, previousFlag *entity.FeatureFlag, changeType entity.FlagChangeType, actorID string) {
	if s.historyRepo == nil {
		return
	}

	var changedBy *uuid.UUID
	if actorID != "" {
		if id, err := uuid.Parse(actorID); err == nil {
			changedBy = &id
		}
	}

	history := entity.FlagHistoryFromFlag(flag, changeType, changedBy, previousFlag, "")
	s.historyRepo.Create(ctx, history)
}
