package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// APIKeyCache handles caching for API keys
type APIKeyCache struct {
	client *Client
	ttl    time.Duration
}

// NewAPIKeyCache creates a new APIKeyCache
func NewAPIKeyCache(client *Client) *APIKeyCache {
	return &APIKeyCache{
		client: client,
		ttl:    10 * time.Minute, // API keys change infrequently
	}
}

// apiKeyHashKey generates a cache key for an API key by hash
func apiKeyHashKey(keyHash string) string {
	return fmt.Sprintf("%shash:%s", apiKeyCachePrefix, keyHash)
}

// GetAPIKey gets a cached API key by hash
func (c *APIKeyCache) GetAPIKey(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	var apiKey entity.APIKey
	err := c.client.Get(ctx, apiKeyHashKey(keyHash), &apiKey)
	if err != nil {
		return nil, err
	}
	if apiKey.ID == uuid.Nil {
		return nil, nil
	}
	return &apiKey, nil
}

// SetAPIKey caches an API key
func (c *APIKeyCache) SetAPIKey(ctx context.Context, apiKey *entity.APIKey) error {
	return c.client.Set(ctx, apiKeyHashKey(apiKey.KeyHash), apiKey, c.ttl)
}

// InvalidateAPIKey removes a cached API key
func (c *APIKeyCache) InvalidateAPIKey(ctx context.Context, keyHash string) error {
	return c.client.Delete(ctx, apiKeyHashKey(keyHash))
}
