package service

import (
	"context"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/IzuCas/flagflash/internal/infrastructure/redis"
	"github.com/google/uuid"
)

// EnvironmentService handles environment business logic
type EnvironmentService struct {
	envRepo   repository.EnvironmentRepository
	appRepo   repository.ApplicationRepository
	auditRepo repository.AuditLogRepository
	flagRepo  repository.FeatureFlagRepository
	cache     *redis.EnvironmentCache
}

// NewEnvironmentService creates a new environment service
func NewEnvironmentService(
	envRepo repository.EnvironmentRepository,
	appRepo repository.ApplicationRepository,
	auditRepo repository.AuditLogRepository,
	flagRepo repository.FeatureFlagRepository,
	cache *redis.EnvironmentCache,
) *EnvironmentService {
	return &EnvironmentService{
		envRepo:   envRepo,
		appRepo:   appRepo,
		auditRepo: auditRepo,
		flagRepo:  flagRepo,
		cache:     cache,
	}
}

// Create creates a new environment
func (s *EnvironmentService) Create(ctx context.Context, applicationID uuid.UUID, name, slug, color string, isProduction bool, actorID string) (*entity.Environment, error) {
	// Check if application exists
	app, err := s.appRepo.GetByID(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("application not found: %w", err)
	}

	// Check if slug already exists for this application
	existing, _ := s.envRepo.GetBySlug(ctx, applicationID, slug)
	if existing != nil {
		return nil, fmt.Errorf("environment with slug '%s' already exists in this application", slug)
	}

	env := entity.NewEnvironment(applicationID, name, slug, color, isProduction)
	if err := s.envRepo.Create(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Create audit log
	auditLog := entity.NewAuditLog(
		app.TenantID,
		entity.EntityTypeEnvironment,
		env.ID,
		entity.AuditActionCreate,
		actorID,
		entity.ActorTypeUser,
		nil,
		env,
		nil,
	)
	s.auditRepo.Create(ctx, auditLog)

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateEnvironmentCache(ctx, env)
	}

	return env, nil
}

// GetByID retrieves an environment by ID
func (s *EnvironmentService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error) {
	// Try cache first
	if s.cache != nil {
		if cached, _ := s.cache.GetEnvironment(ctx, id); cached != nil {
			return cached, nil
		}
	}

	env, err := s.envRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("environment not found: %w", err)
	}

	// Cache the result
	if s.cache != nil && env != nil {
		_ = s.cache.SetEnvironment(ctx, env)
	}

	return env, nil
}

// GetBySlug retrieves an environment by application ID and slug
func (s *EnvironmentService) GetBySlug(ctx context.Context, applicationID uuid.UUID, slug string) (*entity.Environment, error) {
	env, err := s.envRepo.GetBySlug(ctx, applicationID, slug)
	if err != nil {
		return nil, fmt.Errorf("environment not found: %w", err)
	}
	return env, nil
}

// Update updates an environment
func (s *EnvironmentService) Update(ctx context.Context, id uuid.UUID, name, color string, isProduction *bool, actorID string) (*entity.Environment, error) {
	env, err := s.envRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("environment not found: %w", err)
	}

	oldEnv := *env
	env.Update(name, color, isProduction)

	if err := s.envRepo.Update(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to update environment: %w", err)
	}

	// Get application for tenant ID
	app, _ := s.appRepo.GetByID(ctx, env.ApplicationID)
	if app != nil {
		auditLog := entity.NewAuditLog(
			app.TenantID,
			entity.EntityTypeEnvironment,
			env.ID,
			entity.AuditActionUpdate,
			actorID,
			entity.ActorTypeUser,
			oldEnv,
			env,
			nil,
		)
		s.auditRepo.Create(ctx, auditLog)
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateEnvironmentCache(ctx, env)
	}

	return env, nil
}

// Delete deletes an environment
func (s *EnvironmentService) Delete(ctx context.Context, id uuid.UUID, actorID string) error {
	env, err := s.envRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("environment not found: %w", err)
	}

	// Prevent deletion of production environment
	if env.IsProduction {
		return fmt.Errorf("cannot delete production environment")
	}

	if err := s.envRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	// Get application for tenant ID
	app, _ := s.appRepo.GetByID(ctx, env.ApplicationID)
	if app != nil {
		auditLog := entity.NewAuditLog(
			app.TenantID,
			entity.EntityTypeEnvironment,
			env.ID,
			entity.AuditActionDelete,
			actorID,
			entity.ActorTypeUser,
			env,
			nil,
			nil,
		)
		s.auditRepo.Create(ctx, auditLog)
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.InvalidateEnvironmentCache(ctx, env)
	}

	return nil
}

// ListByApplication lists all environments for an application
func (s *EnvironmentService) ListByApplication(ctx context.Context, applicationID uuid.UUID) ([]*entity.Environment, error) {
	// Try cache first
	if s.cache != nil {
		if cached, _ := s.cache.GetEnvironmentList(ctx, applicationID); cached != nil && len(cached) > 0 {
			return cached, nil
		}
	}

	envs, err := s.envRepo.ListByApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if s.cache != nil {
		_ = s.cache.SetEnvironmentList(ctx, applicationID, envs)
	}

	return envs, nil
}

// GetByIDWithDetails retrieves an environment with its application info
func (s *EnvironmentService) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*entity.Environment, *entity.Application, error) {
	env, err := s.envRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("environment not found: %w", err)
	}

	app, err := s.appRepo.GetByID(ctx, env.ApplicationID)
	if err != nil {
		return nil, nil, fmt.Errorf("application not found: %w", err)
	}

	return env, app, nil
}

// CopyEnvironment copies an environment including all its feature flags
func (s *EnvironmentService) CopyEnvironment(ctx context.Context, sourceID uuid.UUID, name, slug, description string) (*entity.Environment, error) {
	// Get source environment
	source, err := s.envRepo.GetByID(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("source environment not found: %w", err)
	}

	// Check if slug already exists for this application
	existing, _ := s.envRepo.GetBySlug(ctx, source.ApplicationID, slug)
	if existing != nil {
		return nil, fmt.Errorf("environment with slug '%s' already exists in this application", slug)
	}

	// Create new environment
	newEnv := entity.NewEnvironment(source.ApplicationID, name, slug, source.Color, false)
	newEnv.Description = description

	if err := s.envRepo.Create(ctx, newEnv); err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Copy feature flags
	if s.flagRepo != nil {
		if err := s.flagRepo.CopyFlags(ctx, sourceID, newEnv.ID); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to copy flags: %v\n", err)
		}
	}

	return newEnv, nil
}
