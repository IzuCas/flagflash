package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// EmergencyControlRepository defines the interface for emergency control persistence
type EmergencyControlRepository interface {
	Create(ctx context.Context, control *entity.EmergencyControl) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.EmergencyControl, error)
	GetByType(ctx context.Context, tenantID uuid.UUID, envID *uuid.UUID, controlType entity.EmergencyControlType) (*entity.EmergencyControl, error)
	GetActiveKillSwitch(ctx context.Context, tenantID uuid.UUID, envID *uuid.UUID) (*entity.EmergencyControl, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.EmergencyControl, error)
	ListEnabled(ctx context.Context, tenantID uuid.UUID) ([]*entity.EmergencyControl, error)
	ListExpired(ctx context.Context) ([]*entity.EmergencyControl, error)
	Update(ctx context.Context, control *entity.EmergencyControl) error
	Upsert(ctx context.Context, control *entity.EmergencyControl) error
	Delete(ctx context.Context, id uuid.UUID) error
	DisableExpired(ctx context.Context) (int, error)
}
