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

// NotificationHandler handles notification HTTP requests
type NotificationHandler struct {
	service *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

// RegisterRoutes registers notification routes
func (h *NotificationHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "listNotifications",
		Method:      http.MethodGet,
		Path:        "/users/me/notifications",
		Summary:     "List notifications for current user",
		Tags:        []string{"Notifications"},
	}, h.ListNotifications)

	huma.Register(api, huma.Operation{
		OperationID: "getUnreadCount",
		Method:      http.MethodGet,
		Path:        "/users/me/notifications/unread-count",
		Summary:     "Get unread notification count",
		Tags:        []string{"Notifications"},
	}, h.GetUnreadCount)

	huma.Register(api, huma.Operation{
		OperationID: "markNotificationRead",
		Method:      http.MethodPost,
		Path:        "/users/me/notifications/{notification_id}/read",
		Summary:     "Mark notification as read",
		Tags:        []string{"Notifications"},
	}, h.MarkAsRead)

	huma.Register(api, huma.Operation{
		OperationID: "markAllNotificationsRead",
		Method:      http.MethodPost,
		Path:        "/users/me/notifications/read-all",
		Summary:     "Mark all notifications as read",
		Tags:        []string{"Notifications"},
	}, h.MarkAllAsRead)

	huma.Register(api, huma.Operation{
		OperationID: "deleteNotification",
		Method:      http.MethodDelete,
		Path:        "/users/me/notifications/{notification_id}",
		Summary:     "Delete a notification",
		Tags:        []string{"Notifications"},
	}, h.DeleteNotification)
}

type ListNotificationsRequest struct {
	Page  int `query:"page" default:"1" minimum:"1"`
	Limit int `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type MarkNotificationReadRequest struct {
	NotificationID string `path:"notification_id" format:"uuid"`
}

type DeleteNotificationRequest struct {
	NotificationID string `path:"notification_id" format:"uuid"`
}

func (h *NotificationHandler) ListNotifications(ctx context.Context, req *ListNotificationsRequest) (*dto.NotificationsListResponse, error) {
	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	offset := (req.Page - 1) * req.Limit

	notifications, total, err := h.service.ListByUser(ctx, userID, req.Limit, offset)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list notifications", err)
	}

	notificationDTOs := make([]dto.NotificationDTO, len(notifications))
	for i, n := range notifications {
		notificationDTOs[i] = toNotificationDTO(n)
	}

	totalPages := (total + req.Limit - 1) / req.Limit

	return &dto.NotificationsListResponse{
		Body: struct {
			Notifications []dto.NotificationDTO  `json:"notifications"`
			Pagination    dto.PaginationResponse `json:"pagination"`
		}{
			Notifications: notificationDTOs,
			Pagination: dto.PaginationResponse{
				Page:       req.Page,
				Limit:      req.Limit,
				Total:      int64(total),
				TotalPages: totalPages,
			},
		},
	}, nil
}

func (h *NotificationHandler) GetUnreadCount(ctx context.Context, _ *struct{}) (*dto.UnreadCountResponse, error) {
	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	count, err := h.service.CountUnread(ctx, userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get unread count", err)
	}

	return &dto.UnreadCountResponse{
		Body: struct {
			Count int `json:"count"`
		}{
			Count: count,
		},
	}, nil
}

func (h *NotificationHandler) MarkAsRead(ctx context.Context, req *MarkNotificationReadRequest) (*struct{}, error) {
	notificationID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid notification ID", err)
	}

	// Verify the notification belongs to the current user
	notification, err := h.service.GetByID(ctx, notificationID)
	if err != nil {
		return nil, huma.Error404NotFound("Notification not found")
	}

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	if notification.UserID != userID {
		return nil, huma.Error403Forbidden("Access denied")
	}

	if err := h.service.MarkAsRead(ctx, notificationID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to mark notification as read", err)
	}

	return &struct{}{}, nil
}

func (h *NotificationHandler) MarkAllAsRead(ctx context.Context, _ *struct{}) (*struct{}, error) {
	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	if err := h.service.MarkAllAsRead(ctx, userID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to mark all notifications as read", err)
	}

	return &struct{}{}, nil
}

func (h *NotificationHandler) DeleteNotification(ctx context.Context, req *DeleteNotificationRequest) (*struct{}, error) {
	notificationID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid notification ID", err)
	}

	// Verify the notification belongs to the current user
	notification, err := h.service.GetByID(ctx, notificationID)
	if err != nil {
		return nil, huma.Error404NotFound("Notification not found")
	}

	userIDStr := middleware.GetUserIDFromContext(ctx)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, huma.Error401Unauthorized("Invalid user")
	}

	if notification.UserID != userID {
		return nil, huma.Error403Forbidden("Access denied")
	}

	if err := h.service.Delete(ctx, notificationID); err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete notification", err)
	}

	return &struct{}{}, nil
}

func toNotificationDTO(n *entity.Notification) dto.NotificationDTO {
	return dto.NotificationDTO{
		ID:        n.ID,
		UserID:    n.UserID,
		TenantID:  n.TenantID,
		Type:      string(n.Type),
		Title:     n.Title,
		Message:   n.Message,
		Link:      n.Link,
		Read:      n.Read,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}
