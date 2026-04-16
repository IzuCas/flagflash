package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// InviteTokenRepository defines the interface for invite token persistence
type InviteTokenRepository interface {
	Create(ctx context.Context, invite *entity.InviteToken) error
	GetByToken(ctx context.Context, token string) (*entity.InviteToken, error)
	GetPendingByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*entity.InviteToken, error)
	Update(ctx context.Context, invite *entity.InviteToken) error
	DeleteExpired(ctx context.Context) error
}
