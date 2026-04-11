package service

import (
	"context"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// AuditLogService handles audit log business logic
type AuditLogService struct {
	repo repository.AuditLogRepository
}

// NewAuditLogService creates a new audit log service
func NewAuditLogService(repo repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

// List retrieves audit logs with filtering and pagination
func (s *AuditLogService) List(
	ctx context.Context,
	tenantID uuid.UUID,
	entityType *string,
	entityID *uuid.UUID,
	action *string,
	actorID *string,
	startDate, endDate *time.Time,
	page, limit int,
) ([]*entity.AuditLog, int, error) {
	query := &entity.AuditLogQuery{
		TenantID: &tenantID,
		Limit:    limit,
		Offset:   (page - 1) * limit,
	}

	if entityType != nil && *entityType != "" {
		et := entity.EntityType(*entityType)
		query.EntityType = &et
	}

	if entityID != nil {
		query.EntityID = entityID
	}

	if action != nil && *action != "" {
		a := entity.AuditAction(*action)
		query.Action = &a
	}

	if actorID != nil && *actorID != "" {
		query.ActorID = actorID
	}

	if startDate != nil {
		query.StartDate = startDate
	}

	if endDate != nil {
		query.EndDate = endDate
	}

	return s.repo.List(ctx, query)
}

// GetByID retrieves a single audit log entry
func (s *AuditLogService) GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByEntity retrieves audit logs for a specific entity
func (s *AuditLogService) GetByEntity(ctx context.Context, entityType entity.EntityType, entityID uuid.UUID, limit int) ([]*entity.AuditLog, error) {
	return s.repo.GetByEntity(ctx, entityType, entityID, limit)
}

// Create creates a new audit log entry
func (s *AuditLogService) Create(ctx context.Context, log *entity.AuditLog) error {
	return s.repo.Create(ctx, log)
}

// DeleteOlderThan deletes audit logs older than specified days
func (s *AuditLogService) DeleteOlderThan(ctx context.Context, tenantID uuid.UUID, days int) (int64, error) {
	return s.repo.DeleteOlderThan(ctx, tenantID, days)
}
