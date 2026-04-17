package dto

import (
	"time"

	"github.com/google/uuid"
)

// Pagination
type PaginationRequest struct {
	Page  int `query:"page" default:"1" minimum:"1"`
	Limit int `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Error responses
type ErrorResponse struct {
	Body struct {
		Error   string `json:"error"`
		Message string `json:"message,omitempty"`
	}
}

// ===== Tenant DTOs =====
type CreateTenantRequest struct {
	Body struct {
		Name string `json:"name" minLength:"1" maxLength:"255"`
		Slug string `json:"slug" minLength:"1" maxLength:"100" pattern:"^[a-z0-9-]+$"`
	}
}

type UpdateTenantRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Name string `json:"name" minLength:"1" maxLength:"255"`
	}
}

type TenantResponse struct {
	Body TenantDTO
}

type TenantsListResponse struct {
	Body struct {
		Tenants    []TenantDTO        `json:"tenants"`
		Pagination PaginationResponse `json:"pagination"`
	}
}

// ===== Application DTOs =====
type CreateApplicationRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Name        string `json:"name" minLength:"1" maxLength:"255"`
		Slug        string `json:"slug" minLength:"1" maxLength:"100" pattern:"^[a-z0-9-]+$"`
		Description string `json:"description,omitempty"`
	}
}

type UpdateApplicationRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	Body     struct {
		Name        string `json:"name" minLength:"1" maxLength:"255"`
		Description string `json:"description,omitempty"`
	}
}

type ApplicationResponse struct {
	Body ApplicationDTO
}

type ApplicationsListResponse struct {
	Body struct {
		Applications []ApplicationDTO   `json:"applications"`
		Pagination   PaginationResponse `json:"pagination"`
	}
}

// ===== Environment DTOs =====
type CreateEnvironmentRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	Body     struct {
		Name        string `json:"name" minLength:"1" maxLength:"100"`
		Slug        string `json:"slug" minLength:"1" maxLength:"100" pattern:"^[a-z0-9-]+$"`
		Description string `json:"description,omitempty"`
		Color       string `json:"color,omitempty" pattern:"^#[0-9A-Fa-f]{6}$"`
	}
}

type UpdateEnvironmentRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	Body     struct {
		Name        string `json:"name" minLength:"1" maxLength:"100"`
		Description string `json:"description,omitempty"`
		Color       string `json:"color,omitempty" pattern:"^#[0-9A-Fa-f]{6}$"`
	}
}

type EnvironmentResponse struct {
	Body EnvironmentDTO
}

type EnvironmentsListResponse struct {
	Body struct {
		Environments []EnvironmentDTO `json:"environments"`
	}
}

// ===== Feature Flag DTOs =====
type CreateFeatureFlagRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	Body     struct {
		Key          string   `json:"key" minLength:"1" maxLength:"255" pattern:"^[a-zA-Z][a-zA-Z0-9_-]*$"`
		Name         string   `json:"name" minLength:"1" maxLength:"255"`
		Description  string   `json:"description,omitempty"`
		Type         string   `json:"type" enum:"boolean,string,number,json"`
		DefaultValue any      `json:"default_value"`
		Enabled      bool     `json:"enabled"`
		Tags         []string `json:"tags,omitempty"`
	}
}

type UpdateFeatureFlagRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	Body     struct {
		Name         string   `json:"name,omitempty"`
		Description  string   `json:"description,omitempty"`
		DefaultValue any      `json:"default_value,omitempty"`
		Enabled      *bool    `json:"enabled,omitempty"`
		Tags         []string `json:"tags,omitempty"`
	}
}

type ToggleFeatureFlagRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	Body     struct {
		Enabled bool `json:"enabled"`
	}
}

type CopyFeatureFlagsRequest struct {
	TenantID    string `path:"tenant_id" format:"uuid"`
	AppID       string `path:"app_id" format:"uuid"`
	SourceEnvID string `path:"source_env_id" format:"uuid"`
	Body        struct {
		TargetEnvironmentID uuid.UUID `json:"target_environment_id"`
		FlagKeys            []string  `json:"flag_keys,omitempty"` // Empty = all flags
		Overwrite           bool      `json:"overwrite"`
	}
}

type CopyFeatureFlagsResponse struct {
	Body struct {
		Message string `json:"message"`
	}
}

// FeatureFlagDTO is used for flat flag data in lists (no Body wrapper)
type FeatureFlagDTO struct {
	ID            uuid.UUID `json:"id"`
	EnvironmentID uuid.UUID `json:"environment_id"`
	Key           string    `json:"key"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Type          string    `json:"type"`
	DefaultValue  any       `json:"default_value"`
	Enabled       bool      `json:"enabled"`
	Version       int       `json:"version"`
	Tags          []string  `json:"tags"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type FeatureFlagResponse struct {
	Body FeatureFlagDTO
}

type FeatureFlagsListResponse struct {
	Body struct {
		Flags      []FeatureFlagDTO   `json:"flags"`
		Pagination PaginationResponse `json:"pagination"`
	}
}

// ===== Targeting Rule DTOs =====
type CreateTargetingRuleRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	Body     struct {
		Name        string         `json:"name" minLength:"1" maxLength:"255"`
		Description string         `json:"description,omitempty"`
		Priority    int            `json:"priority,omitempty" minimum:"0"`
		Conditions  []ConditionDTO `json:"conditions" minItems:"1"`
		Value       any            `json:"value"`
		Percentage  int            `json:"percentage,omitempty" minimum:"0" maximum:"100"`
		Enabled     bool           `json:"enabled,omitempty"`
	}
}

type ConditionDTO struct {
	Attribute string `json:"attribute" minLength:"1"`
	Operator  string `json:"operator" enum:"eq,neq,gt,gte,lt,lte,contains,not_contains,starts_with,ends_with,in,not_in,matches,exists"`
	Value     any    `json:"value"`
}

type UpdateTargetingRuleRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	RuleID   string `path:"rule_id" format:"uuid"`
	Body     struct {
		Name        string         `json:"name,omitempty"`
		Description string         `json:"description,omitempty"`
		Priority    int            `json:"priority,omitempty"`
		Conditions  []ConditionDTO `json:"conditions,omitempty"`
		Value       any            `json:"value,omitempty"`
		Percentage  int            `json:"percentage,omitempty" minimum:"0" maximum:"100"`
		Enabled     *bool          `json:"enabled,omitempty"`
	}
}

type TargetingRuleResponse struct {
	Body TargetingRuleDTO
}

type TargetingRulesListResponse struct {
	Body struct {
		Rules []TargetingRuleDTO `json:"rules"`
	}
}

// ===== API Key DTOs =====
type CreateAPIKeyRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Name          string     `json:"name" minLength:"1" maxLength:"255"`
		EnvironmentID uuid.UUID  `json:"environment_id"`
		Permissions   []string   `json:"permissions" minItems:"1" enum:"read,write,admin"`
		ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	}
}

type UpdateAPIKeyRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	KeyID    string `path:"key_id" format:"uuid"`
	Body     struct {
		Name   string `json:"name,omitempty"`
		Active *bool  `json:"active,omitempty"`
	}
}

type APIKeyResponse struct {
	Body APIKeyDTO
}

type APIKeyCreatedResponse struct {
	Body struct {
		APIKeyDTO
		Key string `json:"key"` // Only returned on creation
	}
}

type APIKeysListResponse struct {
	Body struct {
		Keys       []APIKeyDTO        `json:"keys"`
		Pagination PaginationResponse `json:"pagination"`
	}
}

// ===== Evaluation DTOs =====
type EvaluateFlagRequest struct {
	Body struct {
		FlagKey string                 `json:"flag_key" minLength:"1"`
		Context map[string]interface{} `json:"context,omitempty"`
	}
}

type EvaluateFlagResponse struct {
	Body struct {
		FlagKey  string     `json:"flag_key"`
		Value    any        `json:"value"`
		Enabled  bool       `json:"enabled"`
		Version  int        `json:"version"`
		RuleID   *uuid.UUID `json:"rule_id,omitempty"`
		RuleName string     `json:"rule_name,omitempty"`
	}
}

type EvaluateAllFlagsRequest struct {
	Body struct {
		Context map[string]interface{} `json:"context,omitempty"`
	}
}

type EvaluateAllFlagsResponse struct {
	Body struct {
		Flags map[string]EvaluatedFlag `json:"flags"`
	}
}

type EvaluatedFlag struct {
	Value    any        `json:"value"`
	Enabled  bool       `json:"enabled"`
	Version  int        `json:"version"`
	RuleID   *uuid.UUID `json:"rule_id,omitempty"`
	RuleName string     `json:"rule_name,omitempty"`
}

// ===== Auth DTOs =====
type LoginRequest struct {
	Body struct {
		Email    string `json:"email" format:"email"`
		Password string `json:"password" minLength:"1"`
	}
}

type LoginResponse struct {
	Body LoginResponseBody
}

type RefreshTokenRequest struct {
	Body struct {
		Token string `json:"token" required:"true"`
	}
}

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
}

// ===== Audit Log DTOs =====
type AuditLogsListRequest struct {
	TenantID   string `path:"tenant_id" format:"uuid"`
	EntityType string `query:"entity_type,omitempty"`
	EntityID   string `query:"entity_id,omitempty" format:"uuid"`
	Action     string `query:"action,omitempty"`
	ActorID    string `query:"actor_id,omitempty"`
	StartDate  string `query:"start_date,omitempty"`
	EndDate    string `query:"end_date,omitempty"`
	Page       int    `query:"page" default:"1" minimum:"1"`
	Limit      int    `query:"limit" default:"50" minimum:"1" maximum:"100"`
}

type AuditLogDTO struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	EntityType string    `json:"entity_type"`
	EntityID   uuid.UUID `json:"entity_id"`
	Action     string    `json:"action"`
	ActorID    string    `json:"actor_id"`
	ActorName  string    `json:"actor_name,omitempty"`
	ActorType  string    `json:"actor_type"`
	OldValue   any       `json:"old_value,omitempty"`
	NewValue   any       `json:"new_value,omitempty"`
	Metadata   any       `json:"metadata,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditLogResponse struct {
	Body AuditLogDTO
}

