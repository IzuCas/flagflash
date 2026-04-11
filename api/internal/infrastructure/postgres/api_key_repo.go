package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// APIKeyRepo implements repository.APIKeyRepository
type APIKeyRepo struct {
	db *DB
}

// NewAPIKeyRepo creates a new APIKeyRepo
func NewAPIKeyRepo(db *DB) repository.APIKeyRepository {
	return &APIKeyRepo{db: db}
}

// Create creates a new API key
func (r *APIKeyRepo) Create(ctx context.Context, key *entity.APIKey) error {
	query := `INSERT INTO api_keys (id, tenant_id, environment_id, name, key_prefix, key_hash, permissions, active, expires_at, last_used_at, created_at, revoked_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query,
		key.ID, key.TenantID, key.EnvironmentID, key.Name, key.KeyPrefix, key.KeyHash,
		pq.Array(key.Permissions), key.Active, key.ExpiresAt, key.LastUsedAt, key.CreatedAt, key.RevokedAt)
	return err
}

// GetByID retrieves an API key by ID
func (r *APIKeyRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.APIKey, error) {
	query := `SELECT id, tenant_id, environment_id, name, key_prefix, key_hash, permissions, active, 
			  expires_at, last_used_at, created_at, revoked_at 
			  FROM api_keys WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanAPIKey(row)
}

// GetByHash retrieves an API key by its hash
func (r *APIKeyRepo) GetByHash(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	query := `SELECT id, tenant_id, environment_id, name, key_prefix, key_hash, permissions, active, 
			  expires_at, last_used_at, created_at, revoked_at 
			  FROM api_keys WHERE key_hash = $1 AND revoked_at IS NULL AND active = true`
	row := r.db.QueryRowContext(ctx, query, keyHash)
	return r.scanAPIKey(row)
}

// GetByKeyHash retrieves an API key by prefix and hash
func (r *APIKeyRepo) GetByKeyHash(ctx context.Context, keyPrefix, keyHash string) (*entity.APIKey, error) {
	query := `SELECT id, tenant_id, environment_id, name, key_prefix, key_hash, permissions, active, 
			  expires_at, last_used_at, created_at, revoked_at 
			  FROM api_keys WHERE key_prefix = $1 AND key_hash = $2 AND revoked_at IS NULL AND active = true`
	row := r.db.QueryRowContext(ctx, query, keyPrefix, keyHash)
	return r.scanAPIKey(row)
}

// GetByHashWithDetails retrieves an API key with environment and tenant details
func (r *APIKeyRepo) GetByHashWithDetails(ctx context.Context, keyHash string) (*entity.APIKey, *entity.Environment, uuid.UUID, error) {
	query := `SELECT k.id, k.tenant_id, k.environment_id, k.name, k.key_prefix, k.key_hash, k.permissions, k.active, 
			  k.expires_at, k.last_used_at, k.created_at, k.revoked_at,
			  e.id, e.application_id, e.name, e.slug, e.color, e.is_production, e.created_at, e.updated_at
			  FROM api_keys k
			  LEFT JOIN environments e ON k.environment_id = e.id
			  WHERE k.key_hash = $1 AND k.revoked_at IS NULL AND k.active = true`
	row := r.db.QueryRowContext(ctx, query, keyHash)

	var key entity.APIKey
	var env entity.Environment
	var expiresAt, lastUsedAt, revokedAt sql.NullTime
	var environmentID sql.NullString
	var envAppID sql.NullString
	var envIDResult sql.NullString
	var envName, envSlug, envColor sql.NullString
	var envIsProduction sql.NullBool
	var envCreatedAt, envUpdatedAt sql.NullTime

	err := row.Scan(
		&key.ID, &key.TenantID, &environmentID, &key.Name, &key.KeyPrefix, &key.KeyHash,
		pq.Array(&key.Permissions), &key.Active, &expiresAt, &lastUsedAt, &key.CreatedAt, &revokedAt,
		&envIDResult, &envAppID, &envName, &envSlug, &envColor, &envIsProduction, &envCreatedAt, &envUpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, uuid.Nil, fmt.Errorf("API key not found")
		}
		return nil, nil, uuid.Nil, err
	}

	if environmentID.Valid {
		envUUID, _ := uuid.Parse(environmentID.String)
		key.EnvironmentID = &envUUID
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}

	// Build environment if it exists
	var envPtr *entity.Environment
	if envIDResult.Valid {
		env.ID, _ = uuid.Parse(envIDResult.String)
		env.ApplicationID, _ = uuid.Parse(envAppID.String)
		env.Name = envName.String
		env.Slug = envSlug.String
		env.Color = envColor.String
		env.IsProduction = envIsProduction.Bool
		env.CreatedAt = envCreatedAt.Time
		env.UpdatedAt = envUpdatedAt.Time
		envPtr = &env
	}

	return &key, envPtr, key.TenantID, nil
}

