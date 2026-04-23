package sdk

import (
	"encoding/json"
	"os"
	"strconv"
)

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

// BoolValueFromEnv returns the flag value as bool when the flag is enabled.
// If not, it reads the environment variable named envName and parses it as a bool.
// If the env var is absent or cannot be parsed, defaultVal is returned.
func (r *EvalResult) BoolValueFromEnv(envName string, defaultVal bool) bool {
	if r.Enabled && r.Value != nil {
		var v bool
		if err := json.Unmarshal(r.Value, &v); err == nil {
			return v
		}
	}
	if raw, ok := os.LookupEnv(envName); ok {
		if v, err := strconv.ParseBool(raw); err == nil {
			return v
		}
	}
	return defaultVal
}

// StringValueFromEnv returns the flag value as string when the flag is enabled.
// If not, it reads the environment variable named envName.
// If the env var is absent, defaultVal is returned.
func (r *EvalResult) StringValueFromEnv(envName string, defaultVal string) string {
	if r.Enabled && r.Value != nil {
		var v string
		if err := json.Unmarshal(r.Value, &v); err == nil {
			return v
		}
	}
	if raw, ok := os.LookupEnv(envName); ok {
		return raw
	}
	return defaultVal
}

// Float64ValueFromEnv returns the flag value as float64 when the flag is enabled.
// If not, it reads the environment variable named envName and parses it as float64.
// If the env var is absent or cannot be parsed, defaultVal is returned.
func (r *EvalResult) Float64ValueFromEnv(envName string, defaultVal float64) float64 {
	if r.Enabled && r.Value != nil {
		var v float64
		if err := json.Unmarshal(r.Value, &v); err == nil {
			return v
		}
	}
	if raw, ok := os.LookupEnv(envName); ok {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			return v
		}
	}
	return defaultVal
}

// IntValueFromEnv returns the flag value as int when the flag is enabled.
// If not, it reads the environment variable named envName and parses it as int.
// If the env var is absent or cannot be parsed, defaultVal is returned.
func (r *EvalResult) IntValueFromEnv(envName string, defaultVal int) int {
	if r.Enabled && r.Value != nil {
		var v int
		if err := json.Unmarshal(r.Value, &v); err == nil {
			return v
		}
	}
	if raw, ok := os.LookupEnv(envName); ok {
		if v, err := strconv.Atoi(raw); err == nil {
			return v
		}
	}
	return defaultVal
}

// BoolFromEnv returns the flag value as bool when the flag is enabled.
// If not, it reads the environment variable named envName and parses it as bool.
// Returns false if the env var is absent or cannot be parsed.
func (r *EvalResult) BoolFromEnv(envName string) bool {
	if r.Enabled && r.Value != nil {
		var v bool
		if err := json.Unmarshal(r.Value, &v); err == nil {
			return v
		}
	}
	if raw, ok := os.LookupEnv(envName); ok {
		if v, err := strconv.ParseBool(raw); err == nil {
			return v
		}
	}
	return false
}

// StringFromEnv returns the flag value as string when the flag is enabled.
// If not, it reads the environment variable named envName.
// Returns "" if the env var is absent.
func (r *EvalResult) StringFromEnv(envName string) string {
	if r.Enabled && r.Value != nil {
		var v string
		if err := json.Unmarshal(r.Value, &v); err == nil {
			return v
		}
	}
	return os.Getenv(envName)
}

// Float64FromEnv returns the flag value as float64 when the flag is enabled.
// If not, it reads the environment variable named envName and parses it as float64.
// Returns 0 if the env var is absent or cannot be parsed.
func (r *EvalResult) Float64FromEnv(envName string) float64 {
	if r.Enabled && r.Value != nil {
		var v float64
		if err := json.Unmarshal(r.Value, &v); err == nil {
			return v
		}
	}
	if raw, ok := os.LookupEnv(envName); ok {
		if v, err := strconv.ParseFloat(raw, 64); err == nil {
			return v
		}
	}
	return 0
}

// IntFromEnv returns the flag value as int when the flag is enabled.
// If not, it reads the environment variable named envName and parses it as int.
// Returns 0 if the env var is absent or cannot be parsed.
func (r *EvalResult) IntFromEnv(envName string) int {
	if r.Enabled && r.Value != nil {
		var v int
		if err := json.Unmarshal(r.Value, &v); err == nil {
			return v
		}
	}
	if raw, ok := os.LookupEnv(envName); ok {
		if v, err := strconv.Atoi(raw); err == nil {
			return v
		}
	}
	return 0
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