type AuditLogsListResponse struct {
	Body struct {
		Logs       []AuditLogDTO      `json:"logs"`
		Pagination PaginationResponse `json:"pagination"`
	}
}

// ===== Flat DTOs for handlers =====

// TenantDTO represents a tenant for API responses
type TenantDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TenantWithRoleDTO represents a tenant with user's role for API responses
type TenantWithRoleDTO struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ApplicationDTO represents an application for API responses
type ApplicationDTO struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EnvironmentDTO represents an environment for API responses
type EnvironmentDTO struct {
	ID            uuid.UUID `json:"id"`
	ApplicationID uuid.UUID `json:"application_id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description"`
	Color         string    `json:"color"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserDTO represents a user for API responses
type UserDTO struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Role     string    `json:"role"`
}

// LoginResponseBody represents the login response body
type LoginResponseBody struct {
	Token     string              `json:"token"`
	ExpiresAt time.Time           `json:"expires_at"`
	User      UserDTO             `json:"user"`
	Tenants   []TenantWithRoleDTO `json:"tenants"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Body struct {
		Email      string `json:"email" format:"email" required:"true"`
		Password   string `json:"password" minLength:"8" required:"true"`
		Name       string `json:"name" minLength:"1" maxLength:"255" required:"true"`
		TenantName string `json:"tenant_name" minLength:"1" maxLength:"255" required:"true"`
		TenantSlug string `json:"tenant_slug" minLength:"1" maxLength:"100" pattern:"^[a-z0-9-]+$" required:"true"`
	}
}

