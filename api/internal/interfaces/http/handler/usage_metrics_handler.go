package handler

import (
	"context"
	"time"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	"github.com/IzuCas/flagflash/internal/interfaces/http/middleware"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// UsageMetricsHandler handles usage metrics HTTP requests
type UsageMetricsHandler struct {
	service *service.UsageMetricsService
}

// NewUsageMetricsHandler creates a new usage metrics handler
func NewUsageMetricsHandler(service *service.UsageMetricsService) *UsageMetricsHandler {
	return &UsageMetricsHandler{service: service}
}

// RegisterRoutes registers usage metrics routes
func (h *UsageMetricsHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getUsageMetrics",
		Method:      "GET",
		Path:        "/tenants/{tenant_id}/usage-metrics",
		Summary:     "Get usage metrics summary",
		Description: "Returns aggregated usage metrics for a tenant",
		Tags:        []string{"Usage Metrics"},
	}, h.GetUsageMetrics)

	huma.Register(api, huma.Operation{
		OperationID: "getTimeline",
		Method:      "GET",
		Path:        "/tenants/{tenant_id}/usage-metrics/timeline",
		Summary:     "Get usage timeline",
		Description: "Returns time-series data for flag evaluations",
		Tags:        []string{"Usage Metrics"},
	}, h.GetTimeline)

	huma.Register(api, huma.Operation{
		OperationID: "getFlagMetrics",
		Method:      "GET",
		Path:        "/tenants/{tenant_id}/usage-metrics/flags",
		Summary:     "Get metrics by flag",
		Description: "Returns usage metrics broken down by feature flag",
		Tags:        []string{"Usage Metrics"},
	}, h.GetFlagMetrics)

	huma.Register(api, huma.Operation{
		OperationID: "getEnvironmentMetrics",
		Method:      "GET",
		Path:        "/tenants/{tenant_id}/usage-metrics/environments",
		Summary:     "Get metrics by environment",
		Description: "Returns usage metrics broken down by environment",
		Tags:        []string{"Usage Metrics"},
	}, h.GetEnvironmentMetrics)

	huma.Register(api, huma.Operation{
		OperationID: "recordEvaluation",
		Method:      "POST",
		Path:        "/tenants/{tenant_id}/evaluations",
		Summary:     "Record evaluation event",
		Description: "Records a single flag evaluation event",
		Tags:        []string{"Usage Metrics"},
	}, h.RecordEvaluation)

	huma.Register(api, huma.Operation{
		OperationID: "recordEvaluationBatch",
		Method:      "POST",
		Path:        "/tenants/{tenant_id}/evaluations/batch",
		Summary:     "Record evaluation events batch",
		Description: "Records multiple flag evaluation events in a batch",
		Tags:        []string{"Usage Metrics"},
	}, h.RecordEvaluationBatch)
}

// GetUsageMetrics returns aggregated usage metrics
func (h *UsageMetricsHandler) GetUsageMetrics(ctx context.Context, req *dto.UsageMetricsRequest) (*dto.UsageMetricsResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID")
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid start_date format, use RFC3339")
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid end_date format, use RFC3339")
	}

	var environmentID *uuid.UUID
	if req.EnvironmentID != "" {
		envID, err := uuid.Parse(req.EnvironmentID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid environment_id")
		}
		environmentID = &envID
	}

	var flagID *uuid.UUID
	if req.FlagID != "" {
		fID, err := uuid.Parse(req.FlagID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid flag_id")
		}
		flagID = &fID
	}

	granularity := req.Granularity
	if granularity == "" {
		granularity = "hour"
	}

	metrics, err := h.service.GetMetricsSummary(ctx, tenantID, environmentID, flagID, startDate, endDate, granularity)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get usage metrics", err)
	}

	return &dto.UsageMetricsResponse{
		Body: metricsToDTO(metrics),
	}, nil
}

