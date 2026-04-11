package entity

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Operator defines the condition operator
type Operator string

const (
	OperatorEquals       Operator = "equals"
	OperatorNotEquals    Operator = "not_equals"
	OperatorContains     Operator = "contains"
	OperatorNotContains  Operator = "not_contains"
	OperatorStartsWith   Operator = "starts_with"
	OperatorEndsWith     Operator = "ends_with"
	OperatorIn           Operator = "in"
	OperatorNotIn        Operator = "not_in"
	OperatorGreaterThan  Operator = "greater_than"
	OperatorLessThan     Operator = "less_than"
	OperatorGreaterEqual Operator = "greater_equal"
	OperatorLessEqual    Operator = "less_equal"
	OperatorRegex        Operator = "regex"
	OperatorSemverGT     Operator = "semver_gt"
	OperatorSemverLT     Operator = "semver_lt"
	OperatorSemverEQ     Operator = "semver_eq"
)

// Condition represents a targeting condition
type Condition struct {
	Attribute string      `json:"attribute"`
	Operator  Operator    `json:"operator"`
	Value     interface{} `json:"value"`
}

// TargetingRule represents a targeting rule for a feature flag
type TargetingRule struct {
	ID            uuid.UUID       `json:"id"`
	FeatureFlagID uuid.UUID       `json:"feature_flag_id"`
	FlagID        uuid.UUID       `json:"-"` // Alias for FeatureFlagID
	Name          string          `json:"name"`
	Priority      int             `json:"priority"`
	Conditions    []Condition     `json:"conditions"`
	Value         json.RawMessage `json:"value"`
	Percentage    int             `json:"percentage"` // 0-100 for rollout
	Enabled       bool            `json:"enabled"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// EvaluationContext represents the context for evaluating flags
type EvaluationContext struct {
	UserID     string                 `json:"user_id,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Country    string                 `json:"country,omitempty"`
	Region     string                 `json:"region,omitempty"`
	City       string                 `json:"city,omitempty"`
	Version    string                 `json:"version,omitempty"`
	Platform   string                 `json:"platform,omitempty"`
	DeviceType string                 `json:"device_type,omitempty"`
	Custom     map[string]interface{} `json:"custom,omitempty"`
}

// NewTargetingRule creates a new targeting rule
func NewTargetingRule(flagID uuid.UUID, name string, priority int, conditions []Condition, value json.RawMessage, percentage int) *TargetingRule {
	now := time.Now()
	return &TargetingRule{
		ID:            uuid.New(),
		FeatureFlagID: flagID,
		FlagID:        flagID,
		Name:          name,
		Priority:      priority,
		Conditions:    conditions,
		Value:         value,
		Percentage:    percentage,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Update updates the targeting rule
func (r *TargetingRule) Update(name string, priority int, conditions []Condition, value json.RawMessage, percentage int, enabled *bool) {
	if name != "" {
		r.Name = name
	}
	if priority >= 0 {
		r.Priority = priority
	}
	if conditions != nil {
		r.Conditions = conditions
	}
	if value != nil {
		r.Value = value
	}
	if percentage >= 0 && percentage <= 100 {
		r.Percentage = percentage
	}
	if enabled != nil {
		r.Enabled = *enabled
	}
	r.UpdatedAt = time.Now()
}

// Evaluate checks if the context matches the targeting rule
func (r *TargetingRule) Evaluate(ctx *EvaluationContext) bool {
	if !r.Enabled {
		return false
	}

	// All conditions must match (AND logic)
	for _, condition := range r.Conditions {
		if !r.evaluateCondition(condition, ctx) {
			return false
		}
	}

	// Check percentage-based rollout
	if r.Percentage < 100 {
		return r.isInRolloutBucket(ctx)
	}

	return true
}

// evaluateCondition evaluates a single condition
func (r *TargetingRule) evaluateCondition(condition Condition, ctx *EvaluationContext) bool {
	value := r.getAttributeValue(condition.Attribute, ctx)
	if value == nil {
		return false
	}

	switch condition.Operator {
	case OperatorEquals:
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", condition.Value)
	case OperatorNotEquals:
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", condition.Value)
	case OperatorContains:
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", value)), strings.ToLower(fmt.Sprintf("%v", condition.Value)))
	case OperatorNotContains:
		return !strings.Contains(strings.ToLower(fmt.Sprintf("%v", value)), strings.ToLower(fmt.Sprintf("%v", condition.Value)))
	case OperatorStartsWith:
		return strings.HasPrefix(strings.ToLower(fmt.Sprintf("%v", value)), strings.ToLower(fmt.Sprintf("%v", condition.Value)))
	case OperatorEndsWith:
		return strings.HasSuffix(strings.ToLower(fmt.Sprintf("%v", value)), strings.ToLower(fmt.Sprintf("%v", condition.Value)))
	case OperatorIn:
		return r.valueInList(value, condition.Value)
	case OperatorNotIn:
		return !r.valueInList(value, condition.Value)
	case OperatorRegex:
		pattern := fmt.Sprintf("%v", condition.Value)
		matched, _ := regexp.MatchString(pattern, fmt.Sprintf("%v", value))
		return matched
	case OperatorGreaterThan, OperatorLessThan, OperatorGreaterEqual, OperatorLessEqual:
		return r.compareNumbers(value, condition.Value, condition.Operator)
	}

	return false
}

// getAttributeValue gets the value of an attribute from the context
func (r *TargetingRule) getAttributeValue(attribute string, ctx *EvaluationContext) interface{} {
	switch attribute {
	case "user_id":
		return ctx.UserID
	case "email":
		return ctx.Email
	case "country":
		return ctx.Country
	case "region":
		return ctx.Region
	case "city":
		return ctx.City
	case "version":
		return ctx.Version
	case "platform":
		return ctx.Platform
	case "device_type":
		return ctx.DeviceType
	default:
		if ctx.Custom != nil {
			if val, ok := ctx.Custom[attribute]; ok {
				return val
			}
		}
	}
	return nil
}

// valueInList checks if value is in a list
func (r *TargetingRule) valueInList(value, list interface{}) bool {
	strValue := strings.ToLower(fmt.Sprintf("%v", value))
	switch v := list.(type) {
	case []interface{}:
		for _, item := range v {
			if strings.ToLower(fmt.Sprintf("%v", item)) == strValue {
				return true
			}
		}
	case []string:
		for _, item := range v {
			if strings.ToLower(item) == strValue {
				return true
			}
		}
	}
	return false
}

// compareNumbers compares numeric values
func (r *TargetingRule) compareNumbers(value, target interface{}, op Operator) bool {
	v, ok1 := toFloat64(value)
	t, ok2 := toFloat64(target)
	if !ok1 || !ok2 {
		return false
	}
	switch op {
	case OperatorGreaterThan:
		return v > t
	case OperatorLessThan:
		return v < t
	case OperatorGreaterEqual:
		return v >= t
	case OperatorLessEqual:
		return v <= t
	}
	return false
}

// isInRolloutBucket determines if the context falls within the percentage rollout
func (r *TargetingRule) isInRolloutBucket(ctx *EvaluationContext) bool {
	// Use user_id as the primary bucketing key
	key := ctx.UserID
	if key == "" {
		key = ctx.Email
	}
	if key == "" {
		return false // Can't do percentage rollout without a consistent identifier
	}

	// Create a hash to get a consistent bucket for the user
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", r.FeatureFlagID.String(), key)))
	bucket := binary.BigEndian.Uint32(hash[:4]) % 100

	return int(bucket) < r.Percentage
}

// toFloat64 converts interface to float64
func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	}
	return 0, false
}
