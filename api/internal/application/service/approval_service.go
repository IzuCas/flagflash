package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// ApprovalService handles approval workflow logic
type ApprovalService struct {
	settingRepo     repository.ApprovalSettingRepository
	pendingRepo     repository.PendingChangeRepository
	approvalRepo    repository.ApprovalRepository
	flagRepo        repository.FeatureFlagRepository
	ruleRepo        repository.TargetingRuleRepository
	userRepo        repository.UserRepository
	membershipRepo  repository.UserTenantMembershipRepository
	notificationSvc *NotificationService
	webhookSvc      *WebhookService
}

// NewApprovalService creates a new approval service
func NewApprovalService(
	settingRepo repository.ApprovalSettingRepository,
	pendingRepo repository.PendingChangeRepository,
	approvalRepo repository.ApprovalRepository,
	flagRepo repository.FeatureFlagRepository,
	ruleRepo repository.TargetingRuleRepository,
	userRepo repository.UserRepository,
	membershipRepo repository.UserTenantMembershipRepository,
	notificationSvc *NotificationService,
	webhookSvc *WebhookService,
) *ApprovalService {
	return &ApprovalService{
		settingRepo:     settingRepo,
		pendingRepo:     pendingRepo,
		approvalRepo:    approvalRepo,
		flagRepo:        flagRepo,
		ruleRepo:        ruleRepo,
		userRepo:        userRepo,
		membershipRepo:  membershipRepo,
		notificationSvc: notificationSvc,
		webhookSvc:      webhookSvc,
	}
}

// RequiresApproval checks if an action requires approval
func (s *ApprovalService) RequiresApproval(ctx context.Context, tenantID, environmentID uuid.UUID, flagID *uuid.UUID, userRole string) (bool, *entity.ApprovalSetting, error) {
	// Owner role never needs approval
	if userRole == "owner" {
		return false, nil, nil
	}

	// Check for flag-specific setting first
	if flagID != nil {
		setting, err := s.settingRepo.GetByFlag(ctx, tenantID, *flagID)
		if err == nil && setting != nil {
			return setting.RequiresApproval, setting, nil
		}
	}

	// Check for environment-specific setting
	setting, err := s.settingRepo.GetByEnvironment(ctx, tenantID, environmentID)
	if err == nil && setting != nil {
		return setting.RequiresApproval, setting, nil
	}

	// Default: no approval required
	return false, nil, nil
}

// CreatePendingChange creates a new pending change for approval
func (s *ApprovalService) CreatePendingChange(
	ctx context.Context,
	tenantID, environmentID, requesterID uuid.UUID,
	flagID, ruleID *uuid.UUID,
	changeType entity.PendingChangeType,
	entityType entity.PendingChangeEntityType,
	oldValue, newValue interface{},
	comment string,
) (*entity.PendingChange, error) {
	// Get approval settings
	requires, setting, err := s.RequiresApproval(ctx, tenantID, environmentID, flagID, "member")
	if err != nil {
		return nil, fmt.Errorf("failed to check approval requirements: %w", err)
	}

	if !requires {
		return nil, fmt.Errorf("approval not required for this action")
	}

	autoRejectHours := 72
	if setting != nil {
		autoRejectHours = setting.AutoRejectHours
	}

	var oldJSON, newJSON json.RawMessage
	if oldValue != nil {
		oldJSON, _ = json.Marshal(oldValue)
	}
	if newValue != nil {
		newJSON, _ = json.Marshal(newValue)
	}

	change := entity.NewPendingChange(
		tenantID,
		environmentID,
		requesterID,
		flagID,
		ruleID,
		changeType,
		entityType,
		oldJSON,
		newJSON,
		comment,
		autoRejectHours,
	)

	if err := s.pendingRepo.Create(ctx, change); err != nil {
		return nil, fmt.Errorf("failed to create pending change: %w", err)
	}

	// Notify approvers
	if s.notificationSvc != nil && setting != nil && setting.NotifyOnRequest {
		s.notifyApprovers(ctx, tenantID, change, setting.AllowedApproverRoles)
	}

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, tenantID, entity.WebhookEventApprovalRequested, change)
	}

	return change, nil
}

