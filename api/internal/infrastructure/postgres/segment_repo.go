package postgres

import (
"context"
"database/sql"
"encoding/json"
"fmt"
"time"

"github.com/IzuCas/flagflash/internal/domain/entity"
"github.com/IzuCas/flagflash/internal/domain/repository"
"github.com/google/uuid"
"github.com/lib/pq"
)

// SegmentRepo implements repository.SegmentRepository
type SegmentRepo struct {
	db *DB
}

// NewSegmentRepo creates a new segment repository
func NewSegmentRepo(db *DB) repository.SegmentRepository {
	return &SegmentRepo{db: db}
}

// Create creates a new segment
func (r *SegmentRepo) Create(ctx context.Context, segment *entity.Segment) error {
	conditionsJSON, err := json.Marshal(segment.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		INSERT INTO segments (id, tenant_id, name, description, conditions, is_dynamic, included_users, excluded_users, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
segment.ID, segment.TenantID, segment.Name, segment.Description,
conditionsJSON, segment.IsDynamic, pq.Array(segment.IncludedUsers),
pq.Array(segment.ExcludedUsers), segment.CreatedBy, segment.CreatedAt, segment.UpdatedAt,
)
	if err != nil {
		return fmt.Errorf("failed to create segment: %w", err)
	}

	return nil
}

// GetByID retrieves a segment by ID
func (r *SegmentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Segment, error) {
	query := `
		SELECT id, tenant_id, name, description, conditions, is_dynamic, included_users, excluded_users, created_by, created_at, updated_at
		FROM segments
		WHERE id = $1
	`

	return r.scanSegment(r.db.QueryRowContext(ctx, query, id))
}

// GetByName retrieves a segment by name within a tenant
func (r *SegmentRepo) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*entity.Segment, error) {
	query := `
		SELECT id, tenant_id, name, description, conditions, is_dynamic, included_users, excluded_users, created_by, created_at, updated_at
		FROM segments
		WHERE tenant_id = $1 AND name = $2
	`

	return r.scanSegment(r.db.QueryRowContext(ctx, query, tenantID, name))
}

// ListByTenant lists all segments for a tenant
func (r *SegmentRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.Segment, error) {
	query := `
		SELECT id, tenant_id, name, description, conditions, is_dynamic, included_users, excluded_users, created_by, created_at, updated_at
		FROM segments
		WHERE tenant_id = $1
		ORDER BY name
	`

	return r.scanSegments(r.db.QueryContext(ctx, query, tenantID))
}

// ListByTenantPaginated lists segments with pagination
func (r *SegmentRepo) ListByTenantPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int, search string) ([]*entity.Segment, int, error) {
	var countQuery, query string
	var countArgs, queryArgs []interface{}

	if search != "" {
		searchPattern := "%" + search + "%"
		countQuery = `SELECT COUNT(*) FROM segments WHERE tenant_id = $1 AND (name ILIKE $2 OR description ILIKE $2)`
		countArgs = []interface{}{tenantID, searchPattern}
		query = `
			SELECT id, tenant_id, name, description, conditions, is_dynamic, included_users, excluded_users, created_by, created_at, updated_at
			FROM segments
			WHERE tenant_id = $1 AND (name ILIKE $2 OR description ILIKE $2)
			ORDER BY name
			LIMIT $3 OFFSET $4
		`
		queryArgs = []interface{}{tenantID, searchPattern, limit, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM segments WHERE tenant_id = $1`
		countArgs = []interface{}{tenantID}
		query = `
			SELECT id, tenant_id, name, description, conditions, is_dynamic, included_users, excluded_users, created_by, created_at, updated_at
			FROM segments
			WHERE tenant_id = $1
			ORDER BY name
			LIMIT $2 OFFSET $3
		`
		queryArgs = []interface{}{tenantID, limit, offset}
	}

	// Get total count
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count segments: %w", err)
	}

	// Get paginated results
	segments, err := r.scanSegments(r.db.QueryContext(ctx, query, queryArgs...))
	if err != nil {
		return nil, 0, err
	}

	return segments, total, nil
}

// Update updates a segment
func (r *SegmentRepo) Update(ctx context.Context, segment *entity.Segment) error {
	segment.UpdatedAt = time.Now()

	conditionsJSON, err := json.Marshal(segment.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	query := `
		UPDATE segments
		SET name = $1, description = $2, conditions = $3, is_dynamic = $4, included_users = $5, excluded_users = $6, updated_at = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
segment.Name, segment.Description, conditionsJSON, segment.IsDynamic,
pq.Array(segment.IncludedUsers), pq.Array(segment.ExcludedUsers),
segment.UpdatedAt, segment.ID,
)
	if err != nil {
		return fmt.Errorf("failed to update segment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("segment not found")
	}

	return nil
}

// Delete deletes a segment
func (r *SegmentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM segments WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete segment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("segment not found")
	}

	return nil
}

// AddIncludedUser adds a user to the segment's included list
func (r *SegmentRepo) AddIncludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error {
query := `
UPDATE segments
SET included_users = array_append(included_users, $1), updated_at = $2
WHERE id = $3 AND NOT ($1 = ANY(included_users))
`

_, err := r.db.ExecContext(ctx, query, userID, time.Now(), segmentID)
if err != nil {
return fmt.Errorf("failed to add included user: %w", err)
}

return nil
}

// RemoveIncludedUser removes a user from the segment's included list
func (r *SegmentRepo) RemoveIncludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error {
	query := `
		UPDATE segments
		SET included_users = array_remove(included_users, $1), updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, userID, time.Now(), segmentID)
	if err != nil {
		return fmt.Errorf("failed to remove included user: %w", err)
	}

	return nil
}

// AddExcludedUser adds a user to the segment's excluded list
func (r *SegmentRepo) AddExcludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error {
query := `
UPDATE segments
SET excluded_users = array_append(excluded_users, $1), updated_at = $2
WHERE id = $3 AND NOT ($1 = ANY(excluded_users))
`

_, err := r.db.ExecContext(ctx, query, userID, time.Now(), segmentID)
if err != nil {
return fmt.Errorf("failed to add excluded user: %w", err)
}

return nil
}

// RemoveExcludedUser removes a user from the segment's excluded list
func (r *SegmentRepo) RemoveExcludedUser(ctx context.Context, segmentID uuid.UUID, userID string) error {
	query := `
		UPDATE segments
		SET excluded_users = array_remove(excluded_users, $1), updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, userID, time.Now(), segmentID)
	if err != nil {
		return fmt.Errorf("failed to remove excluded user: %w", err)
	}

	return nil
}

