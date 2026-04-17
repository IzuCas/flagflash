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

// ApplicationHandler handles application-related HTTP requests
type ApplicationHandler struct {
	service *service.ApplicationService
}

// NewApplicationHandler creates a new application handler
func NewApplicationHandler(service *service.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{service: service}
}

// RegisterRoutes registers application routes
func (h *ApplicationHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createApplication",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/applications",
		Summary:     "Create a new application",
		Tags:        []string{"Applications"},
	}, h.CreateApplication)

	huma.Register(api, huma.Operation{
		OperationID: "getApplication",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications/{app_id}",
		Summary:     "Get application by ID",
		Tags:        []string{"Applications"},
	}, h.GetApplication)

	huma.Register(api, huma.Operation{
		OperationID: "listApplications",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications",
		Summary:     "List applications for a tenant",
		Tags:        []string{"Applications"},
	}, h.ListApplications)

	huma.Register(api, huma.Operation{
		OperationID: "updateApplication",
		Method:      http.MethodPut,
		Path:        "/tenants/{tenant_id}/applications/{app_id}",
		Summary:     "Update application",
		Tags:        []string{"Applications"},
	}, h.UpdateApplication)

	huma.Register(api, huma.Operation{
		OperationID: "deleteApplication",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/applications/{app_id}",
		Summary:     "Delete application",
		Tags:        []string{"Applications"},
	}, h.DeleteApplication)
}

// CreateApplication creates a new application
func (h *ApplicationHandler) CreateApplication(ctx context.Context, req *dto.CreateApplicationRequest) (*dto.ApplicationResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	app, err := h.service.Create(ctx, tenantID, req.Body.Name, req.Body.Slug, req.Body.Description, userID)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to create application", err)
	}

	return &dto.ApplicationResponse{
		Body: dto.ApplicationDTO{
			ID:          app.ID,
			TenantID:    app.TenantID,
			Name:        app.Name,
			Slug:        app.Slug,
			Description: app.Description,
			CreatedAt:   app.CreatedAt,
			UpdatedAt:   app.UpdatedAt,
		},
	}, nil
}

// GetApplicationRequest represents request for getting an application
type GetApplicationRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
}

// GetApplication retrieves an application by ID
func (h *ApplicationHandler) GetApplication(ctx context.Context, req *GetApplicationRequest) (*dto.ApplicationResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	appID, err := uuid.Parse(req.AppID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid application ID", err)
	}

	app, err := h.service.GetByID(ctx, appID)
	if err != nil {
		return nil, huma.Error404NotFound("Application not found", err)
	}

	// SECURITY: Verify the application belongs to the requested tenant
	reqTenantID, _ := uuid.Parse(req.TenantID)
	if app.TenantID != reqTenantID {
		return nil, huma.Error403Forbidden("Access denied: application belongs to another tenant")
	}

	return &dto.ApplicationResponse{
		Body: dto.ApplicationDTO{
			ID:          app.ID,
			TenantID:    app.TenantID,
			Name:        app.Name,
			Slug:        app.Slug,
			Description: app.Description,
			CreatedAt:   app.CreatedAt,
			UpdatedAt:   app.UpdatedAt,
		},
	}, nil
}

// ListApplicationsRequest represents request for listing applications
type ListApplicationsRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Page     int    `query:"page" default:"1" minimum:"1"`
	Limit    int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

// ListApplications lists applications for a tenant
func (h *ApplicationHandler) ListApplications(ctx context.Context, req *ListApplicationsRequest) (*dto.ApplicationsListResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	apps, total, err := h.service.ListByTenant(ctx, tenantID, req.Limit, (req.Page-1)*req.Limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list applications", err)
	}

	resp := &dto.ApplicationsListResponse{}
	resp.Body.Pagination.Page = req.Page
	resp.Body.Pagination.Limit = req.Limit
	resp.Body.Pagination.Total = int64(total)
	resp.Body.Pagination.TotalPages = (total + req.Limit - 1) / req.Limit

	resp.Body.Applications = make([]dto.ApplicationDTO, 0, len(apps))
	for _, app := range apps {
		resp.Body.Applications = append(resp.Body.Applications, dto.ApplicationDTO{
			ID:          app.ID,
			TenantID:    app.TenantID,
			Name:        app.Name,
			Slug:        app.Slug,
			Description: app.Description,
			CreatedAt:   app.CreatedAt,
			UpdatedAt:   app.UpdatedAt,
		})
	}

	return resp, nil
}

// UpdateApplication updates an application
func (h *ApplicationHandler) UpdateApplication(ctx context.Context, req *dto.UpdateApplicationRequest) (*dto.ApplicationResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	appID, err := uuid.Parse(req.AppID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid application ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	app, err := h.service.Update(ctx, appID, req.Body.Name, req.Body.Description, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to update application", err)
	}

	return &dto.ApplicationResponse{
		Body: dto.ApplicationDTO{
			ID:          app.ID,
			TenantID:    app.TenantID,
			Name:        app.Name,
			Slug:        app.Slug,
			Description: app.Description,
			CreatedAt:   app.CreatedAt,
			UpdatedAt:   app.UpdatedAt,
		},
	}, nil
}

// DeleteApplicationRequest represents request for deleting an application
type DeleteApplicationRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
}

// DeleteApplication deletes an application
func (h *ApplicationHandler) DeleteApplication(ctx context.Context, req *DeleteApplicationRequest) (*struct{}, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	appID, err := uuid.Parse(req.AppID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid application ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	if err := h.service.Delete(ctx, appID, userID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete application", err)
	}

	return &struct{}{}, nil
}
