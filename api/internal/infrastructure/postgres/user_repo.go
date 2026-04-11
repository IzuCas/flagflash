package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
)

// UserRepository implements repository.UserRepository
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (id, tenant_id, email, password_hash, name, role, active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.TenantID, user.Email, user.PasswordHash, user.Name, user.Role,
		user.Active, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `SELECT id, tenant_id, email, password_hash, name, role, active, last_login_at, created_at, updated_at
			  FROM users WHERE id = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRowContext(ctx, query, id))
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, tenant_id, email, password_hash, name, role, active, last_login_at, created_at, updated_at
			  FROM users WHERE email = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRowContext(ctx, query, email))
}

// GetByTenantID retrieves all users for a tenant
func (r *UserRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.User, error) {
	query := `SELECT id, tenant_id, email, password_hash, name, role, active, last_login_at, created_at, updated_at
			  FROM users WHERE tenant_id = $1 ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()
	return r.scanUsers(rows)
}

// GetByIDs retrieves multiple users by IDs
func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*entity.User, error) {
	if len(ids) == 0 {
		return []*entity.User{}, nil
	}

	// Build placeholder string for IN clause
	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`SELECT id, tenant_id, email, password_hash, name, role, active, last_login_at, created_at, updated_at
			  FROM users WHERE id IN (%s) AND deleted_at IS NULL`, placeholders)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}
	defer rows.Close()
	return r.scanUsers(rows)
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `UPDATE users SET name = $1, role = $2, active = $3, password_hash = $4, last_login_at = $5, updated_at = $6 WHERE id = $7`
	_, err := r.db.ExecContext(ctx, query, user.Name, user.Role, user.Active, user.PasswordHash, user.LastLoginAt, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (r *UserRepository) scanUser(row *sql.Row) (*entity.User, error) {
	user := &entity.User{}
	err := row.Scan(
		&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Name,
		&user.Role, &user.Active, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) scanUsers(rows *sql.Rows) ([]*entity.User, error) {
	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		if err := rows.Scan(
			&user.ID, &user.TenantID, &user.Email, &user.PasswordHash, &user.Name,
			&user.Role, &user.Active, &user.LastLoginAt, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return exists, nil
}

// ListByTenant lists all users for a tenant with pagination
func (r *UserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entity.User, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM users WHERE tenant_id = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	query := `SELECT id, tenant_id, email, password_hash, name, role, active, last_login_at, created_at, updated_at
			  FROM users WHERE tenant_id = $1 ORDER BY name LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	users, err := r.scanUsers(rows)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
