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

// AuditLogRepo implements repository.AuditLogRepository
type AuditLogRepo struct {
	db *DB
}

// NewAuditLogRepo creates a new audit log repository
func NewAuditLogRepo(db *DB) repository.AuditLogRepository {
	return &AuditLogRepo{db: db}
}

// Create creates a new audit log entry
func (r *AuditLogRepo) Create(ctx context.Context, log *entity.AuditLog) error {
	metadata, err := json.Marshal(log.Metadata)
	if err != nil {
		metadata = []byte("{}")
	}

	query := `
		INSERT INTO audit_logs (id, tenant_id, entity_type, entity_id, action, actor_id, actor_type, old_value, new_value, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		log.ID, log.TenantID, log.EntityType, log.EntityID, log.Action,
		log.ActorID, log.ActorType, log.OldValue, log.NewValue, metadata, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// CreateBatch creates multiple audit log entries
func (r *AuditLogRepo) CreateBatch(ctx context.Context, logs []*entity.AuditLog) error {
	for _, log := range logs {
		if err := r.Create(ctx, log); err != nil {
			return err
		}
	}
	return nil
}

// GetByID retrieves an audit log by ID
func (r *AuditLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	query := `
		SELECT id, tenant_id, entity_type, entity_id, action, actor_id, actor_type, old_value, new_value, metadata, created_at
		FROM audit_logs
		WHERE id = $1
	`

	var log entity.AuditLog
	var oldValue, newValue, metadata []byte
	var actorID, actorType sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.TenantID, &log.EntityType, &log.EntityID, &log.Action,
		&actorID, &actorType, &oldValue, &newValue, &metadata, &log.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	if actorID.Valid {
		log.ActorID = actorID.String
	}
	if actorType.Valid {
		log.ActorType = entity.ActorType(actorType.String)
	}

	log.OldValue = json.RawMessage(oldValue)
	log.NewValue = json.RawMessage(newValue)
	json.Unmarshal(metadata, &log.Metadata)

	return &log, nil
}

// List queries audit logs with filters
func (r *AuditLogRepo) List(ctx context.Context, query *entity.AuditLogQuery) ([]*entity.AuditLog, int, error) {
	baseQuery := `
		SELECT id, tenant_id, entity_type, entity_id, action, actor_id, actor_type, old_value, new_value, metadata, created_at
		FROM audit_logs
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if query.TenantID != nil {
		baseQuery += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, *query.TenantID)
		argIndex++
	}

	if query.EntityType != nil {
		baseQuery += fmt.Sprintf(" AND entity_type = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND entity_type = $%d", argIndex)
		args = append(args, *query.EntityType)
		argIndex++
	}

	if query.EntityID != nil {
		baseQuery += fmt.Sprintf(" AND entity_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND entity_id = $%d", argIndex)
		args = append(args, *query.EntityID)
		argIndex++
	}

	if query.Action != nil {
		baseQuery += fmt.Sprintf(" AND action = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, *query.Action)
		argIndex++
	}

	if query.StartDate != nil {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *query.StartDate)
		argIndex++
	}

	if query.EndDate != nil {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *query.EndDate)
		argIndex++
	}

	// Get total count
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	baseQuery += " ORDER BY created_at DESC"

	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, query.Offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	return r.scanLogs(rows, total)
}

// GetByEntity retrieves audit logs for a specific entity
func (r *AuditLogRepo) GetByEntity(ctx context.Context, entityType entity.EntityType, entityID uuid.UUID, limit int) ([]*entity.AuditLog, error) {
	query := `
		SELECT id, tenant_id, entity_type, entity_id, action, actor_id, actor_type, old_value, new_value, metadata, created_at
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, entityType, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	logs, _, err := r.scanLogs(rows, 0)
	return logs, err
}

// DeleteOlderThan deletes audit logs older than a specific time
func (r *AuditLogRepo) DeleteOlderThan(ctx context.Context, tenantID uuid.UUID, days int) (int64, error) {
	query := `
		DELETE FROM audit_logs
		WHERE tenant_id = $1 AND created_at < $2
	`

	cutoff := time.Now().AddDate(0, 0, -days)
	result, err := r.db.ExecContext(ctx, query, tenantID, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}

	return result.RowsAffected()
}

// scanLogs scans audit log rows
func (r *AuditLogRepo) scanLogs(rows *sql.Rows, total int) ([]*entity.AuditLog, int, error) {
	var logs []*entity.AuditLog
	for rows.Next() {
		var log entity.AuditLog
		var oldValue, newValue, metadata []byte
		var actorID, actorType sql.NullString

		if err := rows.Scan(
			&log.ID, &log.TenantID, &log.EntityType, &log.EntityID, &log.Action,
			&actorID, &actorType, &oldValue, &newValue, &metadata, &log.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if actorID.Valid {
			log.ActorID = actorID.String
		}
		if actorType.Valid {
			log.ActorType = entity.ActorType(actorType.String)
		}

		log.OldValue = json.RawMessage(oldValue)
		log.NewValue = json.RawMessage(newValue)
		json.Unmarshal(metadata, &log.Metadata)

		logs = append(logs, &log)
	}

	return logs, total, nil
}
