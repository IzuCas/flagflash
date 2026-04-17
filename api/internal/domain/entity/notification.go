package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationTypeApprovalRequest  NotificationType = "approval_request"
	NotificationTypeApprovalDecision NotificationType = "approval_decision"
	NotificationTypeFlagChange       NotificationType = "flag_change"
	NotificationTypeExperimentUpdate NotificationType = "experiment_update"
	NotificationTypeRolloutUpdate    NotificationType = "rollout_update"
	NotificationTypeEmergencyAlert   NotificationType = "emergency_alert"
	NotificationTypeStaleFlag        NotificationType = "stale_flag"
	NotificationTypeScheduledChange  NotificationType = "scheduled_change"
)

// Notification represents a user notification
type Notification struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	TenantID  uuid.UUID        `json:"tenant_id"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message,omitempty"`
	Link      string           `json:"link,omitempty"`
	Read      bool             `json:"read"`
	ReadAt    *time.Time       `json:"read_at,omitempty"`
	Metadata  json.RawMessage  `json:"metadata,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
}

// NewNotification creates a new notification
func NewNotification(userID, tenantID uuid.UUID, notifType NotificationType, title, message, link string, metadata map[string]interface{}) *Notification {
	var metadataJSON json.RawMessage
	if metadata != nil {
		metadataJSON, _ = json.Marshal(metadata)
	}
	return &Notification{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Link:      link,
		Read:      false,
		Metadata:  metadataJSON,
		CreatedAt: time.Now(),
	}
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.Read = true
	n.ReadAt = &now
}

// MarkAsUnread marks the notification as unread
func (n *Notification) MarkAsUnread() {
	n.Read = false
	n.ReadAt = nil
}
