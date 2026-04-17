package service

import (
	"context"
	"fmt"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/google/uuid"
)

// EmergencyControlService handles emergency control operations (kill switches)
type EmergencyControlService struct {
	controlRepo repository.EmergencyControlRepository
	flagRepo    repository.FeatureFlagRepository
	auditRepo   repository.AuditLogRepository
	webhookSvc  *WebhookService
	notifySvc   *NotificationService
}

// NewEmergencyControlService creates a new emergency control service
func NewEmergencyControlService(
	controlRepo repository.EmergencyControlRepository,
	flagRepo repository.FeatureFlagRepository,
	auditRepo repository.AuditLogRepository,
	webhookSvc *WebhookService,
	notifySvc *NotificationService,
) *EmergencyControlService {
	return &EmergencyControlService{
		controlRepo: controlRepo,
		flagRepo:    flagRepo,
		auditRepo:   auditRepo,
		webhookSvc:  webhookSvc,
		notifySvc:   notifySvc,
	}
}

// ActivateKillSwitch activates a kill switch for an environment
func (s *EmergencyControlService) ActivateKillSwitch(
	ctx context.Context,
	tenantID uuid.UUID,
	environmentID *uuid.UUID,
	reason string,
	duration *time.Duration,
	activatedBy uuid.UUID,
) (*entity.EmergencyControl, error) {
	// Check for existing active kill switch
	existing, _ := s.controlRepo.GetActiveKillSwitch(ctx, tenantID, environmentID)
	if existing != nil && existing.Enabled {
		return nil, fmt.Errorf("kill switch is already active")
	}

	control := entity.NewEmergencyControl(tenantID, environmentID, entity.EmergencyControlTypeKillSwitch)
	control.Enable(activatedBy, reason, duration)

	if err := s.controlRepo.Create(ctx, control); err != nil {
		return nil, err
	}

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, tenantID, entity.WebhookEventEmergencyEnabled, control)
	}

	// Send notifications
	if s.notifySvc != nil {
		s.notifySvc.NotifyKillSwitchActivated(ctx, tenantID, control, &activatedBy)
	}

	return control, nil
}

// ActivateReadOnlyMode activates read-only mode for a tenant
func (s *EmergencyControlService) ActivateReadOnlyMode(
	ctx context.Context,
	tenantID uuid.UUID,
	reason string,
	duration *time.Duration,
	activatedBy uuid.UUID,
) (*entity.EmergencyControl, error) {
	// Check for existing active read-only mode
	existing, _ := s.controlRepo.GetByType(ctx, tenantID, nil, entity.EmergencyControlTypeReadOnly)
	if existing != nil && existing.Enabled {
		return nil, fmt.Errorf("read-only mode is already active")
	}

	control := entity.NewEmergencyControl(tenantID, nil, entity.EmergencyControlTypeReadOnly)
	control.Enable(activatedBy, reason, duration)

	if err := s.controlRepo.Create(ctx, control); err != nil {
		return nil, err
	}

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, tenantID, entity.WebhookEventEmergencyEnabled, control)
	}

	return control, nil
}

// ActivateMaintenanceMode activates maintenance mode for a tenant
func (s *EmergencyControlService) ActivateMaintenanceMode(
	ctx context.Context,
	tenantID uuid.UUID,
	reason string,
	duration *time.Duration,
	activatedBy uuid.UUID,
) (*entity.EmergencyControl, error) {
	// Check for existing maintenance mode
	existing, _ := s.controlRepo.GetByType(ctx, tenantID, nil, entity.EmergencyControlTypeMaintenance)
	if existing != nil && existing.Enabled {
		return nil, fmt.Errorf("maintenance mode is already active")
	}

	control := entity.NewEmergencyControl(tenantID, nil, entity.EmergencyControlTypeMaintenance)
	control.Enable(activatedBy, reason, duration)

	if err := s.controlRepo.Create(ctx, control); err != nil {
		return nil, err
	}

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, tenantID, entity.WebhookEventEmergencyEnabled, control)
	}

	return control, nil
}

// Deactivate deactivates an emergency control
func (s *EmergencyControlService) Deactivate(ctx context.Context, id uuid.UUID, deactivatedBy uuid.UUID) error {
	control, err := s.controlRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !control.Enabled {
		return fmt.Errorf("emergency control is not active")
	}

	control.Disable()

	if err := s.controlRepo.Update(ctx, control); err != nil {
		return err
	}

	// Trigger webhook
	if s.webhookSvc != nil {
		s.webhookSvc.TriggerEvent(ctx, control.TenantID, entity.WebhookEventEmergencyDisabled, control)
	}

	return nil
}

