package service

import (
	"context"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// UserService handles user and membership business logic
type UserService struct {
	userRepo       repository.UserRepository
	membershipRepo repository.UserTenantMembershipRepository
	tenantRepo     repository.TenantRepository
	auditRepo      repository.AuditLogRepository
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	membershipRepo repository.UserTenantMembershipRepository,
	tenantRepo repository.TenantRepository,
	auditRepo repository.AuditLogRepository,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		membershipRepo: membershipRepo,
		tenantRepo:     tenantRepo,
		auditRepo:      auditRepo,
	}
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Email    string          `json:"email"`
	Password string          `json:"password"`
	Name     string          `json:"name"`
	TenantID uuid.UUID       `json:"tenant_id"`
	Role     entity.UserRole `json:"role"`
}

// checkHierarchy verifies if the actor has permission to manage the target role
func (s *UserService) checkHierarchy(ctx context.Context, actorID uuid.UUID, tenantID uuid.UUID, targetRole entity.UserRole) error {
	// Get actor's membership to find their role
	actorMembership, err := s.membershipRepo.GetByUserAndTenant(ctx, actorID, tenantID)
	if err != nil {
		return fmt.Errorf("actor membership not found")
	}

	// Check if actor can manage the target role
	if !actorMembership.Role.CanManageRole(targetRole) {
		return fmt.Errorf("insufficient permissions: cannot manage users with role '%s'", targetRole)
	}

	return nil
}

// CreateUser creates a new user and adds them to a tenant
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest, actorID string) (*entity.UserWithMembership, error) {
	// Check hierarchy: actor must be able to create users with the requested role
	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor ID")
	}

	if err := s.checkHierarchy(ctx, actorUUID, req.TenantID, req.Role); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, _ := s.userRepo.ExistsByEmail(ctx, req.Email)
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Check if tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Create user (using tenant_id for backwards compatibility)
	user, err := entity.NewUser(req.TenantID, req.Email, req.Password, req.Name, req.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create membership
	membership := entity.NewUserTenantMembership(user.ID, req.TenantID, req.Role)
	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		// Rollback user creation
		s.userRepo.Delete(ctx, user.ID)
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenant.ID,
		entity.EntityTypeUser,
		user.ID,
		entity.AuditActionCreate,
		actorID,
		entity.ActorTypeUser,
		nil,
		user,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return &entity.UserWithMembership{
		User:       user,
		Membership: membership,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Name     string           `json:"name,omitempty"`
	Role     *entity.UserRole `json:"role,omitempty"`
	Active   *bool            `json:"active,omitempty"`
	TenantID uuid.UUID        `json:"tenant_id"` // Tenant context for role update
}

// UpdateUser updates a user's details
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req *UpdateUserRequest, actorID string) (*entity.UserWithMembership, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get target user's current role in this tenant for hierarchy check
	targetMembership, err := s.membershipRepo.GetByUserAndTenant(ctx, userID, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("user membership not found: %w", err)
	}

	// Owner users are immutable — no one can modify them
	if targetMembership.Role == entity.UserRoleOwner {
		return nil, fmt.Errorf("owner users cannot be modified")
	}

	// Check hierarchy: actor must be able to manage target user's current role
	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor ID")
	}

	if err := s.checkHierarchy(ctx, actorUUID, req.TenantID, targetMembership.Role); err != nil {
		return nil, err
	}

	// If changing role, also check if actor can assign the new role
	if req.Role != nil {
		if err := s.checkHierarchy(ctx, actorUUID, req.TenantID, *req.Role); err != nil {
			return nil, fmt.Errorf("cannot assign role '%s': %w", *req.Role, err)
		}
	}

	oldUser := *user

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Update membership role if role is provided and tenant_id is set
	var membership *entity.UserTenantMembership
	if req.Role != nil && req.TenantID != uuid.Nil {
		membership, err = s.membershipRepo.GetByUserAndTenant(ctx, userID, req.TenantID)
		if err == nil && membership != nil {
			membership.Update(*req.Role)
			s.membershipRepo.Update(ctx, membership)
		}
	} else if req.TenantID != uuid.Nil {
		membership, _ = s.membershipRepo.GetByUserAndTenant(ctx, userID, req.TenantID)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		req.TenantID,
		entity.EntityTypeUser,
		user.ID,
		entity.AuditActionUpdate,
		actorID,
		entity.ActorTypeUser,
		oldUser,
		user,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return &entity.UserWithMembership{
		User:       user,
		Membership: membership,
	}, nil
}