// GetByTenantID retrieves all API keys for a tenant
func (r *APIKeyRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.APIKey, error) {
	query := `SELECT id, tenant_id, environment_id, name, key_prefix, key_hash, permissions, active, 
			  expires_at, last_used_at, created_at, revoked_at 
			  FROM api_keys WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*entity.APIKey
	for rows.Next() {
		key, err := r.scanAPIKeyRows(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// Update updates an API key
func (r *APIKeyRepo) Update(ctx context.Context, key *entity.APIKey) error {
	query := `UPDATE api_keys SET name = $1, permissions = $2, active = $3, expires_at = $4, revoked_at = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, key.Name, pq.Array(key.Permissions), key.Active, key.ExpiresAt, key.RevokedAt, key.ID)
	return err
}

// Revoke revokes an API key
func (r *APIKeyRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET revoked_at = $1, active = false WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// Delete permanently deletes an API key
func (r *APIKeyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListByEnvironment lists all API keys for an environment
func (r *APIKeyRepo) ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.APIKeyInfo, error) {
	query := `SELECT id, name, key_prefix, permissions, expires_at, last_used_at, created_at, revoked_at 
			  FROM api_keys WHERE environment_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, environmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*entity.APIKeyInfo
	for rows.Next() {
		var key entity.APIKeyInfo
		var expiresAt, lastUsedAt, revokedAt sql.NullTime
		err := rows.Scan(&key.ID, &key.Name, &key.KeyPrefix, pq.Array(&key.Permissions),
			&expiresAt, &lastUsedAt, &key.CreatedAt, &revokedAt)
		if err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			key.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			key.LastUsedAt = &lastUsedAt.Time
		}
		key.IsRevoked = revokedAt.Valid
		keys = append(keys, &key)
	}
	return keys, rows.Err()
}

// UpdateLastUsed updates the last used timestamp
func (r *APIKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *APIKeyRepo) scanAPIKey(row *sql.Row) (*entity.APIKey, error) {
	var key entity.APIKey
	var environmentID sql.NullString
	var expiresAt, lastUsedAt, revokedAt sql.NullTime
	err := row.Scan(&key.ID, &key.TenantID, &environmentID, &key.Name, &key.KeyPrefix, &key.KeyHash,
		pq.Array(&key.Permissions), &key.Active, &expiresAt, &lastUsedAt, &key.CreatedAt, &revokedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, err
	}
	if environmentID.Valid {
		envUUID, _ := uuid.Parse(environmentID.String)
		key.EnvironmentID = &envUUID
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}
	return &key, nil
}

func (r *APIKeyRepo) scanAPIKeyRows(rows *sql.Rows) (*entity.APIKey, error) {
	var key entity.APIKey
	var environmentID sql.NullString
	var expiresAt, lastUsedAt, revokedAt sql.NullTime
	err := rows.Scan(&key.ID, &key.TenantID, &environmentID, &key.Name, &key.KeyPrefix, &key.KeyHash,
		pq.Array(&key.Permissions), &key.Active, &expiresAt, &lastUsedAt, &key.CreatedAt, &revokedAt)
	if err != nil {
		return nil, err
	}
	if environmentID.Valid {
		envUUID, _ := uuid.Parse(environmentID.String)
		key.EnvironmentID = &envUUID
	}
	if expiresAt.Valid {
		key.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		key.LastUsedAt = &lastUsedAt.Time
	}
	if revokedAt.Valid {
		key.RevokedAt = &revokedAt.Time
	}
	return &key, nil
}
