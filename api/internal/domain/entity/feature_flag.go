package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FlagType defines the type of feature flag
type FlagType string

const (
	FlagTypeBoolean FlagType = "boolean"
	FlagTypeJSON    FlagType = "json"
	FlagTypeString  FlagType = "string"
	FlagTypeNumber  FlagType = "number"
)

// LifecycleState defines the lifecycle state of a flag
type LifecycleState string

const (
	LifecycleStateDraft      LifecycleState = "draft"
	LifecycleStateActive     LifecycleState = "active"
	LifecycleStateDeprecated LifecycleState = "deprecated"
	LifecycleStateArchived   LifecycleState = "archived"
)

// ScheduleRecurrence defines recurring schedule for flags
type ScheduleRecurrence struct {
	Type       string `json:"type"`                   // "daily", "weekly", "monthly", "cron"
	DaysOfWeek []int  `json:"days_of_week,omitempty"` // 0=Sunday, 6=Saturday
	TimeStart  string `json:"time_start,omitempty"`   // "09:00"
	TimeEnd    string `json:"time_end,omitempty"`     // "18:00"
	CronExpr   string `json:"cron_expr,omitempty"`    // For cron type
}

// FeatureFlag represents a feature flag
type FeatureFlag struct {
	ID            uuid.UUID       `json:"id"`
	EnvironmentID uuid.UUID       `json:"environment_id"`
	Key           string          `json:"key"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Type          FlagType        `json:"type"`
	FlagType      FlagType        `json:"-"` // Alias for Type for backward compatibility
	Enabled       bool            `json:"enabled"`
	Value         json.RawMessage `json:"value"`
	DefaultValue  json.RawMessage `json:"default_value"`
	Version       int             `json:"version"`
	Tags          []string        `json:"tags"`
	// Scheduling fields
	ScheduledEnableAt  *time.Time          `json:"scheduled_enable_at,omitempty"`
	ScheduledDisableAt *time.Time          `json:"scheduled_disable_at,omitempty"`
	ScheduleTimezone   string              `json:"schedule_timezone,omitempty"`
	ScheduleRecurrence *ScheduleRecurrence `json:"schedule_recurrence,omitempty"`
	// Lifecycle
	LifecycleState LifecycleState `json:"lifecycle_state"`
	// Tracking
	LastEvaluatedAt *time.Time `json:"last_evaluated_at,omitempty"`
	// Dependencies
	PrerequisiteFlagID *uuid.UUID `json:"prerequisite_flag_id,omitempty"`
	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewFeatureFlag creates a new feature flag
func NewFeatureFlag(environmentID uuid.UUID, key, name, description string, flagType FlagType, defaultValue json.RawMessage) *FeatureFlag {
	now := time.Now()
	return &FeatureFlag{
		ID:               uuid.New(),
		EnvironmentID:    environmentID,
		Key:              key,
		Name:             name,
		Description:      description,
		Type:             flagType,
		FlagType:         flagType,
		Enabled:          false,
		Value:            defaultValue,
		DefaultValue:     defaultValue,
		Version:          1,
		Tags:             []string{},
		LifecycleState:   LifecycleStateActive,
		ScheduleTimezone: "UTC",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// Enable enables the feature flag
func (f *FeatureFlag) Enable() {
	f.Enabled = true
	f.Version++
	f.UpdatedAt = time.Now()
}

// Disable disables the feature flag
func (f *FeatureFlag) Disable() {
	f.Enabled = false
	f.Version++
	f.UpdatedAt = time.Now()
}

// Toggle toggles the feature flag state
func (f *FeatureFlag) Toggle() {
	f.Enabled = !f.Enabled
	f.Version++
	f.UpdatedAt = time.Now()
}

// Update updates feature flag details
func (f *FeatureFlag) Update(name, description string, defaultValue json.RawMessage, tags []string) {
	if name != "" {
		f.Name = name
	}
	if description != "" {
		f.Description = description
	}
	if defaultValue != nil {
		f.DefaultValue = defaultValue
		f.Value = defaultValue
	}

	if tags != nil {
		f.Tags = tags
	}
	f.Version++
	f.UpdatedAt = time.Now()
}

// SoftDelete marks the flag as deleted
func (f *FeatureFlag) SoftDelete() {
	now := time.Now()
	f.DeletedAt = &now
}

// IsDeleted checks if flag is soft deleted
func (f *FeatureFlag) IsDeleted() bool {
	return f.DeletedAt != nil
}

// GetBoolValue returns the default value as boolean
func (f *FeatureFlag) GetBoolValue() bool {
	if f.FlagType != FlagTypeBoolean {
		return false
	}
	var value bool
	json.Unmarshal(f.DefaultValue, &value)
	return value
}

// GetStringValue returns the default value as string
func (f *FeatureFlag) GetStringValue() string {
	if f.FlagType != FlagTypeString {
		return ""
	}
	var value string
	json.Unmarshal(f.DefaultValue, &value)
	return value
}

// GetNumberValue returns the default value as float64
func (f *FeatureFlag) GetNumberValue() float64 {
	if f.FlagType != FlagTypeNumber {
		return 0
	}
	var value float64
	json.Unmarshal(f.DefaultValue, &value)
	return value
}

// GetJSONValue returns the default value as map
func (f *FeatureFlag) GetJSONValue() map[string]interface{} {
	if f.FlagType != FlagTypeJSON {
		return nil
	}
	var value map[string]interface{}
	json.Unmarshal(f.DefaultValue, &value)
	return value
}

// FlagEvaluation represents the result of evaluating a flag
type FlagEvaluation struct {
	Key       string          `json:"key"`
	Enabled   bool            `json:"enabled"`
	Value     json.RawMessage `json:"value"`
	Version   int             `json:"version"`
	Reason    string          `json:"reason"`
	RuleID    *uuid.UUID      `json:"rule_id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// NewFlagEvaluation creates a new flag evaluation result
func NewFlagEvaluation(flag *FeatureFlag, value json.RawMessage, reason string, ruleID *uuid.UUID) *FlagEvaluation {
	return &FlagEvaluation{
		Key:       flag.Key,
		Enabled:   flag.Enabled,
		Value:     value,
		Version:   flag.Version,
		Reason:    reason,
		RuleID:    ruleID,
		Timestamp: time.Now(),
	}
}
