package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// EnvironmentCache handles caching for environments
type EnvironmentCache struct {
	client *Client
	ttl    time.Duration
}

// NewEnvironmentCache creates a new EnvironmentCache
func NewEnvironmentCache(client *Client) *EnvironmentCache {
	return &EnvironmentCache{
		client: client,
		ttl:    10 * time.Minute,
	}
}

// environmentIDKey generates a cache key for an environment by ID
func environmentIDKey(id uuid.UUID) string {
	return fmt.Sprintf("%sid:%s", environmentCachePrefix, id.String())
}

// environmentListKey generates a cache key for listing environments by application
func environmentListKey(applicationID uuid.UUID) string {
	return fmt.Sprintf("%slist:%s", environmentCachePrefix, applicationID.String())
}

// GetEnvironment gets a cached environment by ID
func (c *EnvironmentCache) GetEnvironment(ctx context.Context, id uuid.UUID) (*entity.Environment, error) {
	var env entity.Environment
	err := c.client.Get(ctx, environmentIDKey(id), &env)
	if err != nil {
		return nil, err
	}
	if env.ID == uuid.Nil {
		return nil, nil
	}
	return &env, nil
}

// SetEnvironment caches an environment
func (c *EnvironmentCache) SetEnvironment(ctx context.Context, env *entity.Environment) error {
	return c.client.Set(ctx, environmentIDKey(env.ID), env, c.ttl)
}

// GetEnvironmentList gets cached environments for an application
func (c *EnvironmentCache) GetEnvironmentList(ctx context.Context, applicationID uuid.UUID) ([]*entity.Environment, error) {
	var envs []*entity.Environment
	err := c.client.Get(ctx, environmentListKey(applicationID), &envs)
	if err != nil {
		return nil, err
	}
	return envs, nil
}

// SetEnvironmentList caches environments for an application
func (c *EnvironmentCache) SetEnvironmentList(ctx context.Context, applicationID uuid.UUID, envs []*entity.Environment) error {
	return c.client.Set(ctx, environmentListKey(applicationID), envs, c.ttl)
}

// InvalidateEnvironmentCache removes a cached environment and invalidates the list
func (c *EnvironmentCache) InvalidateEnvironmentCache(ctx context.Context, env *entity.Environment) error {
	_ = c.client.Delete(ctx, environmentIDKey(env.ID))
	return c.client.Delete(ctx, environmentListKey(env.ApplicationID))
}
