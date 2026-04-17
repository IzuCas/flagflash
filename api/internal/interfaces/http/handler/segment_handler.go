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

// SegmentHandler handles segment-related HTTP requests
type SegmentHandler struct {
	service *service.SegmentService
}

// NewSegmentHandler creates a new segment handler
func NewSegmentHandler(service *service.SegmentService) *SegmentHandler {
	return &SegmentHandler{service: service}
}

// RegisterRoutes registers segment routes
func (h *SegmentHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createSegment",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/segments",
		Summary:     "Create a new segment",
		Tags:        []string{"Segments"},
	}, h.CreateSegment)

	huma.Register(api, huma.Operation{
		OperationID: "getSegment",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/segments/{segment_id}",
		Summary:     "Get segment by ID",
		Tags:        []string{"Segments"},
	}, h.GetSegment)

	huma.Register(api, huma.Operation{
		OperationID: "listSegments",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/segments",
		Summary:     "List segments for a tenant",
		Tags:        []string{"Segments"},
	}, h.ListSegments)

	huma.Register(api, huma.Operation{
		OperationID: "updateSegment",
		Method:      http.MethodPut,
		Path:        "/tenants/{tenant_id}/segments/{segment_id}",
		Summary:     "Update segment",
		Tags:        []string{"Segments"},
	}, h.UpdateSegment)

	huma.Register(api, huma.Operation{
		OperationID: "deleteSegment",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/segments/{segment_id}",
		Summary:     "Delete segment",
		Tags:        []string{"Segments"},
	}, h.DeleteSegment)

	huma.Register(api, huma.Operation{
		OperationID: "addIncludedUser",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/segments/{segment_id}/included-users",
		Summary:     "Add user to segment's included list",
		Tags:        []string{"Segments"},
	}, h.AddIncludedUser)

	huma.Register(api, huma.Operation{
		OperationID: "removeIncludedUser",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/segments/{segment_id}/included-users/{user_id}",
		Summary:     "Remove user from segment's included list",
		Tags:        []string{"Segments"},
	}, h.RemoveIncludedUser)

	huma.Register(api, huma.Operation{
		OperationID: "addExcludedUser",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/segments/{segment_id}/excluded-users",
		Summary:     "Add user to segment's excluded list",
		Tags:        []string{"Segments"},
	}, h.AddExcludedUser)

	huma.Register(api, huma.Operation{
		OperationID: "removeExcludedUser",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/segments/{segment_id}/excluded-users/{user_id}",
		Summary:     "Remove user from segment's excluded list",
		Tags:        []string{"Segments"},
	}, h.RemoveExcludedUser)
}

type GetSegmentRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	SegmentID string `path:"segment_id" format:"uuid"`
}

type DeleteSegmentRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	SegmentID string `path:"segment_id" format:"uuid"`
}

type ListSegmentsRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Page     int    `query:"page" default:"1" minimum:"1"`
	Limit    int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
	Search   string `query:"search,omitempty"`
}

type SegmentUserRemoveRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	SegmentID string `path:"segment_id" format:"uuid"`
	UserID    string `path:"user_id"`
}

func (h *SegmentHandler) CreateSegment(ctx context.Context, req *dto.CreateSegmentRequest) (*dto.SegmentResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	conditions := make([]entity.Condition, len(req.Body.Conditions))
	for i, c := range req.Body.Conditions {
		conditions[i] = entity.Condition{
			Attribute: c.Attribute,
			Operator:  entity.Operator(c.Operator),
			Value:     c.Value,
		}
	}

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}
	segment, err := h.service.Create(ctx, tenantID, req.Body.Name, req.Body.Description, conditions, &userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create segment", err)
	}

	return &dto.SegmentResponse{Body: toSegmentDTO(segment)}, nil
}

func (h *SegmentHandler) GetSegment(ctx context.Context, req *GetSegmentRequest) (*dto.SegmentResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	segmentID, err := uuid.Parse(req.SegmentID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid segment ID", err)
	}

	segment, err := h.service.GetByID(ctx, segmentID)
	if err != nil {
		return nil, huma.Error404NotFound("Segment not found")
	}

	return &dto.SegmentResponse{Body: toSegmentDTO(segment)}, nil
}

