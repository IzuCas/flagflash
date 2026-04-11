package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// UserTenantMembershipRepository implements repository.UserTenantMembershipRepository
type UserTenantMembershipRepository struct {
	db *sql.DB
}

// NewUserTenantMembershipRepository creates a new UserTenantMembershipRepository
func NewUserTenantMembershipRepository(db *sql.DB) *UserTenantMembershipRepository {
	return &UserTenantMembershipRepository{db: db}
}

// Create creates a new membership
func (r *UserTenantMembershipRepository) Create(ctx context.Context, membership *entity.UserTenantMembership) error {
	query := `INSERT INTO user_tenant_memberships (id, user_id, tenant_id, role, active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query,
		membership.ID, membership.UserID, membership.TenantID, membership.Role,
		membership.Active, membership.CreatedAt, membership.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}
	return nil
}

// GetByID retrieves a membership by ID
func (r *UserTenantMembershipRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.UserTenantMembership, error) {
	query := `SELECT id, user_id, tenant_id, role, active, created_at, updated_at
			  FROM user_tenant_memberships WHERE id = $1`
	return r.scanMembership(r.db.QueryRowContext(ctx, query, id))
}

// GetByUserAndTenant retrieves a membership by user and tenant
func (r *UserTenantMembershipRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*entity.UserTenantMembership, error) {
	query := `SELECT id, user_id, tenant_id, role, active, created_at, updated_at
			  FROM user_tenant_memberships WHERE user_id = $1 AND tenant_id = $2`
	return r.scanMembership(r.db.QueryRowContext(ctx, query, userID, tenantID))
}

// Update updates a membership
func (r *UserTenantMembershipRepository) Update(ctx context.Context, membership *entity.UserTenantMembership) error {
	query := `UPDATE user_tenant_memberships SET role = $1, active = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, membership.Role, membership.Active, membership.UpdatedAt, membership.ID)
	if err != nil {
		return fmt.Errorf("failed to update membership: %w", err)
	}
	return nil
}

// Delete deletes a membership
func (r *UserTenantMembershipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM user_tenant_memberships WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}
	return nil
}

// DeleteByUserAndTenant deletes a membership by user and tenant
func (r *UserTenantMembershipRepository) DeleteByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) error {
	query := `DELETE FROM user_tenant_memberships WHERE user_id = $1 AND tenant_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}
	return nil
}

// ListByUser lists all memberships for a user
func (r *UserTenantMembershipRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.UserTenantMembership, error) {
	query := `SELECT id, user_id, tenant_id, role, active, created_at, updated_at
			  FROM user_tenant_memberships WHERE user_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()
	return r.scanMemberships(rows)
}

// ListByTenant lists all memberships for a tenant with pagination
func (r *UserTenantMembershipRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserTenantMembership, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM user_tenant_memberships WHERE tenant_id = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count memberships: %w", err)
	}

	// Get memberships with pagination
	query := `SELECT id, user_id, tenant_id, role, active, created_at, updated_at
			  FROM user_tenant_memberships WHERE tenant_id = $1 ORDER BY created_at LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list memberships: %w", err)
	}
	defer rows.Close()

	memberships, err := r.scanMemberships(rows)
	if err != nil {
		return nil, 0, err
	}
	return memberships, total, nil
}

// ListUsersWithMembershipByTenant lists users with their membership details for a tenant
func (r *UserTenantMembershipRepository) ListUsersWithMembershipByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.UserWithMembership, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM user_tenant_memberships m 
				   JOIN users u ON m.user_id = u.id
				   WHERE m.tenant_id = $1 AND u.deleted_at IS NULL`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with memberships
	query := `SELECT u.id, u.tenant_id, u.email, u.password_hash, u.name, u.role, u.active, u.last_login_at, u.created_at, u.updated_at,
					 m.id, m.user_id, m.tenant_id, m.role, m.active, m.created_at, m.updated_at
			  FROM user_tenant_memberships m
			  JOIN users u ON m.user_id = u.id
			  WHERE m.tenant_id = $1 AND u.deleted_at IS NULL
			  ORDER BY u.name LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var results []*entity.UserWithMembership
	for rows.Next() {
		user := &entity.User{}
		membership := &entity.UserTenantMembership{}
		var userTenantID sql.NullString
		err := rows.Scan(
			&user.ID, &userTenantID, &user.Email, &user.PasswordHash, &user.Name,
			&user.Role, &user.Active, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt,
			&membership.ID, &membership.UserID, &membership.TenantID, &membership.Role,
			&membership.Active, &membership.CreatedAt, &membership.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user with membership: %w", err)
		}
		if userTenantID.Valid {
			user.TenantID, _ = uuid.Parse(userTenantID.String)
		}
		results = append(results, &entity.UserWithMembership{
			User:       user,
			Membership: membership,
		})
	}
	return results, total, nil
}

// ListTenantsForUser lists all tenants for a user with their roles
func (r *UserTenantMembershipRepository) ListTenantsForUser(ctx context.Context, userID uuid.UUID) ([]*entity.TenantWithRole, error) {
	query := `SELECT t.id, t.name, t.slug, t.plan, t.active, t.settings, t.created_at, t.updated_at, t.deleted_at,
					 m.role, m.active
			  FROM user_tenant_memberships m
			  JOIN tenants t ON m.tenant_id = t.id
			  WHERE m.user_id = $1 AND t.deleted_at IS NULL AND m.active = true
			  ORDER BY t.name`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants for user: %w", err)
	}
	defer rows.Close()

	var results []*entity.TenantWithRole
	for rows.Next() {
		tenant := &entity.Tenant{}
		var role entity.UserRole
		var membershipActive bool
		var settings []byte
		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan, &tenant.Active,
			&settings, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
			&role, &membershipActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant with role: %w", err)
		}
		// Parse settings JSON if needed
		results = append(results, &entity.TenantWithRole{
			Tenant: tenant,
			Role:   role,
			Active: membershipActive,
		})
	}
	return results, nil
}

// ExistsByUserAndTenant checks if a membership exists for user and tenant
func (r *UserTenantMembershipRepository) ExistsByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_tenant_memberships WHERE user_id = $1 AND tenant_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check membership existence: %w", err)
	}
	return exists, nil
}

func (r *UserTenantMembershipRepository) scanMembership(row *sql.Row) (*entity.UserTenantMembership, error) {
	membership := &entity.UserTenantMembership{}
	err := row.Scan(
		&membership.ID, &membership.UserID, &membership.TenantID, &membership.Role,
		&membership.Active, &membership.CreatedAt, &membership.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("membership not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan membership: %w", err)
	}
	return membership, nil
}

func (r *UserTenantMembershipRepository) scanMemberships(rows *sql.Rows) ([]*entity.UserTenantMembership, error) {
	var memberships []*entity.UserTenantMembership
	for rows.Next() {
		membership := &entity.UserTenantMembership{}
		if err := rows.Scan(
			&membership.ID, &membership.UserID, &membership.TenantID, &membership.Role,
			&membership.Active, &membership.CreatedAt, &membership.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		memberships = append(memberships, membership)
	}
	return memberships, nil
}
