package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ExperimentStatus defines the status of an experiment
type ExperimentStatus string

const (
	ExperimentStatusDraft     ExperimentStatus = "draft"
	ExperimentStatusRunning   ExperimentStatus = "running"
	ExperimentStatusPaused    ExperimentStatus = "paused"
	ExperimentStatusCompleted ExperimentStatus = "completed"
	ExperimentStatusCancelled ExperimentStatus = "cancelled"
)

// Experiment represents an A/B test experiment
type Experiment struct {
	ID                      uuid.UUID        `json:"id"`
	TenantID                uuid.UUID        `json:"tenant_id"`
	EnvironmentID           uuid.UUID        `json:"environment_id"`
	FeatureFlagID           uuid.UUID        `json:"feature_flag_id"`
	Name                    string           `json:"name"`
	Description             string           `json:"description,omitempty"`
	Hypothesis              string           `json:"hypothesis,omitempty"`
	Status                  ExperimentStatus `json:"status"`
	StartedAt               *time.Time       `json:"started_at,omitempty"`
	EndedAt                 *time.Time       `json:"ended_at,omitempty"`
	TargetSampleSize        *int             `json:"target_sample_size,omitempty"`
	CurrentSampleSize       int              `json:"current_sample_size"`
	WinnerVariant           string           `json:"winner_variant,omitempty"`
	StatisticalSignificance *float64         `json:"statistical_significance,omitempty"`
	CreatedBy               *uuid.UUID       `json:"created_by,omitempty"`
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at"`
	// Populated by joins
	Variants []ExperimentVariant `json:"variants,omitempty"`
	Metrics  []ExperimentMetric  `json:"metrics,omitempty"`
}

