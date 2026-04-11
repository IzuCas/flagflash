package service

import (
	"context"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/IzuCas/flagflash/pkg/auth"
	"github.com/google/uuid"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       repository.UserRepository
	tenantRepo     repository.TenantRepository
	membershipRepo repository.UserTenantMembershipRepository
	auditRepo      repository.AuditLogRepository
	jwtSecret      string
	jwtExpiry      time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	tenantRepo repository.TenantRepository,
	membershipRepo repository.UserTenantMembershipRepository,
	auditRepo repository.AuditLogRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tenantRepo:     tenantRepo,
		membershipRepo: membershipRepo,
		auditRepo:      auditRepo,
		jwtSecret:      jwtSecret,
		jwtExpiry:      jwtExpiry,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string                   `json:"token"`
	ExpiresAt time.Time                `json:"expires_at"`
	User      *entity.User             `json:"user"`
	Tenants   []*entity.TenantWithRole `json:"tenants"`
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.CheckPassword(req.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.Active {
		return nil, fmt.Errorf("user account is disabled")
	}

	// Get user's tenants via membership
	tenants, err := s.membershipRepo.ListTenantsForUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	if len(tenants) == 0 {
		return nil, fmt.Errorf("user has no tenant access")
	}

	// Use first tenant for initial JWT (user will select tenant in frontend)
	defaultTenant := tenants[0]

	// Generate JWT token
	expiresAt := time.Now().Add(s.jwtExpiry)
	claims := &entity.UserClaims{
		UserID:   user.ID,
		TenantID: defaultTenant.Tenant.ID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     defaultTenant.Role,
	}

	token, err := auth.GenerateJWT(claims, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
		Tenants:   tenants,
	}, nil
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
}

// Register creates a new tenant and owner user
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*LoginResponse, error) {
	// Check if email already exists
	exists, _ := s.userRepo.ExistsByEmail(ctx, req.Email)
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Check if tenant slug already exists
	existingTenant, _ := s.tenantRepo.GetBySlug(ctx, req.TenantSlug)
	if existingTenant != nil {
		return nil, fmt.Errorf("tenant slug already exists")
	}

	// Create tenant
	tenant := entity.NewTenant(req.TenantName, req.TenantSlug)
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create owner user
	user, err := entity.NewUser(tenant.ID, req.Email, req.Password, req.Name, entity.UserRoleOwner)
	if err != nil {
		// Rollback tenant creation
		s.tenantRepo.Delete(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		// Rollback tenant creation
		s.tenantRepo.Delete(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create user-tenant membership
	membership := entity.NewUserTenantMembership(user.ID, tenant.ID, entity.UserRoleOwner)
	if err := s.membershipRepo.Create(ctx, membership); err != nil {
		// Rollback user and tenant
		s.userRepo.Delete(ctx, user.ID)
		s.tenantRepo.Delete(ctx, tenant.ID)
		return nil, fmt.Errorf("failed to create membership: %w", err)
	}

	// Create audit logs
	auditLog := entity.NewAuditLog(
		tenant.ID,
		entity.EntityTypeTenant,
		tenant.ID,
		entity.AuditActionCreate,
		user.ID.String(),
		entity.ActorTypeUser,
		nil,
		tenant,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	// Build tenant list for response
	tenants := []*entity.TenantWithRole{
		{
			Tenant: tenant,
			Role:   entity.UserRoleOwner,
			Active: true,
		},
	}

	// Generate JWT token
	expiresAt := time.Now().Add(s.jwtExpiry)
	claims := &entity.UserClaims{
		UserID:   user.ID,
		TenantID: tenant.ID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     entity.UserRoleOwner,
	}

	token, err := auth.GenerateJWT(claims, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
		Tenants:   tenants,
	}, nil
}

// RefreshToken refreshes a JWT token
func (s *AuthService) RefreshToken(ctx context.Context, claims *entity.UserClaims) (*LoginResponse, error) {
	// Verify user still exists and is active
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.Active {
		return nil, fmt.Errorf("user account is disabled")
	}

	// Get user's tenants via membership
	tenants, err := s.membershipRepo.ListTenantsForUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	// Verify user still has access to the current tenant
	currentTenantID := claims.TenantID
	var currentRole entity.UserRole
	hasAccess := false
	for _, t := range tenants {
		if t.Tenant.ID == currentTenantID {
			hasAccess = true
			currentRole = t.Role
			break
		}
	}

	if !hasAccess && len(tenants) > 0 {
		// Fall back to first tenant
		currentTenantID = tenants[0].Tenant.ID
		currentRole = tenants[0].Role
	}

	// Generate new token
	expiresAt := time.Now().Add(s.jwtExpiry)
	newClaims := &entity.UserClaims{
		UserID:   user.ID,
		TenantID: currentTenantID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     currentRole,
	}

	token, err := auth.GenerateJWT(newClaims, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
		Tenants:   tenants,
	}, nil
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return fmt.Errorf("current password is incorrect")
	}

	if err := user.UpdatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

// SwitchTenantRequest represents a switch tenant request
type SwitchTenantRequest struct {
	TenantID uuid.UUID `json:"tenant_id"`
}

// SwitchTenant switches the user's current tenant and generates a new JWT
func (s *AuthService) SwitchTenant(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*LoginResponse, error) {
	// Verify user exists and is active
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if !user.Active {
		return nil, fmt.Errorf("user account is disabled")
	}

	// Check user has access to the target tenant
	hasAccess, err := s.membershipRepo.ExistsByUserAndTenant(ctx, userID, tenantID)
	if err != nil || !hasAccess {
		return nil, fmt.Errorf("access denied to this tenant")
	}

	// Get membership to get role
	membership, err := s.membershipRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get membership: %w", err)
	}

	// Get all user's tenants for the response
	tenants, err := s.membershipRepo.ListTenantsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	// Generate new JWT with the new tenant
	expiresAt := time.Now().Add(s.jwtExpiry)
	claims := &entity.UserClaims{
		UserID:   user.ID,
		TenantID: tenantID,
		Email:    user.Email,
		Name:     user.Name,
		Role:     membership.Role,
	}

	token, err := auth.GenerateJWT(claims, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
		Tenants:   tenants,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetProfile retrieves the current user's profile
func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// UpdateProfileRequest represents an update profile request
type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// UpdateProfile updates the user's profile
func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// GetJWTSecret returns the JWT secret for token validation
func (s *AuthService) GetJWTSecret() string {
	return s.jwtSecret
}
