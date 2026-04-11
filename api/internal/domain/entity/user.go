package entity

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserRole defines the user's role in the system
type UserRole string

const (
	UserRoleOwner  UserRole = "owner"
	UserRoleAdmin  UserRole = "admin"
	UserRoleMember UserRole = "member"
	UserRoleViewer UserRole = "viewer"
)

// RoleHierarchy returns the numeric level of a role (higher = more permissions)
func (r UserRole) Level() int {
	switch r {
	case UserRoleOwner:
		return 100
	case UserRoleAdmin:
		return 75
	case UserRoleMember:
		return 50
	case UserRoleViewer:
		return 25
	default:
		return 0
	}
}

// CanManageRole checks if this role can manage users with the target role
func (r UserRole) CanManageRole(target UserRole) bool {
	// Only higher or equal level can manage (but not equal for non-owners)
	if r == UserRoleOwner {
		return true // Owner can manage everyone
	}
	// Others can only manage roles lower than theirs
	return r.Level() > target.Level()
}

// User represents a dashboard user
type User struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	Name         string     `json:"name"`
	Role         UserRole   `json:"role"`
	Active       bool       `json:"active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

// NewUser creates a new user
func NewUser(tenantID uuid.UUID, email, password, name string, role UserRole) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &User{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.UpdatedAt = time.Now()
	return nil
}

// Update updates user details
func (u *User) Update(name string, role *UserRole) {
	if name != "" {
		u.Name = name
	}
	if role != nil {
		u.Role = *role
	}
	u.UpdatedAt = time.Now()
}

// SoftDelete marks the user as deleted
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
}

// IsDeleted checks if user is soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// CanManageTenants checks if user can manage tenants
func (u *User) CanManageTenants() bool {
	return u.Role == UserRoleOwner
}

// CanManageApplications checks if user can manage applications
func (u *User) CanManageApplications() bool {
	return u.Role == UserRoleOwner || u.Role == UserRoleAdmin
}

// CanManageFlags checks if user can manage feature flags
func (u *User) CanManageFlags() bool {
	return u.Role != UserRoleViewer
}

// CanViewFlags checks if user can view feature flags
func (u *User) CanViewFlags() bool {
	return true
}

// UserClaims represents JWT claims for a user
type UserClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     UserRole  `json:"role"`
}
