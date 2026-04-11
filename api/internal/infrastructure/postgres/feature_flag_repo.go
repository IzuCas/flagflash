package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// FeatureFlagRepository implements repository.FeatureFlagRepository
type FeatureFlagRepository struct {
	db *sql.DB
}

// NewFeatureFlagRepository creates a new FeatureFlagRepository
func NewFeatureFlagRepository(db *sql.DB) *FeatureFlagRepository {
	return &FeatureFlagRepository{db: db}
}

// Create creates a new feature flag
func (r *FeatureFlagRepository) Create(ctx context.Context, flag *entity.FeatureFlag) error {
	valueJSON, _ := json.Marshal(flag.Value)
	defaultJSON, _ := json.Marshal(flag.DefaultValue)

	query := `INSERT INTO feature_flags (id, environment_id, key, name, description, type, value, default_value, enabled, tags, version, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.ExecContext(ctx, query,
		flag.ID, flag.EnvironmentID, flag.Key, flag.Name, flag.Description,
		flag.Type, valueJSON, defaultJSON, flag.Enabled, pq.Array(flag.Tags),
		flag.Version, flag.CreatedAt, flag.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create feature flag: %w", err)
	}
	return nil
}

// GetByID retrieves a feature flag by ID
func (r *FeatureFlagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.FeatureFlag, error) {
	query := `SELECT id, environment_id, key, name, description, type, value, default_value, enabled, tags, version, created_at, updated_at
			  FROM feature_flags WHERE id = $1`
	return r.scanFlag(r.db.QueryRowContext(ctx, query, id))
}

// GetByKey retrieves a feature flag by key
func (r *FeatureFlagRepository) GetByKey(ctx context.Context, environmentID uuid.UUID, key string) (*entity.FeatureFlag, error) {
	query := `SELECT id, environment_id, key, name, description, type, value, default_value, enabled, tags, version, created_at, updated_at
			  FROM feature_flags WHERE environment_id = $1 AND key = $2`
	return r.scanFlag(r.db.QueryRowContext(ctx, query, environmentID, key))
}

// GetByEnvironmentID retrieves all feature flags for an environment
func (r *FeatureFlagRepository) GetByEnvironmentID(ctx context.Context, environmentID uuid.UUID) ([]*entity.FeatureFlag, error) {
	query := `SELECT id, environment_id, key, name, description, type, value, default_value, enabled, tags, version, created_at, updated_at
			  FROM feature_flags WHERE environment_id = $1 ORDER BY key`
	rows, err := r.db.QueryContext(ctx, query, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature flags: %w", err)
	}
	defer rows.Close()

	return r.scanFlags(rows)
}

// GetByEnvironmentIDPaginated retrieves feature flags with pagination
func (r *FeatureFlagRepository) GetByEnvironmentIDPaginated(ctx context.Context, environmentID uuid.UUID, offset, limit int) ([]*entity.FeatureFlag, int, error) {
	countQuery := `SELECT COUNT(*) FROM feature_flags WHERE environment_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, environmentID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count flags: %w", err)
	}

	query := `SELECT id, environment_id, key, name, description, type, value, default_value, enabled, tags, version, created_at, updated_at
			  FROM feature_flags WHERE environment_id = $1 ORDER BY key LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, environmentID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get feature flags: %w", err)
	}
	defer rows.Close()

	flags, err := r.scanFlags(rows)
	return flags, total, err
}

// Update updates a feature flag
func (r *FeatureFlagRepository) Update(ctx context.Context, flag *entity.FeatureFlag) error {
	valueJSON, _ := json.Marshal(flag.Value)
	defaultJSON, _ := json.Marshal(flag.DefaultValue)

	query := `UPDATE feature_flags SET name = $1, description = $2, value = $3, default_value = $4, enabled = $5, tags = $6, version = $7, updated_at = $8 WHERE id = $9`
	_, err := r.db.ExecContext(ctx, query,
		flag.Name, flag.Description, valueJSON, defaultJSON,
		flag.Enabled, pq.Array(flag.Tags), flag.Version, flag.UpdatedAt, flag.ID)
	if err != nil {
		return fmt.Errorf("failed to update feature flag: %w", err)
	}
	return nil
}

// Delete deletes a feature flag
func (r *FeatureFlagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM feature_flags WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete feature flag: %w", err)
	}
	return nil
}

// IncrementVersion increments the flag version
func (r *FeatureFlagRepository) IncrementVersion(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE feature_flags SET version = version + 1, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment version: %w", err)
	}
	return nil
}

// ListByEnvironment lists flags by environment with optional inclusion of deleted flags
func (r *FeatureFlagRepository) ListByEnvironment(ctx context.Context, environmentID uuid.UUID, includeDeleted bool) ([]*entity.FeatureFlag, error) {
	query := `SELECT id, environment_id, key, name, description, type, value, default_value, enabled, tags, version, created_at, updated_at
			  FROM feature_flags WHERE environment_id = $1`
	if !includeDeleted {
		query += ` AND deleted_at IS NULL`
	}
	query += ` ORDER BY key`

	rows, err := r.db.QueryContext(ctx, query, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list feature flags: %w", err)
	}
	defer rows.Close()

	return r.scanFlags(rows)
}

// ListByEnvironmentWithPagination lists flags with pagination and search
func (r *FeatureFlagRepository) ListByEnvironmentWithPagination(ctx context.Context, environmentID uuid.UUID, limit, offset int, search string) ([]*entity.FeatureFlag, int, error) {
	// Enforce safe bounds to prevent DoS via huge result sets
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Count query
	countQuery := `SELECT COUNT(*) FROM feature_flags WHERE environment_id = $1 AND deleted_at IS NULL`
	args := []interface{}{environmentID}

	if search != "" {
		countQuery += ` AND (key ILIKE $2 OR name ILIKE $2)`
		args = append(args, "%"+search+"%")
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count flags: %w", err)
	}

	// Data query — always use parameterized placeholders for LIMIT/OFFSET
	query := `SELECT id, environment_id, key, name, description, type, value, default_value, enabled, tags, version, created_at, updated_at
			  FROM feature_flags WHERE environment_id = $1 AND deleted_at IS NULL`

	if search != "" {
		query += ` AND (key ILIKE $2 OR name ILIKE $2)`
		query += ` ORDER BY key LIMIT $3 OFFSET $4`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY key LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list feature flags: %w", err)
	}
	defer rows.Close()

	flags, err := r.scanFlags(rows)
	return flags, total, err
}

// GetFlagWithTenant retrieves a flag along with its tenant ID (via environment -> application -> tenant)
func (r *FeatureFlagRepository) GetFlagWithTenant(ctx context.Context, id uuid.UUID) (*entity.FeatureFlag, uuid.UUID, error) {
	query := `
		SELECT 
			f.id, f.environment_id, f.key, f.name, f.description, f.type, f.value, f.default_value, f.enabled, f.tags, f.version, f.created_at, f.updated_at,
			a.tenant_id
		FROM feature_flags f
		JOIN environments e ON f.environment_id = e.id
		JOIN applications a ON e.application_id = a.id
		WHERE f.id = $1
	`

	flag := &entity.FeatureFlag{}
	var valueJSON, defaultJSON []byte
	var tenantID uuid.UUID

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&flag.ID, &flag.EnvironmentID, &flag.Key, &flag.Name, &flag.Description,
		&flag.Type, &valueJSON, &defaultJSON, &flag.Enabled, pq.Array(&flag.Tags),
		&flag.Version, &flag.CreatedAt, &flag.UpdatedAt, &tenantID)
	if err == sql.ErrNoRows {
		return nil, uuid.Nil, fmt.Errorf("feature flag not found")
	}
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("failed to get flag with tenant: %w", err)
	}

	json.Unmarshal(valueJSON, &flag.Value)
	json.Unmarshal(defaultJSON, &flag.DefaultValue)

	return flag, tenantID, nil
}

func (r *FeatureFlagRepository) scanFlag(row *sql.Row) (*entity.FeatureFlag, error) {
	flag := &entity.FeatureFlag{}
	var valueJSON, defaultJSON []byte
	err := row.Scan(
		&flag.ID, &flag.EnvironmentID, &flag.Key, &flag.Name, &flag.Description,
		&flag.Type, &valueJSON, &defaultJSON, &flag.Enabled, pq.Array(&flag.Tags),
		&flag.Version, &flag.CreatedAt, &flag.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan flag: %w", err)
	}

	json.Unmarshal(valueJSON, &flag.Value)
	json.Unmarshal(defaultJSON, &flag.DefaultValue)

	return flag, nil
}

func (r *FeatureFlagRepository) scanFlags(rows *sql.Rows) ([]*entity.FeatureFlag, error) {
	var flags []*entity.FeatureFlag
	for rows.Next() {
		flag := &entity.FeatureFlag{}
		var valueJSON, defaultJSON []byte
		if err := rows.Scan(
			&flag.ID, &flag.EnvironmentID, &flag.Key, &flag.Name, &flag.Description,
			&flag.Type, &valueJSON, &defaultJSON, &flag.Enabled, pq.Array(&flag.Tags),
			&flag.Version, &flag.CreatedAt, &flag.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan flag: %w", err)
		}
		json.Unmarshal(valueJSON, &flag.Value)
		json.Unmarshal(defaultJSON, &flag.DefaultValue)
		flags = append(flags, flag)
	}
	return flags, nil
}

// CopyFlags copies all flags from source environment to target environment
func (r *FeatureFlagRepository) CopyFlags(ctx context.Context, sourceEnvID, targetEnvID uuid.UUID) error {
	// Get all flags from source environment
	sourceFlags, err := r.GetByEnvironmentID(ctx, sourceEnvID)
	if err != nil {
		return fmt.Errorf("failed to get source flags: %w", err)
	}

	// Copy each flag to the target environment with new IDs
	for _, srcFlag := range sourceFlags {
		newFlag := &entity.FeatureFlag{
			ID:            uuid.New(),
			EnvironmentID: targetEnvID,
			Key:           srcFlag.Key,
			Name:          srcFlag.Name,
			Description:   srcFlag.Description,
			Type:          srcFlag.Type,
			Enabled:       false, // Copy as disabled by default
			Value:         srcFlag.Value,
			DefaultValue:  srcFlag.DefaultValue,
			Version:       1,
			Tags:          srcFlag.Tags,
			CreatedAt:     srcFlag.CreatedAt,
			UpdatedAt:     srcFlag.UpdatedAt,
		}
		newFlag.CreatedAt = time.Now()
		newFlag.UpdatedAt = time.Now()

		if err := r.Create(ctx, newFlag); err != nil {
			return fmt.Errorf("failed to copy flag %s: %w", srcFlag.Key, err)
		}
	}

	return nil
}
