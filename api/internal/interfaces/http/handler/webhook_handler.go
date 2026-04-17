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

// WebhookHandler handles webhook-related HTTP requests
type WebhookHandler struct {
	service *service.WebhookService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(service *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// RegisterRoutes registers webhook routes
func (h *WebhookHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createWebhook",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/webhooks",
		Summary:     "Create a new webhook",
		Tags:        []string{"Webhooks"},
	}, h.CreateWebhook)

	huma.Register(api, huma.Operation{
		OperationID: "getWebhook",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/webhooks/{webhook_id}",
		Summary:     "Get webhook by ID",
		Tags:        []string{"Webhooks"},
	}, h.GetWebhook)

	huma.Register(api, huma.Operation{
		OperationID: "listWebhooks",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/webhooks",
		Summary:     "List webhooks for a tenant",
		Tags:        []string{"Webhooks"},
	}, h.ListWebhooks)

	huma.Register(api, huma.Operation{
		OperationID: "updateWebhook",
		Method:      http.MethodPut,
		Path:        "/tenants/{tenant_id}/webhooks/{webhook_id}",
		Summary:     "Update webhook",
		Tags:        []string{"Webhooks"},
	}, h.UpdateWebhook)

	huma.Register(api, huma.Operation{
		OperationID: "deleteWebhook",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/webhooks/{webhook_id}",
		Summary:     "Delete webhook",
		Tags:        []string{"Webhooks"},
	}, h.DeleteWebhook)

	huma.Register(api, huma.Operation{
		OperationID: "toggleWebhook",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/webhooks/{webhook_id}/toggle",
		Summary:     "Enable or disable webhook",
		Tags:        []string{"Webhooks"},
	}, h.ToggleWebhook)

	huma.Register(api, huma.Operation{
		OperationID: "listWebhookDeliveries",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/webhooks/{webhook_id}/deliveries",
		Summary:     "List delivery history for a webhook",
		Tags:        []string{"Webhooks"},
	}, h.ListDeliveries)

	huma.Register(api, huma.Operation{
		OperationID: "retryWebhookDelivery",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/webhooks/{webhook_id}/deliveries/{delivery_id}/retry",
		Summary:     "Manually retry a failed delivery",
		Tags:        []string{"Webhooks"},
	}, h.RetryDelivery)

	huma.Register(api, huma.Operation{
		OperationID: "testWebhook",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/webhooks/{webhook_id}/test",
		Summary:     "Send a test event to the webhook",
		Tags:        []string{"Webhooks"},
	}, h.TestWebhook)
}

type GetWebhookRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	WebhookID string `path:"webhook_id" format:"uuid"`
}

type DeleteWebhookRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	WebhookID string `path:"webhook_id" format:"uuid"`
}

type ListWebhooksRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Page     int    `query:"page" default:"1" minimum:"1"`
	Limit    int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type ToggleWebhookRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	WebhookID string `path:"webhook_id" format:"uuid"`
	Body      struct {
		Enabled bool `json:"enabled"`
	}
}

func (h *WebhookHandler) CreateWebhook(ctx context.Context, req *dto.CreateWebhookRequest) (*dto.WebhookResponse, error) {
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

	events := make([]entity.WebhookEvent, len(req.Body.Events))
	for i, e := range req.Body.Events {
		events[i] = entity.WebhookEvent(e)
	}

	webhook := entity.NewWebhook(tenantID, req.Body.Name, req.Body.URL, req.Body.Secret, events)
	webhook.Headers = req.Body.Headers

	if err := h.service.Create(ctx, webhook); err != nil {
		return nil, huma.Error500InternalServerError("Failed to create webhook", err)
	}

	return &dto.WebhookResponse{Body: toWebhookDTO(webhook)}, nil
}

func (h *WebhookHandler) GetWebhook(ctx context.Context, req *GetWebhookRequest) (*dto.WebhookResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	webhookID, err := uuid.Parse(req.WebhookID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid webhook ID", err)
	}

	webhook, err := h.service.GetByID(ctx, webhookID)
	if err != nil {
		return nil, huma.Error404NotFound("Webhook not found")
	}

	return &dto.WebhookResponse{Body: toWebhookDTO(webhook)}, nil
}

func (h *WebhookHandler) ListWebhooks(ctx context.Context, req *ListWebhooksRequest) (*dto.WebhooksListResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	webhooks, err := h.service.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list webhooks", err)
	}

	webhookDTOs := make([]dto.WebhookDTO, len(webhooks))
	for i, w := range webhooks {
		webhookDTOs[i] = toWebhookDTO(w)
	}

	return &dto.WebhooksListResponse{
		Body: struct {
			Webhooks []dto.WebhookDTO `json:"webhooks"`
		}{
			Webhooks: webhookDTOs,
		},
	}, nil
}

func (h *WebhookHandler) UpdateWebhook(ctx context.Context, req *dto.UpdateWebhookRequest) (*dto.WebhookResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}

	webhookID, err := uuid.Parse(req.WebhookID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid webhook ID", err)
	}

	webhook, err := h.service.GetByID(ctx, webhookID)
	if err != nil {
		return nil, huma.Error404NotFound("Webhook not found")
	}

	events := make([]entity.WebhookEvent, len(req.Body.Events))
	for i, e := range req.Body.Events {
		events[i] = entity.WebhookEvent(e)
	}

	webhook.Update(req.Body.Name, req.Body.URL, req.Body.Secret, events, req.Body.Headers)

	if err := h.service.Update(ctx, webhook); err != nil {
		return nil, huma.Error500InternalServerError("Failed to update webhook", err)
	}

	return &dto.WebhookResponse{Body: toWebhookDTO(webhook)}, nil
}

