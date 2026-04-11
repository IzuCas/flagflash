package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// EnvironmentRepository implements repository.EnvironmentRepository
type EnvironmentRepository struct {
	db *sql.DB
}

// NewEnvironmentRepository creates a new EnvironmentRepository
func NewEnvironmentRepository(db *sql.DB) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}

// Create creates a new environment
func (r *EnvironmentRepository) Create(ctx context.Context, env *entity.Environment) error {
	query := `INSERT INTO environments (id, application_id, name, slug, description, color, is_production, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		env.ID, env.ApplicationID, env.Name, env.Slug, env.Description, env.Color, env.IsProduction,
		env.CreatedAt, env.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}
	return nil
}

// CreateBatch creates multiple environments
func (r *EnvironmentRepository) CreateBatch(ctx context.Context, envs []*entity.Environment) error {
	for _, env := range envs {
		if err := r.Create(ctx, env); err != nil {
			return err
		}
	}
	return nil
}

// GetByID retrieves an environment by ID
func (r *EnvironmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error) {
	query := `SELECT id, application_id, name, slug, description, color, is_production, created_at, updated_at
			  FROM environments WHERE id = $1`
	env := &entity.Environment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&env.ID, &env.ApplicationID, &env.Name, &env.Slug, &env.Description, &env.Color, &env.IsProduction,
		&env.CreatedAt, &env.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	return env, nil
}

// GetByApplicationID retrieves all environments for an application
func (r *EnvironmentRepository) GetByApplicationID(ctx context.Context, appID uuid.UUID) ([]*entity.Environment, error) {
	query := `SELECT id, application_id, name, slug, description, color, is_production, created_at, updated_at
			  FROM environments WHERE application_id = $1 ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environments: %w", err)
	}
	defer rows.Close()

	var envs []*entity.Environment
	for rows.Next() {
		env := &entity.Environment{}
		if err := rows.Scan(&env.ID, &env.ApplicationID, &env.Name, &env.Slug, &env.Description, &env.Color, &env.IsProduction,
			&env.CreatedAt, &env.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		envs = append(envs, env)
	}
	return envs, nil
}

// Update updates an environment
func (r *EnvironmentRepository) Update(ctx context.Context, env *entity.Environment) error {
	query := `UPDATE environments SET name = $1, description = $2, color = $3, is_production = $4, updated_at = $5 WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, env.Name, env.Description, env.Color, env.IsProduction, env.UpdatedAt, env.ID)
	if err != nil {
		return fmt.Errorf("failed to update environment: %w", err)
	}
	return nil
}

// Delete deletes an environment
func (r *EnvironmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM environments WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}
	return nil
}

// GetByIDWithTenant retrieves an environment with its tenant info
func (r *EnvironmentRepository) GetByIDWithTenant(ctx context.Context, id uuid.UUID) (*entity.Environment, uuid.UUID, error) {
	query := `SELECT e.id, e.application_id, e.name, e.slug, e.description, e.color, e.is_production, e.created_at, e.updated_at, a.tenant_id
			  FROM environments e
			  JOIN applications a ON e.application_id = a.id
			  WHERE e.id = $1`
	env := &entity.Environment{}
	var tenantID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&env.ID, &env.ApplicationID, &env.Name, &env.Slug, &env.Description, &env.Color, &env.IsProduction,
		&env.CreatedAt, &env.UpdatedAt, &tenantID)
	if err == sql.ErrNoRows {
		return nil, uuid.Nil, nil
	}
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("failed to get environment with tenant: %w", err)
	}
	return env, tenantID, nil
}

// GetBySlug retrieves an environment by application ID and slug
func (r *EnvironmentRepository) GetBySlug(ctx context.Context, applicationID uuid.UUID, slug string) (*entity.Environment, error) {
	query := `SELECT id, application_id, name, slug, description, color, is_production, created_at, updated_at
			  FROM environments WHERE application_id = $1 AND slug = $2`
	env := &entity.Environment{}
	err := r.db.QueryRowContext(ctx, query, applicationID, slug).Scan(
		&env.ID, &env.ApplicationID, &env.Name, &env.Slug, &env.Description, &env.Color, &env.IsProduction,
		&env.CreatedAt, &env.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get environment by slug: %w", err)
	}
	return env, nil
}

// ListByApplication lists all environments for an application
func (r *EnvironmentRepository) ListByApplication(ctx context.Context, applicationID uuid.UUID) ([]*entity.Environment, error) {
	return r.GetByApplicationID(ctx, applicationID)
}
