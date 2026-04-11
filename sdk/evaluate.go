package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
)

// ─── wire types (match server JSON) ─────────────────────────────

type getFlagsResp struct {
	Flags []Flag `json:"flags"`
}

// validFlagKey matches keys that are safe to pass to the server.
// Keys must be 1–256 characters of alphanumerics, underscores, or hyphens.
var validFlagKey = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,256}$`)

func validateFlagKey(key string) error {
	if !validFlagKey.MatchString(key) {
		return fmt.Errorf("flagflash: invalid flag key %q — must match [a-zA-Z0-9_-]{1,256}", key)
	}
	return nil
}

// ─── public API ──────────────────────────────────────────────────

// IsEnabled returns true when the named flag is enabled.
//
// After Connect(), this is served from the local in-memory cache with no
// network call. On a cache miss (unlikely after Connect) it falls back to
// a single HTTP request.
func (c *Client) IsEnabled(ctx context.Context, flagKey string) bool {
	if err := validateFlagKey(flagKey); err != nil {
		return false
	}
	if r, ok := c.getFromCache(flagKey); ok {
		return r.Enabled
	}
	// Cache miss — fetch from server and warm the cache entry
	result, err := c.evaluateHTTP(ctx, flagKey, nil)
	if err != nil {
		return false
	}
	c.setInCache(flagKey, result)
	return result.Enabled
}

// Evaluate returns the full EvalResult for a flag.
//
//   - If evalCtx is nil and the cache is warm, the result comes from the local
//     cache — zero network latency.
//   - If evalCtx is provided, the call always goes to the server because
//     targeting rules are evaluated server-side.
func (c *Client) Evaluate(ctx context.Context, flagKey string, evalCtx EvaluationContext) (*EvalResult, error) {
	if err := validateFlagKey(flagKey); err != nil {
		return nil, err
	}
	if evalCtx == nil {
		if r, ok := c.getFromCache(flagKey); ok {
			return r, nil
		}
	}
	return c.evaluateHTTP(ctx, flagKey, evalCtx)
}

// EvaluateAll returns results for every flag in the environment.
//
//   - If evalCtx is nil and the SDK is connected (cache is warm), results come
//     from the local cache — zero network latency.
//   - If evalCtx is provided, the call always goes to the server.
func (c *Client) EvaluateAll(ctx context.Context, evalCtx EvaluationContext) (AllFlagsResult, error) {
	if evalCtx == nil && c.connected {
		return c.allFromCache(), nil
	}
	return c.evaluateAllHTTP(ctx, evalCtx)
}

// GetFlags returns raw flag descriptors for the environment (always HTTP).
// Useful for tooling; prefer IsEnabled / Evaluate for hot paths.
func (c *Client) GetFlags(ctx context.Context) ([]Flag, error) {
	var resp getFlagsResp
	if err := c.get(ctx, "/api/v1/flagflash/sdk/flags", &resp); err != nil {
		return nil, err
	}
	return resp.Flags, nil
}

// ─── HTTP evaluation (fallback / targeting) ───────────────────────

func (c *Client) evaluateHTTP(ctx context.Context, flagKey string, evalCtx EvaluationContext) (*EvalResult, error) {
	var resp struct {
		FlagKey  string          `json:"flag_key"`
		Value    json.RawMessage `json:"value"`
		Enabled  bool            `json:"enabled"`
		Version  int             `json:"version"`
		RuleID   *string         `json:"rule_id,omitempty"`
		RuleName string          `json:"rule_name,omitempty"`
	}
	err := c.post(ctx, "/api/v1/flagflash/sdk/evaluate",
		map[string]any{"flag_key": flagKey, "context": evalCtx}, &resp)
	if err != nil {
		return &EvalResult{FlagKey: flagKey}, err
	}
	return &EvalResult{
		FlagKey:  resp.FlagKey,
		Value:    resp.Value,
		Enabled:  resp.Enabled,
		Version:  resp.Version,
		RuleID:   resp.RuleID,
		RuleName: resp.RuleName,
	}, nil
}

func (c *Client) evaluateAllHTTP(ctx context.Context, evalCtx EvaluationContext) (AllFlagsResult, error) {
	var resp struct {
		Flags map[string]struct {
			Value    json.RawMessage `json:"value"`
			Enabled  bool            `json:"enabled"`
			Version  int             `json:"version"`
			RuleID   *string         `json:"rule_id,omitempty"`
			RuleName string          `json:"rule_name,omitempty"`
		} `json:"flags"`
	}
	if err := c.post(ctx, "/api/v1/flagflash/sdk/evaluate-all",
		map[string]any{"context": evalCtx}, &resp); err != nil {
		return AllFlagsResult{}, err
	}
	result := make(AllFlagsResult, len(resp.Flags))
	for key, f := range resp.Flags {
		result[key] = &EvalResult{
			FlagKey:  key,
			Value:    f.Value,
			Enabled:  f.Enabled,
			Version:  f.Version,
			RuleID:   f.RuleID,
			RuleName: f.RuleName,
		}
	}
	return result, nil
}
