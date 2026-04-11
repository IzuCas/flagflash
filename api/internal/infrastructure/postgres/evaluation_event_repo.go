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

type evaluationEventRepo struct {
	db *sql.DB
}

// NewEvaluationEventRepository creates a new Postgres evaluation event repository
func NewEvaluationEventRepository(db *sql.DB) repository.EvaluationEventRepository {
	return &evaluationEventRepo{db: db}
}

func (r *evaluationEventRepo) Create(ctx context.Context, event *entity.EvaluationEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.EvaluatedAt.IsZero() {
		event.EvaluatedAt = time.Now()
	}

	valueJSON, err := json.Marshal(event.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	contextJSON, err := json.Marshal(event.Context)
	if err != nil {
		contextJSON = []byte("{}")
	}

	query := `
		INSERT INTO evaluation_events (id, tenant_id, environment_id, feature_flag_id, flag_key, value, user_id, context, sdk_type, sdk_version, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		event.ID,
		event.TenantID,
		event.EnvironmentID,
		event.FeatureFlagID,
		event.FlagKey,
		valueJSON,
		event.UserID,
		contextJSON,
		event.SDKType,
		event.SDKVersion,
		event.EvaluatedAt,
	)

	return err
}

func (r *evaluationEventRepo) CreateBatch(ctx context.Context, events []*entity.EvaluationEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO evaluation_events (id, tenant_id, environment_id, feature_flag_id, flag_key, value, user_id, context, sdk_type, sdk_version, evaluated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, event := range events {
		if event.ID == uuid.Nil {
			event.ID = uuid.New()
		}
		if event.EvaluatedAt.IsZero() {
			event.EvaluatedAt = time.Now()
		}

		valueJSON, _ := json.Marshal(event.Value)
		contextJSON, _ := json.Marshal(event.Context)
		if contextJSON == nil {
			contextJSON = []byte("{}")
		}

		_, err = stmt.ExecContext(ctx, event.ID, event.TenantID, event.EnvironmentID, event.FeatureFlagID, event.FlagKey, valueJSON, event.UserID, contextJSON, event.SDKType, event.SDKVersion, event.EvaluatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *evaluationEventRepo) GetByTenant(ctx context.Context, tenantID uuid.UUID, filters repository.EvaluationFilters) ([]*entity.EvaluationEvent, int, error) {
	var args []interface{}
	args = append(args, tenantID)
	argCount := 1

	whereClause := "WHERE tenant_id = $1"

	if filters.EnvironmentID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND environment_id = $%d", argCount)
		args = append(args, *filters.EnvironmentID)
	}
	if filters.FlagID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND feature_flag_id = $%d", argCount)
		args = append(args, *filters.FlagID)
	}
	if filters.FlagKey != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND flag_key = $%d", argCount)
		args = append(args, *filters.FlagKey)
	}
	if filters.UserID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filters.UserID)
	}
	if filters.StartDate != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND evaluated_at >= $%d", argCount)
		args = append(args, *filters.StartDate)
	}
	if filters.EndDate != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND evaluated_at <= $%d", argCount)
		args = append(args, *filters.EndDate)
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM evaluation_events " + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get events
	limit := filters.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, tenant_id, environment_id, feature_flag_id, flag_key, value, user_id, context, sdk_type, sdk_version, evaluated_at
		FROM evaluation_events
		%s
		ORDER BY evaluated_at DESC
		LIMIT %d OFFSET %d
	`, whereClause, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []*entity.EvaluationEvent
	for rows.Next() {
		event := &entity.EvaluationEvent{}
		var valueJSON, contextJSON []byte
		err := rows.Scan(
			&event.ID,
			&event.TenantID,
			&event.EnvironmentID,
			&event.FeatureFlagID,
			&event.FlagKey,
			&valueJSON,
			&event.UserID,
			&contextJSON,
			&event.SDKType,
			&event.SDKVersion,
			&event.EvaluatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		json.Unmarshal(valueJSON, &event.Value)
		json.Unmarshal(contextJSON, &event.Context)
		events = append(events, event)
	}

	return events, total, nil
}

func (r *evaluationEventRepo) GetByFlag(ctx context.Context, flagID uuid.UUID, filters repository.EvaluationFilters) ([]*entity.EvaluationEvent, int, error) {
	filters.FlagID = &flagID
	// Get tenant from flag first - for now assume we always filter by tenant in GetByTenant
	// This is a simplified version
	var tenantID uuid.UUID
	err := r.db.QueryRowContext(ctx, `
		SELECT e.application_id FROM environments e
		JOIN feature_flags f ON f.environment_id = e.id
		WHERE f.id = $1
	`, flagID).Scan(&tenantID)
	if err != nil {
		return nil, 0, err
	}

	return r.GetByTenant(ctx, tenantID, filters)
}

func (r *evaluationEventRepo) GetSummary(ctx context.Context, tenantID uuid.UUID, filters repository.MetricsFilters) (*entity.UsageMetrics, error) {
	metrics := &entity.UsageMetrics{
		TenantID:  tenantID,
		StartDate: filters.StartDate,
		EndDate:   filters.EndDate,
		Period:    filters.Granularity,
	}

	var args []interface{}
	args = append(args, tenantID, filters.StartDate, filters.EndDate)
	argCount := 3

	whereClause := "WHERE tenant_id = $1 AND evaluated_at >= $2 AND evaluated_at <= $3"

	if filters.EnvironmentID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND environment_id = $%d", argCount)
		args = append(args, *filters.EnvironmentID)
	}
	if filters.FlagID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND feature_flag_id = $%d", argCount)
		args = append(args, *filters.FlagID)
	}

	// Total evaluations
	query := `SELECT COUNT(*) FROM evaluation_events ` + whereClause
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&metrics.TotalEvaluations)
	if err != nil {
		return nil, err
	}

	// Unique flags
	query = `SELECT COUNT(DISTINCT feature_flag_id) FROM evaluation_events ` + whereClause
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&metrics.UniqueFlags)
	if err != nil {
		return nil, err
	}

	// Unique users
	query = `SELECT COUNT(DISTINCT user_id) FROM evaluation_events ` + whereClause + " AND user_id IS NOT NULL"
	err = r.db.QueryRowContext(ctx, query, args...).Scan(&metrics.UniqueUsers)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (r *evaluationEventRepo) GetSummaryByEnvironment(ctx context.Context, tenantID uuid.UUID, filters repository.MetricsFilters) ([]entity.EnvironmentMetrics, error) {
	query := `
		SELECT 
			ee.environment_id,
			COALESCE(e.name, 'Unknown') as environment_name,
			COUNT(*) as evaluations,
			COUNT(DISTINCT ee.feature_flag_id) as unique_flags,
			COUNT(DISTINCT ee.user_id) as unique_users
		FROM evaluation_events ee
		LEFT JOIN environments e ON e.id = ee.environment_id
		WHERE ee.tenant_id = $1 AND ee.evaluated_at >= $2 AND ee.evaluated_at <= $3
		GROUP BY ee.environment_id, e.name
		ORDER BY evaluations DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, filters.StartDate, filters.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []entity.EnvironmentMetrics
	for rows.Next() {
		m := entity.EnvironmentMetrics{}
		err := rows.Scan(&m.EnvironmentID, &m.EnvironmentName, &m.Evaluations, &m.UniqueFlags, &m.UniqueUsers)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (r *evaluationEventRepo) GetSummaryByFlag(ctx context.Context, tenantID uuid.UUID, filters repository.MetricsFilters) ([]entity.FlagMetrics, error) {
	var args []interface{}
	args = append(args, tenantID, filters.StartDate, filters.EndDate)
	argCount := 3

	whereClause := "WHERE ee.tenant_id = $1 AND ee.evaluated_at >= $2 AND ee.evaluated_at <= $3"

	if filters.EnvironmentID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND ee.environment_id = $%d", argCount)
		args = append(args, *filters.EnvironmentID)
	}

	query := fmt.Sprintf(`
		SELECT 
			ee.feature_flag_id,
			ee.flag_key,
			COALESCE(f.name, ee.flag_key) as flag_name,
			ee.environment_id,
			COALESCE(e.name, 'Unknown') as environment_name,
			COUNT(*) as evaluations,
			SUM(CASE WHEN ee.value::text = 'true' THEN 1 ELSE 0 END) as true_count,
			SUM(CASE WHEN ee.value::text = 'false' THEN 1 ELSE 0 END) as false_count,
			COUNT(DISTINCT ee.user_id) as unique_users
		FROM evaluation_events ee
		LEFT JOIN feature_flags f ON f.id = ee.feature_flag_id
		LEFT JOIN environments e ON e.id = ee.environment_id
		%s
		GROUP BY ee.feature_flag_id, ee.flag_key, f.name, ee.environment_id, e.name
		ORDER BY evaluations DESC
		LIMIT 100
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []entity.FlagMetrics
	for rows.Next() {
		m := entity.FlagMetrics{}
		err := rows.Scan(&m.FlagID, &m.FlagKey, &m.FlagName, &m.EnvironmentID, &m.EnvironmentName, &m.Evaluations, &m.TrueCount, &m.FalseCount, &m.UniqueUsers)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

func (r *evaluationEventRepo) GetTimeline(ctx context.Context, tenantID uuid.UUID, filters repository.MetricsFilters) ([]entity.TimelinePoint, error) {
	var interval string
	switch filters.Granularity {
	case "hour":
		interval = "hour"
	case "day":
		interval = "day"
	case "week":
		interval = "week"
	case "month":
		interval = "month"
	default:
		interval = "hour"
	}

	var args []interface{}
	args = append(args, tenantID, filters.StartDate, filters.EndDate)
	argCount := 3

	whereClause := "WHERE tenant_id = $1 AND evaluated_at >= $2 AND evaluated_at <= $3"

	if filters.EnvironmentID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND environment_id = $%d", argCount)
		args = append(args, *filters.EnvironmentID)
	}
	if filters.FlagID != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND feature_flag_id = $%d", argCount)
		args = append(args, *filters.FlagID)
	}

	query := fmt.Sprintf(`
		SELECT 
			date_trunc('%s', evaluated_at) as timestamp,
			COUNT(*) as evaluations,
			SUM(CASE WHEN value::text = 'true' THEN 1 ELSE 0 END) as true_count,
			SUM(CASE WHEN value::text = 'false' THEN 1 ELSE 0 END) as false_count
		FROM evaluation_events
		%s
		GROUP BY date_trunc('%s', evaluated_at)
		ORDER BY timestamp ASC
	`, interval, whereClause, interval)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timeline []entity.TimelinePoint
	for rows.Next() {
		point := entity.TimelinePoint{}
		err := rows.Scan(&point.Timestamp, &point.Evaluations, &point.TrueCount, &point.FalseCount)
		if err != nil {
			return nil, err
		}
		timeline = append(timeline, point)
	}

	return timeline, nil
}

func (r *evaluationEventRepo) DeleteOlderThan(ctx context.Context, tenantID uuid.UUID, before time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM evaluation_events
		WHERE tenant_id = $1 AND evaluated_at < $2
	`, tenantID, before)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
