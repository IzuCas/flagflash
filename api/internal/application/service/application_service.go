package service

import (
	"context"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/IzuCas/flagflash/internal/infrastructure/redis"
	"github.com/google/uuid"
)

// ApplicationService handles application business logic
type ApplicationService struct {
	appRepo   repository.ApplicationRepository
	envRepo   repository.EnvironmentRepository
	auditRepo repository.AuditLogRepository
	cache     *redis.ApplicationCache
}

// NewApplicationService creates a new application service
func NewApplicationService(
	appRepo repository.ApplicationRepository,
	envRepo repository.EnvironmentRepository,
	auditRepo repository.AuditLogRepository,
	cache *redis.ApplicationCache,
) *ApplicationService {
	return &ApplicationService{
		appRepo:   appRepo,
		envRepo:   envRepo,
		auditRepo: auditRepo,
		cache:     cache,
	}
}

// Create creates a new application with default environments
func (s *ApplicationService) Create(ctx context.Context, tenantID uuid.UUID, name, slug, description string, actorID string) (*entity.Application, error) {
	// Check if slug already exists for this tenant
	existing, _ := s.appRepo.GetBySlug(ctx, tenantID, slug)
	if existing != nil {
		return nil, fmt.Errorf("application with slug '%s' already exists in this tenant", slug)
	}

	app := entity.NewApplication(tenantID, name, slug, description)
	if err := s.appRepo.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	// Create default environments
	defaultEnvs := entity.CreateDefaultEnvironments(app.ID)
	if err := s.envRepo.CreateBatch(ctx, defaultEnvs); err != nil {
		// Rollback app creation
		s.appRepo.Delete(ctx, app.ID)
		return nil, fmt.Errorf("failed to create default environments: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		tenantID,
		entity.EntityTypeApplication,
		app.ID,
		entity.AuditActionCreate,
		actorID,
		entity.ActorTypeUser,
		nil,
		app,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	// Invalidate list cache
	if s.cache != nil {
		_ = s.cache.InvalidateApplication(ctx, app)
	}

	return app, nil
}

// GetByID retrieves an application by ID
func (s *ApplicationService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Application, error) {
	// Try cache first
	if s.cache != nil {
		if cached, _ := s.cache.GetApplication(ctx, id); cached != nil {
			return cached, nil
		}
	}

	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	// Cache the result
	if s.cache != nil && app != nil {
		_ = s.cache.SetApplication(ctx, app)
	}

	return app, nil
}

// GetBySlug retrieves an application by tenant ID and slug
func (s *ApplicationService) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*entity.Application, error) {
	app, err := s.appRepo.GetBySlug(ctx, tenantID, slug)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}
	return app, nil
}

// Update updates an application
func (s *ApplicationService) Update(ctx context.Context, id uuid.UUID, name, description string, actorID string) (*entity.Application, error) {
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	oldApp := *app
	app.Update(name, description)

	if err := s.appRepo.Update(ctx, app); err != nil {
		return nil, fmt.Errorf("failed to update application: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		app.TenantID,
		entity.EntityTypeApplication,
		app.ID,
		entity.AuditActionUpdate,
		actorID,
		entity.ActorTypeUser,
		oldApp,
		app,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateApplication(ctx, app)
	}

	return app, nil
}

// Delete soft deletes an application
func (s *ApplicationService) Delete(ctx context.Context, id uuid.UUID, actorID string) error {
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("application not found: %w", err)
	}

	if err := s.appRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete application: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		app.TenantID,
		entity.EntityTypeApplication,
		app.ID,
		entity.AuditActionDelete,
		actorID,
		entity.ActorTypeUser,
		app,
		nil,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateApplication(ctx, app)
	}

	return nil
}

// ListByTenant lists all applications for a tenant
func (s *ApplicationService) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.Application, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Try cache for first page without pagination
	if s.cache != nil && offset == 0 && limit >= 20 {
		if cached, _ := s.cache.GetApplicationList(ctx, tenantID); cached != nil && len(cached) > 0 {
			total := len(cached)
			if limit < total {
				return cached[:limit], total, nil
			}
			return cached, total, nil
		}
	}

	apps, total, err := s.appRepo.ListByTenant(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Cache full list on first page fetch
	if s.cache != nil && offset == 0 {
		_ = s.cache.SetApplicationList(ctx, tenantID, apps)
	}

	return apps, total, nil
}

// ApplicationWithEnvironments represents an application with its environments
type ApplicationWithEnvironments struct {
	*entity.Application
	Environments []*entity.Environment `json:"environments"`
}

// GetByIDWithEnvironments retrieves an application with its environments
func (s *ApplicationService) GetByIDWithEnvironments(ctx context.Context, id uuid.UUID) (*ApplicationWithEnvironments, error) {
	app, err := s.appRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	envs, err := s.envRepo.ListByApplication(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get environments: %w", err)
	}

	return &ApplicationWithEnvironments{
		Application:  app,
		Environments: envs,
	}, nil
}
