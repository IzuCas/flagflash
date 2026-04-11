package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// ApplicationRepo implements repository.ApplicationRepository
type ApplicationRepo struct {
	db *DB
}

// NewApplicationRepo creates a new application repository
func NewApplicationRepo(db *DB) repository.ApplicationRepository {
	return &ApplicationRepo{db: db}
}

// Create creates a new application
func (r *ApplicationRepo) Create(ctx context.Context, app *entity.Application) error {
	query := `
		INSERT INTO applications (id, tenant_id, name, slug, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		app.ID, app.TenantID, app.Name, app.Slug, app.Description, app.CreatedAt, app.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}

	return nil
}

// GetByID retrieves an application by ID
func (r *ApplicationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Application, error) {
	query := `
		SELECT id, tenant_id, name, slug, description, created_at, updated_at, deleted_at
		FROM applications
		WHERE id = $1 AND deleted_at IS NULL
	`

	var app entity.Application
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.Slug, &description,
		&app.CreatedAt, &app.UpdatedAt, &app.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found")
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	if description.Valid {
		app.Description = description.String
	}

	return &app, nil
}

// GetBySlug retrieves an application by tenant ID and slug
func (r *ApplicationRepo) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*entity.Application, error) {
	query := `
		SELECT id, tenant_id, name, slug, description, created_at, updated_at, deleted_at
		FROM applications
		WHERE tenant_id = $1 AND slug = $2 AND deleted_at IS NULL
	`

	var app entity.Application
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID, slug).Scan(
		&app.ID, &app.TenantID, &app.Name, &app.Slug, &description,
		&app.CreatedAt, &app.UpdatedAt, &app.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("application not found")
		}
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	if description.Valid {
		app.Description = description.String
	}

	return &app, nil
}

// Update updates an application
func (r *ApplicationRepo) Update(ctx context.Context, app *entity.Application) error {
	query := `
		UPDATE applications
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		app.Name, app.Description, app.UpdatedAt, app.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update application: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

// Delete soft deletes an application
func (r *ApplicationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE applications
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

// GetByTenantID retrieves all applications for a tenant
func (r *ApplicationRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Application, error) {
	query := `
		SELECT id, tenant_id, name, slug, description, created_at, updated_at, deleted_at
		FROM applications
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list applications: %w", err)
	}
	defer rows.Close()

	var apps []*entity.Application
	for rows.Next() {
		var app entity.Application
		var description sql.NullString

		if err := rows.Scan(
			&app.ID, &app.TenantID, &app.Name, &app.Slug, &description,
			&app.CreatedAt, &app.UpdatedAt, &app.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan application: %w", err)
		}

		if description.Valid {
			app.Description = description.String
		}

		apps = append(apps, &app)
	}

	return apps, nil
}

// ListByTenant lists all applications for a tenant with pagination
func (r *ApplicationRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.Application, int, error) {
	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM applications WHERE tenant_id = $1 AND deleted_at IS NULL`
	if err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count applications: %w", err)
	}

	query := `
		SELECT id, tenant_id, name, slug, description, created_at, updated_at, deleted_at
		FROM applications
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list applications: %w", err)
	}
	defer rows.Close()

	var apps []*entity.Application
	for rows.Next() {
		var app entity.Application
		var description sql.NullString

		if err := rows.Scan(
			&app.ID, &app.TenantID, &app.Name, &app.Slug, &description,
			&app.CreatedAt, &app.UpdatedAt, &app.DeletedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan application: %w", err)
		}

		if description.Valid {
			app.Description = description.String
		}

		apps = append(apps, &app)
	}

	return apps, total, nil
}
