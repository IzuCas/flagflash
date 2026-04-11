package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// ApplicationCache handles caching for applications
type ApplicationCache struct {
	client *Client
	ttl    time.Duration
}

// NewApplicationCache creates a new ApplicationCache
func NewApplicationCache(client *Client) *ApplicationCache {
	return &ApplicationCache{
		client: client,
		ttl:    10 * time.Minute,
	}
}

// applicationIDKey generates a cache key for an application by ID
func applicationIDKey(id uuid.UUID) string {
	return fmt.Sprintf("%sid:%s", applicationCachePrefix, id.String())
}

// applicationListKey generates a cache key for listing applications by tenant
func applicationListKey(tenantID uuid.UUID) string {
	return fmt.Sprintf("%slist:%s", applicationCachePrefix, tenantID.String())
}

// GetApplication gets a cached application by ID
func (c *ApplicationCache) GetApplication(ctx context.Context, id uuid.UUID) (*entity.Application, error) {
	var app entity.Application
	err := c.client.Get(ctx, applicationIDKey(id), &app)
	if err != nil {
		return nil, err
	}
	if app.ID == uuid.Nil {
		return nil, nil
	}
	return &app, nil
}

// SetApplication caches an application
func (c *ApplicationCache) SetApplication(ctx context.Context, app *entity.Application) error {
	return c.client.Set(ctx, applicationIDKey(app.ID), app, c.ttl)
}

// GetApplicationList gets cached applications for a tenant
func (c *ApplicationCache) GetApplicationList(ctx context.Context, tenantID uuid.UUID) ([]*entity.Application, error) {
	var apps []*entity.Application
	err := c.client.Get(ctx, applicationListKey(tenantID), &apps)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

// SetApplicationList caches applications for a tenant
func (c *ApplicationCache) SetApplicationList(ctx context.Context, tenantID uuid.UUID, apps []*entity.Application) error {
	return c.client.Set(ctx, applicationListKey(tenantID), apps, c.ttl)
}

// InvalidateApplication removes a cached application and invalidates the list
func (c *ApplicationCache) InvalidateApplication(ctx context.Context, app *entity.Application) error {
	_ = c.client.Delete(ctx, applicationIDKey(app.ID))
	return c.client.Delete(ctx, applicationListKey(app.TenantID))
}
