package sdk

import "encoding/json"

// ─── Public flag types ──────────────────────────────────────────

// EvalResult holds the result of evaluating a single feature flag.
type EvalResult struct {
	FlagKey   string          `json:"flag_key"`
	Value     json.RawMessage `json:"value"`
	Enabled   bool            `json:"enabled"`
	Version   int             `json:"version"`
	RuleID    *string         `json:"rule_id,omitempty"`
	RuleName  string          `json:"rule_name,omitempty"`
	FromCache bool            `json:"-"` // true when served from local cache
}

// BoolValue returns the flag value as bool, falling back to defaultVal.
func (r *EvalResult) BoolValue(defaultVal bool) bool {
	if !r.Enabled || r.Value == nil {
		return defaultVal
	}
	var v bool
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return defaultVal
	}
	return v
}

// StringValue returns the flag value as string, falling back to defaultVal.
func (r *EvalResult) StringValue(defaultVal string) string {
	if !r.Enabled || r.Value == nil {
		return defaultVal
	}
	var v string
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return defaultVal
	}
	return v
}

// Float64Value returns the flag value as float64, falling back to defaultVal.
func (r *EvalResult) Float64Value(defaultVal float64) float64 {
	if !r.Enabled || r.Value == nil {
		return defaultVal
	}
	var v float64
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return defaultVal
	}
	return v
}

// IntValue returns the flag value as int, falling back to defaultVal.
func (r *EvalResult) IntValue(defaultVal int) int {
	if !r.Enabled || r.Value == nil {
		return defaultVal
	}
	var v int
	if err := json.Unmarshal(r.Value, &v); err != nil {
		return defaultVal
	}
	return v
}

// JSONValue unmarshals the flag value into dest.
func (r *EvalResult) JSONValue(dest any) error {
	return json.Unmarshal(r.Value, dest)
}

// AllFlagsResult maps flag key → EvalResult for bulk evaluation.
type AllFlagsResult map[string]*EvalResult

// Get returns the EvalResult for a flag key, or an empty disabled result if missing.
func (a AllFlagsResult) Get(key string) *EvalResult {
	if r, ok := a[key]; ok {
		return r
	}
	return &EvalResult{FlagKey: key}
}

// Flag is a lightweight descriptor (no evaluation, from GetFlags).
type Flag struct {
	Key          string          `json:"key"`
	Type         string          `json:"type"`
	Enabled      bool            `json:"enabled"`
	DefaultValue json.RawMessage `json:"default_value"`
	Version      int             `json:"version"`
}

// EvaluationContext carries arbitrary attributes for targeting rule evaluation.
type EvaluationContext map[string]any
