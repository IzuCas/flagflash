package service

import (
	"context"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/IzuCas/flagflash/internal/infrastructure/email"
	"github.com/google/uuid"
)

// EmailSender abstracts email sending for testability
type EmailSender interface {
	IsConfigured() bool
	SendInvite(to, inviterName, tenantName, role, acceptURL string) error
}

// UserService handles user and membership business logic
type UserService struct {
	userRepo       repository.UserRepository
	membershipRepo repository.UserTenantMembershipRepository
	tenantRepo     repository.TenantRepository
	auditRepo      repository.AuditLogRepository
	inviteRepo     repository.InviteTokenRepository
	emailService   EmailSender
	appURL         string
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repository.UserRepository,
	membershipRepo repository.UserTenantMembershipRepository,
	tenantRepo repository.TenantRepository,
	auditRepo repository.AuditLogRepository,
	inviteRepo repository.InviteTokenRepository,
	emailSvc *email.Service,
	appURL string,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		membershipRepo: membershipRepo,
		tenantRepo:     tenantRepo,
		auditRepo:      auditRepo,
		inviteRepo:     inviteRepo,
		emailService:   emailSvc,
		appURL:         appURL,
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

// InviteUserToTenant creates an invite token and sends an email
func (s *UserService) InviteUserToTenant(ctx context.Context, req *InviteUserRequest, actorID string) (*InviteResult, error) {
	// Check if tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	actorUUID, err := uuid.Parse(actorID)
	if err != nil {
		return nil, fmt.Errorf("invalid actor ID")
	}

	// Check hierarchy
	if err := s.checkHierarchy(ctx, actorUUID, req.TenantID, req.Role); err != nil {
		return nil, err
	}

	// If user already exists and is already a member, reject
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		exists, _ := s.membershipRepo.ExistsByUserAndTenant(ctx, existingUser.ID, req.TenantID)
		if exists {
			return nil, fmt.Errorf("user is already a member of this tenant")
		}
	}

	// Check if there's already a pending invite
	pending, _ := s.inviteRepo.GetPendingByEmailAndTenant(ctx, req.Email, req.TenantID)
	if pending != nil {
		return nil, fmt.Errorf("an invitation is already pending for this email")
	}

	// Create invite token
	invite, err := entity.NewInviteToken(req.TenantID, req.Email, req.Role, actorUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite token: %w", err)
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	// Get inviter name
	inviter, _ := s.userRepo.GetByID(ctx, actorUUID)
	inviterName := "A team member"
	if inviter != nil {
		inviterName = inviter.Name
	}

	// Send email
	emailSent := false
	if s.emailService != nil && s.emailService.IsConfigured() {
		acceptURL := fmt.Sprintf("%s/accept-invite?token=%s", s.appURL, invite.Token)
		if err := s.emailService.SendInvite(req.Email, inviterName, tenant.Name, string(req.Role), acceptURL); err != nil {
			// Log but don't fail - the invite was created and can be shared manually
			fmt.Printf("WARNING: Failed to send invite email to %s: %v\n", req.Email, err)
		} else {
			emailSent = true
		}
	}

	// Audit log
	auditLog := entity.NewAuditLog(
		tenant.ID,
		entity.EntityTypeUser,
		invite.ID,
		entity.AuditActionCreate,
		actorID,
		entity.ActorTypeUser,
		nil,
		map[string]interface{}{
			"action":     "invite_sent",
			"email":      req.Email,
			"role":       req.Role,
			"email_sent": emailSent,
		},
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return &InviteResult{
		InviteID:  invite.ID,
		Token:     invite.Token,
		Email:     invite.Email,
		Role:      invite.Role,
		ExpiresAt: invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		EmailSent: emailSent,
	}, nil
}

// InviteResult represents the result of creating an invite
type InviteResult struct {
	InviteID  uuid.UUID       `json:"invite_id"`
	Token     string          `json:"token"`
	Email     string          `json:"email"`
	Role      entity.UserRole `json:"role"`
	ExpiresAt string          `json:"expires_at"`
	EmailSent bool            `json:"email_sent"`
}

// ValidateInviteToken validates an invite token and returns its details
func (s *UserService) ValidateInviteToken(ctx context.Context, token string) (*InviteDetails, error) {
	invite, err := s.inviteRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid invite token")
	}

	if invite.IsAccepted() {
		return nil, fmt.Errorf("this invitation has already been accepted")
	}

	if invite.IsExpired() {
		return nil, fmt.Errorf("this invitation has expired")
	}

	tenant, err := s.tenantRepo.GetByID(ctx, invite.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found")
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, invite.Email)
	userExists := existingUser != nil

	return &InviteDetails{
		Email:      invite.Email,
		TenantName: tenant.Name,
		Role:       string(invite.Role),
		ExpiresAt:  invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		UserExists: userExists,
	}, nil
}

// InviteDetails represents invite token details for the frontend
type InviteDetails struct {
	Email      string `json:"email"`
	TenantName string `json:"tenant_name"`
	Role       string `json:"role"`
	ExpiresAt  string `json:"expires_at"`
	UserExists bool   `json:"user_exists"`
}

// AcceptInviteRequest represents a request to accept an invite
type AcceptInviteRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// AcceptInvite accepts an invitation - creates user if needed and adds to tenant
func (s *UserService) AcceptInvite(ctx context.Context, req *AcceptInviteRequest) (*entity.UserWithMembership, error) {
	invite, err := s.inviteRepo.GetByToken(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid invite token")
	}

	if invite.IsAccepted() {
		return nil, fmt.Errorf("this invitation has already been accepted")
	}

	if invite.IsExpired() {
		return nil, fmt.Errorf("this invitation has expired")
	}

	// Check tenant exists
	_, err = s.tenantRepo.GetByID(ctx, invite.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found")
	}

	// Check if user already exists
	user, _ := s.userRepo.GetByEmail(ctx, invite.Email)

	if user == nil {
		// Create new user
		if req.Name == "" {
			return nil, fmt.Errorf("name is required for new users")
		}
		if req.Password == "" {
			return nil, fmt.Errorf("password is required for new users")
		}
		if len(req.Password) < 8 {
			return nil, fmt.Errorf("password must be at least 8 characters")
		}

		user, err = entity.NewUser(invite.TenantID, invite.Email, req.Password, req.Name, invite.Role)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Check if already a member (race condition protection)
	exists, _ := s.membershipRepo.ExistsByUserAndTenant(ctx, user.ID, invite.TenantID)
	if exists {
		// Mark invite as accepted even though user is already a member
		invite.Accept()
		s.inviteRepo.Update(ctx, invite)
		return nil, fmt.Errorf("user is already a member of this tenant")
	}

	// Create membership
	membership := entity.NewUserTenantMembership(user.ID, invite.TenantID, invite.Role)
	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// Mark invite as accepted
	invite.Accept()
	if err := s.inviteRepo.Update(ctx, invite); err != nil {
		// Non-fatal - membership was created
		fmt.Printf("WARNING: Failed to mark invite as accepted: %v\n", err)
	}

	// Audit log
	auditLog := entity.NewAuditLog(
		invite.TenantID,
		entity.EntityTypeUser,
		user.ID,
		entity.AuditActionCreate,
		user.ID.String(),
		entity.ActorTypeUser,
		nil,
		map[string]interface{}{
			"action":    "invite_accepted",
			"email":     invite.Email,
			"role":      invite.Role,
			"invite_id": invite.ID,
		},
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	return &entity.UserWithMembership{
		User:       user,
		Membership: membership,
	}, nil
}
