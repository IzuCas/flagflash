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

// TargetingRuleHandler handles targeting rule HTTP requests
type TargetingRuleHandler struct {
	service *service.FeatureFlagService
}

// NewTargetingRuleHandler creates a new targeting rule handler
func NewTargetingRuleHandler(service *service.FeatureFlagService) *TargetingRuleHandler {
	return &TargetingRuleHandler{service: service}
}

// RegisterRoutes registers targeting rule routes
func (h *TargetingRuleHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createTargetingRule",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}/rules",
		Summary:     "Create a new targeting rule",
		Tags:        []string{"Targeting Rules"},
	}, h.CreateTargetingRule)

	huma.Register(api, huma.Operation{
		OperationID: "getTargetingRule",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}/rules/{rule_id}",
		Summary:     "Get targeting rule by ID",
		Tags:        []string{"Targeting Rules"},
	}, h.GetTargetingRule)

	huma.Register(api, huma.Operation{
		OperationID: "listTargetingRules",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}/rules",
		Summary:     "List targeting rules for a flag",
		Tags:        []string{"Targeting Rules"},
	}, h.ListTargetingRules)

	huma.Register(api, huma.Operation{
		OperationID: "updateTargetingRule",
		Method:      http.MethodPut,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}/rules/{rule_id}",
		Summary:     "Update targeting rule",
		Tags:        []string{"Targeting Rules"},
	}, h.UpdateTargetingRule)

	huma.Register(api, huma.Operation{
		OperationID: "deleteTargetingRule",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}/rules/{rule_id}",
		Summary:     "Delete targeting rule",
		Tags:        []string{"Targeting Rules"},
	}, h.DeleteTargetingRule)

	huma.Register(api, huma.Operation{
		OperationID: "reorderTargetingRules",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/applications/{app_id}/environments/{env_id}/flags/{flag_id}/rules/reorder",
		Summary:     "Reorder targeting rules",
		Tags:        []string{"Targeting Rules"},
	}, h.ReorderTargetingRules)
}

// CreateTargetingRule creates a new targeting rule
func (h *TargetingRuleHandler) CreateTargetingRule(ctx context.Context, req *dto.CreateTargetingRuleRequest) (*dto.TargetingRuleResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	// SECURITY: Only member or higher can create targeting rules
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	// Convert conditions
	conditions := make([]entity.Condition, len(req.Body.Conditions))
	for i, c := range req.Body.Conditions {
		conditions[i] = entity.Condition{
			Attribute: c.Attribute,
			Operator:  entity.Operator(c.Operator),
			Value:     c.Value,
		}
	}

	valueJSON, _ := json.Marshal(req.Body.Value)

	rule := entity.NewTargetingRule(flagID, req.Body.Name, req.Body.Priority, conditions, valueJSON, req.Body.Percentage)

	if err := h.service.CreateTargetingRule(ctx, rule); err != nil {
		return nil, huma.Error400BadRequest("Failed to create targeting rule", err)
	}

	return h.buildRuleResponse(rule), nil
}

// GetTargetingRuleRequest represents request for getting a targeting rule
type GetTargetingRuleRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	RuleID   string `path:"rule_id" format:"uuid"`
}

// GetTargetingRule retrieves a targeting rule by ID
func (h *TargetingRuleHandler) GetTargetingRule(ctx context.Context, req *GetTargetingRuleRequest) (*dto.TargetingRuleResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	ruleID, err := uuid.Parse(req.RuleID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rule ID", err)
	}

	rule, err := h.service.GetTargetingRule(ctx, ruleID)
	if err != nil {
		return nil, huma.Error404NotFound("Targeting rule not found", err)
	}

	return h.buildRuleResponse(rule), nil
}

// ListTargetingRulesRequest represents request for listing targeting rules
type ListTargetingRulesRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
}

// ListTargetingRules lists targeting rules for a flag
func (h *TargetingRuleHandler) ListTargetingRules(ctx context.Context, req *ListTargetingRulesRequest) (*dto.TargetingRulesListResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	rules, err := h.service.ListTargetingRules(ctx, flagID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list targeting rules", err)
	}

	resp := &dto.TargetingRulesListResponse{}
	resp.Body.Rules = make([]dto.TargetingRuleDTO, 0, len(rules))
	for _, rule := range rules {
		resp.Body.Rules = append(resp.Body.Rules, h.toDTO(rule))
	}

	return resp, nil
}

