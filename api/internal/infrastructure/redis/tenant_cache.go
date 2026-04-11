package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// TenantCache handles caching for tenants
type TenantCache struct {
	client *Client
	ttl    time.Duration
}

// NewTenantCache creates a new TenantCache
func NewTenantCache(client *Client) *TenantCache {
	return &TenantCache{
		client: client,
		ttl:    longTTL, // Tenants change infrequently
	}
}

// tenantIDKey generates a cache key for a tenant by ID
func tenantIDKey(id uuid.UUID) string {
	return fmt.Sprintf("%sid:%s", tenantCachePrefix, id.String())
}

// tenantSlugKey generates a cache key for a tenant by slug
func tenantSlugKey(slug string) string {
	return fmt.Sprintf("%sslug:%s", tenantCachePrefix, slug)
}

// GetTenant gets a cached tenant by ID
func (c *TenantCache) GetTenant(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	var tenant entity.Tenant
	err := c.client.Get(ctx, tenantIDKey(id), &tenant)
	if err != nil {
		return nil, err
	}
	if tenant.ID == uuid.Nil {
		return nil, nil
	}
	return &tenant, nil
}

// SetTenant caches a tenant
func (c *TenantCache) SetTenant(ctx context.Context, tenant *entity.Tenant) error {
	// Cache by ID and slug
	if err := c.client.Set(ctx, tenantIDKey(tenant.ID), tenant, c.ttl); err != nil {
		return err
	}
	return c.client.Set(ctx, tenantSlugKey(tenant.Slug), tenant, c.ttl)
}

// InvalidateTenant removes a cached tenant
func (c *TenantCache) InvalidateTenant(ctx context.Context, tenant *entity.Tenant) error {
	_ = c.client.Delete(ctx, tenantIDKey(tenant.ID))
	return c.client.Delete(ctx, tenantSlugKey(tenant.Slug))
}
