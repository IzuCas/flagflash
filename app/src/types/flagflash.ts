// FlagFlash Types

// ==== Core Entities ====

export interface Tenant {
  id: string;
  name: string;
  slug: string;
  plan?: string;
  active?: boolean;
  created_at: string;
  updated_at: string;
}

export interface Application {
  id: string;
  tenant_id: string;
  name: string;
  slug: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface Environment {
  id: string;
  application_id: string;
  name: string;
  slug: string;
  description: string;
  color: string;
  created_at: string;
  updated_at: string;
}

export type FlagType = 'boolean' | 'string' | 'number' | 'json';

export interface FeatureFlag {
  id: string;
  environment_id: string;
  key: string;
  name: string;
  description: string;
  type: FlagType;
  default_value: unknown;
  enabled: boolean;
  version: number;
  tags: string[];
  created_at: string;
  updated_at: string;
}

export type Operator = 
  | 'eq' | 'neq' 
  | 'gt' | 'gte' | 'lt' | 'lte' 
  | 'contains' | 'not_contains' 
  | 'starts_with' | 'ends_with' 
  | 'in' | 'not_in' 
  | 'matches' | 'exists';

export interface Condition {
  attribute: string;
  operator: Operator;
  value: unknown;
}

export interface TargetingRule {
  id: string;
  flag_id: string;
  name: string;
  description: string;
  priority: number;
  conditions: Condition[];
  value: unknown;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface APIKey {
  id: string;
  tenant_id: string;
  environment_id: string;
  name: string;
  key_prefix: string;
  permissions: string[];
  active: boolean;
  last_used_at?: string;
  expires_at?: string;
  created_at: string;
}

export type UserRole = 'owner' | 'admin' | 'member' | 'viewer';

export interface User {
  id: string;
  tenant_id?: string;
  email: string;
  name: string;
  role: UserRole;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserWithMembership {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface TenantWithRole {
  tenant: Tenant;
  role: UserRole;
  active: boolean;
}

// ==== API Responses ====

export interface PaginationResponse {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface TenantsListResponse {
  tenants: Tenant[];
  pagination: PaginationResponse;
}

export interface ApplicationsListResponse {
  applications: Application[];
  pagination: PaginationResponse;
}

export interface EnvironmentsListResponse {
  environments: Environment[];
  pagination: PaginationResponse;
}

export interface FeatureFlagsListResponse {
  flags: FeatureFlag[];
  pagination: PaginationResponse;
}

export interface TargetingRulesListResponse {
  rules: TargetingRule[];
}

export interface APIKeysListResponse {
  keys: APIKey[];
  pagination: PaginationResponse;
}

export interface APIKeyCreatedResponse extends APIKey {
  key: string; // Raw key, only returned on creation
}

export interface UsersListResponse {
  users: UserWithMembership[];
  pagination: PaginationResponse;
}

// ==== Request DTOs ====

export interface CreateTenantRequest {
  name: string;
  slug: string;
}

export interface UpdateTenantRequest {
  name: string;
}

export interface CreateApplicationRequest {
  name: string;
  slug: string;
  description?: string;
}

export interface CreateEnvironmentRequest {
  name: string;
  slug: string;
  description?: string;
  color?: string;
}

export interface CreateFeatureFlagRequest {
  key: string;
  name: string;
  description?: string;
  type: FlagType;
  default_value: unknown;
  enabled: boolean;
  tags?: string[];
}

export interface UpdateFeatureFlagRequest {
  name?: string;
  description?: string;
  default_value?: unknown;
  enabled?: boolean;
  tags?: string[];
}

export interface CreateTargetingRuleRequest {
  name: string;
  description?: string;
  priority: number;
  conditions: Condition[];
  value: unknown;
  enabled: boolean;
}

export interface CreateAPIKeyRequest {
  name: string;
  environment_id: string;
  permissions: string[];
  expires_at?: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  role?: UserRole;
}

export interface UpdateUserRequest {
  name?: string;
  role?: UserRole;
  active?: boolean;
}

export interface InviteUserRequest {
  email: string;
  role?: UserRole;
}

export interface InviteResponse {
  invite_id: string;
  email: string;
  role: string;
  expires_at: string;
  email_sent: boolean;
}

export interface InviteDetails {
  email: string;
  tenant_name: string;
  role: string;
  expires_at: string;
  user_exists: boolean;
}

export interface AcceptInviteRequest {
  token: string;
  name?: string;
  password?: string;
}

export interface AcceptInviteResponse {
  message: string;
  email: string;
}

// ==== Evaluation ====

export interface EvaluationContext {
  [key: string]: unknown;
}

export interface EvaluateFlagResponse {
  flag_key: string;
  value: unknown;
  enabled: boolean;
  version: number;
  rule_id?: string;
  rule_name?: string;
}

export interface EvaluateAllFlagsResponse {
  flags: Record<string, {
    value: unknown;
    enabled: boolean;
    version: number;
    rule_id?: string;
    rule_name?: string;
  }>;
}

// ==== Auth ====

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginTenant {
  id: string;
  name: string;
  slug: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface LoginResponse {
  token: string;
  refresh_token: string;
  expires_at: string;
  user: User;
  tenants: LoginTenant[];
}

// ==== WebSocket Messages ====

export type WSMessageType = 
  | 'flag_update' 
  | 'flag_delete' 
  | 'flags_sync' 
  | 'subscribe' 
  | 'unsubscribe' 
  | 'ping' 
  | 'pong' 
  | 'error';

export interface WSMessage {
  type: WSMessageType;
  environment_id?: string;
  flag?: FeatureFlag;
  flags?: FeatureFlag[];
  flag_key?: string;
  error?: string;
  timestamp: string;
}

// ==== Audit Log Types ====

export type EntityType = 
  | 'tenant' 
  | 'application' 
  | 'environment' 
  | 'feature_flag' 
  | 'targeting_rule' 
  | 'api_key' 
  | 'user';

export type AuditAction = 
  | 'create' 
  | 'update' 
  | 'delete' 
  | 'enable' 
  | 'disable' 
  | 'toggle' 
  | 'revoke' 
  | 'rotate';

export type ActorType = 'user' | 'api_key' | 'system';

export interface AuditLog {
  id: string;
  tenant_id: string;
  entity_type: EntityType;
  entity_id: string;
  action: AuditAction;
  actor_id: string;
  actor_name?: string;
  actor_type: ActorType;
  old_value?: unknown;
  new_value?: unknown;
  metadata?: Record<string, unknown>;
  created_at: string;
}

export interface AuditLogsListResponse {
  logs: AuditLog[];
  pagination: PaginationResponse;
}

export interface AuditLogFilters {
  entity_type?: EntityType;
  entity_id?: string;
  action?: AuditAction;
  actor_id?: string;
  start_date?: string;
  end_date?: string;
}

// ==== Usage Metrics ====

export interface TimelinePoint {
  timestamp: string;
  evaluations: number;
  true_count: number;
  false_count: number;
}

export interface EnvironmentMetric {
  environment_id: string;
  environment_name: string;
  evaluations: number;
  unique_flags: number;
  unique_users: number;
}

export interface FlagMetric {
  flag_id: string;
  flag_key: string;
  flag_name: string;
  environment_id: string;
  environment_name: string;
  evaluations: number;
  true_count: number;
  false_count: number;
  unique_users: number;
}

export interface UsageMetrics {
  tenant_id: string;
  period: string;
  start_date: string;
  end_date: string;
  total_evaluations: number;
  unique_flags: number;
  unique_users: number;
  by_environment?: EnvironmentMetric[];
  by_flag?: FlagMetric[];
  timeline?: TimelinePoint[];
}

export interface UsageMetricsFilters {
  environment_id?: string;
  flag_id?: string;
  start_date: string;
  end_date: string;
  granularity: 'hour' | 'day' | 'week' | 'month';
}
