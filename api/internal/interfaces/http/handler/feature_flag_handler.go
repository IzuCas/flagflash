package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	"github.com/IzuCas/flagflash/internal/interfaces/http/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// FeatureFlagHandler handles feature flag related HTTP requests
type FeatureFlagHandler struct {
	service *service.FeatureFlagService
}

// NewFeatureFlagHandler creates a new feature flag handler
func NewFeatureFlagHandler(service *service.FeatureFlagService) *FeatureFlagHandler {
	return &FeatureFlagHandler{service: service}
}

// RegisterRoutes registers feature flag routes
func (h *FeatureFlagHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createFeatureFlag",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags",
		Summary:     "Create a new feature flag",
		Tags:        []string{"Feature Flags"},
	}, h.CreateFeatureFlag)

	huma.Register(api, huma.Operation{
		OperationID: "getFeatureFlag",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}",
		Summary:     "Get feature flag by ID",
		Tags:        []string{"Feature Flags"},
	}, h.GetFeatureFlag)

	huma.Register(api, huma.Operation{
		OperationID: "getFeatureFlagByKey",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/key/{key}",
		Summary:     "Get feature flag by key",
		Tags:        []string{"Feature Flags"},
	}, h.GetFeatureFlagByKey)

	huma.Register(api, huma.Operation{
		OperationID: "updateFeatureFlag",
		Method:      http.MethodPut,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}",
		Summary:     "Update feature flag",
		Tags:        []string{"Feature Flags"},
	}, h.UpdateFeatureFlag)

	huma.Register(api, huma.Operation{
		OperationID: "toggleFeatureFlag",
		Method:      http.MethodPatch,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}/toggle",
		Summary:     "Toggle feature flag enabled state",
		Tags:        []string{"Feature Flags"},
	}, h.ToggleFeatureFlag)

	huma.Register(api, huma.Operation{
		OperationID: "deleteFeatureFlag",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}",
		Summary:     "Delete feature flag",
		Tags:        []string{"Feature Flags"},
	}, h.DeleteFeatureFlag)

	huma.Register(api, huma.Operation{
		OperationID: "listFeatureFlags",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags",
		Summary:     "List feature flags for an environment",
		Tags:        []string{"Feature Flags"},
	}, h.ListFeatureFlags)

	huma.Register(api, huma.Operation{
		OperationID: "copyFeatureFlags",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{source_env_id}/flags/copy",
		Summary:     "Copy feature flags to another environment",
		Tags:        []string{"Feature Flags"},
	}, h.CopyFeatureFlags)
}

type GetFeatureFlagRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
}

type GetFeatureFlagByKeyRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	Key      string `path:"key"`
}

type DeleteFeatureFlagRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
}

type ListFeatureFlagsRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	Page     int    `query:"page" default:"1" minimum:"1"`
	Limit    int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
	Search   string `query:"search,omitempty"`
	Type     string `query:"type,omitempty" enum:"boolean,string,number,json"`
	Enabled  string `query:"enabled,omitempty" enum:"true,false"`
	Tag      string `query:"tag,omitempty"`
}

func (h *FeatureFlagHandler) CreateFeatureFlag(ctx context.Context, req *dto.CreateFeatureFlagRequest) (*dto.FeatureFlagResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	envID, err := uuid.Parse(req.EnvID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid environment ID", err)
	}

	flagType := entity.FlagTypeBoolean
	switch req.Body.Type {
	case "string":
		flagType = entity.FlagTypeString
	case "number":
		flagType = entity.FlagTypeNumber
	case "json":
		flagType = entity.FlagTypeJSON
	}

	var defaultValue json.RawMessage
	if req.Body.DefaultValue != nil {
		defaultValue, _ = json.Marshal(req.Body.DefaultValue)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	flag, err := h.service.Create(ctx, envID, req.Body.Key, req.Body.Name, req.Body.Description, flagType, defaultValue, req.Body.Tags, userID)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to create feature flag", err)
	}

	if req.Body.Enabled {
		_, _ = h.service.Toggle(ctx, flag.ID, userID)
		flag.Enabled = true
	}

	return h.buildFlagResponse(flag), nil
}

func (h *FeatureFlagHandler) GetFeatureFlag(ctx context.Context, req *GetFeatureFlagRequest) (*dto.FeatureFlagResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	flag, err := h.service.GetByID(ctx, flagID)
	if err != nil {
		return nil, huma.Error404NotFound("Feature flag not found", err)
	}

	return h.buildFlagResponse(flag), nil
}

