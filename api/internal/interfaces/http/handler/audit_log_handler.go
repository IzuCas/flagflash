package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	"github.com/IzuCas/flagflash/internal/interfaces/http/middleware"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// AuditLogHandler handles audit log related HTTP requests
type AuditLogHandler struct {
	service  *service.AuditLogService
	userRepo repository.UserRepository
}

// NewAuditLogHandler creates a new audit log handler
func NewAuditLogHandler(service *service.AuditLogService, userRepo repository.UserRepository) *AuditLogHandler {
	return &AuditLogHandler{service: service, userRepo: userRepo}
}

// RegisterRoutes registers audit log routes
func (h *AuditLogHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "listAuditLogs",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/audit-logs",
		Summary:     "List audit logs",
		Description: "Retrieve audit logs with optional filtering by entity type, action, date range",
		Tags:        []string{"Audit Logs"},
	}, h.ListAuditLogs)

	huma.Register(api, huma.Operation{
		OperationID: "getAuditLog",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/audit-logs/{log_id}",
		Summary:     "Get audit log by ID",
		Tags:        []string{"Audit Logs"},
	}, h.GetAuditLog)
}

// GetAuditLogRequest represents the request for getting a single audit log
type GetAuditLogRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	LogID    string `path:"log_id" format:"uuid"`
}

// ListAuditLogs retrieves audit logs with filtering
func (h *AuditLogHandler) ListAuditLogs(ctx context.Context, req *dto.AuditLogsListRequest) (*dto.AuditLogsListResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	// Parse optional filters
	var entityType *string
	if req.EntityType != "" {
		entityType = &req.EntityType
	}

	var entityID *uuid.UUID
	if req.EntityID != "" {
		id, err := uuid.Parse(req.EntityID)
		if err == nil {
			entityID = &id
		}
	}

	var action *string
	if req.Action != "" {
		action = &req.Action
	}

	var actorID *string
	if req.ActorID != "" {
		actorID = &req.ActorID
	}

	var startDate, endDate *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, req.StartDate); err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, req.EndDate); err == nil {
			endDate = &t
		}
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}

	logs, total, err := h.service.List(ctx, tenantID, entityType, entityID, action, actorID, startDate, endDate, req.Page, limit)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to retrieve audit logs", err)
	}

	// Collect unique user IDs to fetch names
	userIDs := make(map[uuid.UUID]bool)
	for _, log := range logs {
		if log.ActorType == entity.ActorTypeUser && log.ActorID != "" {
			if id, err := uuid.Parse(log.ActorID); err == nil {
				userIDs[id] = true
			}
		}
	}

	// Fetch user names
	userNames := make(map[string]string)
	if len(userIDs) > 0 {
		ids := make([]uuid.UUID, 0, len(userIDs))
		for id := range userIDs {
			ids = append(ids, id)
		}
		users, err := h.userRepo.GetByIDs(ctx, ids)
		if err == nil {
			for _, user := range users {
				userNames[user.ID.String()] = user.Name
			}
		}
	}

	resp := &dto.AuditLogsListResponse{}
	resp.Body.Pagination.Page = req.Page
	resp.Body.Pagination.Limit = limit
	resp.Body.Pagination.Total = int64(total)
	resp.Body.Pagination.TotalPages = (total + limit - 1) / limit

	resp.Body.Logs = make([]dto.AuditLogDTO, 0, len(logs))
	for _, log := range logs {
		logDTO := h.buildAuditLogDTO(log)
		if name, ok := userNames[log.ActorID]; ok {
			logDTO.ActorName = name
		}
		resp.Body.Logs = append(resp.Body.Logs, logDTO)
	}

	return resp, nil
}

// GetAuditLog retrieves a single audit log entry
func (h *AuditLogHandler) GetAuditLog(ctx context.Context, req *GetAuditLogRequest) (*dto.AuditLogResponse, error) {
	// SECURITY: Verify user has access to this tenant
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	logID, err := uuid.Parse(req.LogID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid log ID", err)
	}

	log, err := h.service.GetByID(ctx, logID)
	if err != nil {
		return nil, huma.Error404NotFound("Audit log not found", err)
	}

	logDTO := h.buildAuditLogDTO(log)

	// Fetch actor name if it's a user
	if log.ActorType == entity.ActorTypeUser && log.ActorID != "" {
		if userID, err := uuid.Parse(log.ActorID); err == nil {
			if user, err := h.userRepo.GetByIDs(ctx, []uuid.UUID{userID}); err == nil && len(user) > 0 {
				logDTO.ActorName = user[0].Name
			}
		}
	}

	resp := &dto.AuditLogResponse{
		Body: logDTO,
	}
	return resp, nil
}

func (h *AuditLogHandler) buildAuditLogDTO(log *entity.AuditLog) dto.AuditLogDTO {
	var oldValue, newValue, metadata any

	if len(log.OldValue) > 0 {
		json.Unmarshal(log.OldValue, &oldValue)
	}
	if len(log.NewValue) > 0 {
		json.Unmarshal(log.NewValue, &newValue)
	}
	if log.Metadata != nil {
		metadata = log.Metadata
	}

	return dto.AuditLogDTO{
		ID:         log.ID,
		TenantID:   log.TenantID,
		EntityType: string(log.EntityType),
		EntityID:   log.EntityID,
		Action:     string(log.Action),
		ActorID:    log.ActorID,
		ActorType:  string(log.ActorType),
		OldValue:   oldValue,
		NewValue:   newValue,
		Metadata:   metadata,
		CreatedAt:  log.CreatedAt,
	}
}
