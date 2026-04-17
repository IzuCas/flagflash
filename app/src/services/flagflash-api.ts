import type {
  Tenant,
  Application,
  Environment,
  FeatureFlag,
  TargetingRule,
  APIKey,
  APIKeyCreatedResponse,
  TenantsListResponse,
  ApplicationsListResponse,
  EnvironmentsListResponse,
  FeatureFlagsListResponse,
  TargetingRulesListResponse,
  APIKeysListResponse,
  CreateTenantRequest,
  CreateApplicationRequest,
  CreateEnvironmentRequest,
  CreateFeatureFlagRequest,
  UpdateFeatureFlagRequest,
  CreateTargetingRuleRequest,
  CreateAPIKeyRequest,
  LoginRequest,
  LoginResponse,
  EvaluateFlagResponse,
  EvaluateAllFlagsResponse,
  EvaluationContext,
  AuditLog,
  AuditLogsListResponse,
  AuditLogFilters,
  UsageMetrics,
  UsageMetricsFilters,
  FlagMetric,
  EnvironmentMetric,
  TimelinePoint,
  UserWithMembership,
  UsersListResponse,
  CreateUserRequest,
  UpdateUserRequest,
  InviteUserRequest,
  InviteResponse,
  InviteDetails,
  AcceptInviteRequest,
  AcceptInviteResponse,
  // Advanced features
  Segment,
  CreateSegmentRequest,
  UpdateSegmentRequest,
  Webhook,
  CreateWebhookRequest,
  UpdateWebhookRequest,
  EmergencyControl,
  ActivateEmergencyControlRequest,
  NotificationsListResponse,
  FlagHistoryListResponse,
  FlagHistory,
  FlagComparisonResponse,
  RolloutPlan,
  CreateRolloutRequest,
  RolloutHistoryListResponse,
} from '../types/flagflash';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';
const BASE_URL = `${API_BASE_URL}/api/v1/flagflash`;

// Helper function for fetch requests
async function request<T>(
  endpoint: string,
  options: RequestInit = {},
  customHeaders?: Record<string, string>,
  skipAuthRedirect = false
): Promise<T> {
  const token = localStorage.getItem('flagflash_token');
  
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...customHeaders,
  };
  
  if (token && !customHeaders?.['X-API-Key']) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    if (skipAuthRedirect) {
      throw new Error('Invalid email or password');
    }
    localStorage.removeItem('flagflash_token');
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json();
}

function get<T>(endpoint: string, params?: Record<string, string | number>): Promise<T> {
  const url = params 
    ? `${endpoint}?${new URLSearchParams(Object.entries(params).map(([k, v]) => [k, String(v)]))}`
    : endpoint;
  return request<T>(url);
}

function post<T>(endpoint: string, body?: unknown, headers?: Record<string, string>, skipAuthRedirect = false): Promise<T> {
  return request<T>(endpoint, { method: 'POST', body: body ? JSON.stringify(body) : undefined }, headers, skipAuthRedirect);
}

function put<T>(endpoint: string, body?: unknown): Promise<T> {
  return request<T>(endpoint, { method: 'PUT', body: body ? JSON.stringify(body) : undefined });
}

function patch<T>(endpoint: string, body?: unknown): Promise<T> {
  return request<T>(endpoint, { method: 'PATCH', body: body ? JSON.stringify(body) : undefined });
}

function del<T>(endpoint: string): Promise<T> {
  return request<T>(endpoint, { method: 'DELETE' });
}

// ==== Auth ====
export interface UserProfile {
  id: string;
  tenant_id: string;
  email: string;
  name: string;
  role: string;
}

export const flagflashAuthApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    return post<LoginResponse>('/auth/login', data, undefined, true);
  },

  refreshToken: async (refreshToken: string): Promise<LoginResponse> => {
    return post<LoginResponse>('/auth/refresh', { refresh_token: refreshToken });
  },

  switchTenant: async (tenantId: string): Promise<LoginResponse> => {
    return post<LoginResponse>('/auth/switch-tenant', { tenant_id: tenantId });
  },

  me: async () => {
    return get('/auth/me');
  },

  getProfile: async (): Promise<UserProfile> => {
    return get<UserProfile>('/auth/profile');
  },

  updateProfile: async (data: { name: string }): Promise<UserProfile> => {
    return put<UserProfile>('/auth/profile', data);
  },
};

