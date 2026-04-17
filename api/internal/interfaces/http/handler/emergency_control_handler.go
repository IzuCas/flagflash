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

// EmergencyControlHandler handles emergency control HTTP requests
type EmergencyControlHandler struct {
	service *service.EmergencyControlService
}

// NewEmergencyControlHandler creates a new emergency control handler
func NewEmergencyControlHandler(service *service.EmergencyControlService) *EmergencyControlHandler {
	return &EmergencyControlHandler{service: service}
}

// RegisterRoutes registers emergency control routes
func (h *EmergencyControlHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getKillSwitchStatus",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/emergency-control/kill-switch",
		Summary:     "Get kill switch status",
		Tags:        []string{"Emergency Controls"},
	}, h.GetKillSwitchStatus)

	huma.Register(api, huma.Operation{
		OperationID: "activateKillSwitch",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/emergency-control/kill-switch",
		Summary:     "Activate kill switch (disable all flags)",
		Tags:        []string{"Emergency Controls"},
	}, h.ActivateKillSwitch)

	huma.Register(api, huma.Operation{
		OperationID: "deactivateEmergencyControl",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/emergency-control/{control_id}",
		Summary:     "Deactivate emergency control",
		Tags:        []string{"Emergency Controls"},
	}, h.DeactivateEmergencyControl)

	huma.Register(api, huma.Operation{
		OperationID: "activateMaintenanceMode",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/emergency-control/maintenance",
		Summary:     "Activate maintenance mode",
		Tags:        []string{"Emergency Controls"},
	}, h.ActivateMaintenanceMode)

	huma.Register(api, huma.Operation{
		OperationID: "listEmergencyControls",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/emergency-controls",
		Summary:     "List all emergency controls for tenant",
		Tags:        []string{"Emergency Controls"},
	}, h.ListEmergencyControls)

	huma.Register(api, huma.Operation{
		OperationID: "listActiveEmergencyControls",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/emergency-controls/active",
		Summary:     "List active emergency controls",
		Tags:        []string{"Emergency Controls"},
	}, h.ListActiveEmergencyControls)
}

type GetKillSwitchStatusRequest struct {
	TenantID      string `path:"tenant_id" format:"uuid"`
	EnvironmentID string `query:"environment_id,omitempty" format:"uuid"`
}

type ListEmergencyControlsRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
}

type ActivateKillSwitchRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		EnvironmentID   string `json:"environment_id,omitempty" format:"uuid"`
		Reason          string `json:"reason,omitempty"`
		DurationMinutes int    `json:"duration_minutes,omitempty"`
	}
}

type DeactivateEmergencyControlRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	ControlID string `path:"control_id" format:"uuid"`
}

type ActivateMaintenanceModeRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Reason          string `json:"reason,omitempty"`
		DurationMinutes int    `json:"duration_minutes,omitempty"`
	}
}

func (h *EmergencyControlHandler) GetKillSwitchStatus(ctx context.Context, req *GetKillSwitchStatusRequest) (*dto.EmergencyControlResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	var envID *uuid.UUID
	if req.EnvironmentID != "" {
		eid, err := uuid.Parse(req.EnvironmentID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid environment ID", err)
		}
		envID = &eid
	}

	control, err := h.service.GetActiveKillSwitch(ctx, tenantID, envID)
	if err != nil {
		// Return a default inactive state if no control exists
		return &dto.EmergencyControlResponse{
			Body: dto.EmergencyControlDTO{
				ID:          uuid.Nil,
				TenantID:    tenantID,
				ControlType: string(entity.EmergencyControlTypeKillSwitch),
				Enabled:     false,
			},
		}, nil
	}

	return &dto.EmergencyControlResponse{Body: toEmergencyControlDTO(control)}, nil
}

func (h *EmergencyControlHandler) ActivateKillSwitch(ctx context.Context, req *ActivateKillSwitchRequest) (*dto.EmergencyControlResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	var envID *uuid.UUID
	if req.Body.EnvironmentID != "" {
		eid, err := uuid.Parse(req.Body.EnvironmentID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid environment ID", err)
		}
		envID = &eid
	}

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	control, err := h.service.ActivateKillSwitch(ctx, tenantID, envID, req.Body.Reason, nil, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to activate kill switch", err)
	}

	return &dto.EmergencyControlResponse{Body: toEmergencyControlDTO(control)}, nil
}

func (h *EmergencyControlHandler) DeactivateEmergencyControl(ctx context.Context, req *DeactivateEmergencyControlRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}

	controlID, err := uuid.Parse(req.ControlID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid control ID", err)
	}

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	if err := h.service.Deactivate(ctx, controlID, userID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to deactivate emergency control", err)
	}

	return &struct{}{}, nil
}

func (h *EmergencyControlHandler) ActivateMaintenanceMode(ctx context.Context, req *ActivateMaintenanceModeRequest) (*dto.EmergencyControlResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	control, err := h.service.ActivateMaintenanceMode(ctx, tenantID, req.Body.Reason, nil, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to activate maintenance mode", err)
	}

	return &dto.EmergencyControlResponse{Body: toEmergencyControlDTO(control)}, nil
}

func (h *EmergencyControlHandler) ListEmergencyControls(ctx context.Context, req *ListEmergencyControlsRequest) (*dto.EmergencyControlsListResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	controls, err := h.service.ListAll(ctx, tenantID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list emergency controls", err)
	}

	controlDTOs := make([]dto.EmergencyControlDTO, len(controls))
	for i, c := range controls {
		controlDTOs[i] = toEmergencyControlDTO(c)
	}

	return &dto.EmergencyControlsListResponse{
		Body: struct {
			Controls []dto.EmergencyControlDTO `json:"controls"`
		}{
			Controls: controlDTOs,
		},
	}, nil
}

func (h *EmergencyControlHandler) ListActiveEmergencyControls(ctx context.Context, req *ListEmergencyControlsRequest) (*dto.EmergencyControlsListResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	controls, err := h.service.ListActive(ctx, tenantID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list active emergency controls", err)
	}

	controlDTOs := make([]dto.EmergencyControlDTO, len(controls))
	for i, c := range controls {
		controlDTOs[i] = toEmergencyControlDTO(c)
	}

	return &dto.EmergencyControlsListResponse{
		Body: struct {
			Controls []dto.EmergencyControlDTO `json:"controls"`
		}{
			Controls: controlDTOs,
		},
	}, nil
}

func toEmergencyControlDTO(c *entity.EmergencyControl) dto.EmergencyControlDTO {
	return dto.EmergencyControlDTO{
		ID:            c.ID,
		TenantID:      c.TenantID,
		EnvironmentID: c.EnvironmentID,
		ControlType:   string(c.ControlType),
		Enabled:       c.Enabled,
		Reason:        c.Reason,
		EnabledBy:     c.EnabledBy,
		EnabledAt:     c.EnabledAt,
		ExpiresAt:     c.ExpiresAt,
		CreatedAt:     c.CreatedAt,
	}
}
