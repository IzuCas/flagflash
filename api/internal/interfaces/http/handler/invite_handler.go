package handler

import (
	"context"
	"net/http"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/danielgtaylor/huma/v2"
)

// InviteHandler handles public invite-related HTTP requests
type InviteHandler struct {
	userService *service.UserService
}

// NewInviteHandler creates a new invite handler
func NewInviteHandler(userService *service.UserService) *InviteHandler {
	return &InviteHandler{userService: userService}
}

// RegisterRoutes registers public invite routes
func (h *InviteHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "validateInvite",
		Method:      http.MethodGet,
		Path:        "/auth/invite/{token}",
		Summary:     "Validate an invite token",
		Tags:        []string{"Invites"},
	}, h.ValidateInvite)

	huma.Register(api, huma.Operation{
		OperationID: "acceptInvite",
		Method:      http.MethodPost,
		Path:        "/auth/invite/accept",
		Summary:     "Accept an invitation",
		Tags:        []string{"Invites"},
	}, h.AcceptInvite)
}

// === Request/Response DTOs ===

type ValidateInviteRequest struct {
	Token string `path:"token"`
}

type InviteDetailsResponse struct {
	Body struct {
		Email      string `json:"email"`
		TenantName string `json:"tenant_name"`
		Role       string `json:"role"`
		ExpiresAt  string `json:"expires_at"`
		UserExists bool   `json:"user_exists"`
	}
}

type AcceptInviteRequest struct {
	Body struct {
		Token    string `json:"token" required:"true"`
		Name     string `json:"name,omitempty"`
		Password string `json:"password,omitempty"`
	}
}

type AcceptInviteResponse struct {
	Body struct {
		Message string `json:"message"`
		Email   string `json:"email"`
	}
}

// ValidateInvite validates an invite token and returns details
func (h *InviteHandler) ValidateInvite(ctx context.Context, req *ValidateInviteRequest) (*InviteDetailsResponse, error) {
	details, err := h.userService.ValidateInviteToken(ctx, req.Token)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid invitation", err)
	}

	return &InviteDetailsResponse{
		Body: struct {
			Email      string `json:"email"`
			TenantName string `json:"tenant_name"`
			Role       string `json:"role"`
			ExpiresAt  string `json:"expires_at"`
			UserExists bool   `json:"user_exists"`
		}{
			Email:      details.Email,
			TenantName: details.TenantName,
			Role:       details.Role,
			ExpiresAt:  details.ExpiresAt,
			UserExists: details.UserExists,
		},
	}, nil
}

// AcceptInvite accepts an invitation
func (h *InviteHandler) AcceptInvite(ctx context.Context, req *AcceptInviteRequest) (*AcceptInviteResponse, error) {
	result, err := h.userService.AcceptInvite(ctx, &service.AcceptInviteRequest{
		Token:    req.Body.Token,
		Name:     req.Body.Name,
		Password: req.Body.Password,
	})
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to accept invitation", err)
	}

	return &AcceptInviteResponse{
		Body: struct {
			Message string `json:"message"`
			Email   string `json:"email"`
		}{
			Message: "Invitation accepted successfully",
			Email:   result.User.Email,
		},
	}, nil
}
