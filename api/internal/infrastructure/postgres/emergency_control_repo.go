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

// EmergencyControlRepo implements repository.EmergencyControlRepository
type EmergencyControlRepo struct {
	db *DB
}

// NewEmergencyControlRepo creates a new emergency control repository
func NewEmergencyControlRepo(db *DB) repository.EmergencyControlRepository {
	return &EmergencyControlRepo{db: db}
}

// Create creates a new emergency control
func (r *EmergencyControlRepo) Create(ctx context.Context, control *entity.EmergencyControl) error {
	query := `
		INSERT INTO emergency_controls (id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		control.ID, control.TenantID, control.EnvironmentID, control.ControlType,
		control.Enabled, control.Reason, control.EnabledBy, control.EnabledAt,
		control.ExpiresAt, control.CreatedAt, control.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create emergency control: %w", err)
	}

	return nil
}

// Update updates an emergency control
func (r *EmergencyControlRepo) Update(ctx context.Context, control *entity.EmergencyControl) error {
	control.UpdatedAt = time.Now()

	query := `
		UPDATE emergency_controls
		SET enabled = $1, reason = $2, enabled_by = $3, enabled_at = $4, expires_at = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		control.Enabled, control.Reason, control.EnabledBy, control.EnabledAt,
		control.ExpiresAt, control.UpdatedAt, control.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update emergency control: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("emergency control not found")
	}

	return nil
}

// Upsert creates or updates an emergency control
func (r *EmergencyControlRepo) Upsert(ctx context.Context, control *entity.EmergencyControl) error {
	control.UpdatedAt = time.Now()

	query := `
		INSERT INTO emergency_controls (id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (tenant_id, environment_id, control_type) WHERE environment_id IS NOT NULL
		DO UPDATE SET enabled = $5, reason = $6, enabled_by = $7, enabled_at = $8, expires_at = $9, updated_at = $11
	`

	_, err := r.db.ExecContext(ctx, query,
		control.ID, control.TenantID, control.EnvironmentID, control.ControlType,
		control.Enabled, control.Reason, control.EnabledBy, control.EnabledAt,
		control.ExpiresAt, control.CreatedAt, control.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert emergency control: %w", err)
	}

	return nil
}

// Delete deletes an emergency control
func (r *EmergencyControlRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM emergency_controls WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete emergency control: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("emergency control not found")
	}

	return nil
}

// GetByID retrieves an emergency control by ID
func (r *EmergencyControlRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.EmergencyControl, error) {
	query := `
		SELECT id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at
		FROM emergency_controls
		WHERE id = $1
	`

	return r.scanEmergencyControl(r.db.QueryRowContext(ctx, query, id))
}

// GetByType gets emergency control by type for a tenant/environment
func (r *EmergencyControlRepo) GetByType(ctx context.Context, tenantID uuid.UUID, envID *uuid.UUID, controlType entity.EmergencyControlType) (*entity.EmergencyControl, error) {
	var query string
	var row *sql.Row

	if envID == nil {
		query = `
			SELECT id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at
			FROM emergency_controls
			WHERE tenant_id = $1 AND environment_id IS NULL AND control_type = $2
			LIMIT 1
		`
		row = r.db.QueryRowContext(ctx, query, tenantID, controlType)
	} else {
		query = `
			SELECT id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at
			FROM emergency_controls
			WHERE tenant_id = $1 AND environment_id = $2 AND control_type = $3
			LIMIT 1
		`
		row = r.db.QueryRowContext(ctx, query, tenantID, *envID, controlType)
	}

	return r.scanEmergencyControl(row)
}

// GetActiveKillSwitch gets active kill switch for a tenant/environment
func (r *EmergencyControlRepo) GetActiveKillSwitch(ctx context.Context, tenantID uuid.UUID, envID *uuid.UUID) (*entity.EmergencyControl, error) {
	var query string
	var row *sql.Row

	if envID == nil {
		query = `
			SELECT id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at
			FROM emergency_controls
			WHERE tenant_id = $1 AND environment_id IS NULL AND control_type = 'kill_switch' AND enabled = true
			LIMIT 1
		`
		row = r.db.QueryRowContext(ctx, query, tenantID)
	} else {
		query = `
			SELECT id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at
			FROM emergency_controls
			WHERE tenant_id = $1 AND environment_id = $2 AND control_type = 'kill_switch' AND enabled = true
			LIMIT 1
		`
		row = r.db.QueryRowContext(ctx, query, tenantID, *envID)
	}

	return r.scanEmergencyControl(row)
}

