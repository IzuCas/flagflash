package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// RolloutPlanRepo implements repository.RolloutPlanRepository
type RolloutPlanRepo struct {
	db *DB
}

// NewRolloutPlanRepo creates a new rollout plan repository
func NewRolloutPlanRepo(db *DB) repository.RolloutPlanRepository {
	return &RolloutPlanRepo{db: db}
}

// Create creates a new rollout plan
func (r *RolloutPlanRepo) Create(ctx context.Context, plan *entity.RolloutPlan) error {
	query := `
		INSERT INTO rollout_plans (id, feature_flag_id, name, status, current_percentage, target_percentage, increment_percentage, increment_interval_minutes, auto_rollback, rollback_threshold_error_rate, rollback_threshold_latency_ms, last_increment_at, next_increment_at, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.ExecContext(ctx, query,
		plan.ID, plan.FeatureFlagID, plan.Name, plan.Status, plan.CurrentPercentage,
		plan.TargetPercentage, plan.IncrementPercentage, plan.IncrementIntervalMinutes,
		plan.AutoRollback, plan.RollbackThresholdErrorRate, plan.RollbackThresholdLatencyMs,
		plan.LastIncrementAt, plan.NextIncrementAt, plan.CreatedBy, plan.CreatedAt, plan.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create rollout plan: %w", err)
	}

	return nil
}

// GetByID retrieves a rollout plan by ID
func (r *RolloutPlanRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.RolloutPlan, error) {
	query := `
		SELECT id, feature_flag_id, name, status, current_percentage, target_percentage, increment_percentage, increment_interval_minutes, auto_rollback, rollback_threshold_error_rate, rollback_threshold_latency_ms, last_increment_at, next_increment_at, created_by, created_at, updated_at
		FROM rollout_plans
		WHERE id = $1
	`

	return r.scanPlan(r.db.QueryRowContext(ctx, query, id))
}

// GetByFlag retrieves a rollout plan by flag ID
func (r *RolloutPlanRepo) GetByFlag(ctx context.Context, flagID uuid.UUID) (*entity.RolloutPlan, error) {
	query := `
		SELECT id, feature_flag_id, name, status, current_percentage, target_percentage, increment_percentage, increment_interval_minutes, auto_rollback, rollback_threshold_error_rate, rollback_threshold_latency_ms, last_increment_at, next_increment_at, created_by, created_at, updated_at
		FROM rollout_plans
		WHERE feature_flag_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	return r.scanPlan(r.db.QueryRowContext(ctx, query, flagID))
}

// GetActiveByFlag retrieves the active rollout plan for a flag
func (r *RolloutPlanRepo) GetActiveByFlag(ctx context.Context, flagID uuid.UUID) (*entity.RolloutPlan, error) {
	query := `
		SELECT id, feature_flag_id, name, status, current_percentage, target_percentage, increment_percentage, increment_interval_minutes, auto_rollback, rollback_threshold_error_rate, rollback_threshold_latency_ms, last_increment_at, next_increment_at, created_by, created_at, updated_at
		FROM rollout_plans
		WHERE feature_flag_id = $1 AND status = 'active'
	`

	return r.scanPlan(r.db.QueryRowContext(ctx, query, flagID))
}

// ListByFlag lists all rollout plans for a flag
func (r *RolloutPlanRepo) ListByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.RolloutPlan, error) {
	query := `
		SELECT id, feature_flag_id, name, status, current_percentage, target_percentage, increment_percentage, increment_interval_minutes, auto_rollback, rollback_threshold_error_rate, rollback_threshold_latency_ms, last_increment_at, next_increment_at, created_by, created_at, updated_at
		FROM rollout_plans
		WHERE feature_flag_id = $1
		ORDER BY created_at DESC
	`

	return r.scanPlans(r.db.QueryContext(ctx, query, flagID))
}

// ListActive lists all active rollout plans
func (r *RolloutPlanRepo) ListActive(ctx context.Context) ([]*entity.RolloutPlan, error) {
	query := `
		SELECT id, feature_flag_id, name, status, current_percentage, target_percentage, increment_percentage, increment_interval_minutes, auto_rollback, rollback_threshold_error_rate, rollback_threshold_latency_ms, last_increment_at, next_increment_at, created_by, created_at, updated_at
		FROM rollout_plans
		WHERE status = 'active'
		ORDER BY next_increment_at
	`

	return r.scanPlans(r.db.QueryContext(ctx, query))
}

