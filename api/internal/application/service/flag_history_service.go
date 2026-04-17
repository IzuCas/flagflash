package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// FlagHistoryService handles flag version history operations
type FlagHistoryService struct {
	historyRepo repository.FlagHistoryRepository
	flagRepo    repository.FeatureFlagRepository
}

// NewFlagHistoryService creates a new flag history service
func NewFlagHistoryService(
	historyRepo repository.FlagHistoryRepository,
	flagRepo repository.FeatureFlagRepository,
) *FlagHistoryService {
	return &FlagHistoryService{
		historyRepo: historyRepo,
		flagRepo:    flagRepo,
	}
}

// RecordChange records a change to a flag
func (s *FlagHistoryService) RecordChange(
	ctx context.Context,
	flag *entity.FeatureFlag,
	changedBy *uuid.UUID,
	changeType entity.FlagChangeType,
	comment string,
	previousFlag *entity.FeatureFlag,
) (*entity.FlagHistory, error) {
	var previousState, newState json.RawMessage

	if previousFlag != nil {
		previousState, _ = json.Marshal(previousFlag)
	}
	newState, _ = json.Marshal(flag)

	history := entity.NewFlagHistory(
		flag.ID,
		flag.Version,
		changeType,
		changedBy,
		previousState,
		newState,
		comment,
	)

	if err := s.historyRepo.Create(ctx, history); err != nil {
		return nil, err
	}

	return history, nil
}

// GetHistory gets the history of a flag
func (s *FlagHistoryService) GetHistory(ctx context.Context, flagID uuid.UUID, limit, offset int) ([]*entity.FlagHistory, int, error) {
	return s.historyRepo.ListByFlagPaginated(ctx, flagID, limit, offset)
}

// GetAllHistory gets all history of a flag (no pagination)
func (s *FlagHistoryService) GetAllHistory(ctx context.Context, flagID uuid.UUID) ([]*entity.FlagHistory, error) {
	return s.historyRepo.ListByFlag(ctx, flagID)
}

// GetHistoryByVersion gets a specific version of a flag
func (s *FlagHistoryService) GetHistoryByVersion(ctx context.Context, flagID uuid.UUID, version int) (*entity.FlagHistory, error) {
	return s.historyRepo.GetByVersion(ctx, flagID, version)
}

// GetLatestHistory gets the latest history entry for a flag
func (s *FlagHistoryService) GetLatestHistory(ctx context.Context, flagID uuid.UUID) (*entity.FlagHistory, error) {
	return s.historyRepo.GetLatestByFlag(ctx, flagID)
}

// Compare compares two versions of a flag
func (s *FlagHistoryService) Compare(ctx context.Context, flagID uuid.UUID, version1, version2 int) (*FlagComparison, error) {
	history1, err := s.historyRepo.GetByVersion(ctx, flagID, version1)
	if err != nil {
		return nil, fmt.Errorf("version %d not found: %w", version1, err)
	}

	history2, err := s.historyRepo.GetByVersion(ctx, flagID, version2)
	if err != nil {
		return nil, fmt.Errorf("version %d not found: %w", version2, err)
	}

	// Parse snapshots
	var state1, state2 map[string]interface{}
	json.Unmarshal(history1.NewState, &state1)
	json.Unmarshal(history2.NewState, &state2)

	// Find differences
	differences := s.findDifferences(state1, state2)

	return &FlagComparison{
		Version1:    history1,
		Version2:    history2,
		Differences: differences,
	}, nil
}

// RestoreVersion restores a flag to a previous version
func (s *FlagHistoryService) RestoreVersion(
	ctx context.Context,
	flagID uuid.UUID,
	version int,
	restoredBy *uuid.UUID,
) (*entity.FeatureFlag, error) {
	history, err := s.historyRepo.GetByVersion(ctx, flagID, version)
	if err != nil {
		return nil, fmt.Errorf("version %d not found: %w", version, err)
	}

	flag, err := s.flagRepo.GetByID(ctx, flagID)
	if err != nil {
		return nil, fmt.Errorf("flag not found: %w", err)
	}

	// Parse snapshot and restore state
	var snapshot entity.FeatureFlag
	if err := json.Unmarshal(history.NewState, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}

	// Store previous state for history
	previousFlag := *flag

	// Restore key fields from snapshot
	flag.Name = snapshot.Name
	flag.Description = snapshot.Description
	flag.Enabled = snapshot.Enabled
	flag.DefaultValue = snapshot.DefaultValue
	flag.Type = snapshot.Type
	flag.Tags = snapshot.Tags
	flag.Version++
	flag.UpdatedAt = time.Now()

	if err := s.flagRepo.Update(ctx, flag); err != nil {
		return nil, fmt.Errorf("failed to update flag: %w", err)
	}

	// Record the restoration as a new history entry
	s.RecordChange(ctx, flag, restoredBy, entity.FlagChangeTypeRollback, fmt.Sprintf("Restored from version %d", version), &previousFlag)

	return flag, nil
}

