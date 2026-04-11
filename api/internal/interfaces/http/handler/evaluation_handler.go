package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// EvaluationHandler handles flag evaluation HTTP requests
type EvaluationHandler struct {
	service *service.EvaluationService
}

// NewEvaluationHandler creates a new evaluation handler
func NewEvaluationHandler(service *service.EvaluationService) *EvaluationHandler {
	return &EvaluationHandler{service: service}
}

// RegisterRoutes registers evaluation routes
func (h *EvaluationHandler) RegisterRoutes(api huma.API) {
	// SDK endpoints (authenticated via API Key)
	huma.Register(api, huma.Operation{
		OperationID: "evaluateFlag",
		Method:      http.MethodPost,
		Path:        "/evaluate",
		Summary:     "Evaluate a single feature flag",
		Description: "Evaluates a feature flag for the given context. Requires API Key authentication.",
		Tags:        []string{"SDK"},
	}, h.EvaluateFlag)

	huma.Register(api, huma.Operation{
		OperationID: "evaluateAllFlags",
		Method:      http.MethodPost,
		Path:        "/evaluate-all",
		Summary:     "Evaluate all feature flags",
		Description: "Evaluates all feature flags for the given context. Requires API Key authentication.",
		Tags:        []string{"SDK"},
	}, h.EvaluateAllFlags)

	huma.Register(api, huma.Operation{
		OperationID: "getFlags",
		Method:      http.MethodGet,
		Path:        "/flags",
		Summary:     "Get all feature flags",
		Description: "Returns all feature flags without evaluation. Requires API Key authentication.",
		Tags:        []string{"SDK"},
	}, h.GetFlags)
}

func (h *EvaluationHandler) EvaluateFlag(ctx context.Context, req *dto.EvaluateFlagRequest) (*dto.EvaluateFlagResponse, error) {
	// Get environment ID from context (set by API key middleware)
	environmentID, ok := ctx.Value("environment_id").(uuid.UUID)
	if !ok {
		return nil, huma.Error401Unauthorized("Invalid API key")
	}

	evalCtx := &entity.EvaluationContext{
		Custom: req.Body.Context,
	}

	result, err := h.service.EvaluateFlag(ctx, environmentID, req.Body.FlagKey, evalCtx)
	if err != nil {
		return nil, huma.Error404NotFound("Flag not found or evaluation failed", err)
	}

	resp := &dto.EvaluateFlagResponse{}
	resp.Body.FlagKey = result.Key
	resp.Body.Value = result.Value
	resp.Body.Enabled = result.Enabled
	resp.Body.Version = result.Version
	if result.RuleID != nil {
		resp.Body.RuleID = result.RuleID
		resp.Body.RuleName = result.RuleName
	}
	return resp, nil
}

func (h *EvaluationHandler) EvaluateAllFlags(ctx context.Context, req *dto.EvaluateAllFlagsRequest) (*dto.EvaluateAllFlagsResponse, error) {
	// Get environment ID from context (set by API key middleware)
	environmentID, ok := ctx.Value("environment_id").(uuid.UUID)
	if !ok {
		return nil, huma.Error401Unauthorized("Invalid API key")
	}

	evalCtx := &entity.EvaluationContext{
		Custom: req.Body.Context,
	}

	results, err := h.service.EvaluateAllFlags(ctx, environmentID, evalCtx)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to evaluate flags", err)
	}

	resp := &dto.EvaluateAllFlagsResponse{}
	resp.Body.Flags = make(map[string]dto.EvaluatedFlag)
	for _, result := range results {
		eval := dto.EvaluatedFlag{
			Value:   result.Value,
			Enabled: result.Enabled,
			Version: result.Version,
		}
		if result.RuleID != nil {
			eval.RuleID = result.RuleID
			eval.RuleName = result.RuleName
		}
		resp.Body.Flags[result.Key] = eval
	}
	return resp, nil
}

type sdkFlagItem struct {
	Key          string          `json:"key"`
	Type         string          `json:"type"`
	Enabled      bool            `json:"enabled"`
	DefaultValue json.RawMessage `json:"default_value"`
	Version      int             `json:"version"`
}

type GetFlagsResponse struct {
	Body struct {
		Flags []sdkFlagItem `json:"flags"`
	}
}

func (h *EvaluationHandler) GetFlags(ctx context.Context, _ *struct{}) (*GetFlagsResponse, error) {
	// Get environment ID from context (set by API key middleware)
	ctxNew := context.Background()

	environmentID, ok := ctx.Value("environment_id").(uuid.UUID)
	if !ok {
		return nil, huma.Error401Unauthorized("Invalid API key")
	}

	flags, err := h.service.GetAllFlags(ctxNew, environmentID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get flags", err)
	}

	resp := &GetFlagsResponse{}
	resp.Body.Flags = make([]sdkFlagItem, 0, len(flags))
	for _, f := range flags {
		resp.Body.Flags = append(resp.Body.Flags, sdkFlagItem{
			Key:          f.Key,
			Type:         f.Type,
			Enabled:      f.Enabled,
			DefaultValue: f.DefaultValue,
			Version:      f.Version,
		})
	}
	return resp, nil
}
