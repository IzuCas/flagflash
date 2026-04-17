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

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	service *service.TenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(service *service.TenantService) *TenantHandler {
	return &TenantHandler{service: service}
}

// RegisterRoutes registers tenant routes
func (h *TenantHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "listTenants",
		Method:      http.MethodGet,
		Path:        "/tenants",
		Summary:     "List all tenants",
		Tags:        []string{"Tenants"},
	}, h.ListTenants)

	huma.Register(api, huma.Operation{
		OperationID: "listMyTenants",
		Method:      http.MethodGet,
		Path:        "/tenants/me",
		Summary:     "List tenants for current user",
		Tags:        []string{"Tenants"},
	}, h.ListMyTenants)

	huma.Register(api, huma.Operation{
		OperationID: "getTenant",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}",
		Summary:     "Get tenant by ID",
		Tags:        []string{"Tenants"},
	}, h.GetTenant)

	huma.Register(api, huma.Operation{
		OperationID: "createTenant",
		Method:      http.MethodPost,
		Path:        "/tenants",
		Summary:     "Create a new tenant",
		Tags:        []string{"Tenants"},
	}, h.CreateTenant)

	huma.Register(api, huma.Operation{
		OperationID: "updateTenant",
		Method:      http.MethodPut,
		Path:        "/tenants/{tenant_id}",
		Summary:     "Update tenant",
		Tags:        []string{"Tenants"},
	}, h.UpdateTenant)

	huma.Register(api, huma.Operation{
		OperationID: "deleteTenant",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}",
		Summary:     "Delete tenant",
		Tags:        []string{"Tenants"},
	}, h.DeleteTenant)

	huma.Register(api, huma.Operation{
		OperationID: "getTenantStats",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/stats",
		Summary:     "Get tenant statistics",
		Tags:        []string{"Tenants"},
	}, h.GetTenantStats)
}

// GetTenantRequest represents request for getting a tenant
type GetTenantRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
}

