package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserTenantMembership represents the many-to-many relationship between users and tenants
type UserTenantMembership struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Role      UserRole  `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserWithMembership represents a user with their membership details for a specific tenant
type UserWithMembership struct {
	User       *User                 `json:"user"`
	Membership *UserTenantMembership `json:"membership"`
}

// TenantWithRole represents a tenant with the user's role in that tenant
type TenantWithRole struct {
	Tenant *Tenant  `json:"tenant"`
	Role   UserRole `json:"role"`
	Active bool     `json:"active"`
}

// NewUserTenantMembership creates a new user-tenant membership
func NewUserTenantMembership(userID, tenantID uuid.UUID, role UserRole) *UserTenantMembership {
	now := time.Now()
	return &UserTenantMembership{
		ID:        uuid.New(),
		UserID:    userID,
		TenantID:  tenantID,
		Role:      role,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the membership role
func (m *UserTenantMembership) Update(role UserRole) {
	m.Role = role
	m.UpdatedAt = time.Now()
}

// Deactivate deactivates the membership
func (m *UserTenantMembership) Deactivate() {
	m.Active = false
	m.UpdatedAt = time.Now()
}

// Activate activates the membership
func (m *UserTenantMembership) Activate() {
	m.Active = true
	m.UpdatedAt = time.Now()
}