// RefreshTokenRequestBody represents a refresh token request
type RefreshTokenRequestBody struct {
	Token string `json:"token" required:"true"`
}

// APIKeyDTO represents an API key for API responses
type APIKeyDTO struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	EnvironmentID *uuid.UUID `json:"environment_id,omitempty"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	Permissions   []string   `json:"permissions"`
	Active        bool       `json:"active"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// TargetingRuleDTO represents a targeting rule for API responses
type TargetingRuleDTO struct {
	ID            uuid.UUID      `json:"id"`
	FeatureFlagID uuid.UUID      `json:"feature_flag_id"`
	Name          string         `json:"name"`
	Priority      int            `json:"priority"`
	Conditions    []ConditionDTO `json:"conditions"`
	Value         interface{}    `json:"value"`
	Percentage    int            `json:"percentage"`
	Enabled       bool           `json:"enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// ===== Usage Metrics DTOs =====
type UsageMetricsRequest struct {
	TenantID      string `path:"tenant_id" format:"uuid"`
	EnvironmentID string `query:"environment_id,omitempty" format:"uuid"`
	FlagID        string `query:"flag_id,omitempty" format:"uuid"`
	StartDate     string `query:"start_date"`
	EndDate       string `query:"end_date"`
	Granularity   string `query:"granularity" default:"hour"`
}

type UsageMetricsResponse struct {
	Body UsageMetricsDTO
}

type UsageMetricsDTO struct {
	TenantID         uuid.UUID              `json:"tenant_id"`
	Period           string                 `json:"period"`
	StartDate        time.Time              `json:"start_date"`
	EndDate          time.Time              `json:"end_date"`
	TotalEvaluations int64                  `json:"total_evaluations"`
	UniqueFlags      int                    `json:"unique_flags"`
	UniqueUsers      int                    `json:"unique_users"`
	ByEnvironment    []EnvironmentMetricDTO `json:"by_environment,omitempty"`
	ByFlag           []FlagMetricDTO        `json:"by_flag,omitempty"`
	Timeline         []TimelinePointDTO     `json:"timeline,omitempty"`
}

type EnvironmentMetricDTO struct {
	EnvironmentID   uuid.UUID `json:"environment_id"`
	EnvironmentName string    `json:"environment_name"`
	Evaluations     int64     `json:"evaluations"`
	UniqueFlags     int       `json:"unique_flags"`
	UniqueUsers     int       `json:"unique_users"`
}

type FlagMetricDTO struct {
	FlagID          uuid.UUID `json:"flag_id"`
	FlagKey         string    `json:"flag_key"`
	FlagName        string    `json:"flag_name"`
	EnvironmentID   uuid.UUID `json:"environment_id"`
	EnvironmentName string    `json:"environment_name"`
	Evaluations     int64     `json:"evaluations"`
	TrueCount       int64     `json:"true_count"`
	FalseCount      int64     `json:"false_count"`
	UniqueUsers     int       `json:"unique_users"`
}

type TimelinePointDTO struct {
	Timestamp   time.Time `json:"timestamp"`
	Evaluations int64     `json:"evaluations"`
	TrueCount   int64     `json:"true_count"`
	FalseCount  int64     `json:"false_count"`
}

type TimelineResponse struct {
	Body struct {
		Timeline []TimelinePointDTO `json:"timeline"`
	}
}

type FlagMetricsResponse struct {
	Body struct {
		Flags []FlagMetricDTO `json:"flags"`
	}
}

type EnvironmentMetricsResponse struct {
	Body struct {
		Environments []EnvironmentMetricDTO `json:"environments"`
	}
}

type RecordEvaluationRequest struct {
	Body struct {
		EnvironmentID uuid.UUID              `json:"environment_id" format:"uuid" required:"true"`
		FlagID        uuid.UUID              `json:"flag_id" format:"uuid" required:"true"`
		FlagKey       string                 `json:"flag_key" required:"true"`
		Value         interface{}            `json:"value" required:"true"`
		UserID        *string                `json:"user_id,omitempty"`
		Context       map[string]interface{} `json:"context,omitempty"`
		SDKType       *string                `json:"sdk_type,omitempty"`
		SDKVersion    *string                `json:"sdk_version,omitempty"`
	}
}

type RecordEvaluationBatchRequest struct {
	Body struct {
		Events []EvaluationEventInput `json:"events" required:"true"`
	}
}

type EvaluationEventInput struct {
	EnvironmentID uuid.UUID              `json:"environment_id" format:"uuid" required:"true"`
	FlagID        uuid.UUID              `json:"flag_id" format:"uuid" required:"true"`
	FlagKey       string                 `json:"flag_key" required:"true"`
	Value         interface{}            `json:"value" required:"true"`
	UserID        *string                `json:"user_id,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	SDKType       *string                `json:"sdk_type,omitempty"`
	SDKVersion    *string                `json:"sdk_version,omitempty"`
	EvaluatedAt   *time.Time             `json:"evaluated_at,omitempty"`
}

