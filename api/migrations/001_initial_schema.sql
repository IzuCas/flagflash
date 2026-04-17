-- FlagFlash Database Schema
-- Complete schema with all features
-- Migration: 001_initial_schema.sql

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- CORE TABLES
-- ============================================

-- Tenants (Organizations)
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) DEFAULT 'free',
    active BOOLEAN DEFAULT TRUE,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Users (Dashboard Users)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'member',
    active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- User-Tenant Memberships (Many-to-Many relationship)
CREATE TABLE IF NOT EXISTS user_tenant_memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, tenant_id)
);

-- Invite Tokens
CREATE TABLE IF NOT EXISTS invite_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    token VARCHAR(128) UNIQUE NOT NULL,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Applications (Per Tenant)
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    UNIQUE(tenant_id, slug)
);

-- Environments (Per Application)
CREATE TABLE IF NOT EXISTS environments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#6366f1',
    is_production BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(application_id, slug)
);

-- Feature Flags (with scheduling and lifecycle)
CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL CHECK (type IN ('boolean', 'json', 'string', 'number')),
    enabled BOOLEAN DEFAULT FALSE,
    value JSONB NOT NULL DEFAULT 'false',
    default_value JSONB NOT NULL,
    version INTEGER DEFAULT 1,
    tags TEXT[] DEFAULT '{}',
    -- Scheduling fields
    scheduled_enable_at TIMESTAMP WITH TIME ZONE NULL,
    scheduled_disable_at TIMESTAMP WITH TIME ZONE NULL,
    schedule_timezone VARCHAR(50) DEFAULT 'UTC',
    schedule_recurrence JSONB NULL,
    -- Lifecycle state
    lifecycle_state VARCHAR(20) DEFAULT 'active' CHECK (lifecycle_state IN ('draft', 'active', 'deprecated', 'archived')),
    last_evaluated_at TIMESTAMP WITH TIME ZONE NULL,
    prerequisite_flag_id UUID NULL,
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    UNIQUE(environment_id, key)
);

-- Add self-reference for prerequisite after table exists
ALTER TABLE feature_flags 
ADD CONSTRAINT fk_feature_flags_prerequisite 
FOREIGN KEY (prerequisite_flag_id) REFERENCES feature_flags(id) ON DELETE SET NULL;

-- ============================================
-- SEGMENTS (Reusable User Cohorts)
-- ============================================

CREATE TABLE IF NOT EXISTS segments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    conditions JSONB NOT NULL DEFAULT '[]',
    is_dynamic BOOLEAN DEFAULT FALSE,
    included_users TEXT[] DEFAULT '{}',
    excluded_users TEXT[] DEFAULT '{}',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

-- Targeting Rules (with segment support)
CREATE TABLE IF NOT EXISTS targeting_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    segment_id UUID REFERENCES segments(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    priority INTEGER DEFAULT 0,
    conditions JSONB NOT NULL DEFAULT '[]',
    value JSONB NOT NULL,
    percentage INTEGER DEFAULT 100 CHECK (percentage >= 0 AND percentage <= 100),
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API Keys
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(64) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    permissions TEXT[] DEFAULT ARRAY['read'],
    active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE NULL,
    last_used_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE NULL
);

-- Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255),
    actor_type VARCHAR(50),
    old_value JSONB,
    new_value JSONB,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- EVALUATION & METRICS
-- ============================================

-- Evaluation Events (for usage metrics)
CREATE TABLE IF NOT EXISTS evaluation_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    flag_key VARCHAR(255) NOT NULL,
    value JSONB NOT NULL,
    user_id VARCHAR(255),
    context JSONB DEFAULT '{}',
    sdk_type VARCHAR(50),
    sdk_version VARCHAR(20),
    evaluated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Evaluation Summary (pre-aggregated hourly stats)
CREATE TABLE IF NOT EXISTS evaluation_summary (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    flag_key VARCHAR(255) NOT NULL,
    hour_bucket TIMESTAMP WITH TIME ZONE NOT NULL,
    total_evaluations BIGINT DEFAULT 0,
    true_count BIGINT DEFAULT 0,
    false_count BIGINT DEFAULT 0,
    unique_users INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(environment_id, feature_flag_id, hour_bucket)
);

-- ============================================
-- APPROVAL WORKFLOW
-- ============================================