// GetTimeline returns time-series data
func (h *UsageMetricsHandler) GetTimeline(ctx context.Context, req *dto.UsageMetricsRequest) (*dto.TimelineResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID")
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid start_date format, use RFC3339")
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid end_date format, use RFC3339")
	}

	var environmentID *uuid.UUID
	if req.EnvironmentID != "" {
		envID, err := uuid.Parse(req.EnvironmentID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid environment_id")
		}
		environmentID = &envID
	}

	var flagID *uuid.UUID
	if req.FlagID != "" {
		fID, err := uuid.Parse(req.FlagID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid flag_id")
		}
		flagID = &fID
	}

	granularity := req.Granularity
	if granularity == "" {
		granularity = "hour"
	}

	timeline, err := h.service.GetTimeline(ctx, tenantID, environmentID, flagID, startDate, endDate, granularity)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get timeline", err)
	}

	return &dto.TimelineResponse{
		Body: struct {
			Timeline []dto.TimelinePointDTO `json:"timeline"`
		}{
			Timeline: timelineToDTO(timeline),
		},
	}, nil
}

// GetFlagMetrics returns metrics by flag
func (h *UsageMetricsHandler) GetFlagMetrics(ctx context.Context, req *dto.UsageMetricsRequest) (*dto.FlagMetricsResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID")
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid start_date format, use RFC3339")
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid end_date format, use RFC3339")
	}

	var environmentID *uuid.UUID
	if req.EnvironmentID != "" {
		envID, err := uuid.Parse(req.EnvironmentID)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid environment_id")
		}
		environmentID = &envID
	}

	flags, err := h.service.GetFlagMetrics(ctx, tenantID, environmentID, startDate, endDate)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get flag metrics", err)
	}

	return &dto.FlagMetricsResponse{
		Body: struct {
			Flags []dto.FlagMetricDTO `json:"flags"`
		}{
			Flags: flagMetricsToDTO(flags),
		},
	}, nil
}

// GetEnvironmentMetrics returns metrics by environment
func (h *UsageMetricsHandler) GetEnvironmentMetrics(ctx context.Context, req *struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	StartDate string `query:"start_date"`
	EndDate   string `query:"end_date"`
}) (*dto.EnvironmentMetricsResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID")
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid start_date format, use RFC3339")
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid end_date format, use RFC3339")
	}

	envs, err := h.service.GetEnvironmentMetrics(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get environment metrics", err)
	}

	return &dto.EnvironmentMetricsResponse{
		Body: struct {
			Environments []dto.EnvironmentMetricDTO `json:"environments"`
		}{
			Environments: envMetricsToDTO(envs),
		},
	}, nil
}

// RecordEvaluation records a single evaluation event
func (h *UsageMetricsHandler) RecordEvaluation(ctx context.Context, req *struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		EnvironmentID uuid.UUID              `json:"environment_id" format:"uuid" required:"true"`
		FlagID        uuid.UUID              `json:"flag_id" format:"uuid" required:"true"`
		FlagKey       string                 `json:"flag_key" required:"true"`
		Value         interface{}            `json:"value" required:"true"`
		UserID        *string                `json:"user_id,omitempty"`
		Context       map[string]interface{} `json:"context,omitempty"`
		SDKType       *string                `json:"sdk_type,omitempty"`
		SDKVersion    *string                `json:"sdk_version,omitempty"`
	}
}) (*struct{}, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID")
	}

	event := &entity.EvaluationEvent{
		TenantID:      tenantID,
		EnvironmentID: req.Body.EnvironmentID,
		FeatureFlagID: req.Body.FlagID,
		FlagKey:       req.Body.FlagKey,
		Value:         req.Body.Value,
		UserID:        req.Body.UserID,
		Context:       req.Body.Context,
		SDKType:       req.Body.SDKType,
		SDKVersion:    req.Body.SDKVersion,
		EvaluatedAt:   time.Now(),
	}

	err = h.service.RecordEvaluation(ctx, event)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to record evaluation", err)
	}

	return &struct{}{}, nil
}