// DeleteUser soft deletes a user and all their memberships
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID, actorID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Get target user's role for hierarchy check
	targetMembership, err := s.membershipRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return fmt.Errorf("user membership not found: %w", err)
	}

	// Owner users are immutable — no one can delete them
	if targetMembership.Role == entity.UserRoleOwner {
		return fmt.Errorf("owner users cannot be deleted")
	}

	// Check hierarchy: actor must be able to manage target user's role
	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		return fmt.Errorf("invalid actor ID")
	}

	if err := s.checkHierarchy(ctx, actorUUID, tenantID, targetMembership.Role); err != nil {
		return err
	}

	// Delete membership for this tenant
	if err := s.membershipRepo.DeleteByUserAndTenant(ctx, userID, tenantID); err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	// Check if user has other memberships
	memberships, err := s.membershipRepo.ListByUser(ctx, userID)
	if err != nil || len(memberships) == 0 {
		// No more memberships, soft delete the user
		if err := s.userRepo.Delete(ctx, userID); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenantID,
		entity.EntityTypeUser,
		user.ID,
		entity.AuditActionDelete,
		actorID,
		entity.ActorTypeUser,
		user,
		nil,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// ListUsersByTenant lists all users for a tenant with pagination
func (s *UserService) ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserWithMembership, int, error) {
	users, total, err := s.membershipRepo.ListUsersWithMembershipByTenant(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	return users, total, nil
}

// AddUserToTenantRequest represents a request to add an existing user to a tenant
type AddUserToTenantRequest struct {
	UserID   uuid.UUID       `json:"user_id"`
	TenantID uuid.UUID       `json:"tenant_id"`
	Role     entity.UserRole `json:"role"`
}

// AddUserToTenant adds an existing user to a tenant
func (s *UserService) AddUserToTenant(ctx context.Context, req *AddUserToTenantRequest, actorID string) (*entity.UserTenantMembership, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Check if membership already exists
	exists, _ := s.membershipRepo.ExistsByUserAndTenant(ctx, req.UserID, req.TenantID)
	if exists {
		return nil, fmt.Errorf("user is already a member of this tenant")
	}

	// Create membership
	membership := entity.NewUserTenantMembership(req.UserID, req.TenantID, req.Role)
	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenant.ID,
		entity.EntityTypeUser,
		user.ID,
		entity.AuditActionCreate,
		actorID,
		entity.ActorTypeUser,
		nil,
		map[string]interface{}{
			"user_id":   req.UserID,
			"tenant_id": req.TenantID,
			"role":      req.Role,
		},
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return membership, nil
}

// RemoveUserFromTenant removes a user from a tenant
func (s *UserService) RemoveUserFromTenant(ctx context.Context, userID, tenantID uuid.UUID, actorID string) error {
	// Check if membership exists
	membership, err := s.membershipRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return fmt.Errorf("membership not found: %w", err)
	}

	// Delete the membership
	if err := s.membershipRepo.Delete(ctx, membership.ID); err != nil {
		return fmt.Errorf("failed to remove user from tenant: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenantID,
		entity.EntityTypeUser,
		userID,
		entity.AuditActionDelete,
		actorID,
		entity.ActorTypeUser,
		membership,
		nil,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return nil
}

// UpdateMembershipRequest represents a request to update a membership
type UpdateMembershipRequest struct {
	Role   *entity.UserRole `json:"role,omitempty"`
	Active *bool            `json:"active,omitempty"`
}

// UpdateMembership updates a user's membership in a tenant
func (s *UserService) UpdateMembership(ctx context.Context, userID, tenantID uuid.UUID, req *UpdateMembershipRequest, actorID string) (*entity.UserTenantMembership, error) {
	membership, err := s.membershipRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("membership not found: %w", err)
	}

	// Owners cannot be modified
	if membership.Role == entity.UserRoleOwner {
		return nil, fmt.Errorf("owner users cannot be modified")
	}

	// Verify actor has permission to manage the target's current role
	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor ID")
	}
	if err := s.checkHierarchy(ctx, actorUUID, tenantID, membership.Role); err != nil {
		return nil, err
	}

	oldMembership := *membership

	if req.Role != nil {
		// Actor must also outrank the new role being assigned
		if err := s.checkHierarchy(ctx, actorUUID, tenantID, *req.Role); err != nil {
			return nil, fmt.Errorf("cannot assign role '%s': %w", *req.Role, err)
		}
		membership.Role = *req.Role
	}
	if req.Active != nil {
		membership.Active = *req.Active
	}
	membership.UpdatedAt = membership.CreatedAt // Will be set by Update method

	if err := s.membershipRepo.Update(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to update membership: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenantID,
		entity.EntityTypeUser,
		userID,
		entity.AuditActionUpdate,
		actorID,
		entity.ActorTypeUser,
		oldMembership,
		membership,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return membership, nil
}

// GetTenantsForUser retrieves all tenants for a user
func (s *UserService) GetTenantsForUser(ctx context.Context, userID uuid.UUID) ([]*entity.TenantWithRole, error) {
	tenants, err := s.membershipRepo.ListTenantsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

// InviteUserRequest represents a request to invite a user to a tenant
type InviteUserRequest struct {
	Email    string          `json:"email"`
	TenantID uuid.UUID       `json:"tenant_id"`
	Role     entity.UserRole `json:"role"`
}

// InviteUserToTenant invites a user to a tenant (creates user if doesn't exist, or adds to tenant)
func (s *UserService) InviteUserToTenant(ctx context.Context, req *InviteUserRequest, actorID string) (*entity.UserWithMembership, error) {
	// Check if tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Check if user already exists
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && user != nil {
		// User exists, check if already a member
		exists, _ := s.membershipRepo.ExistsByUserAndTenant(ctx, user.ID, req.TenantID)
		if exists {
			return nil, fmt.Errorf("user is already a member of this tenant")
		}

		// Add user to tenant
		membership := entity.NewUserTenantMembership(user.ID, req.TenantID, req.Role)
		if err := s.membershipRepo.Create(ctx, membership); err != nil {
			return nil, fmt.Errorf("failed to create membership: %w", err)
		}

		// Create audit log
		auditLog := entity.NewAuditLog(
			tenant.ID,
			entity.EntityTypeUser,
			user.ID,
			entity.AuditActionCreate,
			actorID,
			entity.ActorTypeUser,
			nil,
			map[string]interface{}{
				"action": "invited_existing_user",
				"email":  req.Email,
				"role":   req.Role,
			},
			nil,
		)
		s.auditRepo.Create(ctx, auditLog)

		return &entity.UserWithMembership{
			User:       user,
			Membership: membership,
		}, nil
	}

	// User doesn't exist - return error, they need to register first or use CreateUser
	return nil, fmt.Errorf("user with email '%s' not found. Please create a new user instead", req.Email)
}
