package repository

import (
	"context"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// APIKeyRepository defines the interface for API key persistence
type APIKeyRepository interface {
	// Create creates a new API key
	Create(ctx context.Context, key *entity.APIKey) error

	// GetByID retrieves an API key by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.APIKey, error)

	// GetByHash retrieves an API key by its hash
	GetByHash(ctx context.Context, keyHash string) (*entity.APIKey, error)

	// GetByKeyHash retrieves an API key by prefix and hash
	GetByKeyHash(ctx context.Context, keyPrefix, keyHash string) (*entity.APIKey, error)

	// GetByTenantID retrieves all API keys for a tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.APIKey, error)

	// Update updates an API key
	Update(ctx context.Context, key *entity.APIKey) error

	// Revoke revokes an API key
	Revoke(ctx context.Context, id uuid.UUID) error

	// Delete permanently deletes an API key
	Delete(ctx context.Context, id uuid.UUID) error

	// ListByEnvironment lists all API keys for an environment
	ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.APIKeyInfo, error)

	// UpdateLastUsed updates the last used timestamp
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error

	// GetByHashWithDetails retrieves an API key with environment and tenant details
	GetByHashWithDetails(ctx context.Context, keyHash string) (*entity.APIKey, *entity.Environment, uuid.UUID, error)
}
