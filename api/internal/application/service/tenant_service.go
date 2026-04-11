package service

import (
	"context"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// TenantService handles tenant business logic
type TenantService struct {
	tenantRepo     repository.TenantRepository
	membershipRepo repository.UserTenantMembershipRepository
	auditRepo      repository.AuditLogRepository
}

// NewTenantService creates a new tenant service
func NewTenantService(
	tenantRepo repository.TenantRepository,
	membershipRepo repository.UserTenantMembershipRepository,
	auditRepo repository.AuditLogRepository,
) *TenantService {
	return &TenantService{
		tenantRepo:     tenantRepo,
		membershipRepo: membershipRepo,
		auditRepo:      auditRepo,
	}
}

// Create creates a new tenant and associates the creator as owner
func (s *TenantService) Create(ctx context.Context, name, slug string, actorID string) (*entity.Tenant, error) {
	// Check if slug already exists
	existing, _ := s.tenantRepo.GetBySlug(ctx, slug)
	if existing != nil {
		return nil, fmt.Errorf("tenant with slug '%s' already exists", slug)
	}

	tenant := entity.NewTenant(name, slug)
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Associate the creator as owner of the tenant
	if actorID != "" {
		actorUUID, err := uuid.Parse(actorID)
		if err == nil {
			membership := entity.NewUserTenantMembership(actorUUID, tenant.ID, entity.UserRoleOwner)
			if err := s.membershipRepo.Create(ctx, membership); err != nil {
				// Log error but don't fail tenant creation
				fmt.Printf("failed to create owner membership: %v\n", err)
			}
		}
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenant.ID,
		entity.EntityTypeTenant,
		tenant.ID,
		entity.AuditActionCreate,
		actorID,
		entity.ActorTypeUser,
		nil,
		tenant,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return tenant, nil
}

// GetByID retrieves a tenant by ID
func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (s *TenantService) GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return tenant, nil
}

// Update updates a tenant (only owners can update)
func (s *TenantService) Update(ctx context.Context, id uuid.UUID, name string, settings map[string]interface{}, actorID string) (*entity.Tenant, error) {
	// Check if the actor is owner of the tenant
	if actorID != "" {
		actorUUID, err := uuid.Parse(actorID)
		if err != nil {
			return nil, fmt.Errorf("invalid actor ID: %w", err)
		}
		membership, err := s.membershipRepo.GetByUserAndTenant(ctx, actorUUID, id)
		if err != nil {
			return nil, fmt.Errorf("access denied: not a member of this tenant")
		}
		if membership.Role != entity.UserRoleOwner {
			return nil, fmt.Errorf("access denied: only owners can update tenant settings")
		}
	}

	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	oldTenant := *tenant
	tenant.Update(name, settings)

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenant.ID,
		entity.EntityTypeTenant,
		tenant.ID,
		entity.AuditActionUpdate,
		actorID,
		entity.ActorTypeUser,
		oldTenant,
		tenant,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return tenant, nil
}

// Delete soft deletes a tenant (only owners can delete)
func (s *TenantService) Delete(ctx context.Context, id uuid.UUID, actorID string) error {
	// Check if the actor is owner of the tenant
	if actorID != "" {
		actorUUID, err := uuid.Parse(actorID)
		if err != nil {
			return fmt.Errorf("invalid actor ID: %w", err)
		}
		membership, err := s.membershipRepo.GetByUserAndTenant(ctx, actorUUID, id)
		if err != nil {
			return fmt.Errorf("access denied: not a member of this tenant")
		}
		if membership.Role != entity.UserRoleOwner {
			return fmt.Errorf("access denied: only owners can delete tenants")
		}
	}

	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	if err := s.tenantRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenant.ID,
		entity.EntityTypeTenant,
		tenant.ID,
		entity.AuditActionDelete,
		actorID,
		entity.ActorTypeUser,
		tenant,
		nil,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// List lists all tenants with pagination
func (s *TenantService) List(ctx context.Context, limit, offset int) ([]*entity.Tenant, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.tenantRepo.List(ctx, limit, offset)
}

// ListByUser lists all tenants that a user has access to
func (s *TenantService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.TenantWithRole, error) {
	return s.membershipRepo.ListTenantsForUser(ctx, userID)
}

// HasAccess checks if a user has access to a tenant
func (s *TenantService) HasAccess(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	return s.membershipRepo.ExistsByUserAndTenant(ctx, userID, tenantID)
}