// GetByID gets an emergency control by ID
func (s *EmergencyControlService) GetByID(ctx context.Context, id uuid.UUID) (*entity.EmergencyControl, error) {
	return s.controlRepo.GetByID(ctx, id)
}

// ListActive lists all active emergency controls for a tenant
func (s *EmergencyControlService) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*entity.EmergencyControl, error) {
	return s.controlRepo.ListEnabled(ctx, tenantID)
}

// ListAll lists all emergency controls (including inactive) for a tenant
func (s *EmergencyControlService) ListAll(ctx context.Context, tenantID uuid.UUID) ([]*entity.EmergencyControl, error) {
	return s.controlRepo.ListByTenant(ctx, tenantID)
}

// GetActiveKillSwitch gets the active kill switch for an environment
func (s *EmergencyControlService) GetActiveKillSwitch(ctx context.Context, tenantID uuid.UUID, environmentID *uuid.UUID) (*entity.EmergencyControl, error) {
	return s.controlRepo.GetActiveKillSwitch(ctx, tenantID, environmentID)
}

// IsKillSwitchActive checks if a kill switch is active for an environment
func (s *EmergencyControlService) IsKillSwitchActive(ctx context.Context, tenantID uuid.UUID, environmentID *uuid.UUID) (bool, error) {
	control, err := s.controlRepo.GetActiveKillSwitch(ctx, tenantID, environmentID)
	if err != nil {
		return false, nil // No kill switch
	}
	return control.Enabled && !control.IsExpired(), nil
}

// IsReadOnlyMode checks if read-only mode is active for a tenant
func (s *EmergencyControlService) IsReadOnlyMode(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	control, err := s.controlRepo.GetByType(ctx, tenantID, nil, entity.EmergencyControlTypeReadOnly)
	if err != nil {
		return false, nil
	}
	return control.Enabled && !control.IsExpired(), nil
}

// IsMaintenanceMode checks if maintenance mode is active for a tenant
func (s *EmergencyControlService) IsMaintenanceMode(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	control, err := s.controlRepo.GetByType(ctx, tenantID, nil, entity.EmergencyControlTypeMaintenance)
	if err != nil {
		return false, nil
	}
	return control.Enabled && !control.IsExpired(), nil
}

// ProcessExpired processes and deactivates expired emergency controls
func (s *EmergencyControlService) ProcessExpired(ctx context.Context) error {
	_, err := s.controlRepo.DisableExpired(ctx)
	return err
}

// Extend extends the duration of an emergency control
func (s *EmergencyControlService) Extend(ctx context.Context, id uuid.UUID, additionalDuration time.Duration) error {
	control, err := s.controlRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !control.Enabled {
		return fmt.Errorf("cannot extend inactive emergency control")
	}

	if control.ExpiresAt == nil {
		return fmt.Errorf("emergency control has no expiration to extend")
	}

	newExpiry := control.ExpiresAt.Add(additionalDuration)
	control.ExpiresAt = &newExpiry
	control.UpdatedAt = time.Now()

	return s.controlRepo.Update(ctx, control)
}

// GetEmergencyStatus returns the emergency status for a tenant
func (s *EmergencyControlService) GetEmergencyStatus(ctx context.Context, tenantID uuid.UUID) (*EmergencyStatus, error) {
	activeControls, err := s.controlRepo.ListEnabled(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	status := &EmergencyStatus{
		HasActiveEmergencies: len(activeControls) > 0,
		ActiveControls:       activeControls,
		IsReadOnly:           false,
		IsMaintenance:        false,
		ActiveKillSwitches:   make([]*entity.EmergencyControl, 0),
	}

	for _, control := range activeControls {
		switch control.ControlType {
		case entity.EmergencyControlTypeReadOnly:
			status.IsReadOnly = true
		case entity.EmergencyControlTypeMaintenance:
			status.IsMaintenance = true
		case entity.EmergencyControlTypeKillSwitch:
			status.ActiveKillSwitches = append(status.ActiveKillSwitches, control)
		}
	}

	return status, nil
}

// EmergencyStatus represents the emergency status of a tenant
type EmergencyStatus struct {
	HasActiveEmergencies bool                       `json:"has_active_emergencies"`
	ActiveControls       []*entity.EmergencyControl `json:"active_controls"`
	IsReadOnly           bool                       `json:"is_read_only"`
	IsMaintenance        bool                       `json:"is_maintenance"`
	ActiveKillSwitches   []*entity.EmergencyControl `json:"active_kill_switches"`
}
