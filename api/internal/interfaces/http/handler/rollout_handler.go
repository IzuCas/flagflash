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

// RolloutHandler handles rollout plan HTTP requests
type RolloutHandler struct {
	service *service.RolloutService
}

// NewRolloutHandler creates a new rollout handler
func NewRolloutHandler(service *service.RolloutService) *RolloutHandler {
	return &RolloutHandler{service: service}
}

// RegisterRoutes registers rollout routes
func (h *RolloutHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createRolloutPlan",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/apps/{app_id}/envs/{env_id}/flags/{flag_id}/rollouts",
		Summary:     "Create a rollout plan",
		Tags:        []string{"Rollouts"},
	}, h.CreateRolloutPlan)

	huma.Register(api, huma.Operation{
		OperationID: "getRolloutPlan",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/rollouts/{rollout_id}",
		Summary:     "Get rollout plan by ID",
		Tags:        []string{"Rollouts"},
	}, h.GetRolloutPlan)

	huma.Register(api, huma.Operation{
		OperationID: "listRolloutPlans",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/apps/{app_id}/envs/{env_id}/flags/{flag_id}/rollouts",
		Summary:     "List rollout plans for a flag",
		Tags:        []string{"Rollouts"},
	}, h.ListRolloutPlans)

	huma.Register(api, huma.Operation{
		OperationID: "deleteRolloutPlan",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/rollouts/{rollout_id}",
		Summary:     "Delete rollout plan",
		Tags:        []string{"Rollouts"},
	}, h.DeleteRolloutPlan)

	huma.Register(api, huma.Operation{
		OperationID: "startRollout",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/rollouts/{rollout_id}/start",
		Summary:     "Start rollout execution",
		Tags:        []string{"Rollouts"},
	}, h.StartRollout)

	huma.Register(api, huma.Operation{
		OperationID: "pauseRollout",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/rollouts/{rollout_id}/pause",
		Summary:     "Pause rollout execution",
		Tags:        []string{"Rollouts"},
	}, h.PauseRollout)

	huma.Register(api, huma.Operation{
		OperationID: "resumeRollout",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/rollouts/{rollout_id}/resume",
		Summary:     "Resume rollout execution",
		Tags:        []string{"Rollouts"},
	}, h.ResumeRollout)

	huma.Register(api, huma.Operation{
		OperationID: "rollbackRollout",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/rollouts/{rollout_id}/rollback",
		Summary:     "Rollback rollout",
		Tags:        []string{"Rollouts"},
	}, h.RollbackRollout)

	huma.Register(api, huma.Operation{
		OperationID: "getRolloutHistory",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/rollouts/{rollout_id}/history",
		Summary:     "Get rollout execution history",
		Tags:        []string{"Rollouts"},
	}, h.GetRolloutHistory)
}

type GetRolloutPlanRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	RolloutID string `path:"rollout_id" format:"uuid"`
}

type ListRolloutPlansRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
}

type DeleteRolloutPlanRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	RolloutID string `path:"rollout_id" format:"uuid"`
}

type RolloutActionRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	RolloutID string `path:"rollout_id" format:"uuid"`
}

type RollbackRolloutRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	RolloutID string `path:"rollout_id" format:"uuid"`
	Body      struct {
		Reason string `json:"reason,omitempty"`
	}
}

type GetRolloutHistoryRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	RolloutID string `path:"rollout_id" format:"uuid"`
}

func (h *RolloutHandler) CreateRolloutPlan(ctx context.Context, req *dto.CreateRolloutRequest) (*dto.RolloutPlanResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	plan, err := h.service.Create(
		ctx,
		flagID,
		req.Body.Name,
		req.Body.TargetPercentage,
		req.Body.IncrementPercentage,
		req.Body.IncrementIntervalMinutes,
		&userID,
	)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create rollout plan", err)
	}

	return &dto.RolloutPlanResponse{Body: toRolloutPlanDTO(plan)}, nil
}

func (h *RolloutHandler) GetRolloutPlan(ctx context.Context, req *GetRolloutPlanRequest) (*dto.RolloutPlanResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	rolloutID, err := uuid.Parse(req.RolloutID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rollout ID", err)
	}

	plan, err := h.service.GetByID(ctx, rolloutID)
	if err != nil {
		return nil, huma.Error404NotFound("Rollout plan not found")
	}

	return &dto.RolloutPlanResponse{Body: toRolloutPlanDTO(plan)}, nil
}

func (h *RolloutHandler) ListRolloutPlans(ctx context.Context, req *ListRolloutPlansRequest) (*dto.RolloutPlansListResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	plans, err := h.service.ListByFlag(ctx, flagID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list rollout plans", err)
	}

	planDTOs := make([]dto.RolloutPlanDTO, len(plans))
	for i, p := range plans {
		planDTOs[i] = toRolloutPlanDTO(p)
	}

	return &dto.RolloutPlansListResponse{
		Body: struct {
			Plans []dto.RolloutPlanDTO `json:"plans"`
		}{
			Plans: planDTOs,
		},
	}, nil
}

