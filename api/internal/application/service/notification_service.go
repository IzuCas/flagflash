package service

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// NotificationService handles notification operations
type NotificationService struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
	membershipRepo   repository.UserTenantMembershipRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	membershipRepo repository.UserTenantMembershipRepository,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		membershipRepo:   membershipRepo,
	}
}

// Create creates a new notification
func (s *NotificationService) Create(ctx context.Context, notification *entity.Notification) error {
	return s.notificationRepo.Create(ctx, notification)
}

// CreateBatch creates multiple notifications at once
func (s *NotificationService) CreateBatch(ctx context.Context, notifications []*entity.Notification) error {
	return s.notificationRepo.CreateBatch(ctx, notifications)
}

// GetByID gets a notification by ID
func (s *NotificationService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	return s.notificationRepo.GetByID(ctx, id)
}

// ListByUser lists notifications for a user
func (s *NotificationService) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	return s.notificationRepo.ListByUser(ctx, userID, limit, offset)
}

// ListUnread lists unread notifications for a user
func (s *NotificationService) ListUnread(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error) {
	return s.notificationRepo.ListUnread(ctx, userID)
}

// CountUnread counts unread notifications for a user
func (s *NotificationService) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.notificationRepo.CountUnread(ctx, userID)
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.notificationRepo.MarkAsRead(ctx, id)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// Delete deletes a notification
func (s *NotificationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.notificationRepo.Delete(ctx, id)
}

// NotifyUser sends a notification to a user
func (s *NotificationService) NotifyUser(
	ctx context.Context,
	userID, tenantID uuid.UUID,
	notifType entity.NotificationType,
	title, message, link string,
	metadata map[string]interface{},
) error {
	notification := entity.NewNotification(userID, tenantID, notifType, title, message, link, metadata)
	return s.Create(ctx, notification)
}

// NotifyUsers sends a notification to multiple users
func (s *NotificationService) NotifyUsers(
	ctx context.Context,
	userIDs []uuid.UUID,
	tenantID uuid.UUID,
	notifType entity.NotificationType,
	title, message, link string,
	metadata map[string]interface{},
) error {
	notifications := make([]*entity.Notification, len(userIDs))
	for i, userID := range userIDs {
		notifications[i] = entity.NewNotification(userID, tenantID, notifType, title, message, link, metadata)
	}
	return s.CreateBatch(ctx, notifications)
}

// CleanupOld removes old notifications
func (s *NotificationService) CleanupOld(ctx context.Context, days int) (int, error) {
	return s.notificationRepo.DeleteOlderThan(ctx, days)
}

// NotifyTenantAdmins sends a notification to all admins of a tenant
func (s *NotificationService) NotifyTenantAdmins(
	ctx context.Context,
	tenantID uuid.UUID,
	notifType entity.NotificationType,
	title, message string,
	metadata map[string]interface{},
) error {
	if s.membershipRepo == nil {
		return nil
	}

	// Get all memberships for the tenant (using high limit to get all)
	memberships, _, err := s.membershipRepo.ListByTenant(ctx, tenantID, 1000, 0)
	if err != nil {
		return err
	}

	var adminUserIDs []uuid.UUID
	for _, m := range memberships {
		if m.Role == entity.UserRoleOwner || m.Role == entity.UserRoleAdmin {
			adminUserIDs = append(adminUserIDs, m.UserID)
		}
	}

	if len(adminUserIDs) == 0 {
		return nil
	}

	return s.NotifyUsers(ctx, adminUserIDs, tenantID, notifType, title, message, "", metadata)
}

// NotifyKillSwitchActivated sends notification when kill switch is activated
func (s *NotificationService) NotifyKillSwitchActivated(
	ctx context.Context,
	tenantID uuid.UUID,
	control *entity.EmergencyControl,
	activatedBy *uuid.UUID,
) error {
	title := "Kill Switch Activated"
	message := "A kill switch has been activated"
	if control.Reason != "" {
		message += ": " + control.Reason
	}

	metadata := map[string]interface{}{
		"control_id":   control.ID,
		"control_type": control.ControlType,
	}
	if activatedBy != nil {
		metadata["activated_by"] = *activatedBy
	}

	return s.NotifyTenantAdmins(ctx, tenantID, entity.NotificationTypeEmergencyAlert, title, message, metadata)
}
