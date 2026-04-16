package handler

import (
	"context"
	"net/http"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	middleware "github.com/IzuCas/flagflash/internal/interfaces/http/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	service *service.UserService
	appURL  string
}

// NewUserHandler creates a new user handler
func NewUserHandler(service *service.UserService, appURL string) *UserHandler {
	return &UserHandler{service: service, appURL: appURL}
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "listUsers",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/users",
		Summary:     "List all users for a tenant",
		Tags:        []string{"Users"},
	}, h.ListUsers)

	huma.Register(api, huma.Operation{
		OperationID: "getUser",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/users/{user_id}",
		Summary:     "Get user by ID",
		Tags:        []string{"Users"},
	}, h.GetUser)

	huma.Register(api, huma.Operation{
		OperationID: "createUser",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/users",
		Summary:     "Create a new user",
		Tags:        []string{"Users"},
	}, h.CreateUser)

	huma.Register(api, huma.Operation{
		OperationID: "updateUser",
		Method:      http.MethodPut,
		Path:        "/tenants/{tenant_id}/users/{user_id}",
		Summary:     "Update user",
		Tags:        []string{"Users"},
	}, h.UpdateUser)

	huma.Register(api, huma.Operation{
		OperationID: "deleteUser",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/users/{user_id}",
		Summary:     "Delete user from tenant",
		Tags:        []string{"Users"},
	}, h.DeleteUser)

	huma.Register(api, huma.Operation{
		OperationID: "inviteUser",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/users/invite",
		Summary:     "Invite existing user to tenant",
		Tags:        []string{"Users"},
	}, h.InviteUser)

	huma.Register(api, huma.Operation{
		OperationID: "updateUserRole",
		Method:      http.MethodPatch,
		Path:        "/tenants/{tenant_id}/users/{user_id}/role",
		Summary:     "Update user role in tenant",
		Tags:        []string{"Users"},
	}, h.UpdateUserRole)
}

// ===== Request/Response DTOs =====

type ListUsersRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Page     int    `query:"page" default:"1" minimum:"1"`
	Limit    int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type GetUserRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	UserID   string `path:"user_id" format:"uuid"`
}

type CreateUserRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Email    string `json:"email" format:"email" required:"true"`
		Password string `json:"password" minLength:"8" required:"true"`
		Name     string `json:"name" minLength:"1" maxLength:"255" required:"true"`
		Role     string `json:"role" enum:"owner,admin,member,viewer" default:"member"`
	}
}

type UpdateUserRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	UserID   string `path:"user_id" format:"uuid"`
	Body     struct {
		Name   string `json:"name,omitempty" minLength:"1" maxLength:"255"`
		Role   string `json:"role,omitempty" enum:"owner,admin,member,viewer"`
		Active *bool  `json:"active,omitempty"`
	}
}

type DeleteUserRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	UserID   string `path:"user_id" format:"uuid"`
}

type InviteUserRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Email string `json:"email" format:"email" required:"true"`
		Role  string `json:"role" enum:"owner,admin,member,viewer" default:"member"`
	}
}

type UpdateUserRoleRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	UserID   string `path:"user_id" format:"uuid"`
	Body     struct {
		Role string `json:"role" enum:"owner,admin,member,viewer" required:"true"`
	}
}

// UserWithMembershipDTO represents a user with their membership details
type UserWithMembershipDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type UserResponse struct {
	Body UserWithMembershipDTO
}

type UsersListResponse struct {
	Body struct {
		Users      []UserWithMembershipDTO `json:"users"`
		Pagination dto.PaginationResponse  `json:"pagination"`
	}
}

type MessageResponse struct {
	Body struct {
		Message string `json:"message"`
	}
}

// ===== Handlers =====