func (r *SegmentRepo) scanSegment(row *sql.Row) (*entity.Segment, error) {
	var segment entity.Segment
	var description sql.NullString
	var conditionsJSON []byte
	var createdBy sql.NullString
	var includedUsers, excludedUsers pq.StringArray

	err := row.Scan(
&segment.ID, &segment.TenantID, &segment.Name, &description,
		&conditionsJSON, &segment.IsDynamic, &includedUsers, &excludedUsers,
		&createdBy, &segment.CreatedAt, &segment.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("segment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan segment: %w", err)
	}

	if description.Valid {
		segment.Description = description.String
	}
	if len(conditionsJSON) > 0 {
		if err := json.Unmarshal(conditionsJSON, &segment.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}
	}
	if createdBy.Valid {
		id, _ := uuid.Parse(createdBy.String)
		segment.CreatedBy = &id
	}
	segment.IncludedUsers = []string(includedUsers)
	segment.ExcludedUsers = []string(excludedUsers)
	if segment.IncludedUsers == nil {
		segment.IncludedUsers = []string{}
	}
	if segment.ExcludedUsers == nil {
		segment.ExcludedUsers = []string{}
	}

	return &segment, nil
}

func (r *SegmentRepo) scanSegments(rows *sql.Rows, err error) ([]*entity.Segment, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to query segments: %w", err)
	}
	defer rows.Close()

	var segments []*entity.Segment
	for rows.Next() {
		var segment entity.Segment
		var description sql.NullString
		var conditionsJSON []byte
		var createdBy sql.NullString
		var includedUsers, excludedUsers pq.StringArray

		err := rows.Scan(
&segment.ID, &segment.TenantID, &segment.Name, &description,
			&conditionsJSON, &segment.IsDynamic, &includedUsers, &excludedUsers,
			&createdBy, &segment.CreatedAt, &segment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan segment: %w", err)
		}

		if description.Valid {
			segment.Description = description.String
		}
		if len(conditionsJSON) > 0 {
			if err := json.Unmarshal(conditionsJSON, &segment.Conditions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
			}
		}
		if createdBy.Valid {
			id, _ := uuid.Parse(createdBy.String)
			segment.CreatedBy = &id
		}
		segment.IncludedUsers = []string(includedUsers)
		segment.ExcludedUsers = []string(excludedUsers)
		if segment.IncludedUsers == nil {
			segment.IncludedUsers = []string{}
		}
		if segment.ExcludedUsers == nil {
			segment.ExcludedUsers = []string{}
		}

		segments = append(segments, &segment)
	}

	return segments, nil
}
