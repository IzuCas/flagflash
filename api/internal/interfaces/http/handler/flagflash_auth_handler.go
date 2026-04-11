package handler

import (
	"context"
	"net/http"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	"github.com/IzuCas/flagflash/internal/interfaces/http/middleware"
	"github.com/IzuCas/flagflash/pkg/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// FlagFlashAuthHandler handles authentication for FlagFlash
type FlagFlashAuthHandler struct {
	service *service.AuthService
}

// NewFlagFlashAuthHandler creates a new auth handler
func NewFlagFlashAuthHandler(service *service.AuthService) *FlagFlashAuthHandler {
	return &FlagFlashAuthHandler{service: service}
}

// RegisterRoutes registers auth routes
func (h *FlagFlashAuthHandler) RegisterRoutes(api huma.API) {
	// Public routes (no auth required)
	huma.Register(api, huma.Operation{
		OperationID: "flagflashLogin",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Login to FlagFlash",
		Tags:        []string{"Auth"},
	}, h.Login)

	huma.Register(api, huma.Operation{
		OperationID: "flagflashRegister",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register a new tenant and user",
		Tags:        []string{"Auth"},
	}, h.Register)

	huma.Register(api, huma.Operation{
		OperationID: "flagflashRefreshToken",
		Method:      http.MethodPost,
		Path:        "/auth/refresh",
		Summary:     "Refresh authentication token",
		Tags:        []string{"Auth"},
	}, h.RefreshToken)
}

// RegisterProtectedRoutes registers auth routes that require authentication
func (h *FlagFlashAuthHandler) RegisterProtectedRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "flagflashSwitchTenant",
		Method:      http.MethodPost,
		Path:        "/switch-tenant",
		Summary:     "Switch to a different tenant",
		Tags:        []string{"Auth"},
	}, h.SwitchTenant)

	huma.Register(api, huma.Operation{
		OperationID: "flagflashChangePassword",
		Method:      http.MethodPost,
		Path:        "/change-password",
		Summary:     "Change user password",
		Tags:        []string{"Auth"},
	}, h.ChangePassword)

	huma.Register(api, huma.Operation{
		OperationID: "flagflashGetProfile",
		Method:      http.MethodGet,
		Path:        "/profile",
		Summary:     "Get current user profile",
		Tags:        []string{"Auth"},
	}, h.GetProfile)

	huma.Register(api, huma.Operation{
		OperationID: "flagflashUpdateProfile",
		Method:      http.MethodPut,
		Path:        "/profile",
		Summary:     "Update current user profile",
		Tags:        []string{"Auth"},
	}, h.UpdateProfile)
}

// Login handles user login
func (h *FlagFlashAuthHandler) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	result, err := h.service.Login(ctx, &service.LoginRequest{
		Email:    req.Body.Email,
		Password: req.Body.Password,
	})
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid credentials", err)
	}

	// Convert tenants to DTO
	tenantDTOs := make([]dto.TenantWithRoleDTO, len(result.Tenants))
	for i, t := range result.Tenants {
		tenantDTOs[i] = dto.TenantWithRoleDTO{
			ID:        t.Tenant.ID,
			Name:      t.Tenant.Name,
			Slug:      t.Tenant.Slug,
			Role:      string(t.Role),
			CreatedAt: t.Tenant.CreatedAt,
			UpdatedAt: t.Tenant.UpdatedAt,
		}
	}

	return &dto.LoginResponse{
		Body: dto.LoginResponseBody{
			Token:     result.Token,
			ExpiresAt: result.ExpiresAt,
			User: dto.UserDTO{
				ID:       result.User.ID,
				TenantID: result.User.TenantID,
				Email:    result.User.Email,
				Name:     result.User.Name,
				Role:     string(result.User.Role),
			},
			Tenants: tenantDTOs,
		},
	}, nil
}

// Register handles new tenant and user registration
func (h *FlagFlashAuthHandler) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	result, err := h.service.Register(ctx, &service.RegisterRequest{
		Email:      req.Body.Email,
		Password:   req.Body.Password,
		Name:       req.Body.Name,
		TenantName: req.Body.TenantName,
		TenantSlug: req.Body.TenantSlug,
	})
	if err != nil {
		return nil, huma.Error400BadRequest("Registration failed", err)
	}

	// Convert tenants to DTO
	tenantDTOs := make([]dto.TenantWithRoleDTO, len(result.Tenants))
	for i, t := range result.Tenants {
		tenantDTOs[i] = dto.TenantWithRoleDTO{
			ID:        t.Tenant.ID,
			Name:      t.Tenant.Name,
			Slug:      t.Tenant.Slug,
			Role:      string(t.Role),
			CreatedAt: t.Tenant.CreatedAt,
			UpdatedAt: t.Tenant.UpdatedAt,
		}
	}

	return &dto.LoginResponse{
		Body: dto.LoginResponseBody{
			Token:     result.Token,
			ExpiresAt: result.ExpiresAt,
			User: dto.UserDTO{
				ID:       result.User.ID,
				TenantID: result.User.TenantID,
				Email:    result.User.Email,
				Name:     result.User.Name,
				Role:     string(result.User.Role),
			},
			Tenants: tenantDTOs,
		},
	}, nil
}