// ===== Segment DTOs =====
type CreateSegmentRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Name        string         `json:"name" minLength:"1" maxLength:"255"`
		Description string         `json:"description,omitempty"`
		Conditions  []ConditionDTO `json:"conditions"`
		IsDynamic   bool           `json:"is_dynamic"`
	}
}

type UpdateSegmentRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	SegmentID string `path:"segment_id" format:"uuid"`
	Body      struct {
		Name        string         `json:"name,omitempty"`
		Description string         `json:"description,omitempty"`
		Conditions  []ConditionDTO `json:"conditions,omitempty"`
	}
}

type SegmentDTO struct {
	ID            uuid.UUID      `json:"id"`
	TenantID      uuid.UUID      `json:"tenant_id"`
	Name          string         `json:"name"`
	Description   string         `json:"description,omitempty"`
	Conditions    []ConditionDTO `json:"conditions"`
	IsDynamic     bool           `json:"is_dynamic"`
	IncludedUsers []string       `json:"included_users"`
	ExcludedUsers []string       `json:"excluded_users"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type SegmentResponse struct {
	Body SegmentDTO
}

type SegmentsListResponse struct {
	Body struct {
		Segments   []SegmentDTO       `json:"segments"`
		Pagination PaginationResponse `json:"pagination"`
	}
}

type SegmentUserRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	SegmentID string `path:"segment_id" format:"uuid"`
	Body      struct {
		UserID string `json:"user_id" minLength:"1"`
	}
}

// ===== Webhook DTOs =====
type CreateWebhookRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		Name           string            `json:"name" minLength:"1" maxLength:"255"`
		URL            string            `json:"url" format:"uri"`
		Secret         string            `json:"secret,omitempty"`
		Events         []string          `json:"events" minItems:"1"`
		Headers        map[string]string `json:"headers,omitempty"`
		RetryCount     int               `json:"retry_count,omitempty" minimum:"0" maximum:"10"`
		TimeoutSeconds int               `json:"timeout_seconds,omitempty" minimum:"1" maximum:"60"`
	}
}

type UpdateWebhookRequest struct {
	TenantID  string `path:"tenant_id" format:"uuid"`
	WebhookID string `path:"webhook_id" format:"uuid"`
	Body      struct {
		Name    string            `json:"name,omitempty"`
		URL     string            `json:"url,omitempty" format:"uri"`
		Secret  string            `json:"secret,omitempty"`
		Events  []string          `json:"events,omitempty"`
		Headers map[string]string `json:"headers,omitempty"`
		Enabled *bool             `json:"enabled,omitempty"`
	}
}

type WebhookDTO struct {
	ID             uuid.UUID         `json:"id"`
	TenantID       uuid.UUID         `json:"tenant_id"`
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	Events         []string          `json:"events"`
	Headers        map[string]string `json:"headers,omitempty"`
	Enabled        bool              `json:"enabled"`
	RetryCount     int               `json:"retry_count"`
	TimeoutSeconds int               `json:"timeout_seconds"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type WebhookResponse struct {
	Body WebhookDTO
}

type WebhooksListResponse struct {
	Body struct {
		Webhooks []WebhookDTO `json:"webhooks"`
	}
}

type WebhookDeliveryDTO struct {
	ID             uuid.UUID  `json:"id"`
	WebhookID      uuid.UUID  `json:"webhook_id"`
	EventType      string     `json:"event_type"`
	ResponseStatus *int       `json:"response_status,omitempty"`
	DurationMs     int        `json:"duration_ms"`
	Attempt        int        `json:"attempt"`
	Status         string     `json:"status"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type WebhookDeliveriesResponse struct {
	Body struct {
		Deliveries []WebhookDeliveryDTO `json:"deliveries"`
		Pagination PaginationResponse   `json:"pagination"`
	}
}

// ===== Emergency Control DTOs =====
type CreateEmergencyControlRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Body     struct {
		ControlType   string  `json:"control_type" enum:"kill_switch,read_only,maintenance"`
		EnvironmentID *string `json:"environment_id,omitempty" format:"uuid"`
		Reason        string  `json:"reason" minLength:"1"`
		ExpiresIn     *int    `json:"expires_in_minutes,omitempty" minimum:"1"`
	}
}

type EmergencyControlDTO struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	EnvironmentID *uuid.UUID `json:"environment_id,omitempty"`
	ControlType   string     `json:"control_type"`
	Enabled       bool       `json:"enabled"`
	Reason        string     `json:"reason"`
	EnabledBy     *uuid.UUID `json:"enabled_by,omitempty"`
	EnabledAt     *time.Time `json:"enabled_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type EmergencyControlResponse struct {
	Body EmergencyControlDTO
}

type EmergencyControlsListResponse struct {
	Body struct {
		Controls []EmergencyControlDTO `json:"controls"`
	}
}

// ===== Notification DTOs =====
type NotificationDTO struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Message   string     `json:"message,omitempty"`
	Link      string     `json:"link,omitempty"`
	Read      bool       `json:"read"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type NotificationResponse struct {
	Body NotificationDTO
}

type NotificationsListResponse struct {
	Body struct {
		Notifications []NotificationDTO  `json:"notifications"`
		Pagination    PaginationResponse `json:"pagination"`
	}
}

type UnreadCountResponse struct {
	Body struct {
		Count int `json:"count"`
	}
}

// ===== Flag History DTOs =====
type FlagHistoryDTO struct {
	ID            uuid.UUID `json:"id"`
	FeatureFlagID uuid.UUID `json:"feature_flag_id"`
	Version       int       `json:"version"`
	ChangeType    string    `json:"change_type"`
	ChangedBy     *string   `json:"changed_by,omitempty"`
	ChangedByName string    `json:"changed_by_name,omitempty"`
	Comment       string    `json:"comment,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type FlagHistoryResponse struct {
	Body FlagHistoryDTO
}

