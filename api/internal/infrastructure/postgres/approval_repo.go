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

// ApprovalSettingRepo implements repository.ApprovalSettingRepository
type ApprovalSettingRepo struct {
	db *DB
}

// NewApprovalSettingRepo creates a new approval setting repository
func NewApprovalSettingRepo(db *DB) repository.ApprovalSettingRepository {
	return &ApprovalSettingRepo{db: db}
}

// Create creates a new approval setting
func (r *ApprovalSettingRepo) Create(ctx context.Context, setting *entity.ApprovalSetting) error {
	query := `
		INSERT INTO approval_settings (id, tenant_id, environment_id, feature_flag_id, requires_approval, min_approvers, auto_reject_hours, allowed_approver_roles, notify_on_request, notify_on_decision, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.ExecContext(ctx, query,
		setting.ID, setting.TenantID, setting.EnvironmentID, setting.FeatureFlagID,
		setting.RequiresApproval, setting.MinApprovers, setting.AutoRejectHours,
		pq.Array(setting.AllowedApproverRoles), setting.NotifyOnRequest, setting.NotifyOnDecision,
		setting.CreatedAt, setting.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create approval setting: %w", err)
	}

	return nil
}

// Update updates an approval setting
func (r *ApprovalSettingRepo) Update(ctx context.Context, setting *entity.ApprovalSetting) error {
	setting.UpdatedAt = time.Now()

	query := `
		UPDATE approval_settings
		SET requires_approval = $1, min_approvers = $2, auto_reject_hours = $3, allowed_approver_roles = $4, notify_on_request = $5, notify_on_decision = $6, updated_at = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(ctx, query,
		setting.RequiresApproval, setting.MinApprovers, setting.AutoRejectHours,
		pq.Array(setting.AllowedApproverRoles), setting.NotifyOnRequest, setting.NotifyOnDecision,
		setting.UpdatedAt, setting.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update approval setting: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("approval setting not found")
	}

	return nil
}

// Delete deletes an approval setting
func (r *ApprovalSettingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM approval_settings WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete approval setting: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("approval setting not found")
	}

	return nil
}

// GetByID retrieves an approval setting by ID
func (r *ApprovalSettingRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ApprovalSetting, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, requires_approval, min_approvers, auto_reject_hours, allowed_approver_roles, notify_on_request, notify_on_decision, created_at, updated_at
		FROM approval_settings
		WHERE id = $1
	`

	return r.scanSetting(r.db.QueryRowContext(ctx, query, id))
}

// GetByEnvironment retrieves approval setting for an environment
func (r *ApprovalSettingRepo) GetByEnvironment(ctx context.Context, tenantID, environmentID uuid.UUID) (*entity.ApprovalSetting, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, requires_approval, min_approvers, auto_reject_hours, allowed_approver_roles, notify_on_request, notify_on_decision, created_at, updated_at
		FROM approval_settings
		WHERE tenant_id = $1 AND environment_id = $2 AND feature_flag_id IS NULL
	`

	return r.scanSetting(r.db.QueryRowContext(ctx, query, tenantID, environmentID))
}