func (h *SegmentHandler) ListSegments(ctx context.Context, req *ListSegmentsRequest) (*dto.SegmentsListResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	offset := (req.Page - 1) * req.Limit
	segments, total, err := h.service.ListByTenantPaginated(ctx, tenantID, req.Limit, offset, req.Search)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list segments", err)
	}

	segmentDTOs := make([]dto.SegmentDTO, len(segments))
	for i, s := range segments {
		segmentDTOs[i] = toSegmentDTO(s)
	}

	totalPages := (total + req.Limit - 1) / req.Limit

	return &dto.SegmentsListResponse{
		Body: struct {
			Segments   []dto.SegmentDTO       `json:"segments"`
			Pagination dto.PaginationResponse `json:"pagination"`
		}{
			Segments: segmentDTOs,
			Pagination: dto.PaginationResponse{
				Page:       req.Page,
				Limit:      req.Limit,
				Total:      int64(total),
				TotalPages: totalPages,
			},
		},
	}, nil
}

func (h *SegmentHandler) UpdateSegment(ctx context.Context, req *dto.UpdateSegmentRequest) (*dto.SegmentResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	segmentID, err := uuid.Parse(req.SegmentID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid segment ID", err)
	}

	segment, err := h.service.GetByID(ctx, segmentID)
	if err != nil {
		return nil, huma.Error404NotFound("Segment not found")
	}

	conditions := make([]entity.Condition, len(req.Body.Conditions))
	for i, c := range req.Body.Conditions {
		conditions[i] = entity.Condition{
			Attribute: c.Attribute,
			Operator:  entity.Operator(c.Operator),
			Value:     c.Value,
		}
	}

	segment.Update(req.Body.Name, req.Body.Description, conditions)

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}
	if err := h.service.Update(ctx, segment, &userID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to update segment", err)
	}

	return &dto.SegmentResponse{Body: toSegmentDTO(segment)}, nil
}

func (h *SegmentHandler) DeleteSegment(ctx context.Context, req *DeleteSegmentRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}

	segmentID, err := uuid.Parse(req.SegmentID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid segment ID", err)
	}

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}
	if err := h.service.Delete(ctx, segmentID, &userID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete segment", err)
	}

	return &struct{}{}, nil
}

func (h *SegmentHandler) AddIncludedUser(ctx context.Context, req *dto.SegmentUserRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	segmentID, err := uuid.Parse(req.SegmentID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid segment ID", err)
	}

	if err := h.service.AddIncludedUser(ctx, segmentID, req.Body.UserID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to add user", err)
	}

	return &struct{}{}, nil
}

func (h *SegmentHandler) RemoveIncludedUser(ctx context.Context, req *SegmentUserRemoveRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	segmentID, err := uuid.Parse(req.SegmentID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid segment ID", err)
	}

	if err := h.service.RemoveIncludedUser(ctx, segmentID, req.UserID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to remove user", err)
	}

	return &struct{}{}, nil
}

func (h *SegmentHandler) AddExcludedUser(ctx context.Context, req *dto.SegmentUserRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	segmentID, err := uuid.Parse(req.SegmentID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid segment ID", err)
	}

	if err := h.service.AddExcludedUser(ctx, segmentID, req.Body.UserID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to add user", err)
	}

	return &struct{}{}, nil
}

func (h *SegmentHandler) RemoveExcludedUser(ctx context.Context, req *SegmentUserRemoveRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "member"); err != nil {
		return nil, err
	}

	segmentID, err := uuid.Parse(req.SegmentID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid segment ID", err)
	}

	if err := h.service.RemoveExcludedUser(ctx, segmentID, req.UserID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to remove user", err)
	}

	return &struct{}{}, nil
}

func toSegmentDTO(s *entity.Segment) dto.SegmentDTO {
	conditions := make([]dto.ConditionDTO, len(s.Conditions))
	for i, c := range s.Conditions {
		conditions[i] = dto.ConditionDTO{
			Attribute: c.Attribute,
			Operator:  string(c.Operator),
			Value:     c.Value,
		}
	}

	return dto.SegmentDTO{
		ID:            s.ID,
		TenantID:      s.TenantID,
		Name:          s.Name,
		Description:   s.Description,
		Conditions:    conditions,
		IsDynamic:     s.IsDynamic,
		IncludedUsers: s.IncludedUsers,
		ExcludedUsers: s.ExcludedUsers,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}
