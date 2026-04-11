package handler

import (
	"context"
	"net/http"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	"github.com/IzuCas/flagflash/internal/interfaces/http/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// EnvironmentHandler handles environment-related HTTP requests
type EnvironmentHandler struct {
	service *service.EnvironmentService
}

// NewEnvironmentHandler creates a new environment handler
func NewEnvironmentHandler(service *service.EnvironmentService) *EnvironmentHandler {
	return &EnvironmentHandler{service: service}
}

// RegisterRoutes registers environment routes
func (h *EnvironmentHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createEnvironment",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments",
		Summary:     "Create a new environment",
		Tags:        []string{"Environments"},
	}, h.CreateEnvironment)

	huma.Register(api, huma.Operation{
		OperationID: "getEnvironment",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}",
		Summary:     "Get environment by ID",
		Tags:        []string{"Environments"},
	}, h.GetEnvironment)

	huma.Register(api, huma.Operation{
		OperationID: "listEnvironments",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments",
		Summary:     "List environments for an application",
		Tags:        []string{"Environments"},
	}, h.ListEnvironments)

	huma.Register(api, huma.Operation{
		OperationID: "updateEnvironment",
		Method:      http.MethodPut,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}",
		Summary:     "Update environment",
		Tags:        []string{"Environments"},
	}, h.UpdateEnvironment)

	huma.Register(api, huma.Operation{
		OperationID: "deleteEnvironment",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}",
		Summary:     "Delete environment",
		Tags:        []string{"Environments"},
	}, h.DeleteEnvironment)

	huma.Register(api, huma.Operation{
		OperationID: "copyEnvironment",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/copy",
		Summary:     "Copy environment with all flags",
		Tags:        []string{"Environments"},
	}, h.CopyEnvironment)
}

// CreateEnvironment creates a new environment
func (h *EnvironmentHandler) CreateEnvironment(ctx context.Context, req *dto.CreateEnvironmentRequest) (*dto.EnvironmentResponse, error) {
	appID, err := uuid.Parse(req.AppID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid application ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	env, err := h.service.Create(ctx, appID, req.Body.Name, req.Body.Slug, req.Body.Color, false, userID)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to create environment", err)
	}

	return &dto.EnvironmentResponse{
		Body: dto.EnvironmentDTO{
			ID:            env.ID,
			ApplicationID: env.ApplicationID,
			Name:          env.Name,
			Slug:          env.Slug,
			Description:   env.Description,
			Color:         env.Color,
			CreatedAt:     env.CreatedAt,
			UpdatedAt:     env.UpdatedAt,
		},
	}, nil
}

// GetEnvironmentRequest represents request for getting an environment
type GetEnvironmentRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
}

// GetEnvironment retrieves an environment by ID
func (h *EnvironmentHandler) GetEnvironment(ctx context.Context, req *GetEnvironmentRequest) (*dto.EnvironmentResponse, error) {
	envID, err := uuid.Parse(req.EnvID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid environment ID", err)
	}

	env, err := h.service.GetByID(ctx, envID)
	if err != nil {
		return nil, huma.Error404NotFound("Environment not found", err)
	}

	return &dto.EnvironmentResponse{
		Body: dto.EnvironmentDTO{
			ID:            env.ID,
			ApplicationID: env.ApplicationID,
			Name:          env.Name,
			Slug:          env.Slug,
			Description:   env.Description,
			Color:         env.Color,
			CreatedAt:     env.CreatedAt,
			UpdatedAt:     env.UpdatedAt,
		},
	}, nil
}

// ListEnvironmentsRequest represents request for listing environments
type ListEnvironmentsRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
}

// ListEnvironments lists environments for an application
func (h *EnvironmentHandler) ListEnvironments(ctx context.Context, req *ListEnvironmentsRequest) (*dto.EnvironmentsListResponse, error) {
	appID, err := uuid.Parse(req.AppID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid application ID", err)
	}

	envs, err := h.service.ListByApplication(ctx, appID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list environments", err)
	}

	resp := &dto.EnvironmentsListResponse{}
	resp.Body.Environments = make([]dto.EnvironmentDTO, 0, len(envs))
	for _, env := range envs {
		resp.Body.Environments = append(resp.Body.Environments, dto.EnvironmentDTO{
			ID:            env.ID,
			ApplicationID: env.ApplicationID,
			Name:          env.Name,
			Slug:          env.Slug,
			Description:   env.Description,
			Color:         env.Color,
			CreatedAt:     env.CreatedAt,
			UpdatedAt:     env.UpdatedAt,
		})
	}

	return resp, nil
}

// UpdateEnvironment updates an environment
func (h *EnvironmentHandler) UpdateEnvironment(ctx context.Context, req *dto.UpdateEnvironmentRequest) (*dto.EnvironmentResponse, error) {
	envID, err := uuid.Parse(req.EnvID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid environment ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	env, err := h.service.Update(ctx, envID, req.Body.Name, req.Body.Color, nil, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update environment", err)
	}

	return &dto.EnvironmentResponse{
		Body: dto.EnvironmentDTO{
			ID:            env.ID,
			ApplicationID: env.ApplicationID,
			Name:          env.Name,
			Slug:          env.Slug,
			Description:   env.Description,
			Color:         env.Color,
			CreatedAt:     env.CreatedAt,
			UpdatedAt:     env.UpdatedAt,
		},
	}, nil
}

// DeleteEnvironmentRequest represents request for deleting an environment
type DeleteEnvironmentRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
}

// DeleteEnvironment deletes an environment
func (h *EnvironmentHandler) DeleteEnvironment(ctx context.Context, req *DeleteEnvironmentRequest) (*struct{}, error) {
	envID, err := uuid.Parse(req.EnvID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid environment ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	if err := h.service.Delete(ctx, envID, userID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete environment", err)
	}

	return &struct{}{}, nil
}

// CopyEnvironmentRequest represents request for copying an environment
type CopyEnvironmentRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	Body     struct {
		Name        string `json:"name" required:"true"`
		Slug        string `json:"slug" required:"true"`
		Description string `json:"description,omitempty"`
	}
}

// CopyEnvironment copies an environment with all its flags
func (h *EnvironmentHandler) CopyEnvironment(ctx context.Context, req *CopyEnvironmentRequest) (*dto.EnvironmentResponse, error) {
	envID, err := uuid.Parse(req.EnvID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid environment ID", err)
	}

	env, err := h.service.CopyEnvironment(ctx, envID, req.Body.Name, req.Body.Slug, req.Body.Description)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to copy environment", err)
	}

	return &dto.EnvironmentResponse{
		Body: dto.EnvironmentDTO{
			ID:            env.ID,
			ApplicationID: env.ApplicationID,
			Name:          env.Name,
			Slug:          env.Slug,
			Description:   env.Description,
			Color:         env.Color,
			CreatedAt:     env.CreatedAt,
			UpdatedAt:     env.UpdatedAt,
		},
	}, nil
}