type FlagHistoryListResponse struct {
	Body struct {
		History    []FlagHistoryDTO   `json:"history"`
		Pagination PaginationResponse `json:"pagination"`
	}
}

// ===== Rollout Plan DTOs =====
type CreateRolloutRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	AppID    string `path:"app_id" format:"uuid"`
	EnvID    string `path:"env_id" format:"uuid"`
	FlagID   string `path:"flag_id" format:"uuid"`
	Body     struct {
		Name                       string   `json:"name" minLength:"1" maxLength:"255"`
		TargetPercentage           int      `json:"target_percentage" minimum:"1" maximum:"100"`
		IncrementPercentage        int      `json:"increment_percentage" minimum:"1" maximum:"100"`
		IncrementIntervalMinutes   int      `json:"increment_interval_minutes" minimum:"1"`
		AutoRollback               bool     `json:"auto_rollback"`
		RollbackThresholdErrorRate *float64 `json:"rollback_threshold_error_rate,omitempty"`
		RollbackThresholdLatencyMs *int     `json:"rollback_threshold_latency_ms,omitempty"`
	}
}

type RolloutPlanDTO struct {
	ID                         uuid.UUID  `json:"id"`
	FeatureFlagID              uuid.UUID  `json:"feature_flag_id"`
	Name                       string     `json:"name"`
	Status                     string     `json:"status"`
	CurrentPercentage          int        `json:"current_percentage"`
	TargetPercentage           int        `json:"target_percentage"`
	IncrementPercentage        int        `json:"increment_percentage"`
	IncrementIntervalMinutes   int        `json:"increment_interval_minutes"`
	AutoRollback               bool       `json:"auto_rollback"`
	RollbackThresholdErrorRate *float64   `json:"rollback_threshold_error_rate,omitempty"`
	RollbackThresholdLatencyMs *int       `json:"rollback_threshold_latency_ms,omitempty"`
	LastIncrementAt            *time.Time `json:"last_increment_at,omitempty"`
	NextIncrementAt            *time.Time `json:"next_increment_at,omitempty"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
}

type RolloutPlanResponse struct {
	Body RolloutPlanDTO
}

type RolloutPlansListResponse struct {
	Body struct {
		Plans []RolloutPlanDTO `json:"plans"`
	}
}

type RolloutHistoryDTO struct {
	ID             uuid.UUID `json:"id"`
	RolloutPlanID  uuid.UUID `json:"rollout_plan_id"`
	Action         string    `json:"action"`
	FromPercentage int       `json:"from_percentage"`
	ToPercentage   int       `json:"to_percentage"`
	Reason         string    `json:"reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type RolloutHistoryListResponse struct {
	Body struct {
		History    []RolloutHistoryDTO `json:"history"`
		Pagination PaginationResponse  `json:"pagination"`
	}
}
