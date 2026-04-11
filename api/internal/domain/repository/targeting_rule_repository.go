package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// TargetingRuleRepository defines the interface for targeting rule persistence
type TargetingRuleRepository interface {
	// Create creates a new targeting rule
	Create(ctx context.Context, rule *entity.TargetingRule) error

	// GetByID retrieves a targeting rule by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.TargetingRule, error)

	// Update updates a targeting rule
	Update(ctx context.Context, rule *entity.TargetingRule) error

	// Delete deletes a targeting rule
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByFlag lists all targeting rules for a feature flag (ordered by priority)
	ListByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.TargetingRule, error)

	// DeleteByFlag deletes all targeting rules for a feature flag
	DeleteByFlag(ctx context.Context, flagID uuid.UUID) error

	// ReorderRules updates the priority of multiple rules
	ReorderRules(ctx context.Context, rules []*entity.TargetingRule) error
}
