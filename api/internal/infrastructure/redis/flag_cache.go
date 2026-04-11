package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// FlagCache handles caching for feature flags
type FlagCache struct {
	client *Client
	ttl    time.Duration
}

// NewFlagCache creates a new FlagCache
func NewFlagCache(client *Client) *FlagCache {
	return &FlagCache{
		client: client,
		ttl:    defaultTTL,
	}
}

// SetTTL sets the cache TTL
func (c *FlagCache) SetTTL(ttl time.Duration) {
	c.ttl = ttl
}

// flagKey generates a cache key for a flag
func flagKey(environmentID uuid.UUID, key string) string {
	return fmt.Sprintf("%s%s:%s", flagCachePrefix, environmentID.String(), key)
}

// flagsKey generates a cache key for all flags in an environment
func flagsKey(environmentID uuid.UUID) string {
	return fmt.Sprintf("%s%s:all", flagCachePrefix, environmentID.String())
}

// apiKeyKey generates a cache key for an API key (legacy key format used by FlagCache)
func apiKeyKey(keyHash string) string {
	return fmt.Sprintf("%s%s", apiKeyCachePrefix, keyHash)
}

// GetFlag gets a cached flag
func (c *FlagCache) GetFlag(ctx context.Context, environmentID uuid.UUID, key string) (*entity.FeatureFlag, error) {
	var flag entity.FeatureFlag
	err := c.client.Get(ctx, flagKey(environmentID, key), &flag)
	if err != nil {
		return nil, err
	}
	if flag.ID == uuid.Nil {
		return nil, nil
	}
	return &flag, nil
}

// SetFlag caches a flag
func (c *FlagCache) SetFlag(ctx context.Context, environmentID uuid.UUID, flag *entity.FeatureFlag) error {
	return c.client.Set(ctx, flagKey(environmentID, flag.Key), flag, c.ttl)
}

// DeleteFlag removes a cached flag
func (c *FlagCache) DeleteFlag(ctx context.Context, environmentID uuid.UUID, key string) error {
	return c.client.Delete(ctx, flagKey(environmentID, key))
}

// GetFlags gets all cached flags for an environment
func (c *FlagCache) GetFlags(ctx context.Context, environmentID uuid.UUID) ([]*entity.FeatureFlag, error) {
	var flags []*entity.FeatureFlag
	err := c.client.Get(ctx, flagsKey(environmentID), &flags)
	if err != nil {
		return nil, err
	}
	return flags, nil
}

// SetFlags caches all flags for an environment
func (c *FlagCache) SetFlags(ctx context.Context, environmentID uuid.UUID, flags []*entity.FeatureFlag) error {
	return c.client.Set(ctx, flagsKey(environmentID), flags, c.ttl)
}

// InvalidateEnvironment invalidates all cached flags for an environment
func (c *FlagCache) InvalidateEnvironment(ctx context.Context, environmentID uuid.UUID) error {
	pattern := fmt.Sprintf("%s%s:*", flagCachePrefix, environmentID.String())
	return c.client.DeletePattern(ctx, pattern)
}

// InvalidateFlags invalidates all cached flags for an environment (alias for InvalidateEnvironment)
func (c *FlagCache) InvalidateFlags(ctx context.Context, environmentID uuid.UUID) error {
	return c.InvalidateEnvironment(ctx, environmentID)
}

// GetAPIKey gets a cached API key
func (c *FlagCache) GetAPIKey(ctx context.Context, keyHash string) (*entity.APIKey, error) {
	var apiKey entity.APIKey
	err := c.client.Get(ctx, apiKeyKey(keyHash), &apiKey)
	if err != nil {
		return nil, err
	}
	if apiKey.ID == uuid.Nil {
		return nil, nil
	}
	return &apiKey, nil
}

// SetAPIKey caches an API key
func (c *FlagCache) SetAPIKey(ctx context.Context, keyHash string, apiKey *entity.APIKey) error {
	return c.client.Set(ctx, apiKeyKey(keyHash), apiKey, c.ttl)
}

// DeleteAPIKey removes a cached API key
func (c *FlagCache) DeleteAPIKey(ctx context.Context, keyHash string) error {
	return c.client.Delete(ctx, apiKeyKey(keyHash))
}
