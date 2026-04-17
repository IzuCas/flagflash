package entity

import (
	"time"

	"github.com/google/uuid"
)

// EmergencyControlType defines the type of emergency control
type EmergencyControlType string

const (
	EmergencyControlTypeKillSwitch  EmergencyControlType = "kill_switch"
	EmergencyControlTypeReadOnly    EmergencyControlType = "read_only"
	EmergencyControlTypeMaintenance EmergencyControlType = "maintenance"
)

// EmergencyControl represents an emergency control setting
type EmergencyControl struct {
	ID            uuid.UUID            `json:"id"`
	TenantID      uuid.UUID            `json:"tenant_id"`
	EnvironmentID *uuid.UUID           `json:"environment_id,omitempty"`
	ControlType   EmergencyControlType `json:"control_type"`
	Enabled       bool                 `json:"enabled"`
	Reason        string               `json:"reason,omitempty"`
	EnabledBy     *uuid.UUID           `json:"enabled_by,omitempty"`
	EnabledAt     *time.Time           `json:"enabled_at,omitempty"`
	ExpiresAt     *time.Time           `json:"expires_at,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// NewEmergencyControl creates a new emergency control
func NewEmergencyControl(tenantID uuid.UUID, environmentID *uuid.UUID, controlType EmergencyControlType) *EmergencyControl {
	now := time.Now()
	return &EmergencyControl{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EnvironmentID: environmentID,
		ControlType:   controlType,
		Enabled:       false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Enable enables the emergency control
func (e *EmergencyControl) Enable(userID uuid.UUID, reason string, duration *time.Duration) {
	now := time.Now()
	e.Enabled = true
	e.EnabledBy = &userID
	e.EnabledAt = &now
	e.Reason = reason
	if duration != nil {
		expiresAt := now.Add(*duration)
		e.ExpiresAt = &expiresAt
	}
	e.UpdatedAt = now
}

// Disable disables the emergency control
func (e *EmergencyControl) Disable() {
	e.Enabled = false
	e.EnabledBy = nil
	e.EnabledAt = nil
	e.ExpiresAt = nil
	e.Reason = ""
	e.UpdatedAt = time.Now()
}

// IsExpired checks if the control has expired
func (e *EmergencyControl) IsExpired() bool {
	if e.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*e.ExpiresAt)
}

// ShouldDisable checks if the control should be auto-disabled
func (e *EmergencyControl) ShouldDisable() bool {
	return e.Enabled && e.IsExpired()
}