// ListByTenant lists all emergency controls for a tenant
func (r *EmergencyControlRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.EmergencyControl, error) {
	query := `
		SELECT id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at
		FROM emergency_controls
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list emergency controls: %w", err)
	}
	defer rows.Close()

	return r.scanEmergencyControls(rows)
}

// ListEnabled lists enabled emergency controls for a tenant
func (r *EmergencyControlRepo) ListEnabled(ctx context.Context, tenantID uuid.UUID) ([]*entity.EmergencyControl, error) {
	query := `
		SELECT id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at
		FROM emergency_controls
		WHERE tenant_id = $1 AND enabled = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled emergency controls: %w", err)
	}
	defer rows.Close()

	return r.scanEmergencyControls(rows)
}

// ListExpired lists expired emergency controls that are still enabled
func (r *EmergencyControlRepo) ListExpired(ctx context.Context) ([]*entity.EmergencyControl, error) {
	query := `
		SELECT id, tenant_id, environment_id, control_type, enabled, reason, enabled_by, enabled_at, expires_at, created_at, updated_at
		FROM emergency_controls
		WHERE enabled = true AND expires_at IS NOT NULL AND expires_at <= $1
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to list expired emergency controls: %w", err)
	}
	defer rows.Close()

	return r.scanEmergencyControls(rows)
}

// DisableExpired disables all expired emergency controls
func (r *EmergencyControlRepo) DisableExpired(ctx context.Context) (int, error) {
	query := `
		UPDATE emergency_controls
		SET enabled = false, updated_at = $1
		WHERE enabled = true AND expires_at IS NOT NULL AND expires_at <= $1
	`

	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to disable expired controls: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

func (r *EmergencyControlRepo) scanEmergencyControl(row *sql.Row) (*entity.EmergencyControl, error) {
	var control entity.EmergencyControl
	var envID sql.NullString
	var enabledBy sql.NullString
	var enabledAt, expiresAt sql.NullTime

	err := row.Scan(
		&control.ID, &control.TenantID, &envID, &control.ControlType,
		&control.Enabled, &control.Reason, &enabledBy, &enabledAt,
		&expiresAt, &control.CreatedAt, &control.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("emergency control not found")
		}
		return nil, fmt.Errorf("failed to scan emergency control: %w", err)
	}

	if envID.Valid {
		id, _ := uuid.Parse(envID.String)
		control.EnvironmentID = &id
	}
	if enabledBy.Valid {
		id, _ := uuid.Parse(enabledBy.String)
		control.EnabledBy = &id
	}
	if enabledAt.Valid {
		control.EnabledAt = &enabledAt.Time
	}
	if expiresAt.Valid {
		control.ExpiresAt = &expiresAt.Time
	}

	return &control, nil
}

func (r *EmergencyControlRepo) scanEmergencyControls(rows *sql.Rows) ([]*entity.EmergencyControl, error) {
	controls := make([]*entity.EmergencyControl, 0)

	for rows.Next() {
		var control entity.EmergencyControl
		var envID sql.NullString
		var enabledBy sql.NullString
		var enabledAt, expiresAt sql.NullTime

		err := rows.Scan(
			&control.ID, &control.TenantID, &envID, &control.ControlType,
			&control.Enabled, &control.Reason, &enabledBy, &enabledAt,
			&expiresAt, &control.CreatedAt, &control.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan emergency control: %w", err)
		}

		if envID.Valid {
			id, _ := uuid.Parse(envID.String)
			control.EnvironmentID = &id
		}
		if enabledBy.Valid {
			id, _ := uuid.Parse(enabledBy.String)
			control.EnabledBy = &id
		}
		if enabledAt.Valid {
			control.EnabledAt = &enabledAt.Time
}
		if expiresAt.Valid {
			control.ExpiresAt = &expiresAt.Time
		}
		controls = append(controls, &control)
	}

	return controls, nil
}