// GetVersionCount gets the latest version number for a flag
func (s *FlagHistoryService) GetVersionCount(ctx context.Context, flagID uuid.UUID) (int, error) {
	latest, err := s.historyRepo.GetLatestByFlag(ctx, flagID)
	if err != nil {
		return 0, err
	}
	return latest.Version, nil
}

// findDifferences finds differences between two flag states
func (s *FlagHistoryService) findDifferences(state1, state2 map[string]interface{}) []FieldDifference {
	differences := make([]FieldDifference, 0)

	allKeys := make(map[string]bool)
	for k := range state1 {
		allKeys[k] = true
	}
	for k := range state2 {
		allKeys[k] = true
	}

	for key := range allKeys {
		val1, exists1 := state1[key]
		val2, exists2 := state2[key]

		if !exists1 {
			differences = append(differences, FieldDifference{
				Field:    key,
				OldValue: nil,
				NewValue: val2,
				Type:     "added",
			})
		} else if !exists2 {
			differences = append(differences, FieldDifference{
				Field:    key,
				OldValue: val1,
				NewValue: nil,
				Type:     "removed",
			})
		} else if !deepEqual(val1, val2) {
			differences = append(differences, FieldDifference{
				Field:    key,
				OldValue: val1,
				NewValue: val2,
				Type:     "changed",
			})
		}
	}

	return differences
}

// deepEqual compares two interface{} values
func deepEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

// FlagComparison represents a comparison between two flag versions
type FlagComparison struct {
	Version1    *entity.FlagHistory `json:"version_1"`
	Version2    *entity.FlagHistory `json:"version_2"`
	Differences []FieldDifference   `json:"differences"`
}

// FieldDifference represents a difference in a single field
type FieldDifference struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Type     string      `json:"type"` // "added", "removed", "changed"
}

// RecordCreate records flag creation
func (s *FlagHistoryService) RecordCreate(ctx context.Context, flag *entity.FeatureFlag, createdBy *uuid.UUID) (*entity.FlagHistory, error) {
	return s.RecordChange(ctx, flag, createdBy, entity.FlagChangeTypeCreated, "Flag created", nil)
}

// RecordUpdate records flag update
func (s *FlagHistoryService) RecordUpdate(ctx context.Context, flag *entity.FeatureFlag, updatedBy *uuid.UUID, description string, previousFlag *entity.FeatureFlag) (*entity.FlagHistory, error) {
	return s.RecordChange(ctx, flag, updatedBy, entity.FlagChangeTypeUpdated, description, previousFlag)
}

// RecordToggle records flag enable/disable
func (s *FlagHistoryService) RecordToggle(ctx context.Context, flag *entity.FeatureFlag, toggledBy *uuid.UUID, previousFlag *entity.FeatureFlag) (*entity.FlagHistory, error) {
	changeType := entity.FlagChangeTypeDisabled
	action := "disabled"
	if flag.Enabled {
		changeType = entity.FlagChangeTypeEnabled
		action = "enabled"
	}
	return s.RecordChange(ctx, flag, toggledBy, changeType, fmt.Sprintf("Flag %s", action), previousFlag)
}

// RecordDelete records flag deletion (soft delete marker)
func (s *FlagHistoryService) RecordDelete(ctx context.Context, flag *entity.FeatureFlag, deletedBy *uuid.UUID) (*entity.FlagHistory, error) {
	return s.RecordChange(ctx, flag, deletedBy, entity.FlagChangeTypeDeleted, "Flag deleted", nil)
}
