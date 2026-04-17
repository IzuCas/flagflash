package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ApprovalSetting defines approval requirements for an environment or flag
type ApprovalSetting struct {
	ID                   uuid.UUID  `json:"id"`
	TenantID             uuid.UUID  `json:"tenant_id"`
	EnvironmentID        *uuid.UUID `json:"environment_id,omitempty"`
	FeatureFlagID        *uuid.UUID `json:"feature_flag_id,omitempty"`
	RequiresApproval     bool       `json:"requires_approval"`
	MinApprovers         int        `json:"min_approvers"`
	AutoRejectHours      int        `json:"auto_reject_hours"`
	AllowedApproverRoles []string   `json:"allowed_approver_roles"`
	NotifyOnRequest      bool       `json:"notify_on_request"`
	NotifyOnDecision     bool       `json:"notify_on_decision"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// NewApprovalSetting creates a new approval setting
func NewApprovalSetting(tenantID uuid.UUID, environmentID, flagID *uuid.UUID) *ApprovalSetting {
	now := time.Now()
	return &ApprovalSetting{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		EnvironmentID:        environmentID,
		FeatureFlagID:        flagID,
		RequiresApproval:     true,
		MinApprovers:         1,
		AutoRejectHours:      72,
		AllowedApproverRoles: []string{"owner", "admin"},
		NotifyOnRequest:      true,
		NotifyOnDecision:     true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// PendingChangeType defines types of pending changes
type PendingChangeType string

const (
	PendingChangeTypeCreate  PendingChangeType = "create"
	PendingChangeTypeUpdate  PendingChangeType = "update"
	PendingChangeTypeDelete  PendingChangeType = "delete"
	PendingChangeTypeEnable  PendingChangeType = "enable"
	PendingChangeTypeDisable PendingChangeType = "disable"
)

// PendingChangeEntityType defines entity types for pending changes
type PendingChangeEntityType string

const (
	PendingChangeEntityFlag    PendingChangeEntityType = "flag"
	PendingChangeEntityRule    PendingChangeEntityType = "rule"
	PendingChangeEntitySegment PendingChangeEntityType = "segment"
)

// PendingChangeStatus defines status of pending changes
type PendingChangeStatus string

const (
	PendingChangeStatusPending   PendingChangeStatus = "pending"
	PendingChangeStatusApproved  PendingChangeStatus = "approved"
	PendingChangeStatusRejected  PendingChangeStatus = "rejected"
	PendingChangeStatusExpired   PendingChangeStatus = "expired"
	PendingChangeStatusCancelled PendingChangeStatus = "cancelled"
)

// PendingChange represents a change awaiting approval
type PendingChange struct {
	ID              uuid.UUID               `json:"id"`
	TenantID        uuid.UUID               `json:"tenant_id"`
	EnvironmentID   uuid.UUID               `json:"environment_id"`
	FeatureFlagID   *uuid.UUID              `json:"feature_flag_id,omitempty"`
	TargetingRuleID *uuid.UUID              `json:"targeting_rule_id,omitempty"`
	ChangeType      PendingChangeType       `json:"change_type"`
	EntityType      PendingChangeEntityType `json:"entity_type"`
	OldValue        json.RawMessage         `json:"old_value,omitempty"`
	NewValue        json.RawMessage         `json:"new_value,omitempty"`
	Status          PendingChangeStatus     `json:"status"`
	RequestedBy     uuid.UUID               `json:"requested_by"`
	RequestComment  string                  `json:"request_comment,omitempty"`
	DecidedAt       *time.Time              `json:"decided_at,omitempty"`
	ExpiresAt       *time.Time              `json:"expires_at,omitempty"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	// Populated by joins
	Approvals     []Approval `json:"approvals,omitempty"`
	RequesterName string     `json:"requester_name,omitempty"`
}

// NewPendingChange creates a new pending change
func NewPendingChange(
	tenantID, environmentID, requestedBy uuid.UUID,
	flagID, ruleID *uuid.UUID,
	changeType PendingChangeType,
	entityType PendingChangeEntityType,
	oldValue, newValue json.RawMessage,
	comment string,
	autoRejectHours int,
) *PendingChange {
	now := time.Now()
	expiresAt := now.Add(time.Duration(autoRejectHours) * time.Hour)
	return &PendingChange{
		ID:              uuid.New(),
		TenantID:        tenantID,
		EnvironmentID:   environmentID,
		FeatureFlagID:   flagID,
		TargetingRuleID: ruleID,
		ChangeType:      changeType,
		EntityType:      entityType,
		OldValue:        oldValue,
		NewValue:        newValue,
		Status:          PendingChangeStatusPending,
		RequestedBy:     requestedBy,
		RequestComment:  comment,
		ExpiresAt:       &expiresAt,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// ApprovalDecision defines the decision types
type ApprovalDecision string

const (
	ApprovalDecisionApproved     ApprovalDecision = "approved"
	ApprovalDecisionRejected     ApprovalDecision = "rejected"
	ApprovalDecisionNeedsChanges ApprovalDecision = "needs_changes"
)

// Approval represents an individual approval decision
type Approval struct {
	ID              uuid.UUID        `json:"id"`
	PendingChangeID uuid.UUID        `json:"pending_change_id"`
	ApproverID      uuid.UUID        `json:"approver_id"`
	Decision        ApprovalDecision `json:"decision"`
	Comment         string           `json:"comment,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	// Populated by joins
	ApproverName string `json:"approver_name,omitempty"`
}

// NewApproval creates a new approval
func NewApproval(pendingChangeID, approverID uuid.UUID, decision ApprovalDecision, comment string) *Approval {
	return &Approval{
		ID:              uuid.New(),
		PendingChangeID: pendingChangeID,
		ApproverID:      approverID,
		Decision:        decision,
		Comment:         comment,
		CreatedAt:       time.Now(),
	}
}

// Approve marks the pending change as approved
func (pc *PendingChange) Approve() {
	now := time.Now()
	pc.Status = PendingChangeStatusApproved
	pc.DecidedAt = &now
	pc.UpdatedAt = now
}

// Reject marks the pending change as rejected
func (pc *PendingChange) Reject() {
	now := time.Now()
	pc.Status = PendingChangeStatusRejected
	pc.DecidedAt = &now
	pc.UpdatedAt = now
}

// Cancel cancels the pending change
func (pc *PendingChange) Cancel() {
	now := time.Now()
	pc.Status = PendingChangeStatusCancelled
	pc.DecidedAt = &now
	pc.UpdatedAt = now
}

// Expire marks the pending change as expired
func (pc *PendingChange) Expire() {
	now := time.Now()
	pc.Status = PendingChangeStatusExpired
	pc.DecidedAt = &now
	pc.UpdatedAt = now
}

// IsPending checks if the change is still pending
func (pc *PendingChange) IsPending() bool {
	return pc.Status == PendingChangeStatusPending
}

// IsApproved checks if the change is approved
func (pc *PendingChange) IsApproved() bool {
	return pc.Status == PendingChangeStatusApproved
}

// CountApprovals counts the number of approvals by decision type
func (pc *PendingChange) CountApprovals(decision ApprovalDecision) int {
	count := 0
	for _, a := range pc.Approvals {
		if a.Decision == decision {
			count++
		}
	}
	return count
}

// HasEnoughApprovals checks if the change has enough approvals
func (pc *PendingChange) HasEnoughApprovals(minRequired int) bool {
	return pc.CountApprovals(ApprovalDecisionApproved) >= minRequired
}