// RefreshToken handles token refresh
func (h *FlagFlashAuthHandler) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	// First validate the token to get claims
	claims, err := auth.ValidateJWT(req.Body.Token, h.service.GetJWTSecret())
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid token", err)
	}

	// Convert FlagFlashClaims to UserClaims
	userClaims := &entity.UserClaims{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Email:    claims.Email,
		Name:     claims.Name,
		Role:     claims.Role,
	}

	result, err := h.service.RefreshToken(ctx, userClaims)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid token", err)
	}

	// Convert tenants to DTO
	tenantDTOs := make([]dto.TenantWithRoleDTO, len(result.Tenants))
	for i, t := range result.Tenants {
		tenantDTOs[i] = dto.TenantWithRoleDTO{
			ID:        t.Tenant.ID,
			Name:      t.Tenant.Name,
			Slug:      t.Tenant.Slug,
			Role:      string(t.Role),
			CreatedAt: t.Tenant.CreatedAt,
			UpdatedAt: t.Tenant.UpdatedAt,
		}
	}

	return &dto.LoginResponse{
		Body: dto.LoginResponseBody{
			Token:     result.Token,
			ExpiresAt: result.ExpiresAt,
			User: dto.UserDTO{
				ID:       result.User.ID,
				TenantID: result.User.TenantID,
				Email:    result.User.Email,
				Name:     result.User.Name,
				Role:     string(result.User.Role),
			},
			Tenants: tenantDTOs,
		},
	}, nil
}

// SwitchTenantRequestDTO represents a switch tenant request
type SwitchTenantRequestDTO struct {
	Body struct {
		TenantID uuid.UUID `json:"tenant_id" required:"true" format:"uuid"`
	}
}

// SwitchTenant handles tenant switching
func (h *FlagFlashAuthHandler) SwitchTenant(ctx context.Context, req *SwitchTenantRequestDTO) (*dto.LoginResponse, error) {
	// Get user ID from context (requires authentication)
	userIDStr := middleware.GetUserIDFromContext(ctx)
	if userIDStr == "" {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID")
	}

	result, err := h.service.SwitchTenant(ctx, userID, req.Body.TenantID)
	if err != nil {
		return nil, huma.Error403Forbidden("Cannot switch to this tenant", err)
	}

	// Convert tenants to DTO
	tenantDTOs := make([]dto.TenantWithRoleDTO, len(result.Tenants))
	for i, t := range result.Tenants {
		tenantDTOs[i] = dto.TenantWithRoleDTO{
			ID:        t.Tenant.ID,
			Name:      t.Tenant.Name,
			Slug:      t.Tenant.Slug,
			Role:      string(t.Role),
			CreatedAt: t.Tenant.CreatedAt,
			UpdatedAt: t.Tenant.UpdatedAt,
		}
	}

	return &dto.LoginResponse{
		Body: dto.LoginResponseBody{
			Token:     result.Token,
			ExpiresAt: result.ExpiresAt,
			User: dto.UserDTO{
				ID:       result.User.ID,
				TenantID: result.User.TenantID,
				Email:    result.User.Email,
				Name:     result.User.Name,
				Role:     string(result.User.Role),
			},
			Tenants: tenantDTOs,
		},
	}, nil
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequestDTO struct {
	Body struct {
		UserID      uuid.UUID `json:"user_id" required:"true" format:"uuid"`
		OldPassword string    `json:"old_password" required:"true"`
		NewPassword string    `json:"new_password" required:"true" minLength:"8"`
	}
}

// ChangePassword handles password change
func (h *FlagFlashAuthHandler) ChangePassword(ctx context.Context, req *ChangePasswordRequestDTO) (*struct{}, error) {
	if err := h.service.ChangePassword(ctx, req.Body.UserID, &service.ChangePasswordRequest{
		CurrentPassword: req.Body.OldPassword,
		NewPassword:     req.Body.NewPassword,
	}); err != nil {
		return nil, huma.Error400BadRequest("Password change failed", err)
	}

	return &struct{}{}, nil
}

// ProfileResponseDTO represents a profile response
type ProfileResponseDTO struct {
	Body dto.UserDTO
}

// GetProfile returns the current user's profile
func (h *FlagFlashAuthHandler) GetProfile(ctx context.Context, req *struct{}) (*ProfileResponseDTO, error) {
	userIDStr := middleware.GetUserIDFromContext(ctx)
	if userIDStr == "" {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID")
	}

	user, err := h.service.GetProfile(ctx, userID)
	if err != nil {
		return nil, huma.Error404NotFound("User not found", err)
	}

	return &ProfileResponseDTO{
		Body: dto.UserDTO{
			ID:       user.ID,
			TenantID: user.TenantID,
			Email:    user.Email,
			Name:     user.Name,
			Role:     string(user.Role),
		},
	}, nil
}

// UpdateProfileRequestDTO represents an update profile request
type UpdateProfileRequestDTO struct {
	Body struct {
		Name string `json:"name" minLength:"1" maxLength:"255" required:"true"`
	}
}

// UpdateProfile updates the current user's profile
func (h *FlagFlashAuthHandler) UpdateProfile(ctx context.Context, req *UpdateProfileRequestDTO) (*ProfileResponseDTO, error) {
	userIDStr := middleware.GetUserIDFromContext(ctx)
	if userIDStr == "" {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID")
	}

	user, err := h.service.UpdateProfile(ctx, userID, &service.UpdateProfileRequest{
		Name: req.Body.Name,
	})
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to update profile", err)
	}

	return &ProfileResponseDTO{
		Body: dto.UserDTO{
			ID:       user.ID,
			TenantID: user.TenantID,
			Email:    user.Email,
			Name:     user.Name,
			Role:     string(user.Role),
		},
	}, nil
}
