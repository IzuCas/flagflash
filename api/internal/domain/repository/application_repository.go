package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// ApplicationRepository defines the interface for application persistence
type ApplicationRepository interface {
	Create(ctx context.Context, app *entity.Application) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Application, error)
	GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*entity.Application, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Application, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.Application, int, error)
	Update(ctx context.Context, app *entity.Application) error
	Delete(ctx context.Context, id uuid.UUID) error
}
