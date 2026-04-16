package entity

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// InviteToken represents an invitation to join a tenant
type InviteToken struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenant_id"`
	Email      string     `json:"email"`
	Role       UserRole   `json:"role"`
	Token      string     `json:"token"`
	InvitedBy  uuid.UUID  `json:"invited_by"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// NewInviteToken creates a new invite token
func NewInviteToken(tenantID uuid.UUID, email string, role UserRole, invitedBy uuid.UUID) (*InviteToken, error) {
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &InviteToken{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Email:     email,
		Role:      role,
		Token:     token,
		InvitedBy: invitedBy,
		ExpiresAt: now.Add(7 * 24 * time.Hour), // 7 days
		CreatedAt: now,
	}, nil
}

// IsExpired checks if the token has expired
func (t *InviteToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsAccepted checks if the invite has been accepted
func (t *InviteToken) IsAccepted() bool {
	return t.AcceptedAt != nil
}

// Accept marks the invite as accepted
func (t *InviteToken) Accept() {
	now := time.Now()
	t.AcceptedAt = &now
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
