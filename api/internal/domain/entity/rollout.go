package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RolloutStatus defines the status of a rollout plan
type RolloutStatus string

const (
	RolloutStatusDraft     RolloutStatus = "draft"
	RolloutStatusActive    RolloutStatus = "active"
	RolloutStatusPaused    RolloutStatus = "paused"
	RolloutStatusCompleted RolloutStatus = "completed"
	RolloutStatusFailed    RolloutStatus = "failed"
)

// RolloutPlan represents a progressive rollout plan
type RolloutPlan struct {
	ID                         uuid.UUID     `json:"id"`
	FeatureFlagID              uuid.UUID     `json:"feature_flag_id"`
	Name                       string        `json:"name"`
	Status                     RolloutStatus `json:"status"`
	CurrentPercentage          int           `json:"current_percentage"`
	TargetPercentage           int           `json:"target_percentage"`
	IncrementPercentage        int           `json:"increment_percentage"`
	IncrementIntervalMinutes   int           `json:"increment_interval_minutes"`
	AutoRollback               bool          `json:"auto_rollback"`
	RollbackThresholdErrorRate *float64      `json:"rollback_threshold_error_rate,omitempty"`
	RollbackThresholdLatencyMs *int          `json:"rollback_threshold_latency_ms,omitempty"`
	LastIncrementAt            *time.Time    `json:"last_increment_at,omitempty"`
	NextIncrementAt            *time.Time    `json:"next_increment_at,omitempty"`
	CreatedBy                  *uuid.UUID    `json:"created_by,omitempty"`
	CreatedAt                  time.Time     `json:"created_at"`
	UpdatedAt                  time.Time     `json:"updated_at"`
}

// NewRolloutPlan creates a new rollout plan
func NewRolloutPlan(flagID uuid.UUID, name string, targetPct, incrementPct, intervalMinutes int, createdBy *uuid.UUID) *RolloutPlan {
	now := time.Now()
	return &RolloutPlan{
		ID:                       uuid.New(),
		FeatureFlagID:            flagID,
		Name:                     name,
		Status:                   RolloutStatusDraft,
		CurrentPercentage:        0,
		TargetPercentage:         targetPct,
		IncrementPercentage:      incrementPct,
		IncrementIntervalMinutes: intervalMinutes,
		AutoRollback:             true,
		CreatedBy:                createdBy,
		CreatedAt:                now,
		UpdatedAt:                now,
	}
}

// Start starts the rollout
func (r *RolloutPlan) Start() {
	now := time.Now()
	r.Status = RolloutStatusActive
	r.LastIncrementAt = &now
	nextIncrement := now.Add(time.Duration(r.IncrementIntervalMinutes) * time.Minute)
	r.NextIncrementAt = &nextIncrement
	r.UpdatedAt = now
}

// Pause pauses the rollout
func (r *RolloutPlan) Pause() {
	r.Status = RolloutStatusPaused
	r.NextIncrementAt = nil
	r.UpdatedAt = time.Now()
}

// Resume resumes the rollout
func (r *RolloutPlan) Resume() {
	now := time.Now()
	r.Status = RolloutStatusActive
	nextIncrement := now.Add(time.Duration(r.IncrementIntervalMinutes) * time.Minute)
	r.NextIncrementAt = &nextIncrement
	r.UpdatedAt = now
}

// Increment increments the rollout percentage
func (r *RolloutPlan) Increment() bool {
	if r.CurrentPercentage >= r.TargetPercentage {
		r.Complete()
		return false
	}

	now := time.Now()
	r.CurrentPercentage += r.IncrementPercentage
	if r.CurrentPercentage > r.TargetPercentage {
		r.CurrentPercentage = r.TargetPercentage
	}

	r.LastIncrementAt = &now

	if r.CurrentPercentage >= r.TargetPercentage {
		r.Complete()
	} else {
		nextIncrement := now.Add(time.Duration(r.IncrementIntervalMinutes) * time.Minute)
		r.NextIncrementAt = &nextIncrement
	}

	r.UpdatedAt = now
	return true
}

// Rollback rolls back the percentage
func (r *RolloutPlan) Rollback(reason string) {
	r.CurrentPercentage = 0
	r.Status = RolloutStatusFailed
	r.NextIncrementAt = nil
	r.UpdatedAt = time.Now()
}

// Complete marks the rollout as completed
func (r *RolloutPlan) Complete() {
	r.Status = RolloutStatusCompleted
	r.CurrentPercentage = r.TargetPercentage
	r.NextIncrementAt = nil
	r.UpdatedAt = time.Now()
}

// IsActive checks if the rollout is active
func (r *RolloutPlan) IsActive() bool {
	return r.Status == RolloutStatusActive
}

// NeedsIncrement checks if it's time for an increment
func (r *RolloutPlan) NeedsIncrement() bool {
	if !r.IsActive() || r.NextIncrementAt == nil {
		return false
	}
	return time.Now().After(*r.NextIncrementAt)
}

// RolloutAction defines the type of rollout action
type RolloutAction string

const (
	RolloutActionIncrement RolloutAction = "increment"
	RolloutActionRollback  RolloutAction = "rollback"
	RolloutActionPause     RolloutAction = "pause"
	RolloutActionResume    RolloutAction = "resume"
	RolloutActionComplete  RolloutAction = "complete"
)

// RolloutHistory represents a rollout history entry
type RolloutHistory struct {
	ID             uuid.UUID       `json:"id"`
	RolloutPlanID  uuid.UUID       `json:"rollout_plan_id"`
	FromPercentage int             `json:"from_percentage"`
	ToPercentage   int             `json:"to_percentage"`
	Action         RolloutAction   `json:"action"`
	Reason         string          `json:"reason,omitempty"`
	Metrics        json.RawMessage `json:"metrics,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// NewRolloutHistory creates a new rollout history entry
func NewRolloutHistory(planID uuid.UUID, fromPct, toPct int, action RolloutAction, reason string, metrics json.RawMessage) *RolloutHistory {
	return &RolloutHistory{
		ID:             uuid.New(),
		RolloutPlanID:  planID,
		FromPercentage: fromPct,
		ToPercentage:   toPct,
		Action:         action,
		Reason:         reason,
		Metrics:        metrics,
		CreatedAt:      time.Now(),
	}
}