// notifyApprovers sends notifications to all users who can approve
func (s *ApprovalService) notifyApprovers(ctx context.Context, tenantID uuid.UUID, change *entity.PendingChange, roles []string) {
	// Get all members for the tenant
	allMembers, _, _ := s.membershipRepo.ListByTenant(ctx, tenantID, 1000, 0)

	// Filter by allowed roles
	for _, member := range allMembers {
		// Skip the requester
		if member.UserID == change.RequestedBy {
			continue
		}

		// Check if member's role is in the allowed roles list
		memberRoleStr := string(member.Role)
		roleAllowed := false
		for _, role := range roles {
			if memberRoleStr == role {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			continue
		}

		user, _ := s.userRepo.GetByID(ctx, member.UserID)
		if user != nil {
			notification := entity.NewNotification(
				user.ID,
				tenantID,
				entity.NotificationTypeApprovalRequest,
				"New approval request",
				fmt.Sprintf("A %s change requires your approval", change.EntityType),
				fmt.Sprintf("/pending-changes/%s", change.ID),
				map[string]interface{}{
					"change_id":   change.ID,
					"change_type": change.ChangeType,
					"entity_type": change.EntityType,
				},
			)
			s.notificationSvc.Create(ctx, notification)
		}
	}
}

// Approve adds an approval to a pending change
func (s *ApprovalService) Approve(ctx context.Context, changeID, approverID uuid.UUID, comment string) (*entity.Approval, error) {
	return s.decide(ctx, changeID, approverID, entity.ApprovalDecisionApproved, comment)
}

// Reject adds a rejection to a pending change
func (s *ApprovalService) Reject(ctx context.Context, changeID, approverID uuid.UUID, comment string) (*entity.Approval, error) {
	return s.decide(ctx, changeID, approverID, entity.ApprovalDecisionRejected, comment)
}

// RequestChanges adds a "needs changes" decision
func (s *ApprovalService) RequestChanges(ctx context.Context, changeID, approverID uuid.UUID, comment string) (*entity.Approval, error) {
	return s.decide(ctx, changeID, approverID, entity.ApprovalDecisionNeedsChanges, comment)
}

func (s *ApprovalService) decide(ctx context.Context, changeID, approverID uuid.UUID, decision entity.ApprovalDecision, comment string) (*entity.Approval, error) {
	// Get pending change
	change, err := s.pendingRepo.GetByIDWithApprovals(ctx, changeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending change: %w", err)
	}

	if !change.IsPending() {
		return nil, fmt.Errorf("change is no longer pending")
	}

	// Check if user already decided
	existing, _ := s.approvalRepo.GetByChangeAndApprover(ctx, changeID, approverID)
	if existing != nil {
		return nil, fmt.Errorf("you have already made a decision on this change")
	}

	// Verify user has permission to approve
	if err := s.verifyApproverPermission(ctx, change.TenantID, approverID); err != nil {
		return nil, err
	}

	// Create approval
	approval := entity.NewApproval(changeID, approverID, decision, comment)
	if err := s.approvalRepo.Create(ctx, approval); err != nil {
		return nil, fmt.Errorf("failed to create approval: %w", err)
	}

	// Check if change should be auto-approved or rejected
	if err := s.checkAndFinalizeChange(ctx, change); err != nil {
		return nil, err
	}

	return approval, nil
}

func (s *ApprovalService) verifyApproverPermission(ctx context.Context, tenantID, userID uuid.UUID) error {
	membership, err := s.membershipRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil || membership == nil {
		return fmt.Errorf("user is not a member of this tenant")
	}

	// Check if user role is allowed to approve
	allowedRoles := []entity.UserRole{entity.UserRoleOwner, entity.UserRoleAdmin}
	for _, role := range allowedRoles {
		if membership.Role == role {
			return nil
		}
	}

	return fmt.Errorf("user does not have permission to approve")
}

func (s *ApprovalService) checkAndFinalizeChange(ctx context.Context, change *entity.PendingChange) error {
	// Get fresh data with approvals
	change, err := s.pendingRepo.GetByIDWithApprovals(ctx, change.ID)
	if err != nil {
		return err
	}

	// Get approval settings
	setting, _ := s.settingRepo.GetEffective(ctx, change.TenantID, change.EnvironmentID, uuid.Nil)
	minApprovers := 1
	if setting != nil {
		minApprovers = setting.MinApprovers
	}

	// Check for rejection
	if change.CountApprovals(entity.ApprovalDecisionRejected) > 0 {
		change.Reject()
		if err := s.pendingRepo.Update(ctx, change); err != nil {
			return err
		}
		s.notifyDecision(ctx, change, "rejected")
		return nil
	}

	// Check for enough approvals
	if change.HasEnoughApprovals(minApprovers) {
		change.Approve()
		if err := s.pendingRepo.Update(ctx, change); err != nil {
			return err
		}
		// Apply the change
		if err := s.applyChange(ctx, change); err != nil {
			return err
		}
		s.notifyDecision(ctx, change, "approved")
	}

	return nil
}

func (s *ApprovalService) applyChange(ctx context.Context, change *entity.PendingChange) error {
	switch change.EntityType {
	case entity.PendingChangeEntityFlag:
		return s.applyFlagChange(ctx, change)
	case entity.PendingChangeEntityRule:
		return s.applyRuleChange(ctx, change)
	default:
		return fmt.Errorf("unknown entity type: %s", change.EntityType)
	}
}

func (s *ApprovalService) applyFlagChange(ctx context.Context, change *entity.PendingChange) error {
	if change.FeatureFlagID == nil {
		return fmt.Errorf("flag ID is required")
	}

	switch change.ChangeType {
	case entity.PendingChangeTypeEnable:
		flag, err := s.flagRepo.GetByID(ctx, *change.FeatureFlagID)
		if err != nil {
			return err
		}
		flag.Enable()
		return s.flagRepo.Update(ctx, flag)

	case entity.PendingChangeTypeDisable:
		flag, err := s.flagRepo.GetByID(ctx, *change.FeatureFlagID)
		if err != nil {
			return err
		}
		flag.Disable()
		return s.flagRepo.Update(ctx, flag)

	case entity.PendingChangeTypeUpdate:
		if change.NewValue == nil {
			return fmt.Errorf("new value is required for update")
		}
		var flagUpdate entity.FeatureFlag
		if err := json.Unmarshal(change.NewValue, &flagUpdate); err != nil {
			return err
		}
		flag, err := s.flagRepo.GetByID(ctx, *change.FeatureFlagID)
		if err != nil {
			return err
		}
		flag.Update(flagUpdate.Name, flagUpdate.Description, flagUpdate.DefaultValue, flagUpdate.Tags)
		return s.flagRepo.Update(ctx, flag)

	case entity.PendingChangeTypeDelete:
		return s.flagRepo.Delete(ctx, *change.FeatureFlagID)

	default:
		return fmt.Errorf("unsupported change type: %s", change.ChangeType)
	}
}

func (s *ApprovalService) applyRuleChange(ctx context.Context, change *entity.PendingChange) error {
	// Similar implementation for targeting rules
	return nil
}

func (s *ApprovalService) notifyDecision(ctx context.Context, change *entity.PendingChange, decision string) {
	if s.notificationSvc == nil {
		return
	}

	notification := entity.NewNotification(
		change.RequestedBy,
		change.TenantID,
		entity.NotificationTypeApprovalDecision,
		fmt.Sprintf("Your change was %s", decision),
		fmt.Sprintf("Your %s change has been %s", change.EntityType, decision),
		"",
		map[string]interface{}{
			"change_id": change.ID,
			"decision":  decision,
		},
	)
	s.notificationSvc.Create(ctx, notification)

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, change.TenantID, entity.WebhookEventApprovalDecided, map[string]interface{}{
			"change":   change,
			"decision": decision,
		})
	}
}

