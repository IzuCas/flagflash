package handler

import (
	"context"
	"net/http"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	"github.com/IzuCas/flagflash/internal/interfaces/http/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// FlagHistoryHandler handles flag history HTTP requests
type FlagHistoryHandler struct {
	service *service.FlagHistoryService
}

// NewFlagHistoryHandler creates a new flag history handler
func NewFlagHistoryHandler(service *service.FlagHistoryService) *FlagHistoryHandler {
	return &FlagHistoryHandler{service: service}
}

// RegisterRoutes registers flag history routes
func (h *FlagHistoryHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "listFlagHistory",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/apps/{app_id}/envs/{env_id}/flags/{flag_id}/history",
		Summary:     "Get history for a feature flag",
		Tags:        []string{"Flag History"},
	}, h.ListFlagHistory)

	huma.Register(api, huma.Operation{
		OperationID: "getFlagHistoryVersion",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/apps/{app_id}/envs/{env_id}/flags/{flag_id}/history/version/{version}",
		Summary:     "Get a specific version of a flag",
		Tags:        []string{"Flag History"},
	}, h.GetHistoryVersion)

	huma.Register(api, huma.Operation{
		OperationID: "compareFlagVersions",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/apps/{app_id}/envs/{env_id}/flags/{flag_id}/history/compare",
		Summary:     "Compare two flag versions",
		Tags:        []string{"Flag History"},
	}, h.CompareVersions)
}

type ListFlagHistoryRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	Page     int    `query:"page" default:"1" minimum:"1"`
	Limit    int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type GetHistoryVersionRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	Version  int    `path:"version" minimum:"1"`
}

type CompareVersionsRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	Version1 int    `query:"version1" minimum:"1"`
	Version2 int    `query:"version2" minimum:"1"`
}

type FieldDifferenceDTO struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Type     string      `json:"type"`
}

type CompareVersionsResponse struct {
	Body struct {
		Version1    dto.FlagHistoryDTO   `json:"version1"`
		Version2    dto.FlagHistoryDTO   `json:"version2"`
		Differences []FieldDifferenceDTO `json:"differences,omitempty"`
	}
}

func (h *FlagHistoryHandler) ListFlagHistory(ctx context.Context, req *ListFlagHistoryRequest) (*dto.FlagHistoryListResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	offset := (req.Page - 1) * req.Limit
	history, total, err := h.service.GetHistory(ctx, flagID, req.Limit, offset)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list flag history", err)
	}

	historyDTOs := make([]dto.FlagHistoryDTO, len(history))
	for i, h := range history {
		historyDTOs[i] = toFlagHistoryDTO(h)
	}

	totalPages := (total + req.Limit - 1) / req.Limit

	return &dto.FlagHistoryListResponse{
		Body: struct {
			History    []dto.FlagHistoryDTO   `json:"history"`
			Pagination dto.PaginationResponse `json:"pagination"`
		}{
			History: historyDTOs,
			Pagination: dto.PaginationResponse{
				Page:       req.Page,
				Limit:      req.Limit,
				Total:      int64(total),
				TotalPages: totalPages,
			},
		},
	}, nil
}

func (h *FlagHistoryHandler) GetHistoryVersion(ctx context.Context, req *GetHistoryVersionRequest) (*dto.FlagHistoryResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	entry, err := h.service.GetHistoryByVersion(ctx, flagID, req.Version)
	if err != nil {
		return nil, huma.Error404NotFound("History version not found")
	}

	return &dto.FlagHistoryResponse{Body: toFlagHistoryDTO(entry)}, nil
}

func (h *FlagHistoryHandler) CompareVersions(ctx context.Context, req *CompareVersionsRequest) (*CompareVersionsResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	comparison, err := h.service.Compare(ctx, flagID, req.Version1, req.Version2)
	if err != nil {
		return nil, huma.Error404NotFound("Could not compare versions")
	}

	v1, _ := h.service.GetHistoryByVersion(ctx, flagID, req.Version1)
	v2, _ := h.service.GetHistoryByVersion(ctx, flagID, req.Version2)

	differences := make([]FieldDifferenceDTO, len(comparison.Differences))
	for i, d := range comparison.Differences {
		differences[i] = FieldDifferenceDTO{
			Field:    d.Field,
			OldValue: d.OldValue,
			NewValue: d.NewValue,
			Type:     d.Type,
		}
	}

	return &CompareVersionsResponse{
		Body: struct {
			Version1    dto.FlagHistoryDTO   `json:"version1"`
			Version2    dto.FlagHistoryDTO   `json:"version2"`
			Differences []FieldDifferenceDTO `json:"differences,omitempty"`
		}{
			Version1:    toFlagHistoryDTO(v1),
			Version2:    toFlagHistoryDTO(v2),
			Differences: differences,
		},
	}, nil
}

func toFlagHistoryDTO(h *entity.FlagHistory) dto.FlagHistoryDTO {
	var changedByStr *string
	if h.ChangedBy != nil {
		str := h.ChangedBy.String()
		changedByStr = &str
	}

	return dto.FlagHistoryDTO{
		ID:            h.ID,
		FeatureFlagID: h.FeatureFlagID,
		Version:       h.Version,
		ChangeType:    string(h.ChangeType),
		ChangedBy:     changedByStr,
		ChangedByName: h.ChangedByName,
		Comment:       h.Comment,
		CreatedAt:     h.CreatedAt,
	}
}
