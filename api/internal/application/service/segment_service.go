package service

import (
	"context"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// SegmentService handles segment operations
type SegmentService struct {
	segmentRepo repository.SegmentRepository
	auditRepo   repository.AuditLogRepository
}

// NewSegmentService creates a new segment service
func NewSegmentService(
	segmentRepo repository.SegmentRepository,
	auditRepo repository.AuditLogRepository,
) *SegmentService {
	return &SegmentService{
		segmentRepo: segmentRepo,
		auditRepo:   auditRepo,
	}
}

// Create creates a new segment
func (s *SegmentService) Create(ctx context.Context, tenantID uuid.UUID, name, description string, conditions []entity.Condition, createdBy *uuid.UUID) (*entity.Segment, error) {
	// Check if segment with same name exists
	existing, _ := s.segmentRepo.GetByName(ctx, tenantID, name)
	if existing != nil {
		return nil, fmt.Errorf("segment with name '%s' already exists", name)
	}

	segment := entity.NewSegment(tenantID, name, description, conditions, createdBy)

	if err := s.segmentRepo.Create(ctx, segment); err != nil {
		return nil, fmt.Errorf("failed to create segment: %w", err)
	}

	return segment, nil
}

// GetByID gets a segment by ID
func (s *SegmentService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Segment, error) {
	return s.segmentRepo.GetByID(ctx, id)
}

// ListByTenant lists all segments for a tenant
func (s *SegmentService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Segment, error) {
	return s.segmentRepo.ListByTenant(ctx, tenantID)
}

// ListByTenantPaginated lists segments with pagination
func (s *SegmentService) ListByTenantPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int, search string) ([]*entity.Segment, int, error) {
	return s.segmentRepo.ListByTenantPaginated(ctx, tenantID, limit, offset, search)
}

// Update updates a segment
func (s *SegmentService) Update(ctx context.Context, segment *entity.Segment, actorID *uuid.UUID) error {
	if err := s.segmentRepo.Update(ctx, segment); err != nil {
		return fmt.Errorf("failed to update segment: %w", err)
	}

	return nil
}

// Delete deletes a segment
func (s *SegmentService) Delete(ctx context.Context, id uuid.UUID, actorID *uuid.UUID) error {
	_, err := s.segmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.segmentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete segment: %w", err)
	}

	return nil
}

// AddIncludedUser adds a user to the segment's included list
func (s *SegmentService) AddIncludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error {
	return s.segmentRepo.AddIncludedUser(ctx, segmentID, userID)
}

// RemoveIncludedUser removes a user from the segment's included list
func (s *SegmentService) RemoveIncludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error {
	return s.segmentRepo.RemoveIncludedUser(ctx, segmentID, userID)
}

// AddExcludedUser adds a user to the segment's excluded list
func (s *SegmentService) AddExcludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error {
	return s.segmentRepo.AddExcludedUser(ctx, segmentID, userID)
}

// RemoveExcludedUser removes a user from the segment's excluded list
func (s *SegmentService) RemoveExcludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error {
	return s.segmentRepo.RemoveExcludedUser(ctx, segmentID, userID)
}

// EvaluateContext evaluates if a context matches a segment
func (s *SegmentService) EvaluateContext(ctx context.Context, segmentID uuid.UUID, evalCtx *entity.EvaluationContext) (bool, error) {
	segment, err := s.segmentRepo.GetByID(ctx, segmentID)
	if err != nil {
		return false, err
	}

	return segment.MatchesContext(evalCtx), nil
}