// ListTenantsRequest represents request for listing tenants
type ListTenantsRequest struct {
	Page  int `query:"page" default:"1" minimum:"1"`
	Limit int `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

// ListTenants lists tenants accessible to the current user
// SECURITY: Removed ability to list ALL tenants - now redirects to ListMyTenants
func (h *TenantHandler) ListTenants(ctx context.Context, req *ListTenantsRequest) (*dto.TenantsListResponse, error) {
	// SECURITY FIX: Users should only see their own tenants, not all tenants in the system
	// Redirect to user-specific tenant list
	userIDStr := middleware.GetUserIDFromContext(ctx)
	if userIDStr == "" {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID")
	}

	tenants, err := h.service.ListByUser(ctx, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list tenants", err)
	}

	tenantDTOs := make([]dto.TenantDTO, len(tenants))
	for i, t := range tenants {
		tenantDTOs[i] = dto.TenantDTO{
			ID:        t.Tenant.ID,
			Name:      t.Tenant.Name,
			Slug:      t.Tenant.Slug,
			CreatedAt: t.Tenant.CreatedAt,
			UpdatedAt: t.Tenant.UpdatedAt,
		}
	}

	return &dto.TenantsListResponse{
		Body: struct {
			Tenants    []dto.TenantDTO        `json:"tenants"`
			Pagination dto.PaginationResponse `json:"pagination"`
		}{
			Tenants: tenantDTOs,
			Pagination: dto.PaginationResponse{
				Page:       1,
				Limit:      len(tenantDTOs),
				Total:      int64(len(tenantDTOs)),
				TotalPages: 1,
			},
		},
	}, nil
}

// ListMyTenantsRequest represents request for listing user's tenants
type ListMyTenantsRequest struct{}

// MyTenantsListResponse represents the response for listing user's tenants
type MyTenantsListResponse struct {
	Body struct {
		Tenants []dto.TenantWithRoleDTO `json:"tenants"`
	}
}

// ListMyTenants lists all tenants that the current user has access to
func (h *TenantHandler) ListMyTenants(ctx context.Context, req *ListMyTenantsRequest) (*MyTenantsListResponse, error) {
	userIDStr := middleware.GetUserIDFromContext(ctx)
	if userIDStr == "" {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid user ID")
	}

	tenants, err := h.service.ListByUser(ctx, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list tenants", err)
	}

	tenantDTOs := make([]dto.TenantWithRoleDTO, len(tenants))
	for i, t := range tenants {
		tenantDTOs[i] = dto.TenantWithRoleDTO{
			ID:        t.Tenant.ID,
			Name:      t.Tenant.Name,
			Slug:      t.Tenant.Slug,
			Role:      string(t.Role),
			CreatedAt: t.Tenant.CreatedAt,
			UpdatedAt: t.Tenant.UpdatedAt,
		}
	}

	return &MyTenantsListResponse{
		Body: struct {
			Tenants []dto.TenantWithRoleDTO `json:"tenants"`
		}{
			Tenants: tenantDTOs,
		},
	}, nil
}

// CreateTenant creates a new tenant
func (h *TenantHandler) CreateTenant(ctx context.Context, req *dto.CreateTenantRequest) (*dto.TenantResponse, error) {
	userID := middleware.GetUserIDFromContext(ctx)

	tenant, err := h.service.Create(ctx, req.Body.Name, req.Body.Slug, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create tenant", err)
	}

	return &dto.TenantResponse{
		Body: dto.TenantDTO{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Slug:      tenant.Slug,
			CreatedAt: tenant.CreatedAt,
			UpdatedAt: tenant.UpdatedAt,
		},
	}, nil
}

// GetTenant retrieves a tenant by ID
func (h *TenantHandler) GetTenant(ctx context.Context, req *GetTenantRequest) (*dto.TenantResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	tenant, err := h.service.GetByID(ctx, id)
	if err != nil {
		return nil, huma.Error404NotFound("Tenant not found", err)
	}

	return &dto.TenantResponse{
		Body: dto.TenantDTO{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Slug:      tenant.Slug,
			CreatedAt: tenant.CreatedAt,
			UpdatedAt: tenant.UpdatedAt,
		},
	}, nil
}

// UpdateTenant updates a tenant
func (h *TenantHandler) UpdateTenant(ctx context.Context, req *dto.UpdateTenantRequest) (*dto.TenantResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)

	tenant, err := h.service.Update(ctx, id, req.Body.Name, nil, userID)
	if err != nil {
		if err.Error() == "access denied: only owners can update tenant settings" || err.Error() == "access denied: not a member of this tenant" {
			return nil, huma.Error403Forbidden(err.Error())
		}
		return nil, huma.Error500InternalServerError("Failed to update tenant", err)
	}

	return &dto.TenantResponse{
		Body: dto.TenantDTO{
			ID:        tenant.ID,
			Name:      tenant.Name,
			Slug:      tenant.Slug,
			CreatedAt: tenant.CreatedAt,
			UpdatedAt: tenant.UpdatedAt,
		},
	}, nil
}

// DeleteTenantRequest represents request for deleting a tenant
type DeleteTenantRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
}

// DeleteTenant deletes a tenant
func (h *TenantHandler) DeleteTenant(ctx context.Context, req *DeleteTenantRequest) (*struct{}, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)

	if err := h.service.Delete(ctx, id, userID); err != nil {
		if err.Error() == "access denied: only owners can delete tenants" || err.Error() == "access denied: not a member of this tenant" {
			return nil, huma.Error403Forbidden(err.Error())
		}
		return nil, huma.Error500InternalServerError("Failed to delete tenant", err)
	}

	return &struct{}{}, nil
}

// GetTenantStatsRequest represents request for tenant stats
type GetTenantStatsRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
}

// TenantStatsResponse represents tenant statistics response
type TenantStatsResponse struct {
	Body TenantStats
}

// TenantStats contains tenant statistics
type TenantStats struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	Applications int       `json:"applications"`
	Environments int       `json:"environments"`
	Flags        int       `json:"flags"`
	Users        int       `json:"users"`
	APIKeys      int       `json:"api_keys"`
}

// GetTenantStats retrieves tenant statistics
func (h *TenantHandler) GetTenantStats(ctx context.Context, req *GetTenantStatsRequest) (*TenantStatsResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	// TODO: Implement GetTenantStats in TenantService
	return &TenantStatsResponse{
		Body: TenantStats{
			TenantID:     id,
			Applications: 0,
			Environments: 0,
			Flags:        0,
			Users:        0,
			APIKeys:      0,
		},
	}, nil
}