func (h *FeatureFlagHandler) GetFeatureFlagByKey(ctx context.Context, req *GetFeatureFlagByKeyRequest) (*dto.FeatureFlagResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	envID, err := uuid.Parse(req.EnvID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid environment ID", err)
	}

	flag, err := h.service.GetByKey(ctx, envID, req.Key)
	if err != nil {
		return nil, huma.Error404NotFound("Feature flag not found", err)
	}

	return h.buildFlagResponse(flag), nil
}

func (h *FeatureFlagHandler) UpdateFeatureFlag(ctx context.Context, req *dto.UpdateFeatureFlagRequest) (*dto.FeatureFlagResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	flag, err := h.service.GetByID(ctx, flagID)
	if err != nil {
		return nil, huma.Error404NotFound("Feature flag not found", err)
	}

	name := req.Body.Name
	description := req.Body.Description
	var defaultValue json.RawMessage
	if req.Body.DefaultValue != nil {
		defaultValue, _ = json.Marshal(req.Body.DefaultValue)
	}
	var tags []string
	if req.Body.Tags != nil {
		tags = req.Body.Tags
	}

	userID := middleware.GetUserIDFromContext(ctx)
	flag, err = h.service.Update(ctx, flagID, name, description, defaultValue, tags, userID)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to update feature flag", err)
	}

	return h.buildFlagResponse(flag), nil
}

func (h *FeatureFlagHandler) ToggleFeatureFlag(ctx context.Context, req *dto.ToggleFeatureFlagRequest) (*dto.FeatureFlagResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	flag, err := h.service.GetByID(ctx, flagID)
	if err != nil {
		return nil, huma.Error404NotFound("Feature flag not found", err)
	}

	// Only toggle if current state differs from requested
	if flag.Enabled != req.Body.Enabled {
		userID := middleware.GetUserIDFromContext(ctx)
		flag, err = h.service.Toggle(ctx, flagID, userID)
		if err != nil {
			return nil, huma.Error400BadRequest("Failed to toggle feature flag", err)
		}
	}

	return h.buildFlagResponse(flag), nil
}

func (h *FeatureFlagHandler) DeleteFeatureFlag(ctx context.Context, req *DeleteFeatureFlagRequest) (*struct{}, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	if err := h.service.Delete(ctx, flagID, userID); err != nil {
		return nil, huma.Error400BadRequest("Failed to delete feature flag", err)
	}

	return &struct{}{}, nil
}

func (h *FeatureFlagHandler) ListFeatureFlags(ctx context.Context, req *ListFeatureFlagsRequest) (*dto.FeatureFlagsListResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	envID, err := uuid.Parse(req.EnvID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid environment ID", err)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (req.Page - 1) * limit

	flags, total, err := h.service.ListByEnvironmentWithPagination(ctx, envID, limit, offset, "")
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list feature flags", err)
	}

	resp := &dto.FeatureFlagsListResponse{}
	resp.Body.Pagination.Page = req.Page
	resp.Body.Pagination.Limit = limit
	resp.Body.Pagination.Total = int64(total)
	resp.Body.Pagination.TotalPages = (total + limit - 1) / limit

	resp.Body.Flags = make([]dto.FeatureFlagDTO, 0, len(flags))
	for _, flag := range flags {
		resp.Body.Flags = append(resp.Body.Flags, h.buildFlagDTO(flag))
	}

	return resp, nil
}

func (h *FeatureFlagHandler) CopyFeatureFlags(ctx context.Context, req *dto.CopyFeatureFlagsRequest) (*dto.CopyFeatureFlagsResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	sourceEnvID, err := uuid.Parse(req.SourceEnvID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid source environment ID", err)
	}

	userID := middleware.GetUserIDFromContext(ctx)
	err = h.service.CopyFlags(ctx, sourceEnvID, req.Body.TargetEnvironmentID, userID)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to copy feature flags", err)
	}

	return &dto.CopyFeatureFlagsResponse{}, nil
}

func (h *FeatureFlagHandler) buildFlagDTO(flag *entity.FeatureFlag) dto.FeatureFlagDTO {
	return dto.FeatureFlagDTO{
		ID:            flag.ID,
		EnvironmentID: flag.EnvironmentID,
		Key:           flag.Key,
		Name:          flag.Name,
		Description:   flag.Description,
		Type:          string(flag.Type),
		DefaultValue:  flag.DefaultValue,
		Enabled:       flag.Enabled,
		Version:       flag.Version,
		Tags:          flag.Tags,
		CreatedAt:     flag.CreatedAt,
		UpdatedAt:     flag.UpdatedAt,
	}
}

func (h *FeatureFlagHandler) buildFlagResponse(flag *entity.FeatureFlag) *dto.FeatureFlagResponse {
	return &dto.FeatureFlagResponse{
		Body: h.buildFlagDTO(flag),
	}
}