// ListNeedingIncrement lists plans that need percentage increment
func (r *RolloutPlanRepo) ListNeedingIncrement(ctx context.Context) ([]*entity.RolloutPlan, error) {
	query := `
		SELECT id, feature_flag_id, name, status, current_percentage, target_percentage, increment_percentage, increment_interval_minutes, auto_rollback, rollback_threshold_error_rate, rollback_threshold_latency_ms, last_increment_at, next_increment_at, created_by, created_at, updated_at
		FROM rollout_plans
		WHERE status = 'active' AND next_increment_at <= NOW() AND current_percentage < target_percentage
		ORDER BY next_increment_at
	`

	return r.scanPlans(r.db.QueryContext(ctx, query))
}

// Update updates a rollout plan
func (r *RolloutPlanRepo) Update(ctx context.Context, plan *entity.RolloutPlan) error {
	plan.UpdatedAt = time.Now()

	query := `
		UPDATE rollout_plans
		SET name = $1, status = $2, current_percentage = $3, target_percentage = $4, increment_percentage = $5, increment_interval_minutes = $6, auto_rollback = $7, rollback_threshold_error_rate = $8, rollback_threshold_latency_ms = $9, last_increment_at = $10, next_increment_at = $11, updated_at = $12
		WHERE id = $13
	`

	result, err := r.db.ExecContext(ctx, query,
		plan.Name, plan.Status, plan.CurrentPercentage, plan.TargetPercentage,
		plan.IncrementPercentage, plan.IncrementIntervalMinutes, plan.AutoRollback,
		plan.RollbackThresholdErrorRate, plan.RollbackThresholdLatencyMs,
		plan.LastIncrementAt, plan.NextIncrementAt, plan.UpdatedAt, plan.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update rollout plan: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rollout plan not found")
	}

	return nil
}

// Delete deletes a rollout plan
func (r *RolloutPlanRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM rollout_plans WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete rollout plan: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rollout plan not found")
	}

	return nil
}

func (r *RolloutPlanRepo) scanPlan(row *sql.Row) (*entity.RolloutPlan, error) {
	var plan entity.RolloutPlan
	var rollbackErrorRate sql.NullFloat64
	var rollbackLatencyMs sql.NullInt32
	var lastIncrementAt, nextIncrementAt sql.NullTime
	var createdBy sql.NullString

	err := row.Scan(
		&plan.ID, &plan.FeatureFlagID, &plan.Name, &plan.Status,
		&plan.CurrentPercentage, &plan.TargetPercentage, &plan.IncrementPercentage,
		&plan.IncrementIntervalMinutes, &plan.AutoRollback, &rollbackErrorRate,
		&rollbackLatencyMs, &lastIncrementAt, &nextIncrementAt, &createdBy,
		&plan.CreatedAt, &plan.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rollout plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan rollout plan: %w", err)
	}

	if rollbackErrorRate.Valid {
		plan.RollbackThresholdErrorRate = &rollbackErrorRate.Float64
	}
	if rollbackLatencyMs.Valid {
		ms := int(rollbackLatencyMs.Int32)
		plan.RollbackThresholdLatencyMs = &ms
	}
	if lastIncrementAt.Valid {
		plan.LastIncrementAt = &lastIncrementAt.Time
	}
	if nextIncrementAt.Valid {
		plan.NextIncrementAt = &nextIncrementAt.Time
	}
	if createdBy.Valid {
		id, _ := uuid.Parse(createdBy.String)
		plan.CreatedBy = &id
	}

	return &plan, nil
}

func (r *RolloutPlanRepo) scanPlans(rows *sql.Rows, err error) ([]*entity.RolloutPlan, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to query rollout plans: %w", err)
	}
	defer rows.Close()

	var plans []*entity.RolloutPlan
	for rows.Next() {
		var plan entity.RolloutPlan
		var rollbackErrorRate sql.NullFloat64
		var rollbackLatencyMs sql.NullInt32
		var lastIncrementAt, nextIncrementAt sql.NullTime
		var createdBy sql.NullString

		err := rows.Scan(
			&plan.ID, &plan.FeatureFlagID, &plan.Name, &plan.Status,
			&plan.CurrentPercentage, &plan.TargetPercentage, &plan.IncrementPercentage,
			&plan.IncrementIntervalMinutes, &plan.AutoRollback, &rollbackErrorRate,
			&rollbackLatencyMs, &lastIncrementAt, &nextIncrementAt, &createdBy,
			&plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollout plan: %w", err)
		}

		if rollbackErrorRate.Valid {
			plan.RollbackThresholdErrorRate = &rollbackErrorRate.Float64
		}
		if rollbackLatencyMs.Valid {
			ms := int(rollbackLatencyMs.Int32)
			plan.RollbackThresholdLatencyMs = &ms
		}
		if lastIncrementAt.Valid {
			plan.LastIncrementAt = &lastIncrementAt.Time
		}
		if nextIncrementAt.Valid {
			plan.NextIncrementAt = &nextIncrementAt.Time
		}
		if createdBy.Valid {
			id, _ := uuid.Parse(createdBy.String)
			plan.CreatedBy = &id
		}

		plans = append(plans, &plan)
	}

	return plans, nil
}

