package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// EvaluationService handles flag evaluation logic
type EvaluationService struct {
	flagRepo      repository.FeatureFlagRepository
	targetingRepo repository.TargetingRuleRepository
	cache         FlagCache
}

// NewEvaluationService creates a new evaluation service
func NewEvaluationService(
	flagRepo repository.FeatureFlagRepository,
	targetingRepo repository.TargetingRuleRepository,
	cache FlagCache,
) *EvaluationService {
	return &EvaluationService{
		flagRepo:      flagRepo,
		targetingRepo: targetingRepo,
		cache:         cache,
	}
}

// EvaluationType defines the reason for the evaluation result
type EvaluationType string

const (
	EvaluationTypeDefault   EvaluationType = "DEFAULT"
	EvaluationTypeDisabled  EvaluationType = "DISABLED"
	EvaluationTypeTargeting EvaluationType = "TARGETING"
	EvaluationTypeNotFound  EvaluationType = "NOT_FOUND"
	EvaluationTypeError     EvaluationType = "ERROR"
)

// EvaluationResult represents the result of evaluating a flag
type EvaluationResult struct {
	Key      string          `json:"key"`
	Enabled  bool            `json:"enabled"`
	Value    json.RawMessage `json:"value"`
	Type     EvaluationType  `json:"type"`
	Version  int             `json:"version"`
	RuleID   *uuid.UUID      `json:"rule_id,omitempty"`
	RuleName string          `json:"rule_name,omitempty"`
}

// EvaluateFlag evaluates a single flag with the given context
func (s *EvaluationService) EvaluateFlag(ctx context.Context, environmentID uuid.UUID, key string, evalCtx *entity.EvaluationContext) (*EvaluationResult, error) {
	log.Printf("[EVAL] Evaluating flag %q for env %s, context: %+v", key, environmentID, evalCtx)

	// Get flag from cache or database
	var flag *entity.FeatureFlag
	var err error

	if s.cache != nil {
		flag, err = s.cache.GetFlag(ctx, environmentID, key)
	}

	if flag == nil || err != nil {
		flag, err = s.flagRepo.GetByKey(ctx, environmentID, key)
		if err != nil {
			log.Printf("[EVAL] Flag %q NOT FOUND", key)
			return &EvaluationResult{
				Key:   key,
				Type:  EvaluationTypeNotFound,
				Value: json.RawMessage("null"),
			}, nil
		}

		// Cache the flag
		if s.cache != nil {
			s.cache.SetFlag(ctx, environmentID, flag)
		}
	}

	log.Printf("[EVAL] Flag %q found: Enabled=%v, DefaultValue=%s", key, flag.Enabled, string(flag.DefaultValue))

	// If flag is disabled, return default value
	if !flag.Enabled {
		log.Printf("[EVAL] Flag %q is DISABLED", key)
		return &EvaluationResult{
			Key:     key,
			Enabled: false,
			Value:   flag.DefaultValue,
			Type:    EvaluationTypeDisabled,
			Version: flag.Version,
		}, nil
	}

	// Get targeting rules
	rules, err := s.targetingRepo.ListByFlag(ctx, flag.ID)
	if err != nil {
		log.Printf("[EVAL] Error getting targeting rules for flag %q: %v", key, err)
		return &EvaluationResult{
			Key:     key,
			Enabled: true,
			Value:   flag.DefaultValue,
			Type:    EvaluationTypeDefault,
			Version: flag.Version,
		}, nil
	}

	log.Printf("[EVAL] Found %d targeting rules for flag %q", len(rules), key)

	// Evaluate targeting rules in priority order
	for i, rule := range rules {
		log.Printf("[EVAL] Rule %d: %q (enabled=%v, priority=%d, conditions=%d)", i, rule.Name, rule.Enabled, rule.Priority, len(rule.Conditions))
		for j, cond := range rule.Conditions {
			log.Printf("[EVAL]   Condition %d: %s %s %v", j, cond.Attribute, cond.Operator, cond.Value)
		}

		if rule.Evaluate(evalCtx) {
			log.Printf("[EVAL] Rule %q MATCHED! Returning value: %s", rule.Name, string(rule.Value))
			return &EvaluationResult{
				Key:      key,
				Enabled:  true,
				Value:    rule.Value,
				Type:     EvaluationTypeTargeting,
				Version:  flag.Version,
				RuleID:   &rule.ID,
				RuleName: rule.Name,
			}, nil
		}
		log.Printf("[EVAL] Rule %q did NOT match", rule.Name)
	}

	// No rule matched, return default value
	log.Printf("[EVAL] No rules matched for flag %q, returning default", key)
	return &EvaluationResult{
		Key:     key,
		Enabled: true,
		Value:   flag.DefaultValue,
		Type:    EvaluationTypeDefault,
		Version: flag.Version,
	}, nil
}

