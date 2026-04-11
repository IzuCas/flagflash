package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// AuditLogRepository defines the interface for audit log persistence
type AuditLogRepository interface {
	// Create creates a new audit log entry
	Create(ctx context.Context, log *entity.AuditLog) error

	// CreateBatch creates multiple audit log entries
	CreateBatch(ctx context.Context, logs []*entity.AuditLog) error

	// GetByID retrieves an audit log by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.AuditLog, error)

	// List queries audit logs with filters
	List(ctx context.Context, query *entity.AuditLogQuery) ([]*entity.AuditLog, int, error)

	// GetByEntity retrieves audit logs for a specific entity
	GetByEntity(ctx context.Context, entityType entity.EntityType, entityID uuid.UUID, limit int) ([]*entity.AuditLog, error)

	// DeleteOlderThan deletes audit logs older than a specific time
	DeleteOlderThan(ctx context.Context, tenantID uuid.UUID, days int) (int64, error)
}