// ==== Tenants ====
export interface TenantWithRole {
  id: string;
  name: string;
  slug: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface MyTenantsListResponse {
  tenants: TenantWithRole[];
}

export const tenantsApi = {
  list: async (page = 1, limit = 20): Promise<TenantsListResponse> => {
    return get<TenantsListResponse>('/manage/tenants', { page, limit });
  },

  listMyTenants: async (): Promise<MyTenantsListResponse> => {
    return get<MyTenantsListResponse>('/manage/tenants/me');
  },

  get: async (tenantId: string): Promise<Tenant> => {
    return get<Tenant>(`/manage/tenants/${tenantId}`);
  },

  create: async (data: CreateTenantRequest): Promise<Tenant> => {
    return post<Tenant>('/manage/tenants', data);
  },

  update: async (tenantId: string, data: { name: string }): Promise<Tenant> => {
    return put<Tenant>(`/manage/tenants/${tenantId}`, data);
  },

  delete: async (tenantId: string): Promise<void> => {
    return del(`/manage/tenants/${tenantId}`);
  },
};

// ==== Applications ====
export const applicationsApi = {
  list: async (tenantId: string, page = 1, limit = 20): Promise<ApplicationsListResponse> => {
    return get<ApplicationsListResponse>(`/manage/tenants/${tenantId}/applications`, { page, limit });
  },

  get: async (tenantId: string, appId: string): Promise<Application> => {
    return get<Application>(`/manage/tenants/${tenantId}/applications/${appId}`);
  },

  create: async (tenantId: string, data: CreateApplicationRequest): Promise<Application> => {
    return post<Application>(`/manage/tenants/${tenantId}/applications`, data);
  },

  update: async (tenantId: string, appId: string, data: CreateApplicationRequest): Promise<Application> => {
    return put<Application>(`/manage/tenants/${tenantId}/applications/${appId}`, data);
  },

  delete: async (tenantId: string, appId: string): Promise<void> => {
    return del(`/manage/tenants/${tenantId}/applications/${appId}`);
  },
};

// ==== Environments ====
export const environmentsApi = {
  list: async (tenantId: string, appId: string, page = 1, limit = 20): Promise<EnvironmentsListResponse> => {
    return get<EnvironmentsListResponse>(`/manage/tenants/${tenantId}/applications/${appId}/environments`, { page, limit });
  },

  get: async (tenantId: string, appId: string, envId: string): Promise<Environment> => {
    return get<Environment>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}`);
  },

  create: async (tenantId: string, appId: string, data: CreateEnvironmentRequest): Promise<Environment> => {
    return post<Environment>(`/manage/tenants/${tenantId}/applications/${appId}/environments`, data);
  },

  update: async (tenantId: string, appId: string, envId: string, data: CreateEnvironmentRequest): Promise<Environment> => {
    return put<Environment>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}`, data);
  },

  delete: async (tenantId: string, appId: string, envId: string): Promise<void> => {
    return del(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}`);
  },
};

// ==== Feature Flags ====
export const featureFlagsApi = {
  list: async (tenantId: string, appId: string, envId: string, page = 1, limit = 20): Promise<FeatureFlagsListResponse> => {
    return get<FeatureFlagsListResponse>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags`, { page, limit });
  },

  get: async (tenantId: string, appId: string, envId: string, flagId: string): Promise<FeatureFlag> => {
    return get<FeatureFlag>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}`);
  },

  getByKey: async (tenantId: string, appId: string, envId: string, key: string): Promise<FeatureFlag> => {
    return get<FeatureFlag>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/key/${key}`);
  },

  create: async (tenantId: string, appId: string, envId: string, data: CreateFeatureFlagRequest): Promise<FeatureFlag> => {
    return post<FeatureFlag>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags`, data);
  },

  update: async (tenantId: string, appId: string, envId: string, flagId: string, data: UpdateFeatureFlagRequest): Promise<FeatureFlag> => {
    return put<FeatureFlag>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}`, data);
  },

  toggle: async (tenantId: string, appId: string, envId: string, flagId: string, enabled: boolean): Promise<FeatureFlag> => {
    return patch<FeatureFlag>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}/toggle`, { enabled });
  },

  delete: async (tenantId: string, appId: string, envId: string, flagId: string): Promise<void> => {
    return del(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}`);
  },

  copy: async (tenantId: string, appId: string, sourceEnvId: string, targetEnvId: string, flagKeys?: string[]): Promise<FeatureFlagsListResponse> => {
    return post<FeatureFlagsListResponse>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${sourceEnvId}/flags/copy`, {
      target_environment_id: targetEnvId,
      flag_keys: flagKeys,
      overwrite: true,
    });
  },
};

// ==== Targeting Rules ====
export const targetingRulesApi = {
  list: async (tenantId: string, appId: string, envId: string, flagId: string): Promise<TargetingRulesListResponse> => {
    return get<TargetingRulesListResponse>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}/rules`);
  },

  get: async (tenantId: string, appId: string, envId: string, flagId: string, ruleId: string): Promise<TargetingRule> => {
    return get<TargetingRule>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}/rules/${ruleId}`);
  },

  create: async (tenantId: string, appId: string, envId: string, flagId: string, data: CreateTargetingRuleRequest): Promise<TargetingRule> => {
    return post<TargetingRule>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}/rules`, data);
  },

  update: async (tenantId: string, appId: string, envId: string, flagId: string, ruleId: string, data: Partial<CreateTargetingRuleRequest>): Promise<TargetingRule> => {
    return put<TargetingRule>(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}/rules/${ruleId}`, data);
  },

  delete: async (tenantId: string, appId: string, envId: string, flagId: string, ruleId: string): Promise<void> => {
    return del(`/manage/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags/${flagId}/rules/${ruleId}`);
  },
};

// ==== API Keys ====
export const apiKeysApi = {
  list: async (tenantId: string, page = 1, limit = 20): Promise<APIKeysListResponse> => {
    return get<APIKeysListResponse>(`/manage/tenants/${tenantId}/api-keys`, { page, limit });
  },

  get: async (tenantId: string, keyId: string): Promise<APIKey> => {
    return get<APIKey>(`/manage/tenants/${tenantId}/api-keys/${keyId}`);
  },

  create: async (tenantId: string, data: CreateAPIKeyRequest): Promise<APIKeyCreatedResponse> => {
    return post<APIKeyCreatedResponse>(`/manage/tenants/${tenantId}/api-keys`, data);
  },

  update: async (tenantId: string, keyId: string, data: { name?: string; active?: boolean }): Promise<APIKey> => {
    return put<APIKey>(`/manage/tenants/${tenantId}/api-keys/${keyId}`, data);
  },

  revoke: async (tenantId: string, keyId: string): Promise<APIKey> => {
    return post<APIKey>(`/manage/tenants/${tenantId}/api-keys/${keyId}/revoke`);
  },

  delete: async (tenantId: string, keyId: string): Promise<void> => {
    return del(`/manage/tenants/${tenantId}/api-keys/${keyId}`);
  },
};

// ==== Users ====
export const usersApi = {
  list: async (tenantId: string, page = 1, limit = 20): Promise<UsersListResponse> => {
    return get<UsersListResponse>(`/manage/tenants/${tenantId}/users`, { page, limit });
  },

  get: async (tenantId: string, userId: string): Promise<UserWithMembership> => {
    return get<UserWithMembership>(`/manage/tenants/${tenantId}/users/${userId}`);
  },

  create: async (tenantId: string, data: CreateUserRequest): Promise<UserWithMembership> => {
    return post<UserWithMembership>(`/manage/tenants/${tenantId}/users`, data);
  },

  update: async (tenantId: string, userId: string, data: UpdateUserRequest): Promise<UserWithMembership> => {
    return put<UserWithMembership>(`/manage/tenants/${tenantId}/users/${userId}`, data);
  },

  delete: async (tenantId: string, userId: string): Promise<void> => {
    return del(`/manage/tenants/${tenantId}/users/${userId}`);
  },

  invite: async (tenantId: string, data: InviteUserRequest): Promise<InviteResponse> => {
    return post<InviteResponse>(`/manage/tenants/${tenantId}/users/invite`, data);
  },

  updateRole: async (tenantId: string, userId: string, role: string): Promise<UserWithMembership> => {
    return patch<UserWithMembership>(`/manage/tenants/${tenantId}/users/${userId}/role`, { role });
  },
};

// ==== Invites (Public) ====
export const inviteApi = {
  validate: async (token: string): Promise<InviteDetails> => {
    return get<InviteDetails>(`/auth/invite/${encodeURIComponent(token)}`);
  },

  accept: async (data: AcceptInviteRequest): Promise<AcceptInviteResponse> => {
    return post<AcceptInviteResponse>('/auth/invite/accept', data);
  },
};

// ==== SDK (for testing) ====
export const sdkApi = {
  evaluateFlag: async (apiKey: string, flagKey: string, context?: EvaluationContext): Promise<EvaluateFlagResponse> => {
    return post<EvaluateFlagResponse>('/sdk/evaluate', { flag_key: flagKey, context }, { 'X-API-Key': apiKey });
  },

  evaluateAllFlags: async (apiKey: string, context?: EvaluationContext): Promise<EvaluateAllFlagsResponse> => {
    return post<EvaluateAllFlagsResponse>('/sdk/evaluate-all', { context }, { 'X-API-Key': apiKey });
  },

  getFlags: async (apiKey: string): Promise<FeatureFlag[]> => {
    const response = await fetch(`${BASE_URL}/sdk/flags`, {
      headers: { 'X-API-Key': apiKey },
    });
    const data = await response.json();
    return data.flags;
  },
};

// ==== Audit Logs ====
export const auditLogsApi = {
  list: async (tenantId: string, filters?: AuditLogFilters, page = 1, limit = 50): Promise<AuditLogsListResponse> => {
    const params: Record<string, string | number> = { page, limit };
    if (filters?.entity_type) params.entity_type = filters.entity_type;
    if (filters?.entity_id) params.entity_id = filters.entity_id;
    if (filters?.action) params.action = filters.action;
    if (filters?.actor_id) params.actor_id = filters.actor_id;
    if (filters?.start_date) params.start_date = filters.start_date;
    if (filters?.end_date) params.end_date = filters.end_date;
    return get<AuditLogsListResponse>(`/manage/tenants/${tenantId}/audit-logs`, params);
  },

  get: async (tenantId: string, logId: string): Promise<AuditLog> => {
    return get<AuditLog>(`/manage/tenants/${tenantId}/audit-logs/${logId}`);
  },
};

// ==== Usage Metrics ====
export const usageMetricsApi = {
  getSummary: async (tenantId: string, filters: UsageMetricsFilters): Promise<UsageMetrics> => {
    const params: Record<string, string> = {
      start_date: filters.start_date,
      end_date: filters.end_date,
      granularity: filters.granularity,
    };
    if (filters.environment_id) params.environment_id = filters.environment_id;
    if (filters.flag_id) params.flag_id = filters.flag_id;
    return get<UsageMetrics>(`/manage/tenants/${tenantId}/usage-metrics`, params);
  },

  getTimeline: async (tenantId: string, filters: UsageMetricsFilters): Promise<{ timeline: TimelinePoint[] }> => {
    const params: Record<string, string> = {
      start_date: filters.start_date,
      end_date: filters.end_date,
      granularity: filters.granularity,
    };
    if (filters.environment_id) params.environment_id = filters.environment_id;
    if (filters.flag_id) params.flag_id = filters.flag_id;
    return get<{ timeline: TimelinePoint[] }>(`/manage/tenants/${tenantId}/usage-metrics/timeline`, params);
  },

  getFlagMetrics: async (tenantId: string, filters: UsageMetricsFilters): Promise<{ flags: FlagMetric[] }> => {
    const params: Record<string, string> = {
      start_date: filters.start_date,
      end_date: filters.end_date,
      granularity: filters.granularity,
    };
    if (filters.environment_id) params.environment_id = filters.environment_id;
    return get<{ flags: FlagMetric[] }>(`/manage/tenants/${tenantId}/usage-metrics/flags`, params);
  },

  getEnvironmentMetrics: async (tenantId: string, startDate: string, endDate: string): Promise<{ environments: EnvironmentMetric[] }> => {
    return get<{ environments: EnvironmentMetric[] }>(`/manage/tenants/${tenantId}/usage-metrics/environments`, {
      start_date: startDate,
      end_date: endDate,
    });
  },
};

// ==== Segments ====
export const segmentsApi = {
  list: async (tenantId: string): Promise<{ segments: Segment[] }> => {
    return get<{ segments: Segment[] }>(`/manage/tenants/${tenantId}/segments`);
  },

  get: async (tenantId: string, segmentId: string): Promise<Segment> => {
    return get<Segment>(`/manage/tenants/${tenantId}/segments/${segmentId}`);
  },

  create: async (tenantId: string, data: CreateSegmentRequest): Promise<Segment> => {
    return post<Segment>(`/manage/tenants/${tenantId}/segments`, data);
  },

  update: async (tenantId: string, segmentId: string, data: UpdateSegmentRequest): Promise<Segment> => {
    return put<Segment>(`/manage/tenants/${tenantId}/segments/${segmentId}`, data);
  },

  delete: async (tenantId: string, segmentId: string): Promise<void> => {
    return del<void>(`/manage/tenants/${tenantId}/segments/${segmentId}`);
  },
};

// ==== Webhooks ====
export const webhooksApi = {
  list: async (tenantId: string): Promise<{ webhooks: Webhook[] }> => {
    return get<{ webhooks: Webhook[] }>(`/manage/tenants/${tenantId}/webhooks`);
  },

  get: async (tenantId: string, webhookId: string): Promise<Webhook> => {
    return get<Webhook>(`/manage/tenants/${tenantId}/webhooks/${webhookId}`);
  },

  create: async (tenantId: string, data: CreateWebhookRequest): Promise<Webhook> => {
    return post<Webhook>(`/manage/tenants/${tenantId}/webhooks`, data);
  },

  update: async (tenantId: string, webhookId: string, data: UpdateWebhookRequest): Promise<Webhook> => {
    return put<Webhook>(`/manage/tenants/${tenantId}/webhooks/${webhookId}`, data);
  },

  delete: async (tenantId: string, webhookId: string): Promise<void> => {
    return del<void>(`/manage/tenants/${tenantId}/webhooks/${webhookId}`);
  },
};

// ==== Emergency Controls ====
export const emergencyControlsApi = {
  list: async (tenantId: string): Promise<{ controls: EmergencyControl[] }> => {
    return get<{ controls: EmergencyControl[] }>(`/manage/tenants/${tenantId}/emergency-controls`);
  },

  listActive: async (tenantId: string): Promise<{ controls: EmergencyControl[] }> => {
    return get<{ controls: EmergencyControl[] }>(`/manage/tenants/${tenantId}/emergency-controls/active`);
  },

  activateKillSwitch: async (tenantId: string, data: ActivateEmergencyControlRequest): Promise<EmergencyControl> => {
    return post<EmergencyControl>(`/manage/tenants/${tenantId}/emergency-controls/kill-switch`, data);
  },

  activateMaintenance: async (tenantId: string, data: ActivateEmergencyControlRequest): Promise<EmergencyControl> => {
    return post<EmergencyControl>(`/manage/tenants/${tenantId}/emergency-controls/maintenance`, data);
  },

  deactivate: async (tenantId: string, controlId: string): Promise<void> => {
    return post<void>(`/manage/tenants/${tenantId}/emergency-controls/${controlId}/deactivate`);
  },

  checkKillSwitch: async (tenantId: string, envId?: string): Promise<{ active: boolean; control?: EmergencyControl }> => {
    const params = envId ? { environment_id: envId } : undefined;
    return get<{ active: boolean; control?: EmergencyControl }>(`/manage/tenants/${tenantId}/emergency-controls/kill-switch/check`, params);
  },
};

// ==== Notifications ====
export const notificationsApi = {
  list: async (tenantId: string, page = 1, limit = 20): Promise<NotificationsListResponse> => {
    return get<NotificationsListResponse>(`/manage/tenants/${tenantId}/notifications`, { page, limit });
  },

  getUnreadCount: async (tenantId: string): Promise<{ count: number }> => {
    return get<{ count: number }>(`/manage/tenants/${tenantId}/notifications/unread-count`);
  },

  markAsRead: async (tenantId: string, notificationId: string): Promise<void> => {
    return post<void>(`/manage/tenants/${tenantId}/notifications/${notificationId}/read`);
  },

  markAllAsRead: async (tenantId: string): Promise<void> => {
    return post<void>(`/manage/tenants/${tenantId}/notifications/mark-all-read`);
  },
};

// ==== Flag History ====
export const flagHistoryApi = {
  list: async (tenantId: string, appId: string, envId: string, flagId: string, page = 1, limit = 20): Promise<FlagHistoryListResponse> => {
    return get<FlagHistoryListResponse>(`/manage/tenants/${tenantId}/apps/${appId}/envs/${envId}/flags/${flagId}/history`, { page, limit });
  },

  getVersion: async (tenantId: string, appId: string, envId: string, flagId: string, version: number): Promise<FlagHistory> => {
    return get<FlagHistory>(`/manage/tenants/${tenantId}/apps/${appId}/envs/${envId}/flags/${flagId}/history/version/${version}`);
  },

  compare: async (tenantId: string, appId: string, envId: string, flagId: string, v1: number, v2: number): Promise<FlagComparisonResponse> => {
    return get<FlagComparisonResponse>(`/manage/tenants/${tenantId}/apps/${appId}/envs/${envId}/flags/${flagId}/history/compare`, { version1: v1, version2: v2 });
  },
};

// ==== Rollout Plans ====
export const rolloutsApi = {
  list: async (tenantId: string, appId: string, envId: string, flagId: string): Promise<{ plans: RolloutPlan[] }> => {
    return get<{ plans: RolloutPlan[] }>(`/manage/tenants/${tenantId}/apps/${appId}/envs/${envId}/flags/${flagId}/rollouts`);
  },

  get: async (tenantId: string, rolloutId: string): Promise<RolloutPlan> => {
    return get<RolloutPlan>(`/manage/tenants/${tenantId}/rollouts/${rolloutId}`);
  },

  create: async (tenantId: string, appId: string, envId: string, flagId: string, data: CreateRolloutRequest): Promise<RolloutPlan> => {
    return post<RolloutPlan>(`/manage/tenants/${tenantId}/apps/${appId}/envs/${envId}/flags/${flagId}/rollouts`, data);
  },

  start: async (tenantId: string, rolloutId: string): Promise<RolloutPlan> => {
    return post<RolloutPlan>(`/manage/tenants/${tenantId}/rollouts/${rolloutId}/start`);
  },

  pause: async (tenantId: string, rolloutId: string): Promise<RolloutPlan> => {
    return post<RolloutPlan>(`/manage/tenants/${tenantId}/rollouts/${rolloutId}/pause`);
  },

  resume: async (tenantId: string, rolloutId: string): Promise<RolloutPlan> => {
    return post<RolloutPlan>(`/manage/tenants/${tenantId}/rollouts/${rolloutId}/resume`);
  },

  rollback: async (tenantId: string, rolloutId: string, reason?: string): Promise<RolloutPlan> => {
    return post<RolloutPlan>(`/manage/tenants/${tenantId}/rollouts/${rolloutId}/rollback`, { reason });
  },

  delete: async (tenantId: string, rolloutId: string): Promise<void> => {
    return del<void>(`/manage/tenants/${tenantId}/rollouts/${rolloutId}`);
  },

  getHistory: async (tenantId: string, rolloutId: string): Promise<RolloutHistoryListResponse> => {
    return get<RolloutHistoryListResponse>(`/manage/tenants/${tenantId}/rollouts/${rolloutId}/history`);
  },
};