// NewExperiment creates a new experiment
func NewExperiment(tenantID, environmentID, flagID uuid.UUID, name, description, hypothesis string, createdBy *uuid.UUID) *Experiment {
	now := time.Now()
	return &Experiment{
		ID:                uuid.New(),
		TenantID:          tenantID,
		EnvironmentID:     environmentID,
		FeatureFlagID:     flagID,
		Name:              name,
		Description:       description,
		Hypothesis:        hypothesis,
		Status:            ExperimentStatusDraft,
		CurrentSampleSize: 0,
		CreatedBy:         createdBy,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// Start starts the experiment
func (e *Experiment) Start() {
	now := time.Now()
	e.Status = ExperimentStatusRunning
	e.StartedAt = &now
	e.UpdatedAt = now
}

// Pause pauses the experiment
func (e *Experiment) Pause() {
	e.Status = ExperimentStatusPaused
	e.UpdatedAt = time.Now()
}

// Resume resumes the experiment
func (e *Experiment) Resume() {
	e.Status = ExperimentStatusRunning
	e.UpdatedAt = time.Now()
}

// Complete marks the experiment as completed
func (e *Experiment) Complete(winnerVariant string, significance float64) {
	now := time.Now()
	e.Status = ExperimentStatusCompleted
	e.EndedAt = &now
	e.WinnerVariant = winnerVariant
	e.StatisticalSignificance = &significance
	e.UpdatedAt = now
}

// Cancel cancels the experiment
func (e *Experiment) Cancel() {
	now := time.Now()
	e.Status = ExperimentStatusCancelled
	e.EndedAt = &now
	e.UpdatedAt = now
}

// IncrementSampleSize increments the sample count
func (e *Experiment) IncrementSampleSize(count int) {
	e.CurrentSampleSize += count
	e.UpdatedAt = time.Now()
}

// IsRunning checks if experiment is running
func (e *Experiment) IsRunning() bool {
	return e.Status == ExperimentStatusRunning
}

// ExperimentVariant represents a variant in an experiment
type ExperimentVariant struct {
	ID           uuid.UUID       `json:"id"`
	ExperimentID uuid.UUID       `json:"experiment_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	Value        json.RawMessage `json:"value"`
	Weight       int             `json:"weight"` // Percentage of traffic
	IsControl    bool            `json:"is_control"`
	CreatedAt    time.Time       `json:"created_at"`
}

// NewExperimentVariant creates a new experiment variant
func NewExperimentVariant(experimentID uuid.UUID, name, description string, value json.RawMessage, weight int, isControl bool) *ExperimentVariant {
	return &ExperimentVariant{
		ID:           uuid.New(),
		ExperimentID: experimentID,
		Name:         name,
		Description:  description,
		Value:        value,
		Weight:       weight,
		IsControl:    isControl,
		CreatedAt:    time.Now(),
	}
}

// MetricType defines the type of metric
type MetricType string

const (
	MetricTypeConversion MetricType = "conversion"
	MetricTypeCount      MetricType = "count"
	MetricTypeSum        MetricType = "sum"
	MetricTypeAverage    MetricType = "average"
)

// GoalDirection defines whether we want to increase or decrease the metric
type GoalDirection string

const (
	GoalDirectionIncrease GoalDirection = "increase"
	GoalDirectionDecrease GoalDirection = "decrease"
)

// ExperimentMetric represents a metric to track in an experiment
type ExperimentMetric struct {
	ID            uuid.UUID     `json:"id"`
	ExperimentID  uuid.UUID     `json:"experiment_id"`
	Name          string        `json:"name"`
	MetricType    MetricType    `json:"metric_type"`
	IsPrimary     bool          `json:"is_primary"`
	GoalDirection GoalDirection `json:"goal_direction,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// NewExperimentMetric creates a new experiment metric
func NewExperimentMetric(experimentID uuid.UUID, name string, metricType MetricType, isPrimary bool, goalDirection GoalDirection) *ExperimentMetric {
	return &ExperimentMetric{
		ID:            uuid.New(),
		ExperimentID:  experimentID,
		Name:          name,
		MetricType:    metricType,
		IsPrimary:     isPrimary,
		GoalDirection: goalDirection,
		CreatedAt:     time.Now(),
	}
}

// ExperimentResult represents the results for a variant-metric pair
type ExperimentResult struct {
	ID                     uuid.UUID `json:"id"`
	ExperimentID           uuid.UUID `json:"experiment_id"`
	VariantID              uuid.UUID `json:"variant_id"`
	MetricID               uuid.UUID `json:"metric_id"`
	SampleCount            int       `json:"sample_count"`
	ConversionCount        int       `json:"conversion_count"`
	SumValue               float64   `json:"sum_value"`
	MeanValue              *float64  `json:"mean_value,omitempty"`
	Variance               *float64  `json:"variance,omitempty"`
	ConfidenceIntervalLow  *float64  `json:"confidence_interval_low,omitempty"`
	ConfidenceIntervalHigh *float64  `json:"confidence_interval_high,omitempty"`
	PValue                 *float64  `json:"p_value,omitempty"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// NewExperimentResult creates a new experiment result
func NewExperimentResult(experimentID, variantID, metricID uuid.UUID) *ExperimentResult {
	return &ExperimentResult{
		ID:              uuid.New(),
		ExperimentID:    experimentID,
		VariantID:       variantID,
		MetricID:        metricID,
		SampleCount:     0,
		ConversionCount: 0,
		SumValue:        0,
		UpdatedAt:       time.Now(),
	}
}

// ConversionRate calculates the conversion rate
func (r *ExperimentResult) ConversionRate() float64 {
	if r.SampleCount == 0 {
		return 0
	}
	return float64(r.ConversionCount) / float64(r.SampleCount)
}

// AddSample adds a sample to the result
func (r *ExperimentResult) AddSample(value float64, isConversion bool) {
	r.SampleCount++
	r.SumValue += value
	if isConversion {
		r.ConversionCount++
	}
	// Update mean
	mean := r.SumValue / float64(r.SampleCount)
	r.MeanValue = &mean
	r.UpdatedAt = time.Now()
}
