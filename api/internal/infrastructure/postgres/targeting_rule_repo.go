package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// TargetingRuleRepository implements repository.TargetingRuleRepository
type TargetingRuleRepository struct {
	db *sql.DB
}

// NewTargetingRuleRepository creates a new TargetingRuleRepository
func NewTargetingRuleRepository(db *sql.DB) *TargetingRuleRepository {
	return &TargetingRuleRepository{db: db}
}

// Create creates a new targeting rule
func (r *TargetingRuleRepository) Create(ctx context.Context, rule *entity.TargetingRule) error {
	conditionsJSON, _ := json.Marshal(rule.Conditions)
	valueJSON, _ := json.Marshal(rule.Value)

	query := `INSERT INTO targeting_rules (id, feature_flag_id, name, priority, conditions, value, enabled, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.FlagID, rule.Name, rule.Priority, conditionsJSON, valueJSON,
		rule.Enabled, rule.CreatedAt, rule.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create targeting rule: %w", err)
	}
	return nil
}

// GetByID retrieves a targeting rule by ID
func (r *TargetingRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.TargetingRule, error) {
	query := `SELECT id, feature_flag_id, name, priority, conditions, value, enabled, created_at, updated_at
			  FROM targeting_rules WHERE id = $1`
	return r.scanRule(r.db.QueryRowContext(ctx, query, id))
}

// GetByFlagID retrieves all targeting rules for a flag
func (r *TargetingRuleRepository) GetByFlagID(ctx context.Context, flagID uuid.UUID) ([]*entity.TargetingRule, error) {
	query := `SELECT id, feature_flag_id, name, priority, conditions, value, enabled, created_at, updated_at
			  FROM targeting_rules WHERE feature_flag_id = $1 ORDER BY priority`
	rows, err := r.db.QueryContext(ctx, query, flagID)
	if err != nil {
		return nil, fmt.Errorf("failed to get targeting rules: %w", err)
	}
	defer rows.Close()
	return r.scanRules(rows)
}

// Update updates a targeting rule
func (r *TargetingRuleRepository) Update(ctx context.Context, rule *entity.TargetingRule) error {
	conditionsJSON, _ := json.Marshal(rule.Conditions)
	valueJSON, _ := json.Marshal(rule.Value)

	query := `UPDATE targeting_rules SET name = $1, priority = $2, conditions = $3, value = $4, enabled = $5, updated_at = $6 WHERE id = $7`
	_, err := r.db.ExecContext(ctx, query, rule.Name, rule.Priority, conditionsJSON, valueJSON, rule.Enabled, rule.UpdatedAt, rule.ID)
	if err != nil {
		return fmt.Errorf("failed to update targeting rule: %w", err)
	}
	return nil
}

// Delete deletes a targeting rule
func (r *TargetingRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM targeting_rules WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete targeting rule: %w", err)
	}
	return nil
}

// DeleteByFlag deletes all targeting rules for a feature flag
func (r *TargetingRuleRepository) DeleteByFlag(ctx context.Context, flagID uuid.UUID) error {
	query := `DELETE FROM targeting_rules WHERE feature_flag_id = $1`
	_, err := r.db.ExecContext(ctx, query, flagID)
	if err != nil {
		return fmt.Errorf("failed to delete targeting rules for flag: %w", err)
	}
	return nil
}

func (r *TargetingRuleRepository) scanRule(row *sql.Row) (*entity.TargetingRule, error) {
	rule := &entity.TargetingRule{}
	var conditionsJSON, valueJSON []byte
	err := row.Scan(
		&rule.ID, &rule.FlagID, &rule.Name, &rule.Priority, &conditionsJSON, &valueJSON,
		&rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan targeting rule: %w", err)
	}
	json.Unmarshal(conditionsJSON, &rule.Conditions)
	json.Unmarshal(valueJSON, &rule.Value)
	return rule, nil
}

func (r *TargetingRuleRepository) scanRules(rows *sql.Rows) ([]*entity.TargetingRule, error) {
	var rules []*entity.TargetingRule
	for rows.Next() {
		rule := &entity.TargetingRule{}
		var conditionsJSON, valueJSON []byte
		if err := rows.Scan(
			&rule.ID, &rule.FlagID, &rule.Name, &rule.Priority, &conditionsJSON, &valueJSON,
			&rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan targeting rule: %w", err)
		}
		json.Unmarshal(conditionsJSON, &rule.Conditions)
		json.Unmarshal(valueJSON, &rule.Value)
		rules = append(rules, rule)
	}
	return rules, nil
}

// ListByFlag lists all targeting rules for a feature flag (ordered by priority)
func (r *TargetingRuleRepository) ListByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.TargetingRule, error) {
	return r.GetByFlagID(ctx, flagID)
}

// ReorderRules updates the priority of multiple rules
func (r *TargetingRuleRepository) ReorderRules(ctx context.Context, rules []*entity.TargetingRule) error {
	for _, rule := range rules {
		query := `UPDATE targeting_rules SET priority = $1, updated_at = $2 WHERE id = $3`
		_, err := r.db.ExecContext(ctx, query, rule.Priority, rule.UpdatedAt, rule.ID)
		if err != nil {
			return fmt.Errorf("failed to reorder targeting rule: %w", err)
		}
	}
	return nil
}
