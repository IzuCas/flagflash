package postgres

import (
"context"
"database/sql"
"encoding/json"
"fmt"
"time"

"github.com/IzuCas/flagflash/internal/domain/entity"
"github.com/IzuCas/flagflash/internal/domain/repository"
"github.com/google/uuid"
)

// ExperimentRepo implements repository.ExperimentRepository
type ExperimentRepo struct {
	db *DB
}

// NewExperimentRepo creates a new experiment repository
func NewExperimentRepo(db *DB) repository.ExperimentRepository {
	return &ExperimentRepo{db: db}
}

// Create creates a new experiment
func (r *ExperimentRepo) Create(ctx context.Context, exp *entity.Experiment) error {
	query := `
		INSERT INTO experiments (id, tenant_id, environment_id, feature_flag_id, name, description, hypothesis, status, started_at, ended_at, winner_variant, statistical_significance, target_sample_size, current_sample_size, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.ExecContext(ctx, query,
exp.ID, exp.TenantID, exp.EnvironmentID, exp.FeatureFlagID, exp.Name, exp.Description,
exp.Hypothesis, exp.Status, exp.StartedAt, exp.EndedAt, exp.WinnerVariant,
exp.StatisticalSignificance, exp.TargetSampleSize, exp.CurrentSampleSize, exp.CreatedBy, exp.CreatedAt, exp.UpdatedAt,
)
	if err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	return nil
}

// Update updates an experiment
func (r *ExperimentRepo) Update(ctx context.Context, exp *entity.Experiment) error {
	exp.UpdatedAt = time.Now()

	query := `
		UPDATE experiments
		SET name = $1, description = $2, hypothesis = $3, status = $4, started_at = $5, ended_at = $6, winner_variant = $7, statistical_significance = $8, target_sample_size = $9, current_sample_size = $10, updated_at = $11
		WHERE id = $12
	`

	result, err := r.db.ExecContext(ctx, query,
exp.Name, exp.Description, exp.Hypothesis, exp.Status, exp.StartedAt,
exp.EndedAt, exp.WinnerVariant, exp.StatisticalSignificance, exp.TargetSampleSize, exp.CurrentSampleSize,
exp.UpdatedAt, exp.ID,
)
	if err != nil {
		return fmt.Errorf("failed to update experiment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("experiment not found")
	}

	return nil
}

// Delete deletes an experiment
func (r *ExperimentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM experiments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete experiment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("experiment not found")
	}

	return nil
}

// GetByID retrieves an experiment by ID
func (r *ExperimentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Experiment, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, name, description, hypothesis, status, started_at, ended_at, winner_variant, statistical_significance, target_sample_size, current_sample_size, created_by, created_at, updated_at
		FROM experiments
		WHERE id = $1
	`

	return r.scanExperiment(r.db.QueryRowContext(ctx, query, id))
}

// GetByFlag retrieves an experiment by flag ID
func (r *ExperimentRepo) GetByFlag(ctx context.Context, flagID uuid.UUID) (*entity.Experiment, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, name, description, hypothesis, status, started_at, ended_at, winner_variant, statistical_significance, target_sample_size, current_sample_size, created_by, created_at, updated_at
		FROM experiments
		WHERE feature_flag_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	return r.scanExperiment(r.db.QueryRowContext(ctx, query, flagID))
}

// GetActiveByFlag retrieves the active experiment for a flag
func (r *ExperimentRepo) GetActiveByFlag(ctx context.Context, flagID uuid.UUID) (*entity.Experiment, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, name, description, hypothesis, status, started_at, ended_at, winner_variant, statistical_significance, target_sample_size, current_sample_size, created_by, created_at, updated_at
		FROM experiments
		WHERE feature_flag_id = $1 AND status = 'running'
	`

	return r.scanExperiment(r.db.QueryRowContext(ctx, query, flagID))
}

// GetByIDWithDetails retrieves an experiment with all related data
func (r *ExperimentRepo) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.Experiment, error) {
	exp, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load variants
	variantQuery := `
		SELECT id, experiment_id, name, description, value, weight, is_control, created_at
		FROM experiment_variants
		WHERE experiment_id = $1
		ORDER BY is_control DESC, name
	`

	rows, err := r.db.QueryContext(ctx, variantQuery, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var variant entity.ExperimentVariant
			var description sql.NullString
			var value []byte

			err := rows.Scan(&variant.ID, &variant.ExperimentID, &variant.Name, &description, &value, &variant.Weight, &variant.IsControl, &variant.CreatedAt)
			if err == nil {
				if description.Valid {
					variant.Description = description.String
				}
				variant.Value = json.RawMessage(value)
				exp.Variants = append(exp.Variants, variant)
			}
		}
	}

	// Load metrics
	metricQuery := `
		SELECT id, experiment_id, name, metric_type, is_primary, goal_direction, created_at
		FROM experiment_metrics
		WHERE experiment_id = $1
		ORDER BY is_primary DESC, name
	`

	rows, err = r.db.QueryContext(ctx, metricQuery, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var metric entity.ExperimentMetric
			err := rows.Scan(&metric.ID, &metric.ExperimentID, &metric.Name, &metric.MetricType, &metric.IsPrimary, &metric.GoalDirection, &metric.CreatedAt)
			if err == nil {
				exp.Metrics = append(exp.Metrics, metric)
			}
		}
	}

	return exp, nil
}

// ListByTenant lists all experiments for a tenant
func (r *ExperimentRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Experiment, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, name, description, hypothesis, status, started_at, ended_at, winner_variant, statistical_significance, target_sample_size, current_sample_size, created_by, created_at, updated_at
		FROM experiments
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	return r.scanExperiments(r.db.QueryContext(ctx, query, tenantID))
}

// ListByEnvironment lists all experiments for an environment
func (r *ExperimentRepo) ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.Experiment, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, name, description, hypothesis, status, started_at, ended_at, winner_variant, statistical_significance, target_sample_size, current_sample_size, created_by, created_at, updated_at
		FROM experiments
		WHERE environment_id = $1
		ORDER BY created_at DESC
	`

	return r.scanExperiments(r.db.QueryContext(ctx, query, environmentID))
}