// GetPendingChange gets a pending change by ID
func (s *ApprovalService) GetPendingChange(ctx context.Context, id uuid.UUID) (*entity.PendingChange, error) {
	return s.pendingRepo.GetByIDWithApprovals(ctx, id)
}

// ListPendingChanges lists pending changes for a tenant
func (s *ApprovalService) ListPendingChanges(ctx context.Context, tenantID uuid.UUID) ([]*entity.PendingChange, error) {
	return s.pendingRepo.ListPending(ctx, tenantID)
}

// ListMyPendingChanges lists pending changes created by a user
func (s *ApprovalService) ListMyPendingChanges(ctx context.Context, userID uuid.UUID) ([]*entity.PendingChange, error) {
	return s.pendingRepo.ListByRequester(ctx, userID)
}

// CancelPendingChange cancels a pending change
func (s *ApprovalService) CancelPendingChange(ctx context.Context, changeID, userID uuid.UUID) error {
	change, err := s.pendingRepo.GetByID(ctx, changeID)
	if err != nil {
		return err
	}

	if change.RequestedBy != userID {
		return fmt.Errorf("only the requester can cancel this change")
	}

	if !change.IsPending() {
		return fmt.Errorf("change is no longer pending")
	}

	change.Cancel()
	return s.pendingRepo.Update(ctx, change)
}

// CreateOrUpdateSetting creates or updates approval settings
func (s *ApprovalService) CreateOrUpdateSetting(ctx context.Context, setting *entity.ApprovalSetting) error {
	existing, _ := s.settingRepo.GetByID(ctx, setting.ID)
	if existing != nil {
		return s.settingRepo.Update(ctx, setting)
	}
	return s.settingRepo.Create(ctx, setting)
}

// GetApprovalSettings gets approval settings for an environment
func (s *ApprovalService) GetApprovalSettings(ctx context.Context, tenantID, environmentID uuid.UUID) (*entity.ApprovalSetting, error) {
	return s.settingRepo.GetByEnvironment(ctx, tenantID, environmentID)
}

// ListApprovalSettings lists all approval settings for a tenant
func (s *ApprovalService) ListApprovalSettings(ctx context.Context, tenantID uuid.UUID) ([]*entity.ApprovalSetting, error) {
	return s.settingRepo.ListByTenant(ctx, tenantID)
}
