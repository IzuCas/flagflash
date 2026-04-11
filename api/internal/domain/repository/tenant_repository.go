package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// TenantRepository defines the interface for tenant persistence
type TenantRepository interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *entity.Tenant) error

	// GetByID retrieves a tenant by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)

	// GetBySlug retrieves a tenant by slug
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)

	// Update updates a tenant
	Update(ctx context.Context, tenant *entity.Tenant) error

	// Delete soft deletes a tenant
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists all tenants with pagination
	List(ctx context.Context, limit, offset int) ([]*entity.Tenant, int, error)
}
