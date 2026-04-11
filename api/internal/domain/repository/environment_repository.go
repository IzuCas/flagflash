package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// EnvironmentRepository defines the interface for environment persistence
type EnvironmentRepository interface {
	// Create creates a new environment
	Create(ctx context.Context, env *entity.Environment) error

	// CreateBatch creates multiple environments
	CreateBatch(ctx context.Context, envs []*entity.Environment) error

	// GetByID retrieves an environment by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error)

	// GetBySlug retrieves an environment by application ID and slug
	GetBySlug(ctx context.Context, applicationID uuid.UUID, slug string) (*entity.Environment, error)

	// Update updates an environment
	Update(ctx context.Context, env *entity.Environment) error

	// Delete deletes an environment
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByApplication lists all environments for an application
	ListByApplication(ctx context.Context, applicationID uuid.UUID) ([]*entity.Environment, error)

	// GetByIDWithTenant retrieves an environment with its tenant info
	GetByIDWithTenant(ctx context.Context, id uuid.UUID) (*entity.Environment, uuid.UUID, error)
}