CREATE TABLE IF NOT EXISTS approval_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    feature_flag_id UUID REFERENCES feature_flags(id) ON DELETE CASCADE,
    requires_approval BOOLEAN DEFAULT TRUE,
    min_approvers INTEGER DEFAULT 1,
    auto_reject_hours INTEGER DEFAULT 72,
    allowed_approver_roles TEXT[] DEFAULT ARRAY['owner', 'admin'],
    notify_on_request BOOLEAN DEFAULT TRUE,
    notify_on_decision BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pending_changes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    feature_flag_id UUID REFERENCES feature_flags(id) ON DELETE CASCADE,
    targeting_rule_id UUID REFERENCES targeting_rules(id) ON DELETE CASCADE,
    change_type VARCHAR(20) NOT NULL CHECK (change_type IN ('create', 'update', 'delete', 'enable', 'disable')),
    entity_type VARCHAR(20) NOT NULL CHECK (entity_type IN ('flag', 'rule', 'segment')),
    old_value JSONB,
    new_value JSONB,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'expired', 'cancelled')),
    requested_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_comment TEXT,
    decided_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pending_change_id UUID NOT NULL REFERENCES pending_changes(id) ON DELETE CASCADE,
    approver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    decision VARCHAR(20) NOT NULL CHECK (decision IN ('approved', 'rejected', 'needs_changes')),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(pending_change_id, approver_id)
);

-- ============================================
-- WEBHOOKS
-- ============================================

CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(2048) NOT NULL,
    secret VARCHAR(255),
    events TEXT[] NOT NULL DEFAULT ARRAY['flag.updated'],
    headers JSONB DEFAULT '{}',
    enabled BOOLEAN DEFAULT TRUE,
    retry_count INTEGER DEFAULT 3,
    timeout_seconds INTEGER DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    response_status INTEGER,
    response_body TEXT,
    response_headers JSONB,
    duration_ms INTEGER,
    attempt INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'success', 'failed', 'retrying')),
    error_message TEXT,
    next_retry_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- EXPERIMENTS / A/B TESTING
-- ============================================

CREATE TABLE IF NOT EXISTS experiments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    hypothesis TEXT,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'running', 'paused', 'completed', 'cancelled')),
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    target_sample_size INTEGER,
    current_sample_size INTEGER DEFAULT 0,
    winner_variant VARCHAR(100),
    statistical_significance DECIMAL(5,4),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS experiment_variants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    value JSONB NOT NULL,
    weight INTEGER DEFAULT 50,
    is_control BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(experiment_id, name)
);

CREATE TABLE IF NOT EXISTS experiment_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    metric_type VARCHAR(20) NOT NULL CHECK (metric_type IN ('conversion', 'count', 'sum', 'average')),
    is_primary BOOLEAN DEFAULT FALSE,
    goal_direction VARCHAR(10) CHECK (goal_direction IN ('increase', 'decrease')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(experiment_id, name)
);

CREATE TABLE IF NOT EXISTS experiment_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    variant_id UUID NOT NULL REFERENCES experiment_variants(id) ON DELETE CASCADE,
    metric_id UUID NOT NULL REFERENCES experiment_metrics(id) ON DELETE CASCADE,
    sample_count INTEGER DEFAULT 0,
    conversion_count INTEGER DEFAULT 0,
    sum_value DECIMAL(20,4) DEFAULT 0,
    mean_value DECIMAL(20,4),
    variance DECIMAL(20,8),
    confidence_interval_low DECIMAL(20,4),
    confidence_interval_high DECIMAL(20,4),
    p_value DECIMAL(10,8),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(experiment_id, variant_id, metric_id)
);

-- ============================================
-- PROGRESSIVE ROLLOUT
-- ============================================

CREATE TABLE IF NOT EXISTS rollout_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('draft', 'active', 'paused', 'completed', 'failed')),
    current_percentage INTEGER DEFAULT 0,
    target_percentage INTEGER DEFAULT 100,
    increment_percentage INTEGER DEFAULT 10,
    increment_interval_minutes INTEGER DEFAULT 60,
    auto_rollback BOOLEAN DEFAULT TRUE,
    rollback_threshold_error_rate DECIMAL(5,2),
    rollback_threshold_latency_ms INTEGER,
    last_increment_at TIMESTAMP WITH TIME ZONE,
    next_increment_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rollout_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rollout_plan_id UUID NOT NULL REFERENCES rollout_plans(id) ON DELETE CASCADE,
    from_percentage INTEGER NOT NULL,
    to_percentage INTEGER NOT NULL,
    action VARCHAR(20) NOT NULL CHECK (action IN ('increment', 'rollback', 'pause', 'resume', 'complete')),
    reason TEXT,
    metrics JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- EMERGENCY CONTROLS
-- ============================================

