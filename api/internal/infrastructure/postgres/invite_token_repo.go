package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// InviteTokenRepository implements repository.InviteTokenRepository
type InviteTokenRepository struct {
	db *sql.DB
}

// NewInviteTokenRepository creates a new InviteTokenRepository
func NewInviteTokenRepository(db *sql.DB) *InviteTokenRepository {
	return &InviteTokenRepository{db: db}
}

// Create creates a new invite token
func (r *InviteTokenRepository) Create(ctx context.Context, invite *entity.InviteToken) error {
	query := `INSERT INTO invite_tokens (id, tenant_id, email, role, token, invited_by, expires_at, accepted_at, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		invite.ID, invite.TenantID, invite.Email, invite.Role, invite.Token,
		invite.InvitedBy, invite.ExpiresAt, invite.AcceptedAt, invite.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create invite token: %w", err)
	}
	return nil
}

// GetByToken retrieves an invite token by its token string
func (r *InviteTokenRepository) GetByToken(ctx context.Context, token string) (*entity.InviteToken, error) {
	query := `SELECT id, tenant_id, email, role, token, invited_by, expires_at, accepted_at, created_at
			  FROM invite_tokens WHERE token = $1`
	return r.scanInvite(r.db.QueryRowContext(ctx, query, token))
}

// GetPendingByEmailAndTenant retrieves a pending (non-accepted, non-expired) invite
func (r *InviteTokenRepository) GetPendingByEmailAndTenant(ctx context.Context, email string, tenantID uuid.UUID) (*entity.InviteToken, error) {
	query := `SELECT id, tenant_id, email, role, token, invited_by, expires_at, accepted_at, created_at
			  FROM invite_tokens
			  WHERE email = $1 AND tenant_id = $2 AND accepted_at IS NULL AND expires_at > NOW()
			  ORDER BY created_at DESC LIMIT 1`
	return r.scanInvite(r.db.QueryRowContext(ctx, query, email, tenantID))
}

// Update updates an invite token
func (r *InviteTokenRepository) Update(ctx context.Context, invite *entity.InviteToken) error {
	query := `UPDATE invite_tokens SET accepted_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, invite.AcceptedAt, invite.ID)
	if err != nil {
		return fmt.Errorf("failed to update invite token: %w", err)
	}
	return nil
}

// DeleteExpired removes expired and unaccepted tokens
func (r *InviteTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM invite_tokens WHERE expires_at < NOW() AND accepted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}
	return nil
}

func (r *InviteTokenRepository) scanInvite(row *sql.Row) (*entity.InviteToken, error) {
	invite := &entity.InviteToken{}
	err := row.Scan(
		&invite.ID, &invite.TenantID, &invite.Email, &invite.Role, &invite.Token,
		&invite.InvitedBy, &invite.ExpiresAt, &invite.AcceptedAt, &invite.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invite token not found")
		}
		return nil, fmt.Errorf("failed to scan invite token: %w", err)
	}
	return invite, nil
}
