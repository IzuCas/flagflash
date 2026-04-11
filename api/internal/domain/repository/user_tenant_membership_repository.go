package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// UserTenantMembershipRepository defines the interface for user-tenant membership persistence
type UserTenantMembershipRepository interface {
	// Create creates a new membership
	Create(ctx context.Context, membership *entity.UserTenantMembership) error

	// GetByID retrieves a membership by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.UserTenantMembership, error)

	// GetByUserAndTenant retrieves a membership by user and tenant
	GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*entity.UserTenantMembership, error)

	// Update updates a membership
	Update(ctx context.Context, membership *entity.UserTenantMembership) error

	// Delete deletes a membership
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByUserAndTenant deletes a membership by user and tenant
	DeleteByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) error

	// ListByUser lists all memberships for a user
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.UserTenantMembership, error)

	// ListByTenant lists all memberships for a tenant with pagination
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserTenantMembership, int, error)

	// ListUsersWithMembershipByTenant lists users with their membership details for a tenant
	ListUsersWithMembershipByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserWithMembership, int, error)

	// ListTenantsForUser lists all tenants for a user with their roles
	ListTenantsForUser(ctx context.Context, userID uuid.UUID) ([]*entity.TenantWithRole, error)

	// ExistsByUserAndTenant checks if a membership exists for user and tenant
	ExistsByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (bool, error)
}
