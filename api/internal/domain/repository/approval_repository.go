package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// ApprovalSettingRepository defines the interface for approval settings persistence
type ApprovalSettingRepository interface {
	Create(ctx context.Context, setting *entity.ApprovalSetting) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ApprovalSetting, error)
	GetByEnvironment(ctx context.Context, tenantID, environmentID uuid.UUID) (*entity.ApprovalSetting, error)
	GetByFlag(ctx context.Context, tenantID, flagID uuid.UUID) (*entity.ApprovalSetting, error)
	GetEffective(ctx context.Context, tenantID, environmentID, flagID uuid.UUID) (*entity.ApprovalSetting, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.ApprovalSetting, error)
	Update(ctx context.Context, setting *entity.ApprovalSetting) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// PendingChangeRepository defines the interface for pending changes persistence
type PendingChangeRepository interface {
	Create(ctx context.Context, change *entity.PendingChange) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PendingChange, error)
	GetByIDWithApprovals(ctx context.Context, id uuid.UUID) (*entity.PendingChange, error)
	ListPending(ctx context.Context, tenantID uuid.UUID) ([]*entity.PendingChange, error)
	ListPendingByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.PendingChange, error)
	ListPendingByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.PendingChange, error)
	ListByRequester(ctx context.Context, userID uuid.UUID) ([]*entity.PendingChange, error)
	ListExpired(ctx context.Context) ([]*entity.PendingChange, error)
	Update(ctx context.Context, change *entity.PendingChange) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExpireOld(ctx context.Context) (int, error)
}

// ApprovalRepository defines the interface for approvals persistence
type ApprovalRepository interface {
	Create(ctx context.Context, approval *entity.Approval) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Approval, error)
	GetByChangeAndApprover(ctx context.Context, changeID, approverID uuid.UUID) (*entity.Approval, error)
	ListByChange(ctx context.Context, changeID uuid.UUID) ([]*entity.Approval, error)
	CountByDecision(ctx context.Context, changeID uuid.UUID, decision entity.ApprovalDecision) (int, error)
	Update(ctx context.Context, approval *entity.Approval) error
	Delete(ctx context.Context, id uuid.UUID) error
}