CREATE TABLE IF NOT EXISTS emergency_controls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    control_type VARCHAR(20) NOT NULL CHECK (control_type IN ('kill_switch', 'maintenance_mode')),
    is_active BOOLEAN DEFAULT FALSE,
    reason TEXT,
    activated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    activated_at TIMESTAMP WITH TIME ZONE,
    deactivated_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- NOTIFICATIONS
-- ============================================

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    data JSONB DEFAULT '{}',
    read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- FLAG HISTORY (for versioning)
-- ============================================

CREATE TABLE IF NOT EXISTS flag_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    enabled BOOLEAN NOT NULL,
    value JSONB NOT NULL,
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    change_type VARCHAR(20) NOT NULL,
    change_summary TEXT,
    snapshot JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- INDEXES - CORE
-- ============================================

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_tenant_memberships_user_id ON user_tenant_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tenant_memberships_tenant_id ON user_tenant_memberships(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_tenant_memberships_role ON user_tenant_memberships(tenant_id, role);
CREATE INDEX IF NOT EXISTS idx_invite_tokens_token ON invite_tokens(token);
CREATE INDEX IF NOT EXISTS idx_invite_tokens_email ON invite_tokens(email);
CREATE INDEX IF NOT EXISTS idx_invite_tokens_tenant ON invite_tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_applications_tenant_id ON applications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_applications_slug ON applications(tenant_id, slug);
CREATE INDEX IF NOT EXISTS idx_environments_application_id ON environments(application_id);
CREATE INDEX IF NOT EXISTS idx_environments_slug ON environments(application_id, slug);

-- Feature flags indexes
CREATE INDEX IF NOT EXISTS idx_feature_flags_environment_id ON feature_flags(environment_id);
CREATE INDEX IF NOT EXISTS idx_feature_flags_key ON feature_flags(environment_id, key);
CREATE INDEX IF NOT EXISTS idx_feature_flags_enabled ON feature_flags(environment_id, enabled) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_feature_flags_scheduled_enable ON feature_flags(scheduled_enable_at) WHERE scheduled_enable_at IS NOT NULL AND enabled = FALSE;
CREATE INDEX IF NOT EXISTS idx_feature_flags_scheduled_disable ON feature_flags(scheduled_disable_at) WHERE scheduled_disable_at IS NOT NULL AND enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_feature_flags_lifecycle ON feature_flags(lifecycle_state);
CREATE INDEX IF NOT EXISTS idx_feature_flags_last_evaluated ON feature_flags(last_evaluated_at);
CREATE INDEX IF NOT EXISTS idx_feature_flags_prerequisite ON feature_flags(prerequisite_flag_id);

-- Segments and targeting rules indexes
CREATE INDEX IF NOT EXISTS idx_segments_tenant_id ON segments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_flag_id ON targeting_rules(feature_flag_id);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_priority ON targeting_rules(feature_flag_id, priority);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_segment ON targeting_rules(segment_id);

-- API keys indexes
CREATE INDEX IF NOT EXISTS idx_api_keys_environment_id ON api_keys(environment_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_entity ON audit_logs(tenant_id, entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- Evaluation indexes
CREATE INDEX IF NOT EXISTS idx_evaluation_events_tenant_id ON evaluation_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_events_env_id ON evaluation_events(environment_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_events_flag_id ON evaluation_events(feature_flag_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_events_evaluated_at ON evaluation_events(evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_evaluation_events_flag_time ON evaluation_events(feature_flag_id, evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_evaluation_events_tenant_time ON evaluation_events(tenant_id, evaluated_at DESC);
CREATE INDEX IF NOT EXISTS idx_evaluation_summary_tenant_id ON evaluation_summary(tenant_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_summary_env_id ON evaluation_summary(environment_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_summary_flag_id ON evaluation_summary(feature_flag_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_summary_hour ON evaluation_summary(hour_bucket DESC);
CREATE INDEX IF NOT EXISTS idx_evaluation_summary_flag_hour ON evaluation_summary(feature_flag_id, hour_bucket DESC);

-- Approval workflow indexes
CREATE INDEX IF NOT EXISTS idx_approval_settings_tenant ON approval_settings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_approval_settings_env ON approval_settings(environment_id);
CREATE INDEX IF NOT EXISTS idx_approval_settings_flag ON approval_settings(feature_flag_id);
CREATE INDEX IF NOT EXISTS idx_pending_changes_tenant ON pending_changes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_pending_changes_env ON pending_changes(environment_id);
CREATE INDEX IF NOT EXISTS idx_pending_changes_flag ON pending_changes(feature_flag_id);
CREATE INDEX IF NOT EXISTS idx_pending_changes_status ON pending_changes(status) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_pending_changes_requester ON pending_changes(requested_by);
CREATE INDEX IF NOT EXISTS idx_approvals_pending ON approvals(pending_change_id);
CREATE INDEX IF NOT EXISTS idx_approvals_user ON approvals(approver_id);

-- Webhooks indexes
CREATE INDEX IF NOT EXISTS idx_webhooks_tenant ON webhooks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_enabled ON webhooks(tenant_id, enabled) WHERE enabled = TRUE;
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_status ON webhook_deliveries(status) WHERE status IN ('pending', 'retrying');
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_retry ON webhook_deliveries(next_retry_at) WHERE status = 'retrying';

-- Experiments indexes
CREATE INDEX IF NOT EXISTS idx_experiments_tenant ON experiments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_experiments_env ON experiments(environment_id);
CREATE INDEX IF NOT EXISTS idx_experiments_flag ON experiments(feature_flag_id);
CREATE INDEX IF NOT EXISTS idx_experiments_status ON experiments(status);
CREATE INDEX IF NOT EXISTS idx_experiment_variants_exp ON experiment_variants(experiment_id);
CREATE INDEX IF NOT EXISTS idx_experiment_metrics_exp ON experiment_metrics(experiment_id);
CREATE INDEX IF NOT EXISTS idx_experiment_results_exp ON experiment_results(experiment_id);

-- Rollouts indexes
CREATE INDEX IF NOT EXISTS idx_rollout_plans_flag ON rollout_plans(feature_flag_id);
CREATE INDEX IF NOT EXISTS idx_rollout_plans_status ON rollout_plans(status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_rollout_plans_next ON rollout_plans(next_increment_at) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_rollout_history_plan ON rollout_history(rollout_plan_id);

-- Emergency controls indexes
CREATE INDEX IF NOT EXISTS idx_emergency_controls_tenant ON emergency_controls(tenant_id);
CREATE INDEX IF NOT EXISTS idx_emergency_controls_env ON emergency_controls(environment_id);
CREATE INDEX IF NOT EXISTS idx_emergency_controls_active ON emergency_controls(is_active) WHERE is_active = TRUE;

-- Notifications indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_tenant ON notifications(tenant_id);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, read) WHERE read = FALSE;

-- Flag history indexes
CREATE INDEX IF NOT EXISTS idx_flag_history_flag ON flag_history(feature_flag_id);
CREATE INDEX IF NOT EXISTS idx_flag_history_version ON flag_history(feature_flag_id, version);

-- Trigger function for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_user_tenant_memberships_updated_at ON user_tenant_memberships;
CREATE TRIGGER update_user_tenant_memberships_updated_at
    BEFORE UPDATE ON user_tenant_memberships
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_applications_updated_at ON applications;
CREATE TRIGGER update_applications_updated_at
    BEFORE UPDATE ON applications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_environments_updated_at ON environments;
CREATE TRIGGER update_environments_updated_at
    BEFORE UPDATE ON environments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_feature_flags_updated_at ON feature_flags;
CREATE TRIGGER update_feature_flags_updated_at
    BEFORE UPDATE ON feature_flags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_targeting_rules_updated_at ON targeting_rules;
CREATE TRIGGER update_targeting_rules_updated_at
    BEFORE UPDATE ON targeting_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_evaluation_summary_updated_at ON evaluation_summary;
CREATE TRIGGER update_evaluation_summary_updated_at
    BEFORE UPDATE ON evaluation_summary
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- DEFAULT DATA FOR FIRST ACCESS
-- =====================================================

-- Default tenant
INSERT INTO tenants (id, name, slug, plan, active) 
VALUES ('00000000-0000-0000-0000-000000000001', 'Default Organization', 'default', 'free', true)
ON CONFLICT (slug) DO NOTHING;

-- Default admin user
-- Email: admin@flagflash.io
-- Password: admin123 (bcrypt hash with cost 10)
INSERT INTO users (id, tenant_id, email, password_hash, name, role, active)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'admin@flagflash.io',
    '$2a$10$VvwCl2CItO/xxOiqEyEaDO7UXfhEanjCqbsT19A2ZuL79/VlnNnKS',
    'Admin',
    'admin',
    true
)
ON CONFLICT (email) DO NOTHING;

-- Default admin membership (links admin user to default tenant)
INSERT INTO user_tenant_memberships (id, user_id, tenant_id, role, active)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'owner',
    true
)
ON CONFLICT (user_id, tenant_id) DO NOTHING;
