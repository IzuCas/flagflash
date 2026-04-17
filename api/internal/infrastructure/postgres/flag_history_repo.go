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
)

// FlagHistoryRepo implements repository.FlagHistoryRepository
type FlagHistoryRepo struct {
	db *DB
}

// NewFlagHistoryRepo creates a new flag history repository
func NewFlagHistoryRepo(db *DB) repository.FlagHistoryRepository {
	return &FlagHistoryRepo{db: db}
}

// Create creates a new flag history entry
func (r *FlagHistoryRepo) Create(ctx context.Context, history *entity.FlagHistory) error {
	// Extract enabled and value from NewState if present
	var enabled bool
	var value json.RawMessage = []byte("{}")
	if len(history.NewState) > 0 {
		var state map[string]interface{}
		if err := json.Unmarshal(history.NewState, &state); err == nil {
			if e, ok := state["enabled"].(bool); ok {
				enabled = e
			}
			if v, ok := state["value"]; ok {
				value, _ = json.Marshal(v)
			}
		}
	}

	query := `
		INSERT INTO flag_history (id, feature_flag_id, version, change_type, changed_by, enabled, value, change_summary, snapshot, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		history.ID, history.FeatureFlagID, history.Version, history.ChangeType,
		history.ChangedBy, enabled, value, history.Comment, history.NewState, history.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create flag history: %w", err)
	}

	return nil
}

// GetByID retrieves a history entry by ID
func (r *FlagHistoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.FlagHistory, error) {
	query := `
		SELECT id, feature_flag_id, version, change_type, changed_by, change_summary, snapshot, created_at
		FROM flag_history
		WHERE id = $1
	`

	return r.scanHistory(r.db.QueryRowContext(ctx, query, id))
}

// ListByFlag lists all history for a flag
func (r *FlagHistoryRepo) ListByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.FlagHistory, error) {
	query := `
		SELECT h.id, h.feature_flag_id, h.version, h.change_type, h.changed_by, h.change_summary, h.snapshot, h.created_at, COALESCE(u.email, '') as changed_by_name
		FROM flag_history h
		LEFT JOIN users u ON h.changed_by = u.id
		WHERE h.feature_flag_id = $1
		ORDER BY h.version DESC
	`

	return r.scanHistories(r.db.QueryContext(ctx, query, flagID))
}

// ListByFlagPaginated lists history with pagination
func (r *FlagHistoryRepo) ListByFlagPaginated(ctx context.Context, flagID uuid.UUID, limit, offset int) ([]*entity.FlagHistory, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM flag_history WHERE feature_flag_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, flagID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count flag history: %w", err)
	}

	// Get paginated results
	query := `
		SELECT h.id, h.feature_flag_id, h.version, h.change_type, h.changed_by, h.change_summary, h.snapshot, h.created_at, COALESCE(u.email, '') as changed_by_name
		FROM flag_history h
		LEFT JOIN users u ON h.changed_by = u.id
		WHERE h.feature_flag_id = $1
		ORDER BY h.version DESC
		LIMIT $2 OFFSET $3
	`

	histories, err := r.scanHistories(r.db.QueryContext(ctx, query, flagID, limit, offset))
	if err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

// GetLatestByFlag gets the latest history entry for a flag
func (r *FlagHistoryRepo) GetLatestByFlag(ctx context.Context, flagID uuid.UUID) (*entity.FlagHistory, error) {
	query := `
		SELECT id, feature_flag_id, version, change_type, changed_by, change_summary, snapshot, created_at
		FROM flag_history
		WHERE feature_flag_id = $1
		ORDER BY version DESC
		LIMIT 1
	`

	return r.scanHistory(r.db.QueryRowContext(ctx, query, flagID))
}

// GetByVersion gets a history entry for a specific version
func (r *FlagHistoryRepo) GetByVersion(ctx context.Context, flagID uuid.UUID, version int) (*entity.FlagHistory, error) {
	query := `
		SELECT id, feature_flag_id, version, change_type, changed_by, change_summary, snapshot, created_at
		FROM flag_history
		WHERE feature_flag_id = $1 AND version = $2
	`

	return r.scanHistory(r.db.QueryRowContext(ctx, query, flagID, version))
}

// Delete deletes a history entry
func (r *FlagHistoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM flag_history WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete flag history: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("flag history not found")
	}

	return nil
}

// DeleteByFlag deletes all history for a flag
func (r *FlagHistoryRepo) DeleteByFlag(ctx context.Context, flagID uuid.UUID) error {
	query := `DELETE FROM flag_history WHERE feature_flag_id = $1`

	_, err := r.db.ExecContext(ctx, query, flagID)
	if err != nil {
		return fmt.Errorf("failed to delete flag history: %w", err)
	}

	return nil
}

// DeleteOlderThan deletes history older than specified days
func (r *FlagHistoryRepo) DeleteOlderThan(ctx context.Context, days int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	query := `DELETE FROM flag_history WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old flag history: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

func (r *FlagHistoryRepo) scanHistory(row *sql.Row) (*entity.FlagHistory, error) {
	var history entity.FlagHistory
	var changedBy sql.NullString
	var snapshot []byte
	var changeSummary sql.NullString

	err := row.Scan(
		&history.ID, &history.FeatureFlagID, &history.Version, &history.ChangeType,
		&changedBy, &changeSummary, &snapshot, &history.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("flag history not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan flag history: %w", err)
	}

	if changedBy.Valid {
		id, _ := uuid.Parse(changedBy.String)
		history.ChangedBy = &id
	}
	if len(snapshot) > 0 {
		history.NewState = json.RawMessage(snapshot)
	}
	if changeSummary.Valid {
		history.Comment = changeSummary.String
	}

	return &history, nil
}

func (r *FlagHistoryRepo) scanHistories(rows *sql.Rows, err error) ([]*entity.FlagHistory, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to query flag history: %w", err)
	}
	defer rows.Close()

	var histories []*entity.FlagHistory
	for rows.Next() {
		var history entity.FlagHistory
		var changedBy sql.NullString
		var snapshot []byte
		var changeSummary sql.NullString
		var changedByName string

		err := rows.Scan(
			&history.ID, &history.FeatureFlagID, &history.Version, &history.ChangeType,
			&changedBy, &changeSummary, &snapshot, &history.CreatedAt, &changedByName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flag history: %w", err)
		}

		if changedBy.Valid {
			id, _ := uuid.Parse(changedBy.String)
			history.ChangedBy = &id
		}
		if len(snapshot) > 0 {
			history.NewState = json.RawMessage(snapshot)
		}
		if changeSummary.Valid {
			history.Comment = changeSummary.String
		}
		history.ChangedByName = changedByName

		histories = append(histories, &history)
	}

	return histories, nil
}
