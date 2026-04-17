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

// NotificationRepo implements repository.NotificationRepository
type NotificationRepo struct {
	db *DB
}

// NewNotificationRepo creates a new notification repository
func NewNotificationRepo(db *DB) repository.NotificationRepository {
	return &NotificationRepo{db: db}
}

// Create creates a new notification
func (r *NotificationRepo) Create(ctx context.Context, notification *entity.Notification) error {
	query := `
		INSERT INTO notifications (id, tenant_id, user_id, type, title, message, link, read, read_at, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
notification.ID, notification.TenantID, notification.UserID, notification.Type,
notification.Title, notification.Message, notification.Link, notification.Read,
notification.ReadAt, notification.Metadata, notification.CreatedAt,
)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

// CreateBatch creates multiple notifications
func (r *NotificationRepo) CreateBatch(ctx context.Context, notifications []*entity.Notification) error {
	for _, n := range notifications {
		if err := r.Create(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

// GetByID retrieves a notification by ID
func (r *NotificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	query := `
		SELECT id, tenant_id, user_id, type, title, message, link, read, read_at, metadata, created_at
		FROM notifications
		WHERE id = $1
	`

	return r.scanNotification(r.db.QueryRowContext(ctx, query, id))
}

// ListByUser lists notifications for a user with pagination
func (r *NotificationRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, tenant_id, user_id, type, title, message, link, read, read_at, metadata, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifications, err := r.scanNotifications(rows)
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// ListUnread lists unread notifications for a user
func (r *NotificationRepo) ListUnread(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error) {
	query := `
		SELECT id, tenant_id, user_id, type, title, message, link, read, read_at, metadata, created_at
		FROM notifications
		WHERE user_id = $1 AND read = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list unread notifications: %w", err)
	}
	defer rows.Close()

	return r.scanNotifications(rows)
}

// CountUnread counts unread notifications for a user
func (r *NotificationRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	query := `UPDATE notifications SET read = true, read_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE notifications SET read = true, read_at = $1 WHERE user_id = $2 AND read = false`

	_, err := r.db.ExecContext(ctx, query, now, userID)
	return err
}

// Delete deletes a notification
func (r *NotificationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}

// DeleteOlderThan deletes notifications older than specified days
func (r *NotificationRepo) DeleteOlderThan(ctx context.Context, days int) (int, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	query := `DELETE FROM notifications WHERE created_at < $1`

	result, err := r.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old notifications: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

func (r *NotificationRepo) scanNotification(row *sql.Row) (*entity.Notification, error) {
	var notification entity.Notification
	var message, link sql.NullString
	var readAt sql.NullTime
	var metadata []byte

	err := row.Scan(
&notification.ID, &notification.TenantID, &notification.UserID,
		&notification.Type, &notification.Title, &message,
		&link, &notification.Read, &readAt, &metadata, &notification.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to scan notification: %w", err)
	}

	if message.Valid {
		notification.Message = message.String
	}
	if link.Valid {
		notification.Link = link.String
	}
	if readAt.Valid {
		notification.ReadAt = &readAt.Time
	}
	if len(metadata) > 0 {
		notification.Metadata = json.RawMessage(metadata)
	}

	return &notification, nil
}

func (r *NotificationRepo) scanNotifications(rows *sql.Rows) ([]*entity.Notification, error) {
	notifications := make([]*entity.Notification, 0)

	for rows.Next() {
		var notification entity.Notification
		var message, link sql.NullString
		var readAt sql.NullTime
		var metadata []byte

		err := rows.Scan(
&notification.ID, &notification.TenantID, &notification.UserID,
			&notification.Type, &notification.Title, &message,
			&link, &notification.Read, &readAt, &metadata, &notification.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if message.Valid {
			notification.Message = message.String
		}
		if link.Valid {
			notification.Link = link.String
		}
		if readAt.Valid {
			notification.ReadAt = &readAt.Time
		}
		if len(metadata) > 0 {
			notification.Metadata = json.RawMessage(metadata)
		}

		notifications = append(notifications, &notification)
	}

	return notifications, nil
}
