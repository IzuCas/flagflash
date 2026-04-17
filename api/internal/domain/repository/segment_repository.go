package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// SegmentRepository defines the interface for segment persistence
type SegmentRepository interface {
	Create(ctx context.Context, segment *entity.Segment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Segment, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*entity.Segment, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Segment, error)
	ListByTenantPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int, search string) ([]*entity.Segment, int, error)
	Update(ctx context.Context, segment *entity.Segment) error
	Delete(ctx context.Context, id uuid.UUID) error
	AddIncludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error
	RemoveIncludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error
	AddExcludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error
	RemoveExcludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error
}
