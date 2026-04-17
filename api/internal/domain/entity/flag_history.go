package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FlagChangeType defines the type of flag change
type FlagChangeType string

const (
	FlagChangeTypeCreated         FlagChangeType = "created"
	FlagChangeTypeUpdated         FlagChangeType = "updated"
	FlagChangeTypeEnabled         FlagChangeType = "enabled"
	FlagChangeTypeDisabled        FlagChangeType = "disabled"
	FlagChangeTypeDeleted         FlagChangeType = "deleted"
	FlagChangeTypeScheduled       FlagChangeType = "scheduled"
	FlagChangeTypeRollout         FlagChangeType = "rollout"
	FlagChangeTypeRollback        FlagChangeType = "rollback"
	FlagChangeTypeLifecycleChange FlagChangeType = "lifecycle_change"
)

// FlagHistory represents a historical record of flag changes
type FlagHistory struct {
	ID            uuid.UUID       `json:"id"`
	FeatureFlagID uuid.UUID       `json:"feature_flag_id"`
	Version       int             `json:"version"`
	ChangeType    FlagChangeType  `json:"change_type"`
	ChangedBy     *uuid.UUID      `json:"changed_by,omitempty"`
	PreviousState json.RawMessage `json:"previous_state,omitempty"`
	NewState      json.RawMessage `json:"new_state,omitempty"`
	Comment       string          `json:"comment,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	// Populated by joins
	ChangedByName string `json:"changed_by_name,omitempty"`
}

// NewFlagHistory creates a new flag history entry
func NewFlagHistory(flagID uuid.UUID, version int, changeType FlagChangeType, changedBy *uuid.UUID, previousState, newState json.RawMessage, comment string) *FlagHistory {
	return &FlagHistory{
		ID:            uuid.New(),
		FeatureFlagID: flagID,
		Version:       version,
		ChangeType:    changeType,
		ChangedBy:     changedBy,
		PreviousState: previousState,
		NewState:      newState,
		Comment:       comment,
		CreatedAt:     time.Now(),
	}
}

// FlagHistoryFromFlag creates a history entry from a flag state
func FlagHistoryFromFlag(flag *FeatureFlag, changeType FlagChangeType, changedBy *uuid.UUID, previousFlag *FeatureFlag, comment string) *FlagHistory {
	var previousState, newState json.RawMessage

	if previousFlag != nil {
		previousState, _ = json.Marshal(previousFlag)
	}
	newState, _ = json.Marshal(flag)

	return NewFlagHistory(flag.ID, flag.Version, changeType, changedBy, previousState, newState, comment)
}