// RecordEvaluationBatch records multiple evaluation events
func (h *UsageMetricsHandler) RecordEvaluationBatch(ctx context.Context, req *struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Events []struct {
			EnvironmentID uuid.UUID              `json:"environment_id" format:"uuid" required:"true"`
			FlagID        uuid.UUID              `json:"flag_id" format:"uuid" required:"true"`
			FlagKey       string                 `json:"flag_key" required:"true"`
			Value         interface{}            `json:"value" required:"true"`
			UserID        *string                `json:"user_id,omitempty"`
			Context       map[string]interface{} `json:"context,omitempty"`
			SDKType       *string                `json:"sdk_type,omitempty"`
			SDKVersion    *string                `json:"sdk_version,omitempty"`
			EvaluatedAt   *time.Time             `json:"evaluated_at,omitempty"`
		} `json:"events" required:"true"`
	}
}) (*struct{}, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID")
	}

	var events []*entity.EvaluationEvent
	for _, e := range req.Body.Events {
		evaluatedAt := time.Now()
		if e.EvaluatedAt != nil {
			evaluatedAt = *e.EvaluatedAt
		}

		events = append(events, &entity.EvaluationEvent{
			TenantID:      tenantID,
			EnvironmentID: e.EnvironmentID,
			FeatureFlagID: e.FlagID,
			FlagKey:       e.FlagKey,
			Value:         e.Value,
			UserID:        e.UserID,
			Context:       e.Context,
			SDKType:       e.SDKType,
			SDKVersion:    e.SDKVersion,
			EvaluatedAt:   evaluatedAt,
		})
	}

	err = h.service.RecordEvaluationBatch(ctx, events)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to record evaluations", err)
	}

	return &struct{}{}, nil
}

// Helper functions
func metricsToDTO(m *entity.UsageMetrics) dto.UsageMetricsDTO {
	return dto.UsageMetricsDTO{
		TenantID:         m.TenantID,
		Period:           m.Period,
		StartDate:        m.StartDate,
		EndDate:          m.EndDate,
		TotalEvaluations: m.TotalEvaluations,
		UniqueFlags:      m.UniqueFlags,
		UniqueUsers:      m.UniqueUsers,
		ByEnvironment:    envMetricsToDTO(m.ByEnvironment),
		ByFlag:           flagMetricsToDTO(m.ByFlag),
		Timeline:         timelineToDTO(m.Timeline),
	}
}

func envMetricsToDTO(envs []entity.EnvironmentMetrics) []dto.EnvironmentMetricDTO {
	result := make([]dto.EnvironmentMetricDTO, len(envs))
	for i, e := range envs {
		result[i] = dto.EnvironmentMetricDTO{
			EnvironmentID:   e.EnvironmentID,
			EnvironmentName: e.EnvironmentName,
			Evaluations:     e.Evaluations,
			UniqueFlags:     e.UniqueFlags,
			UniqueUsers:     e.UniqueUsers,
		}
	}
	return result
}

func flagMetricsToDTO(flags []entity.FlagMetrics) []dto.FlagMetricDTO {
	result := make([]dto.FlagMetricDTO, len(flags))
	for i, f := range flags {
		result[i] = dto.FlagMetricDTO{
			FlagID:          f.FlagID,
			FlagKey:         f.FlagKey,
			FlagName:        f.FlagName,
			EnvironmentID:   f.EnvironmentID,
			EnvironmentName: f.EnvironmentName,
			Evaluations:     f.Evaluations,
			TrueCount:       f.TrueCount,
			FalseCount:      f.FalseCount,
			UniqueUsers:     f.UniqueUsers,
		}
	}
	return result
}

func timelineToDTO(timeline []entity.TimelinePoint) []dto.TimelinePointDTO {
	result := make([]dto.TimelinePointDTO, len(timeline))
	for i, t := range timeline {
		result[i] = dto.TimelinePointDTO{
			Timestamp:   t.Timestamp,
			Evaluations: t.Evaluations,
			TrueCount:   t.TrueCount,
			FalseCount:  t.FalseCount,
		}
	}
	return result
}