// EvaluateAllFlags evaluates all flags for an environment
func (s *EvaluationService) EvaluateAllFlags(ctx context.Context, environmentID uuid.UUID, evalCtx *entity.EvaluationContext) (map[string]*EvaluationResult, error) {
	// Get all flags
	var flags []*entity.FeatureFlag
	var err error

	if s.cache != nil {
		flags, err = s.cache.GetFlags(ctx, environmentID)
	}

	if flags == nil || err != nil {
		flags, err = s.flagRepo.ListByEnvironment(ctx, environmentID, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get flags: %w", err)
		}

		// Cache the flags
		if s.cache != nil {
			s.cache.SetFlags(ctx, environmentID, flags)
		}
	}

	results := make(map[string]*EvaluationResult)

	for _, flag := range flags {
		result, _ := s.EvaluateFlag(ctx, environmentID, flag.Key, evalCtx)
		results[flag.Key] = result
	}

	return results, nil
}

// GetAllFlags returns all flags for an environment without evaluation
func (s *EvaluationService) GetAllFlags(ctx context.Context, environmentID uuid.UUID) ([]*SimpleFlagValue, error) {
	var flags []*entity.FeatureFlag
	var err error

	if s.cache != nil {
		flags, err = s.cache.GetFlags(ctx, environmentID)
	}

	if flags == nil || err != nil {
		flags, err = s.flagRepo.ListByEnvironment(ctx, environmentID, false)
		if err != nil {
			return nil, fmt.Errorf("failed to get flags: %w", err)
		}

		// Cache the flags
		if s.cache != nil {
			s.cache.SetFlags(ctx, environmentID, flags)
		}
	}

	results := make([]*SimpleFlagValue, len(flags))
	for i, flag := range flags {
		results[i] = &SimpleFlagValue{
			Key:          flag.Key,
			Enabled:      flag.Enabled,
			DefaultValue: flag.DefaultValue,
			Type:         string(flag.FlagType),
			Version:      flag.Version,
		}
	}

	return results, nil
}

// SimpleFlagValue represents a simple flag value for SDK consumption
type SimpleFlagValue struct {
	Key          string          `json:"key"`
	Enabled      bool            `json:"enabled"`
	DefaultValue json.RawMessage `json:"default_value"`
	Type         string          `json:"type"`
	Version      int             `json:"version"`
}

// BulkEvaluationRequest represents a request to evaluate multiple flags
type BulkEvaluationRequest struct {
	Keys    []string                  `json:"keys,omitempty"`
	Context *entity.EvaluationContext `json:"context,omitempty"`
}

// EvaluateBulk evaluates multiple specific flags
func (s *EvaluationService) EvaluateBulk(ctx context.Context, environmentID uuid.UUID, req *BulkEvaluationRequest) (map[string]*EvaluationResult, error) {
	results := make(map[string]*EvaluationResult)

	if len(req.Keys) == 0 {
		return s.EvaluateAllFlags(ctx, environmentID, req.Context)
	}

	for _, key := range req.Keys {
		result, _ := s.EvaluateFlag(ctx, environmentID, key, req.Context)
		results[key] = result
	}

	return results, nil
}