// RolloutHistoryRepo implements repository.RolloutHistoryRepository
type RolloutHistoryRepo struct {
	db *DB
}

// NewRolloutHistoryRepo creates a new rollout history repository
func NewRolloutHistoryRepo(db *DB) repository.RolloutHistoryRepository {
	return &RolloutHistoryRepo{db: db}
}

// Create creates a new rollout history entry
func (r *RolloutHistoryRepo) Create(ctx context.Context, history *entity.RolloutHistory) error {
	query := `
		INSERT INTO rollout_history (id, rollout_plan_id, action, from_percentage, to_percentage, reason, metrics, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		history.ID, history.RolloutPlanID, history.Action, history.FromPercentage,
		history.ToPercentage, history.Reason, history.Metrics, history.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create rollout history: %w", err)
	}

	return nil
}

// GetByID retrieves a rollout history entry by ID
func (r *RolloutHistoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.RolloutHistory, error) {
	query := `
		SELECT id, rollout_plan_id, action, from_percentage, to_percentage, reason, metrics, created_at
		FROM rollout_history
		WHERE id = $1
	`

	var history entity.RolloutHistory
	var reason sql.NullString
	var metrics []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&history.ID, &history.RolloutPlanID, &history.Action, &history.FromPercentage,
		&history.ToPercentage, &reason, &metrics, &history.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rollout history not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan rollout history: %w", err)
	}

	if reason.Valid {
		history.Reason = reason.String
	}
	if len(metrics) > 0 {
		history.Metrics = metrics
	}

	return &history, nil
}

// ListByPlan lists all history for a rollout plan
func (r *RolloutHistoryRepo) ListByPlan(ctx context.Context, planID uuid.UUID) ([]*entity.RolloutHistory, error) {
	query := `
		SELECT id, rollout_plan_id, action, from_percentage, to_percentage, reason, metrics, created_at
		FROM rollout_history
		WHERE rollout_plan_id = $1
		ORDER BY created_at DESC
	`

	return r.scanHistories(r.db.QueryContext(ctx, query, planID))
}

// ListByPlanPaginated lists history with pagination
func (r *RolloutHistoryRepo) ListByPlanPaginated(ctx context.Context, planID uuid.UUID, limit, offset int) ([]*entity.RolloutHistory, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM rollout_history WHERE rollout_plan_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, planID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count rollout history: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, rollout_plan_id, action, from_percentage, to_percentage, reason, metrics, created_at
		FROM rollout_history
		WHERE rollout_plan_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	histories, err := r.scanHistories(r.db.QueryContext(ctx, query, planID, limit, offset))
	if err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// Delete deletes a rollout history entry
func (r *RolloutHistoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM rollout_history WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete rollout history: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("rollout history not found")
	}

	return nil
}

// DeleteByPlan deletes all history for a rollout plan
func (r *RolloutHistoryRepo) DeleteByPlan(ctx context.Context, planID uuid.UUID) error {
	query := `DELETE FROM rollout_history WHERE rollout_plan_id = $1`

	_, err := r.db.ExecContext(ctx, query, planID)
	if err != nil {
		return fmt.Errorf("failed to delete rollout history: %w", err)
	}

	return nil
}

func (r *RolloutHistoryRepo) scanHistories(rows *sql.Rows, err error) ([]*entity.RolloutHistory, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to query rollout history: %w", err)
	}
	defer rows.Close()

	var histories []*entity.RolloutHistory
	for rows.Next() {
		var history entity.RolloutHistory
		var reason sql.NullString
		var metrics []byte

		err := rows.Scan(
			&history.ID, &history.RolloutPlanID, &history.Action, &history.FromPercentage,
			&history.ToPercentage, &reason, &metrics, &history.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rollout history: %w", err)
		}

		if reason.Valid {
			history.Reason = reason.String
		}
		if len(metrics) > 0 {
			history.Metrics = metrics
		}

		histories = append(histories, &history)
	}

	return histories, nil
}
