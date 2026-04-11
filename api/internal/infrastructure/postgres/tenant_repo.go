package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// TenantRepo implements repository.TenantRepository
type TenantRepo struct {
	db *DB
}

// NewTenantRepo creates a new TenantRepo
func NewTenantRepo(db *DB) repository.TenantRepository {
	return &TenantRepo{db: db}
}

// Create creates a new tenant
func (r *TenantRepo) Create(ctx context.Context, tenant *entity.Tenant) error {
	query := `INSERT INTO tenants (id, name, slug, plan, active, settings, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Plan, tenant.Active,
		"{}",
		tenant.CreatedAt, tenant.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}

// GetByID retrieves a tenant by ID
func (r *TenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	query := `SELECT id, name, slug, plan, active, created_at, updated_at, deleted_at
			  FROM tenants WHERE id = $1 AND deleted_at IS NULL`
	tenant := &entity.Tenant{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan, &tenant.Active,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (r *TenantRepo) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	query := `SELECT id, name, slug, plan, active, created_at, updated_at, deleted_at
			  FROM tenants WHERE slug = $1 AND deleted_at IS NULL`
	tenant := &entity.Tenant{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan, &tenant.Active,
		&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return tenant, nil
}

// GetAll retrieves all tenants
func (r *TenantRepo) GetAll(ctx context.Context) ([]*entity.Tenant, error) {
	query := `SELECT id, name, slug, plan, active, created_at, updated_at, deleted_at
			  FROM tenants WHERE deleted_at IS NULL ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*entity.Tenant
	for rows.Next() {
		tenant := &entity.Tenant{}
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan, &tenant.Active,
			&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}
	return tenants, nil
}

// List lists all tenants with pagination
func (r *TenantRepo) List(ctx context.Context, limit, offset int) ([]*entity.Tenant, int, error) {
	countQuery := `SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	query := `SELECT id, name, slug, plan, active, created_at, updated_at, deleted_at
			  FROM tenants WHERE deleted_at IS NULL ORDER BY name LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*entity.Tenant
	for rows.Next() {
		tenant := &entity.Tenant{}
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan, &tenant.Active,
			&tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}
	return tenants, total, nil
}

// Update updates a tenant
func (r *TenantRepo) Update(ctx context.Context, tenant *entity.Tenant) error {
	query := `UPDATE tenants SET name = $1, plan = $2, active = $3, updated_at = $4 WHERE id = $5`
	_, err := r.db.ExecContext(ctx, query, tenant.Name, tenant.Plan, tenant.Active, tenant.UpdatedAt, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	return nil
}

// Delete soft deletes a tenant
func (r *TenantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE tenants SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}