// UpdateTargetingRule updates a targeting rule
func (h *TargetingRuleHandler) UpdateTargetingRule(ctx context.Context, req *dto.UpdateTargetingRuleRequest) (*dto.TargetingRuleResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	// SECURITY: Only member or higher can update targeting rules
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	ruleID, err := uuid.Parse(req.RuleID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rule ID", err)
	}

	rule, err := h.service.GetTargetingRule(ctx, ruleID)
	if err != nil {
		return nil, huma.Error404NotFound("Targeting rule not found", err)
	}

	// Convert conditions if provided
	var conditions []entity.Condition
	if req.Body.Conditions != nil {
		conditions = make([]entity.Condition, len(req.Body.Conditions))
		for i, c := range req.Body.Conditions {
			conditions[i] = entity.Condition{
				Attribute: c.Attribute,
				Operator:  entity.Operator(c.Operator),
				Value:     c.Value,
			}
		}
	}

	var valueJSON json.RawMessage
	if req.Body.Value != nil {
		valueJSON, _ = json.Marshal(req.Body.Value)
	}

	rule.Update(req.Body.Name, req.Body.Priority, conditions, valueJSON, req.Body.Percentage, req.Body.Enabled)

	if err := h.service.UpdateTargetingRule(ctx, rule); err != nil {
		return nil, huma.Error500InternalServerError("Failed to update targeting rule", err)
	}

	return h.buildRuleResponse(rule), nil
}

// DeleteTargetingRuleRequest represents request for deleting a targeting rule
type DeleteTargetingRuleRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	RuleID   string `path:"rule_id" format:"uuid"`
}

// DeleteTargetingRule deletes a targeting rule
func (h *TargetingRuleHandler) DeleteTargetingRule(ctx context.Context, req *DeleteTargetingRuleRequest) (*struct{}, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	// SECURITY: Only admin or owner can delete targeting rules
	if err := middleware.RequireAdminOrOwner(ctx); err != nil {
		return nil, err
	}

	ruleID, err := uuid.Parse(req.RuleID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid rule ID", err)
	}

	if err := h.service.DeleteTargetingRule(ctx, ruleID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete targeting rule", err)
	}

	return &struct{}{}, nil
}

// ReorderTargetingRulesRequest represents request for reordering targeting rules
type ReorderTargetingRulesRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	Body     struct {
		RuleIDs []string `json:"rule_ids" required:"true"`
	}
}

// ReorderTargetingRules reorders targeting rules
func (h *TargetingRuleHandler) ReorderTargetingRules(ctx context.Context, req *ReorderTargetingRulesRequest) (*dto.TargetingRulesListResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	// SECURITY: Only member or higher can reorder targeting rules
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	flagID, err := uuid.Parse(req.FlagID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid flag ID", err)
	}

	ruleIDs := make([]uuid.UUID, len(req.Body.RuleIDs))
	for i, idStr := range req.Body.RuleIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return nil, huma.Error400BadRequest("Invalid rule ID", err)
		}
		ruleIDs[i] = id
	}

	if err := h.service.ReorderTargetingRules(ctx, flagID, ruleIDs); err != nil {
		return nil, huma.Error500InternalServerError("Failed to reorder targeting rules", err)
	}

	rules, err := h.service.ListTargetingRules(ctx, flagID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list targeting rules", err)
	}

	resp := &dto.TargetingRulesListResponse{}
	resp.Body.Rules = make([]dto.TargetingRuleDTO, 0, len(rules))
	for _, rule := range rules {
		resp.Body.Rules = append(resp.Body.Rules, h.toDTO(rule))
	}

	return resp, nil
}

func (h *TargetingRuleHandler) buildRuleResponse(rule *entity.TargetingRule) *dto.TargetingRuleResponse {
	resp := &dto.TargetingRuleResponse{}
	resp.Body = h.toDTO(rule)
	return resp
}

func (h *TargetingRuleHandler) toDTO(rule *entity.TargetingRule) dto.TargetingRuleDTO {
	conditions := make([]dto.ConditionDTO, len(rule.Conditions))
	for i, c := range rule.Conditions {
		conditions[i] = dto.ConditionDTO{
			Attribute: c.Attribute,
			Operator:  string(c.Operator),
			Value:     c.Value,
		}
	}

	var value interface{}
	json.Unmarshal(rule.Value, &value)

	return dto.TargetingRuleDTO{
		ID:            rule.ID,
		FeatureFlagID: rule.FeatureFlagID,
		Name:          rule.Name,
		Priority:      rule.Priority,
		Conditions:    conditions,
		Value:         value,
		Percentage:    rule.Percentage,
		Enabled:       rule.Enabled,
		CreatedAt:     rule.CreatedAt,
		UpdatedAt:     rule.UpdatedAt,
	}
}
