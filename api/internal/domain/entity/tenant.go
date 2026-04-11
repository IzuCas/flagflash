package entity

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents an organization in the system
type Tenant struct {
	ID        uuid.UUID              `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Plan      string                 `json:"plan"`
	Active    bool                   `json:"active"`
	Settings  map[string]interface{} `json:"settings"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
}

// NewTenant creates a new tenant with default values
func NewTenant(name, slug string) *Tenant {
	now := time.Now()
	return &Tenant{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		Plan:      "free",
		Active:    true,
		Settings:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates tenant details
func (t *Tenant) Update(name string, settings map[string]interface{}) {
	if name != "" {
		t.Name = name
	}
	if settings != nil {
		t.Settings = settings
	}
	t.UpdatedAt = time.Now()
}

// SoftDelete marks the tenant as deleted
func (t *Tenant) SoftDelete() {
	now := time.Now()
	t.DeletedAt = &now
}

// IsDeleted checks if tenant is soft deleted
func (t *Tenant) IsDeleted() bool {
	return t.DeletedAt != nil
}
