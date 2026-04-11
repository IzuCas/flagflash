package entity

import (
	"time"

	"github.com/google/uuid"
)

// EvaluationEvent represents a single flag evaluation event
type EvaluationEvent struct {
	ID            uuid.UUID              `json:"id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	EnvironmentID uuid.UUID              `json:"environment_id"`
	FeatureFlagID uuid.UUID              `json:"feature_flag_id"`
	FlagKey       string                 `json:"flag_key"`
	Value         interface{}            `json:"value"`
	UserID        *string                `json:"user_id,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	SDKType       *string                `json:"sdk_type,omitempty"`
	SDKVersion    *string                `json:"sdk_version,omitempty"`
	EvaluatedAt   time.Time              `json:"evaluated_at"`
}

// EvaluationSummary represents pre-aggregated hourly stats
type EvaluationSummary struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	EnvironmentID    uuid.UUID `json:"environment_id"`
	FeatureFlagID    uuid.UUID `json:"feature_flag_id"`
	FlagKey          string    `json:"flag_key"`
	HourBucket       time.Time `json:"hour_bucket"`
	TotalEvaluations int64     `json:"total_evaluations"`
	TrueCount        int64     `json:"true_count"`
	FalseCount       int64     `json:"false_count"`
	UniqueUsers      int       `json:"unique_users"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// UsageMetrics represents aggregated usage statistics
type UsageMetrics struct {
	TenantID         uuid.UUID            `json:"tenant_id"`
	Period           string               `json:"period"` // hour, day, week, month
	StartDate        time.Time            `json:"start_date"`
	EndDate          time.Time            `json:"end_date"`
	TotalEvaluations int64                `json:"total_evaluations"`
	UniqueFlags      int                  `json:"unique_flags"`
	UniqueUsers      int                  `json:"unique_users"`
	ByEnvironment    []EnvironmentMetrics `json:"by_environment,omitempty"`
	ByFlag           []FlagMetrics        `json:"by_flag,omitempty"`
	Timeline         []TimelinePoint      `json:"timeline,omitempty"`
}

// EnvironmentMetrics represents metrics per environment
type EnvironmentMetrics struct {
	EnvironmentID   uuid.UUID `json:"environment_id"`
	EnvironmentName string    `json:"environment_name"`
	Evaluations     int64     `json:"evaluations"`
	UniqueFlags     int       `json:"unique_flags"`
	UniqueUsers     int       `json:"unique_users"`
}

// FlagMetrics represents metrics per feature flag
type FlagMetrics struct {
	FlagID          uuid.UUID `json:"flag_id"`
	FlagKey         string    `json:"flag_key"`
	FlagName        string    `json:"flag_name"`
	EnvironmentID   uuid.UUID `json:"environment_id"`
	EnvironmentName string    `json:"environment_name"`
	Evaluations     int64     `json:"evaluations"`
	TrueCount       int64     `json:"true_count"`
	FalseCount      int64     `json:"false_count"`
	UniqueUsers     int       `json:"unique_users"`
}

// TimelinePoint represents a data point in a time series
type TimelinePoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Evaluations int64     `json:"evaluations"`
	TrueCount   int64     `json:"true_count"`
	FalseCount  int64     `json:"false_count"`
}
