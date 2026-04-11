package entity

import (
	"time"

	"github.com/google/uuid"
)

// Environment represents a deployment environment (dev, staging, prod)
type Environment struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"application_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	Color         string    `json:"color"`
	IsProduction  bool      `json:"is_production"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DefaultEnvironments defines standard environments
var DefaultEnvironments = []struct {
	Name         string
	Slug         string
	Color        string
	IsProduction bool
}{
	{Name: "Development", Slug: "dev", Color: "#22c55e", IsProduction: false},
	{Name: "Staging", Slug: "staging", Color: "#f59e0b", IsProduction: false},
	{Name: "Production", Slug: "prod", Color: "#ef4444", IsProduction: true},
}

// NewEnvironment creates a new environment
func NewEnvironment(applicationID uuid.UUID, name, slug, color string, isProduction bool) *Environment {
	now := time.Now()
	if color == "" {
		color = "#6366f1"
	}
	return &Environment{
		ID:            uuid.New(),
		ApplicationID: applicationID,
		Name:          name,
		Slug:          slug,
		Color:         color,
		IsProduction:  isProduction,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Update updates environment details
func (e *Environment) Update(name, color string, isProduction *bool) {
	if name != "" {
		e.Name = name
	}
	if color != "" {
		e.Color = color
	}
	if isProduction != nil {
		e.IsProduction = *isProduction
	}
	e.UpdatedAt = time.Now()
}

// CreateDefaultEnvironments creates the standard set of environments for an application
func CreateDefaultEnvironments(applicationID uuid.UUID) []*Environment {
	envs := make([]*Environment, len(DefaultEnvironments))
	for i, def := range DefaultEnvironments {
		envs[i] = NewEnvironment(applicationID, def.Name, def.Slug, def.Color, def.IsProduction)
	}
	return envs
}