// ListByStatus lists experiments by status
func (r *ExperimentRepo) ListByStatus(ctx context.Context, tenantID uuid.UUID, status entity.ExperimentStatus) ([]*entity.Experiment, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, name, description, hypothesis, status, started_at, ended_at, winner_variant, statistical_significance, target_sample_size, current_sample_size, created_by, created_at, updated_at
		FROM experiments
		WHERE tenant_id = $1 AND status = $2
		ORDER BY created_at DESC
	`

	return r.scanExperiments(r.db.QueryContext(ctx, query, tenantID, status))
}

// ListRunning lists all running experiments
func (r *ExperimentRepo) ListRunning(ctx context.Context) ([]*entity.Experiment, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, name, description, hypothesis, status, started_at, ended_at, winner_variant, statistical_significance, target_sample_size, current_sample_size, created_by, created_at, updated_at
		FROM experiments
		WHERE status = 'running'
		ORDER BY started_at
	`

	return r.scanExperiments(r.db.QueryContext(ctx, query))
}

func (r *ExperimentRepo) scanExperiment(row *sql.Row) (*entity.Experiment, error) {
	var exp entity.Experiment
	var description, hypothesis, winnerVariant sql.NullString
	var startedAt, endedAt sql.NullTime
	var significance sql.NullFloat64
	var targetSampleSize sql.NullInt32
	var createdBy sql.NullString

	err := row.Scan(
&exp.ID, &exp.TenantID, &exp.EnvironmentID, &exp.FeatureFlagID, &exp.Name, &description,
		&hypothesis, &exp.Status, &startedAt, &endedAt, &winnerVariant, &significance,
		&targetSampleSize, &exp.CurrentSampleSize, &createdBy, &exp.CreatedAt, &exp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("experiment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan experiment: %w", err)
	}

	if description.Valid {
		exp.Description = description.String
	}
	if hypothesis.Valid {
		exp.Hypothesis = hypothesis.String
	}
	if winnerVariant.Valid {
		exp.WinnerVariant = winnerVariant.String
	}
	if startedAt.Valid {
		exp.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		exp.EndedAt = &endedAt.Time
	}
	if significance.Valid {
		exp.StatisticalSignificance = &significance.Float64
	}
	if targetSampleSize.Valid {
		size := int(targetSampleSize.Int32)
		exp.TargetSampleSize = &size
	}
	if createdBy.Valid {
		id, _ := uuid.Parse(createdBy.String)
		exp.CreatedBy = &id
	}

	return &exp, nil
}

func (r *ExperimentRepo) scanExperiments(rows *sql.Rows, err error) ([]*entity.Experiment, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to query experiments: %w", err)
	}
	defer rows.Close()

	var experiments []*entity.Experiment
	for rows.Next() {
		var exp entity.Experiment
		var description, hypothesis, winnerVariant sql.NullString
		var startedAt, endedAt sql.NullTime
		var significance sql.NullFloat64
		var targetSampleSize sql.NullInt32
		var createdBy sql.NullString

		err := rows.Scan(
&exp.ID, &exp.TenantID, &exp.EnvironmentID, &exp.FeatureFlagID, &exp.Name, &description,
			&hypothesis, &exp.Status, &startedAt, &endedAt, &winnerVariant, &significance,
			&targetSampleSize, &exp.CurrentSampleSize, &createdBy, &exp.CreatedAt, &exp.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan experiment: %w", err)
		}

		if description.Valid {
			exp.Description = description.String
		}
		if hypothesis.Valid {
			exp.Hypothesis = hypothesis.String
		}
		if winnerVariant.Valid {
			exp.WinnerVariant = winnerVariant.String
		}
		if startedAt.Valid {
			exp.StartedAt = &startedAt.Time
		}
		if endedAt.Valid {
			exp.EndedAt = &endedAt.Time
		}
		if significance.Valid {
			exp.StatisticalSignificance = &significance.Float64
		}
		if targetSampleSize.Valid {
			size := int(targetSampleSize.Int32)
			exp.TargetSampleSize = &size
		}
		if createdBy.Valid {
			id, _ := uuid.Parse(createdBy.String)
			exp.CreatedBy = &id
		}

		experiments = append(experiments, &exp)
	}

	return experiments, nil
}