// GetByFlag retrieves approval setting for a flag
func (r *ApprovalSettingRepo) GetByFlag(ctx context.Context, tenantID, flagID uuid.UUID) (*entity.ApprovalSetting, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, requires_approval, min_approvers, auto_reject_hours, allowed_approver_roles, notify_on_request, notify_on_decision, created_at, updated_at
		FROM approval_settings
		WHERE tenant_id = $1 AND feature_flag_id = $2
	`

	return r.scanSetting(r.db.QueryRowContext(ctx, query, tenantID, flagID))
}

// GetEffective gets the effective approval setting (flag > environment > tenant)
func (r *ApprovalSettingRepo) GetEffective(ctx context.Context, tenantID, environmentID, flagID uuid.UUID) (*entity.ApprovalSetting, error) {
	// Try flag-specific first
	setting, err := r.GetByFlag(ctx, tenantID, flagID)
	if err == nil {
		return setting, nil
	}

	// Try environment-specific
	setting, err = r.GetByEnvironment(ctx, tenantID, environmentID)
	if err == nil {
		return setting, nil
	}

	return nil, fmt.Errorf("no approval setting found")
}

// ListByTenant lists all approval settings for a tenant
func (r *ApprovalSettingRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.ApprovalSetting, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, requires_approval, min_approvers, auto_reject_hours, allowed_approver_roles, notify_on_request, notify_on_decision, created_at, updated_at
		FROM approval_settings
		WHERE tenant_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list approval settings: %w", err)
	}
	defer rows.Close()

	return r.scanSettings(rows)
}

func (r *ApprovalSettingRepo) scanSetting(row *sql.Row) (*entity.ApprovalSetting, error) {
	var setting entity.ApprovalSetting
	var envID, flagID sql.NullString
	var roles pq.StringArray

	err := row.Scan(
		&setting.ID, &setting.TenantID, &envID, &flagID,
		&setting.RequiresApproval, &setting.MinApprovers, &setting.AutoRejectHours,
		&roles, &setting.NotifyOnRequest, &setting.NotifyOnDecision,
		&setting.CreatedAt, &setting.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("approval setting not found")
		}
		return nil, fmt.Errorf("failed to scan approval setting: %w", err)
	}

	if envID.Valid {
		id, _ := uuid.Parse(envID.String)
		setting.EnvironmentID = &id
	}
	if flagID.Valid {
		id, _ := uuid.Parse(flagID.String)
		setting.FeatureFlagID = &id
	}
	setting.AllowedApproverRoles = []string(roles)

	return &setting, nil
}

func (r *ApprovalSettingRepo) scanSettings(rows *sql.Rows) ([]*entity.ApprovalSetting, error) {
	settings := make([]*entity.ApprovalSetting, 0)

	for rows.Next() {
		var setting entity.ApprovalSetting
		var envID, flagID sql.NullString
		var roles pq.StringArray

		err := rows.Scan(
			&setting.ID, &setting.TenantID, &envID, &flagID,
			&setting.RequiresApproval, &setting.MinApprovers, &setting.AutoRejectHours,
			&roles, &setting.NotifyOnRequest, &setting.NotifyOnDecision,
			&setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval setting: %w", err)
		}

		if envID.Valid {
			id, _ := uuid.Parse(envID.String)
			setting.EnvironmentID = &id
		}
		if flagID.Valid {
			id, _ := uuid.Parse(flagID.String)
			setting.FeatureFlagID = &id
		}
		setting.AllowedApproverRoles = []string(roles)

		settings = append(settings, &setting)
	}

	return settings, nil
}

// PendingChangeRepo implements repository.PendingChangeRepository
type PendingChangeRepo struct {
	db *DB
}

// NewPendingChangeRepo creates a new pending change repository
func NewPendingChangeRepo(db *DB) repository.PendingChangeRepository {
	return &PendingChangeRepo{db: db}
}

// Create creates a new pending change
func (r *PendingChangeRepo) Create(ctx context.Context, change *entity.PendingChange) error {
	query := `
		INSERT INTO pending_changes (id, tenant_id, environment_id, feature_flag_id, targeting_rule_id, change_type, entity_type, old_value, new_value, status, requested_by, request_comment, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.ExecContext(ctx, query,
		change.ID, change.TenantID, change.EnvironmentID, change.FeatureFlagID,
		change.TargetingRuleID, change.ChangeType, change.EntityType, change.OldValue,
		change.NewValue, change.Status, change.RequestedBy, change.RequestComment,
		change.ExpiresAt, change.CreatedAt, change.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create pending change: %w", err)
	}

	return nil
}

// Update updates a pending change
func (r *PendingChangeRepo) Update(ctx context.Context, change *entity.PendingChange) error {
	change.UpdatedAt = time.Now()

	query := `
		UPDATE pending_changes
		SET status = $1, decided_at = $2, updated_at = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		change.Status, change.DecidedAt, change.UpdatedAt, change.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update pending change: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("pending change not found")
	}

	return nil
}

// Delete deletes a pending change
func (r *PendingChangeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM pending_changes WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetByID retrieves a pending change by ID
func (r *PendingChangeRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.PendingChange, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, targeting_rule_id, change_type, entity_type, old_value, new_value, status, requested_by, request_comment, decided_at, expires_at, created_at, updated_at
		FROM pending_changes
		WHERE id = $1
	`

	return r.scanChange(r.db.QueryRowContext(ctx, query, id))
}

// GetByIDWithApprovals retrieves a pending change with all approvals
func (r *PendingChangeRepo) GetByIDWithApprovals(ctx context.Context, id uuid.UUID) (*entity.PendingChange, error) {
	change, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load approvals
	approvalQuery := `
		SELECT id, pending_change_id, approver_id, decision, comment, created_at
		FROM approvals
		WHERE pending_change_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, approvalQuery, id)
	if err != nil {
		return change, nil
	}
	defer rows.Close()

	for rows.Next() {
		var approval entity.Approval
		var comment sql.NullString

		if err := rows.Scan(&approval.ID, &approval.PendingChangeID, &approval.ApproverID, &approval.Decision, &comment, &approval.CreatedAt); err != nil {
			continue
		}

		if comment.Valid {
			approval.Comment = comment.String
		}

		change.Approvals = append(change.Approvals, approval)
	}

	return change, nil
}

// ListPending lists all pending changes for a tenant
func (r *PendingChangeRepo) ListPending(ctx context.Context, tenantID uuid.UUID) ([]*entity.PendingChange, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, targeting_rule_id, change_type, entity_type, old_value, new_value, status, requested_by, request_comment, decided_at, expires_at, created_at, updated_at
		FROM pending_changes
		WHERE tenant_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	return r.listChanges(ctx, query, tenantID)
}

// ListPendingByEnvironment lists pending changes for an environment
func (r *PendingChangeRepo) ListPendingByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.PendingChange, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, targeting_rule_id, change_type, entity_type, old_value, new_value, status, requested_by, request_comment, decided_at, expires_at, created_at, updated_at
		FROM pending_changes
		WHERE environment_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	return r.listChanges(ctx, query, environmentID)
}

// ListPendingByFlag lists pending changes for a flag
func (r *PendingChangeRepo) ListPendingByFlag(ctx context.Context, flagID uuid.UUID) ([]*entity.PendingChange, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, targeting_rule_id, change_type, entity_type, old_value, new_value, status, requested_by, request_comment, decided_at, expires_at, created_at, updated_at
		FROM pending_changes
		WHERE feature_flag_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	return r.listChanges(ctx, query, flagID)
}

// ListByRequester lists pending changes requested by a user
func (r *PendingChangeRepo) ListByRequester(ctx context.Context, userID uuid.UUID) ([]*entity.PendingChange, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, targeting_rule_id, change_type, entity_type, old_value, new_value, status, requested_by, request_comment, decided_at, expires_at, created_at, updated_at
		FROM pending_changes
		WHERE requested_by = $1
		ORDER BY created_at DESC
	`

	return r.listChanges(ctx, query, userID)
}

// ListExpired lists expired pending changes
func (r *PendingChangeRepo) ListExpired(ctx context.Context) ([]*entity.PendingChange, error) {
	query := `
		SELECT id, tenant_id, environment_id, feature_flag_id, targeting_rule_id, change_type, entity_type, old_value, new_value, status, requested_by, request_comment, decided_at, expires_at, created_at, updated_at
		FROM pending_changes
		WHERE status = 'pending' AND expires_at IS NOT NULL AND expires_at <= $1
		ORDER BY expires_at
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to list expired changes: %w", err)
	}
	defer rows.Close()

	return r.scanChanges(rows)
}

// ExpireOld marks old pending changes as expired
func (r *PendingChangeRepo) ExpireOld(ctx context.Context) (int, error) {
	query := `
		UPDATE pending_changes
		SET status = 'expired', decided_at = $1, updated_at = $1
		WHERE status = 'pending' AND expires_at IS NOT NULL AND expires_at <= $1
	`

	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to expire old changes: %w", err)
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

func (r *PendingChangeRepo) listChanges(ctx context.Context, query string, arg interface{}) ([]*entity.PendingChange, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending changes: %w", err)
	}
	defer rows.Close()

	return r.scanChanges(rows)
}

func (r *PendingChangeRepo) scanChange(row *sql.Row) (*entity.PendingChange, error) {
	var change entity.PendingChange
	var flagID, ruleID sql.NullString
	var oldValue, newValue []byte
	var comment sql.NullString
	var decidedAt, expiresAt sql.NullTime

	err := row.Scan(
		&change.ID, &change.TenantID, &change.EnvironmentID, &flagID, &ruleID,
		&change.ChangeType, &change.EntityType, &oldValue, &newValue, &change.Status,
		&change.RequestedBy, &comment, &decidedAt, &expiresAt, &change.CreatedAt, &change.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pending change not found")
		}
		return nil, fmt.Errorf("failed to scan pending change: %w", err)
	}

	if flagID.Valid {
		id, _ := uuid.Parse(flagID.String)
		change.FeatureFlagID = &id
	}
	if ruleID.Valid {
		id, _ := uuid.Parse(ruleID.String)
		change.TargetingRuleID = &id
	}
	if comment.Valid {
		change.RequestComment = comment.String
	}
	if decidedAt.Valid {
		change.DecidedAt = &decidedAt.Time
	}
	if expiresAt.Valid {
		change.ExpiresAt = &expiresAt.Time
	}
	change.OldValue = json.RawMessage(oldValue)
	change.NewValue = json.RawMessage(newValue)

	return &change, nil
}

func (r *PendingChangeRepo) scanChanges(rows *sql.Rows) ([]*entity.PendingChange, error) {
	changes := make([]*entity.PendingChange, 0)

	for rows.Next() {
		var change entity.PendingChange
		var flagID, ruleID sql.NullString
		var oldValue, newValue []byte
		var comment sql.NullString
		var decidedAt, expiresAt sql.NullTime

		err := rows.Scan(
			&change.ID, &change.TenantID, &change.EnvironmentID, &flagID, &ruleID,
			&change.ChangeType, &change.EntityType, &oldValue, &newValue, &change.Status,
			&change.RequestedBy, &comment, &decidedAt, &expiresAt, &change.CreatedAt, &change.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending change: %w", err)
		}

		if flagID.Valid {
			id, _ := uuid.Parse(flagID.String)
			change.FeatureFlagID = &id
		}
		if ruleID.Valid {
			id, _ := uuid.Parse(ruleID.String)
			change.TargetingRuleID = &id
		}
		if comment.Valid {
			change.RequestComment = comment.String
		}
		if decidedAt.Valid {
			change.DecidedAt = &decidedAt.Time
		}
		if expiresAt.Valid {
			change.ExpiresAt = &expiresAt.Time
		}
		change.OldValue = json.RawMessage(oldValue)
		change.NewValue = json.RawMessage(newValue)

		changes = append(changes, &change)
	}

	return changes, nil
}

// ApprovalRepo implements repository.ApprovalRepository
type ApprovalRepo struct {
	db *DB
}

// NewApprovalRepo creates a new approval repository
func NewApprovalRepo(db *DB) repository.ApprovalRepository {
	return &ApprovalRepo{db: db}
}

// Create creates a new approval
func (r *ApprovalRepo) Create(ctx context.Context, approval *entity.Approval) error {
	query := `
		INSERT INTO approvals (id, pending_change_id, approver_id, decision, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		approval.ID, approval.PendingChangeID, approval.ApproverID,
		approval.Decision, approval.Comment, approval.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create approval: %w", err)
	}

	return nil
}

// GetByID retrieves an approval by ID
func (r *ApprovalRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Approval, error) {
	query := `
		SELECT id, pending_change_id, approver_id, decision, comment, created_at
		FROM approvals
		WHERE id = $1
	`

	return r.scanApproval(r.db.QueryRowContext(ctx, query, id))
}

// GetByChangeAndApprover gets approval by approver for a pending change
func (r *ApprovalRepo) GetByChangeAndApprover(ctx context.Context, changeID, approverID uuid.UUID) (*entity.Approval, error) {
	query := `
		SELECT id, pending_change_id, approver_id, decision, comment, created_at
		FROM approvals
		WHERE pending_change_id = $1 AND approver_id = $2
	`

	return r.scanApproval(r.db.QueryRowContext(ctx, query, changeID, approverID))
}

// ListByChange lists approvals for a pending change
func (r *ApprovalRepo) ListByChange(ctx context.Context, changeID uuid.UUID) ([]*entity.Approval, error) {
	query := `
		SELECT id, pending_change_id, approver_id, decision, comment, created_at
		FROM approvals
		WHERE pending_change_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, changeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list approvals: %w", err)
	}
	defer rows.Close()

	return r.scanApprovals(rows)
}

// CountByDecision counts approvals by decision for a pending change
func (r *ApprovalRepo) CountByDecision(ctx context.Context, changeID uuid.UUID, decision entity.ApprovalDecision) (int, error) {
	query := `SELECT COUNT(*) FROM approvals WHERE pending_change_id = $1 AND decision = $2`

	var count int
	err := r.db.QueryRowContext(ctx, query, changeID, decision).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count approvals: %w", err)
	}

	return count, nil
}

// Update updates an approval
func (r *ApprovalRepo) Update(ctx context.Context, approval *entity.Approval) error {
	query := `
		UPDATE approvals
		SET decision = $1, comment = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, approval.Decision, approval.Comment, approval.ID)
	if err != nil {
		return fmt.Errorf("failed to update approval: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("approval not found")
	}

	return nil
}

// Delete deletes an approval
func (r *ApprovalRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM approvals WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ApprovalRepo) scanApproval(row *sql.Row) (*entity.Approval, error) {
	var approval entity.Approval
	var comment sql.NullString

	err := row.Scan(&approval.ID, &approval.PendingChangeID, &approval.ApproverID, &approval.Decision, &comment, &approval.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("approval not found")
		}
		return nil, fmt.Errorf("failed to scan approval: %w", err)
	}

	if comment.Valid {
		approval.Comment = comment.String
	}

	return &approval, nil
}

func (r *ApprovalRepo) scanApprovals(rows *sql.Rows) ([]*entity.Approval, error) {
	approvals := make([]*entity.Approval, 0)

	for rows.Next() {
		var approval entity.Approval
		var comment sql.NullString

		if err := rows.Scan(&approval.ID, &approval.PendingChangeID, &approval.ApproverID, &approval.Decision, &comment, &approval.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan approval: %w", err)
		}

		if comment.Valid {
			approval.Comment = comment.String
		}

		approvals = append(approvals, &approval)
	}

	return approvals, nil
}