func (h *WebhookHandler) DeleteWebhook(ctx context.Context, req *DeleteWebhookRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}

	webhookID, err := uuid.Parse(req.WebhookID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid webhook ID", err)
	}

	if err := h.service.Delete(ctx, webhookID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete webhook", err)
	}

	return &struct{}{}, nil
}

func (h *WebhookHandler) ToggleWebhook(ctx context.Context, req *ToggleWebhookRequest) (*dto.WebhookResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}
	if err := middleware.RequireRole(ctx, "admin"); err != nil {
		return nil, err
	}

	webhookID, err := uuid.Parse(req.WebhookID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid webhook ID", err)
	}

	webhook, err := h.service.GetByID(ctx, webhookID)
	if err != nil {
		return nil, huma.Error404NotFound("Webhook not found")
	}

	if req.Body.Enabled {
		webhook.Enable()
	} else {
		webhook.Disable()
	}

	if err := h.service.Update(ctx, webhook); err != nil {
		return nil, huma.Error500InternalServerError("Failed to toggle webhook", err)
	}

	return &dto.WebhookResponse{Body: toWebhookDTO(webhook)}, nil
}

type ListDeliveriesRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	WebhookID string `path:"webhook_id" format:"uuid"`
	Page      int    `query:"page" default:"1" minimum:"1"`
	Limit     int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type RetryDeliveryRequest struct {
	TenantID   string `path:"tenant_id" format:"uuid"`
	WebhookID  string `path:"webhook_id" format:"uuid"`
	DeliveryID string `path:"delivery_id" format:"uuid"`
}

type TestWebhookRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	WebhookID string `path:"webhook_id" format:"uuid"`
}

func (h *WebhookHandler) ListDeliveries(ctx context.Context, req *ListDeliveriesRequest) (*dto.WebhookDeliveriesResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	webhookID, err := uuid.Parse(req.WebhookID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid webhook ID", err)
	}

	offset := (req.Page - 1) * req.Limit
	deliveries, total, err := h.service.GetDeliveries(ctx, webhookID, req.Limit, offset)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list deliveries", err)
	}

	deliveryDTOs := make([]dto.WebhookDeliveryDTO, len(deliveries))
	for i, d := range deliveries {
		deliveryDTOs[i] = toWebhookDeliveryDTO(d)
	}

	totalPages := (total + req.Limit - 1) / req.Limit

	return &dto.WebhookDeliveriesResponse{
		Body: struct {
			Deliveries []dto.WebhookDeliveryDTO `json:"deliveries"`
			Pagination dto.PaginationResponse   `json:"pagination"`
		}{
			Deliveries: deliveryDTOs,
			Pagination: dto.PaginationResponse{
				Page:       req.Page,
				Limit:      req.Limit,
				Total:      int64(total),
				TotalPages: totalPages,
			},
		},
	}, nil
}

func (h *WebhookHandler) RetryDelivery(ctx context.Context, req *RetryDeliveryRequest) (*struct{}, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	deliveryID, err := uuid.Parse(req.DeliveryID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid delivery ID", err)
	}

	if err := h.service.RetryDelivery(ctx, deliveryID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to retry delivery", err)
	}

	return &struct{}{}, nil
}

func (h *WebhookHandler) TestWebhook(ctx context.Context, req *TestWebhookRequest) (*dto.WebhookDeliveryResponse, error) {
	if err := middleware.RequireTenantAccess(ctx, req.TenantID); err != nil {
		return nil, err
	}

	webhookID, err := uuid.Parse(req.WebhookID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid webhook ID", err)
	}

	webhook, err := h.service.GetByID(ctx, webhookID)
	if err != nil {
		return nil, huma.Error404NotFound("Webhook not found")
	}

	delivery, err := h.service.TestWebhook(ctx, webhook)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to send test event", err)
	}

	return &dto.WebhookDeliveryResponse{Body: toWebhookDeliveryDTO(delivery)}, nil
}

func toWebhookDeliveryDTO(d *entity.WebhookDelivery) dto.WebhookDeliveryDTO {
	return dto.WebhookDeliveryDTO{
		ID:             d.ID,
		WebhookID:      d.WebhookID,
		EventType:      string(d.EventType),
		ResponseStatus: d.ResponseStatus,
		DurationMs:     d.DurationMs,
		Attempt:        d.Attempt,
		Status:         string(d.Status),
		ErrorMessage:   d.ErrorMessage,
		DeliveredAt:    d.DeliveredAt,
		CreatedAt:      d.CreatedAt,
	}
}

func toWebhookDTO(w *entity.Webhook) dto.WebhookDTO {
	events := make([]string, len(w.Events))
	for i, e := range w.Events {
		events[i] = string(e)
	}

	return dto.WebhookDTO{
		ID:             w.ID,
		TenantID:       w.TenantID,
		Name:           w.Name,
		URL:            w.URL,
		Events:         events,
		Headers:        w.Headers,
		Enabled:        w.Enabled,
		RetryCount:     w.RetryCount,
		TimeoutSeconds: w.TimeoutSeconds,
		CreatedAt:      w.CreatedAt,
		UpdatedAt:      w.UpdatedAt,
	}
}
