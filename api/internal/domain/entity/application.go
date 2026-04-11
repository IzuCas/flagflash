package entity

import (
	"time"

	"github.com/google/uuid"
)

// Application represents an application within a tenant
type Application struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// NewApplication creates a new application
func NewApplication(tenantID uuid.UUID, name, slug, description string) *Application {
	now := time.Now()
	return &Application{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Slug:        slug,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Update updates application details
func (a *Application) Update(name, description string) {
	if name != "" {
		a.Name = name
	}
	if description != "" {
		a.Description = description
	}
	a.UpdatedAt = time.Now()
}