func (h *RolloutHandler) DeleteRolloutPlan(ctx context.Context, req *DeleteRolloutPlanRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}

	rolloutID, err := uuid.Parse(req.RolloutID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rollout ID", err)
	}

	if err := h.service.Delete(ctx, rolloutID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete rollout plan", err)
	}

	return &struct{}{}, nil
}

func (h *RolloutHandler) StartRollout(ctx context.Context, req *RolloutActionRequest) (*dto.RolloutPlanResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	rolloutID, err := uuid.Parse(req.RolloutID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rollout ID", err)
	}

	if err := h.service.Start(ctx, rolloutID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to start rollout", err)
	}

	plan, _ := h.service.GetByID(ctx, rolloutID)
	return &dto.RolloutPlanResponse{Body: toRolloutPlanDTO(plan)}, nil
}

func (h *RolloutHandler) PauseRollout(ctx context.Context, req *RolloutActionRequest) (*dto.RolloutPlanResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	rolloutID, err := uuid.Parse(req.RolloutID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rollout ID", err)
	}

	if err := h.service.Pause(ctx, rolloutID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to pause rollout", err)
	}

	plan, _ := h.service.GetByID(ctx, rolloutID)
	return &dto.RolloutPlanResponse{Body: toRolloutPlanDTO(plan)}, nil
}

func (h *RolloutHandler) ResumeRollout(ctx context.Context, req *RolloutActionRequest) (*dto.RolloutPlanResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	rolloutID, err := uuid.Parse(req.RolloutID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rollout ID", err)
	}

	if err := h.service.Resume(ctx, rolloutID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to resume rollout", err)
	}

	plan, _ := h.service.GetByID(ctx, rolloutID)
	return &dto.RolloutPlanResponse{Body: toRolloutPlanDTO(plan)}, nil
}

func (h *RolloutHandler) RollbackRollout(ctx context.Context, req *RollbackRolloutRequest) (*dto.RolloutPlanResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	rolloutID, err := uuid.Parse(req.RolloutID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rollout ID", err)
	}

	if err := h.service.Rollback(ctx, rolloutID, req.Body.Reason); err != nil {
		return nil, huma.Error500InternalServerError("Failed to rollback", err)
	}

	plan, _ := h.service.GetByID(ctx, rolloutID)
	return &dto.RolloutPlanResponse{Body: toRolloutPlanDTO(plan)}, nil
}

func (h *RolloutHandler) GetRolloutHistory(ctx context.Context, req *GetRolloutHistoryRequest) (*dto.RolloutHistoryListResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	rolloutID, err := uuid.Parse(req.RolloutID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rollout ID", err)
	}

	history, err := h.service.GetHistory(ctx, rolloutID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get rollout history", err)
	}

	historyDTOs := make([]dto.RolloutHistoryDTO, len(history))
	for i, h := range history {
		historyDTOs[i] = toRolloutHistoryDTO(h)
	}

	return &dto.RolloutHistoryListResponse{
		Body: struct {
			History    []dto.RolloutHistoryDTO `json:"history"`
			Pagination dto.PaginationResponse  `json:"pagination"`
		}{
			History: historyDTOs,
			Pagination: dto.PaginationResponse{
				Page:       1,
				Limit:      len(historyDTOs),
				Total:      int64(len(historyDTOs)),
				TotalPages: 1,
			},
		},
	}, nil
}

func toRolloutPlanDTO(p *entity.RolloutPlan) dto.RolloutPlanDTO {
	return dto.RolloutPlanDTO{
		ID:                         p.ID,
		FeatureFlagID:              p.FeatureFlagID,
		Name:                       p.Name,
		Status:                     string(p.Status),
		CurrentPercentage:          p.CurrentPercentage,
		TargetPercentage:           p.TargetPercentage,
		IncrementPercentage:        p.IncrementPercentage,
		IncrementIntervalMinutes:   p.IncrementIntervalMinutes,
		AutoRollback:               p.AutoRollback,
		RollbackThresholdErrorRate: p.RollbackThresholdErrorRate,
		RollbackThresholdLatencyMs: p.RollbackThresholdLatencyMs,
		LastIncrementAt:            p.LastIncrementAt,
		NextIncrementAt:            p.NextIncrementAt,
		CreatedAt:                  p.CreatedAt,
		UpdatedAt:                  p.UpdatedAt,
	}
}

func toRolloutHistoryDTO(h *entity.RolloutHistory) dto.RolloutHistoryDTO {
	return dto.RolloutHistoryDTO{
		ID:             h.ID,
		RolloutPlanID:  h.RolloutPlanID,
		Action:         string(h.Action),
		FromPercentage: h.FromPercentage,
		ToPercentage:   h.ToPercentage,
		Reason:         h.Reason,
		CreatedAt:      h.CreatedAt,
	}
}