// ListUsers lists all users for a tenant with pagination
func (h *UserHandler) ListUsers(ctx context.Context, req *ListUsersRequest) (*UsersListResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	offset := (req.Page - 1) * req.Limit
	users, total, err := h.service.ListUsersByTenant(ctx, tenantID, req.Limit, offset)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list users", err)
	}

	userDTOs := make([]UserWithMembershipDTO, len(users))
	for i, u := range users {
		role := string(u.User.Role)
		if u.Membership != nil {
			role = string(u.Membership.Role)
		}
		userDTOs[i] = UserWithMembershipDTO{
			ID:        u.User.ID,
			Email:     u.User.Email,
			Name:      u.User.Name,
			Role:      role,
			Active:    u.User.Active,
			CreatedAt: u.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: u.User.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	totalPages := (total + req.Limit - 1) / req.Limit
	return &UsersListResponse{
		Body: struct {
			Users      []UserWithMembershipDTO `json:"users"`
			Pagination dto.PaginationResponse  `json:"pagination"`
		}{
			Users: userDTOs,
			Pagination: dto.PaginationResponse{
				Page:       req.Page,
				Limit:      req.Limit,
				Total:      int64(total),
				TotalPages: totalPages,
			},
		},
	}, nil
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(ctx context.Context, req *GetUserRequest) (*UserResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID", err)
	}

	user, err := h.service.GetUserByID(ctx, userID)
	if err != nil {
		return nil, huma.Error404NotFound("User not found", err)
	}

	return &UserResponse{
		Body: UserWithMembershipDTO{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      string(user.Role),
			Active:    user.Active,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	role := entity.UserRole(req.Body.Role)
	if role == "" {
		role = entity.UserRoleMember
	}

	createReq := &service.CreateUserRequest{
		Email:    req.Body.Email,
		Password: req.Body.Password,
		Name:     req.Body.Name,
		TenantID: tenantID,
		Role:     role,
	}

	result, err := h.service.CreateUser(ctx, createReq, middleware.GetUserIDFromContext(ctx))
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create user", err)
	}

	membershipRole := string(result.User.Role)
	if result.Membership != nil {
		membershipRole = string(result.Membership.Role)
	}

	return &UserResponse{
		Body: UserWithMembershipDTO{
			ID:        result.User.ID,
			Email:     result.User.Email,
			Name:      result.User.Name,
			Role:      membershipRole,
			Active:    result.User.Active,
			CreatedAt: result.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: result.User.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UserResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID", err)
	}

	var role *entity.UserRole
	if req.Body.Role != "" {
		r := entity.UserRole(req.Body.Role)
		role = &r
	}

	updateReq := &service.UpdateUserRequest{
		Name:     req.Body.Name,
		Role:     role,
		Active:   req.Body.Active,
		TenantID: tenantID,
	}

	result, err := h.service.UpdateUser(ctx, userID, updateReq, middleware.GetUserIDFromContext(ctx))
	if err != nil {
		return nil, huma.Error403Forbidden("Failed to update user", err)
	}

	membershipRole := string(result.User.Role)
	if result.Membership != nil {
		membershipRole = string(result.Membership.Role)
	}

	return &UserResponse{
		Body: UserWithMembershipDTO{
			ID:        result.User.ID,
			Email:     result.User.Email,
			Name:      result.User.Name,
			Role:      membershipRole,
			Active:    result.User.Active,
			CreatedAt: result.User.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: result.User.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// DeleteUser removes a user from a tenant
func (h *UserHandler) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*MessageResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID", err)
	}

	if err := h.service.DeleteUser(ctx, userID, tenantID, middleware.GetUserIDFromContext(ctx)); err != nil {
		return nil, huma.Error403Forbidden("Failed to delete user", err)
	}

	return &MessageResponse{
		Body: struct {
			Message string `json:"message"`
		}{
			Message: "User removed from tenant successfully",
		},
	}, nil
}

type InviteResponseDTO struct {
	InviteID   string `json:"invite_id"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	ExpiresAt  string `json:"expires_at"`
	EmailSent  bool   `json:"email_sent"`
	InviteLink string `json:"invite_link"`
}

type InviteResponse struct {
	Body InviteResponseDTO
}

// InviteUser invites an existing user to a tenant
func (h *UserHandler) InviteUser(ctx context.Context, req *InviteUserRequest) (*InviteResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	role := entity.UserRole(req.Body.Role)
	if role == "" {
		role = entity.UserRoleMember
	}

	inviteReq := &service.InviteUserRequest{
		Email:    req.Body.Email,
		TenantID: tenantID,
		Role:     role,
	}

	result, err := h.service.InviteUserToTenant(ctx, inviteReq, middleware.GetUserIDFromContext(ctx))
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to invite user", err)
	}

	inviteLink := h.appURL + "/accept-invite?token=" + result.Token

	return &InviteResponse{
		Body: InviteResponseDTO{
			InviteID:   result.InviteID.String(),
			Email:      result.Email,
			Role:       string(result.Role),
			ExpiresAt:  result.ExpiresAt,
			EmailSent:  result.EmailSent,
			InviteLink: inviteLink,
		},
	}, nil
}

// UpdateUserRole updates a user's role in a tenant
func (h *UserHandler) UpdateUserRole(ctx context.Context, req *UpdateUserRoleRequest) (*UserResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID", err)
	}

	role := entity.UserRole(req.Body.Role)
	updateReq := &service.UpdateMembershipRequest{
		Role: &role,
	}

	membership, err := h.service.UpdateMembership(ctx, userID, tenantID, updateReq, middleware.GetUserIDFromContext(ctx))
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update user role", err)
	}

	user, err := h.service.GetUserByID(ctx, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get user", err)
	}

	return &UserResponse{
		Body: UserWithMembershipDTO{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			Role:      string(membership.Role),
			Active:    user.Active,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}
