package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditAction defines the type of action performed
type AuditAction string

const (
	AuditActionCreate  AuditAction = "create"
	AuditActionUpdate  AuditAction = "update"
	AuditActionDelete  AuditAction = "delete"
	AuditActionEnable  AuditAction = "enable"
	AuditActionDisable AuditAction = "disable"
	AuditActionToggle  AuditAction = "toggle"
	AuditActionRevoke  AuditAction = "revoke"
	AuditActionRotate  AuditAction = "rotate"
)

// EntityType defines the type of entity
type EntityType string

const (
	EntityTypeTenant        EntityType = "tenant"
	EntityTypeApplication   EntityType = "application"
	EntityTypeEnvironment   EntityType = "environment"
	EntityTypeFeatureFlag   EntityType = "feature_flag"
	EntityTypeTargetingRule EntityType = "targeting_rule"
	EntityTypeAPIKey        EntityType = "api_key"
	EntityTypeUser          EntityType = "user"
)

// ActorType defines who performed the action
type ActorType string

const (
	ActorTypeUser   ActorType = "user"
	ActorTypeAPIKey ActorType = "api_key"
	ActorTypeSystem ActorType = "system"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         uuid.UUID              `json:"id"`
	TenantID   uuid.UUID              `json:"tenant_id"`
	EntityType EntityType             `json:"entity_type"`
	EntityID   uuid.UUID              `json:"entity_id"`
	Action     AuditAction            `json:"action"`
	ActorID    string                 `json:"actor_id"`
	ActorType  ActorType              `json:"actor_type"`
	OldValue   json.RawMessage        `json:"old_value,omitempty"`
	NewValue   json.RawMessage        `json:"new_value,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(
	tenantID uuid.UUID,
	entityType EntityType,
	entityID uuid.UUID,
	action AuditAction,
	actorID string,
	actorType ActorType,
	oldValue, newValue interface{},
	metadata map[string]interface{},
) *AuditLog {
	var oldJSON, newJSON json.RawMessage

	if oldValue != nil {
		oldJSON, _ = json.Marshal(oldValue)
	}
	if newValue != nil {
		newJSON, _ = json.Marshal(newValue)
	}

	return &AuditLog{
		ID:         uuid.New(),
		TenantID:   tenantID,
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		ActorID:    actorID,
		ActorType:  actorType,
		OldValue:   oldJSON,
		NewValue:   newJSON,
		Metadata:   metadata,
		CreatedAt:  time.Now(),
	}
}

// AuditLogQuery represents query parameters for audit logs
type AuditLogQuery struct {
	TenantID   *uuid.UUID
	EntityType *EntityType
	EntityID   *uuid.UUID
	Action     *AuditAction
	ActorID    *string
	ActorType  *ActorType
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}

// GetDiff returns the differences between old and new values
func (a *AuditLog) GetDiff() map[string]interface{} {
	diff := make(map[string]interface{})

	var oldMap, newMap map[string]interface{}
	json.Unmarshal(a.OldValue, &oldMap)
	json.Unmarshal(a.NewValue, &newMap)

	// Find changed fields
	for key, newVal := range newMap {
		if oldVal, exists := oldMap[key]; exists {
			if newVal != oldVal {
				diff[key] = map[string]interface{}{
					"old": oldVal,
					"new": newVal,
				}
			}
		} else {
			diff[key] = map[string]interface{}{
				"old": nil,
				"new": newVal,
			}
		}
	}

	// Find removed fields
	for key, oldVal := range oldMap {
		if _, exists := newMap[key]; !exists {
			diff[key] = map[string]interface{}{
				"old": oldVal,
				"new": nil,
			}
		}
	}

	return diff
}
