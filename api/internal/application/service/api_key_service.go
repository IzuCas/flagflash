package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/IzuCas/flagflash/internal/infrastructure/redis"
	"github.com/google/uuid"
)

// APIKeyService handles API key business logic
type APIKeyService struct {
	apiKeyRepo repository.APIKeyRepository
	tenantRepo repository.TenantRepository
	cache      *redis.APIKeyCache
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(apiKeyRepo repository.APIKeyRepository, tenantRepo repository.TenantRepository, cache *redis.APIKeyCache) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo: apiKeyRepo,
		tenantRepo: tenantRepo,
		cache:      cache,
	}
}

// APIKeyDetails represents API key details returned to clients
type APIKeyDetails struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	EnvironmentID *uuid.UUID `json:"environment_id,omitempty"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	Permissions   []string   `json:"permissions"`
	Active        bool       `json:"active"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// CreateAPIKey creates a new API key
func (s *APIKeyService) CreateAPIKey(ctx context.Context, tenantID uuid.UUID, environmentID *uuid.UUID, name string, permissions []string, expiresAt *time.Time) (*APIKeyDetails, string, error) {
	// Verify tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil || tenant == nil {
		return nil, "", fmt.Errorf("tenant not found")
	}

	// Generate random key
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Hash the key for storage
	keyHash := hashAPIKey(rawKey)
	keyPrefix := rawKey[:8]

	apiKey := entity.NewAPIKey(tenantID, environmentID, name, keyHash, keyPrefix, permissions, expiresAt)

	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	details := &APIKeyDetails{
		ID:            apiKey.ID,
		TenantID:      apiKey.TenantID,
		EnvironmentID: apiKey.EnvironmentID,
		Name:          apiKey.Name,
		KeyPrefix:     apiKey.KeyPrefix,
		Permissions:   apiKey.Permissions,
		Active:        apiKey.Active,
		ExpiresAt:     apiKey.ExpiresAt,
		CreatedAt:     apiKey.CreatedAt,
	}

	return details, rawKey, nil
}

// GetAPIKey retrieves an API key by ID
func (s *APIKeyService) GetAPIKey(ctx context.Context, id uuid.UUID) (*APIKeyDetails, error) {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}
	if apiKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	return s.toDetails(apiKey), nil
}

// ListAPIKeysByTenant lists all API keys for a tenant
func (s *APIKeyService) ListAPIKeysByTenant(ctx context.Context, tenantID uuid.UUID) ([]*APIKeyDetails, error) {
	apiKeys, err := s.apiKeyRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	details := make([]*APIKeyDetails, len(apiKeys))
	for i, key := range apiKeys {
		details[i] = s.toDetails(key)
	}

	return details, nil
}

// ValidateAPIKey validates an API key and returns the associated key details
func (s *APIKeyService) ValidateAPIKey(ctx context.Context, rawKey string) (*APIKeyDetails, error) {
	if len(rawKey) < 8 {
		return nil, fmt.Errorf("invalid key format")
	}

	keyPrefix := rawKey[:8]
	keyHash := hashAPIKey(rawKey)

	// Try cache first
	if s.cache != nil {
		if cachedKey, _ := s.cache.GetAPIKey(ctx, keyHash); cachedKey != nil {
			if !cachedKey.Active {
				return nil, fmt.Errorf("API key is revoked")
			}
			if cachedKey.ExpiresAt != nil && cachedKey.ExpiresAt.Before(time.Now()) {
				return nil, fmt.Errorf("API key has expired")
			}
			// Update last used asynchronously (don't block response)
			go func() {
				bgCtx := context.Background()
				cachedKey.RecordUsage()
				_ = s.apiKeyRepo.Update(bgCtx, cachedKey)
			}()
			return s.toDetails(cachedKey), nil
		}
	}

	apiKey, err := s.apiKeyRepo.GetByKeyHash(ctx, keyPrefix, keyHash)
	if err != nil {
		return nil, fmt.Errorf("failed to validate key: %w", err)
	}
	if apiKey == nil {
		return nil, fmt.Errorf("invalid API key")
	}

	if !apiKey.Active {
		return nil, fmt.Errorf("API key is revoked")
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Cache the API key
	if s.cache != nil {
		_ = s.cache.SetAPIKey(ctx, apiKey)
	}

	// Update last used timestamp asynchronously
	go func() {
		bgCtx := context.Background()
		apiKey.RecordUsage()
		_ = s.apiKeyRepo.Update(bgCtx, apiKey)
	}()

	return s.toDetails(apiKey), nil
}

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, id uuid.UUID) error {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}
	if apiKey == nil {
		return fmt.Errorf("API key not found")
	}

	apiKey.Revoke()

	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateAPIKey(ctx, apiKey.KeyHash)
	}

	return nil
}

// DeleteAPIKey deletes an API key
func (s *APIKeyService) DeleteAPIKey(ctx context.Context, id uuid.UUID) error {
	// Get key first to invalidate cache
	apiKey, _ := s.apiKeyRepo.GetByID(ctx, id)

	if err := s.apiKeyRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	// Invalidate cache
	if s.cache != nil && apiKey != nil {
		_ = s.cache.InvalidateAPIKey(ctx, apiKey.KeyHash)
	}

	return nil
}

// UpdateLastUsed updates the last used timestamp for an API key
func (s *APIKeyService) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	apiKey, err := s.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}
	if apiKey == nil {
		return fmt.Errorf("API key not found")
	}

	apiKey.RecordUsage()

	if err := s.apiKeyRepo.Update(ctx, apiKey); err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	return nil
}

func (s *APIKeyService) toDetails(apiKey *entity.APIKey) *APIKeyDetails {
	return &APIKeyDetails{
		ID:            apiKey.ID,
		TenantID:      apiKey.TenantID,
		EnvironmentID: apiKey.EnvironmentID,
		Name:          apiKey.Name,
		KeyPrefix:     apiKey.KeyPrefix,
		Permissions:   apiKey.Permissions,
		Active:        apiKey.Active,
		LastUsedAt:    apiKey.LastUsedAt,
		ExpiresAt:     apiKey.ExpiresAt,
		CreatedAt:     apiKey.CreatedAt,
	}
}

// generateAPIKey generates a random API key
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "ff_" + hex.EncodeToString(bytes), nil
}

// hashAPIKey creates a hash of the API key
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
