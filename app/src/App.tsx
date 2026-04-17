import { Routes, Route, NavLink, Navigate, useLocation, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import {
  LogOut,
  User,
  Building2,
  Key,
  Zap,
  Activity,
  BarChart3,
  Layers,
  Loader2,
  RefreshCw,
  Plus,
  Users,
  ChevronRight,
  Webhook,
  ShieldAlert,
  UsersRound,
  Bell,
  GitBranch,
} from 'lucide-react';

import LoginPage from './pages/Login';
import ChangePasswordPage from './pages/ChangePassword';
import AcceptInvitePage from './pages/AcceptInvite';
import SettingsPage from './pages/flagflash/Settings';
// FlagFlash Pages
import FlagFlashDashboard from './pages/flagflash/Dashboard';
import TenantsPage from './pages/flagflash/Tenants';
import ApplicationsPage from './pages/flagflash/Applications';
import EnvironmentsPage from './pages/flagflash/Environments';
import FeatureFlagsPage from './pages/flagflash/FeatureFlags';
import APIKeysPage from './pages/flagflash/APIKeys';
import AuditLogPage from './pages/flagflash/AuditLog';
import UsageMetricsPage from './pages/flagflash/UsageMetrics';
import SelectTenantPage from './pages/flagflash/SelectTenant';
import UsersPage from './pages/flagflash/Users';
import WebhooksPage from './pages/flagflash/Webhooks';
import EmergencyControlsPage from './pages/flagflash/EmergencyControls';
import SegmentsPage from './pages/flagflash/Segments';
import NotificationsPage from './pages/flagflash/Notifications';
import RolloutsPage from './pages/flagflash/Rollouts';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { applicationsApi } from './services/flagflash-api';
import type { Application } from './types/flagflash';

function SidebarApps({ tenantId }: { tenantId: string }) {
  const [apps, setApps] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);

  const MAX_VISIBLE = 3;

  useEffect(() => {
    setLoading(true);
    applicationsApi.list(tenantId, 1, 50)
      .then(r => setApps(r.applications || []))
      .catch(() => setApps([]))
      .finally(() => setLoading(false));
  }, [tenantId]);

  useEffect(() => {
    const refetch = () => {
      applicationsApi.list(tenantId, 1, 50)
        .then(r => setApps(r.applications || []))
        .catch(() => setApps([]));
    };
    window.addEventListener('appschanged', refetch);
    return () => window.removeEventListener('appschanged', refetch);
  }, [tenantId]);

  const visibleApps = apps.slice(0, MAX_VISIBLE);
  const hasMore = apps.length > MAX_VISIBLE;

  return (
    <div className="space-y-0.5">
      {/* New Application button */}
      <NavLink
        to={`/tenants/${tenantId}/applications?new=true`}
        className="flex items-center gap-3 py-2 px-3 rounded-lg text-sm font-medium transition-all duration-200 text-text-secondary hover:bg-bg-tertiary hover:text-text-primary"
      >
        <Plus size={16} className="shrink-0" />
        <span className="truncate">Nova Aplicação</span>
      </NavLink>

      {loading ? (
        <div className="flex justify-center py-2">
          <Loader2 size={16} className="animate-spin text-text-secondary" />
        </div>
      ) : apps.length === 0 ? (
        <p className="px-3 py-1 text-xs text-text-secondary italic">Nenhuma aplicação</p>
      ) : (
        <>
          {visibleApps.map(app => (
            <NavLink
              key={app.id}
              to={`/tenants/${tenantId}/applications/${app.id}/environments`}
              className={({ isActive }) =>
                `flex items-center gap-3 py-2 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                  isActive
                    ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                    : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
                }`
              }
            >
              <Layers size={16} className="shrink-0" />
              <span className="truncate">{app.name}</span>
            </NavLink>
          ))}
          {hasMore && (
            <NavLink
              to={`/tenants/${tenantId}/applications`}
              className={({ isActive }) =>
                `flex items-center gap-3 py-2 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                  isActive
                    ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                    : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
                }`
              }
            >
              <ChevronRight size={16} className="shrink-0" />
              <span className="truncate">Ver todos ({apps.length})</span>
            </NavLink>
          )}
        </>
      )}
    </div>
  );
}

function ProtectedApp() {
  const { isAuthenticated, user, logout, selectedTenant, clearTenant } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // Render tenant selection page without sidebar
  if (location.pathname === '/select-tenant') {
    return <SelectTenantPage />;
  }

  // Force tenant selection before accessing any other page
  if (!selectedTenant) {
    return <Navigate to="/select-tenant" replace />;
  }

  const handleChangeTenant = () => {
    clearTenant();
    navigate('/select-tenant');
  };

  return (
    <div className="flex min-h-screen bg-bg-primary">
      <aside className="w-64 bg-bg-secondary border-r border-border flex flex-col fixed h-screen">
        {/* Logo */}
        <div className="flex items-center gap-3 px-4 py-5 border-b border-border">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-accent-purple to-pink-500 flex items-center justify-center">
            <Zap size={22} className="text-white" />
          </div>
          <div>
            <h1 className="text-lg font-bold text-text-primary">FlagFlash</h1>
            <p className="text-xs text-text-secondary">Feature Flags Platform</p>
          </div>
        </div>

        {/* Selected Tenant */}
        <div className="px-4 py-3 border-b border-border">
          <p className="text-xs font-semibold text-text-secondary uppercase tracking-wider mb-2">Tenant Ativo</p>
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <div className="w-7 h-7 rounded-lg bg-accent-purple/20 flex items-center justify-center shrink-0">
                <Building2 size={14} className="text-accent-purple" />
              </div>
              <span className="text-sm font-medium text-text-primary truncate">{selectedTenant.name}</span>
            </div>
            <button
              onClick={handleChangeTenant}
              title="Mudar tenant"
              className="p-1.5 hover:bg-bg-tertiary rounded-lg transition-colors text-text-secondary hover:text-text-primary shrink-0"
            >
              <RefreshCw size={14} />
            </button>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 py-4 px-3 space-y-1 overflow-y-auto">
          <div className="text-xs font-semibold text-text-secondary uppercase tracking-wider px-3 mb-2">
            Overview
          </div>
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <Zap size={18} />
            Dashboard
          </NavLink>

          <div className="text-xs font-semibold text-text-secondary uppercase tracking-wider px-3 mt-6 mb-2">
            Management
          </div>
          <NavLink
            to={`/tenants/${selectedTenant.id}/api-keys`}
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <Key size={18} />
            API Keys
          </NavLink>
          <NavLink
            to={`/tenants/${selectedTenant.id}/users`}
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <Users size={18} />
            Users
          </NavLink>
          <NavLink
            to={`/tenants/${selectedTenant.id}/segments`}
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <UsersRound size={18} />
            Segments
          </NavLink>
          <NavLink
            to={`/tenants/${selectedTenant.id}/webhooks`}
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <Webhook size={18} />
            Webhooks
          </NavLink>
          <NavLink
            to={`/tenants/${selectedTenant.id}/emergency`}
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-red-500/10 text-red-400 border-l-2 border-red-400'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <ShieldAlert size={18} />
            Emergency
          </NavLink>
          <NavLink
            to="/notifications"
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <Bell size={18} />
            Notifications
          </NavLink>

          <div className="text-xs font-semibold text-text-secondary uppercase tracking-wider px-3 mt-6 mb-2">
            Applications
          </div>
          <SidebarApps tenantId={selectedTenant.id} />

          <div className="text-xs font-semibold text-text-secondary uppercase tracking-wider px-3 mt-6 mb-2">
            Analytics
          </div>
          <NavLink
            to="/analytics"
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <BarChart3 size={18} />
            Usage Metrics
          </NavLink>
          <NavLink
            to="/audit-log"
            className={({ isActive }) =>
              `flex items-center gap-3 py-2.5 px-3 rounded-lg text-sm font-medium transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <Activity size={18} />
            Audit Log
          </NavLink>
        </nav>

        {/* Footer */}
        <div className="p-4 border-t border-border space-y-1">
          <NavLink
            to="/settings"
            className={({ isActive }) =>
              `flex items-center gap-2 px-2 py-2 rounded-lg text-sm transition-all duration-200 ${
                isActive
                  ? 'bg-accent-purple/10 text-accent-purple'
                  : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
              }`
            }
          >
            <div className="w-7 h-7 rounded-full bg-accent-purple/20 flex items-center justify-center shrink-0">
              <User size={14} className="text-accent-purple" />
            </div>
            <span className="font-medium truncate">{user?.name || user?.email}</span>
          </NavLink>
          <button
            onClick={logout}
            className="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm text-text-secondary hover:bg-bg-tertiary hover:text-red-400 transition-all duration-200"
          >
            <LogOut size={16} />
            Sign out
          </button>
        </div>
      </aside>

      <main className="flex-1 ml-64 p-6 bg-bg-primary">
        <Routes>
          <Route path="/" element={<FlagFlashDashboard />} />
          <Route path="/tenants" element={<TenantsPage />} />
          <Route path="/tenants/:tenantId/applications" element={<ApplicationsPage />} />
          <Route path="/tenants/:tenantId/applications/:appId/environments" element={<EnvironmentsPage />} />
          <Route path="/tenants/:tenantId/applications/:appId/environments/:envId/flags" element={<FeatureFlagsPage />} />
          <Route path="/tenants/:tenantId/api-keys" element={<APIKeysPage />} />
          <Route path="/tenants/:tenantId/users" element={<UsersPage />} />
          <Route path="/tenants/:tenantId/segments" element={<SegmentsPage />} />
          <Route path="/tenants/:tenantId/webhooks" element={<WebhooksPage />} />
          <Route path="/tenants/:tenantId/emergency" element={<EmergencyControlsPage />} />
          <Route path="/tenants/:tenantId/rollouts" element={<RolloutsPage />} />
          <Route path="/notifications" element={<NotificationsPage />} />
          <Route path="/analytics" element={<UsageMetricsPage />} />
          <Route path="/audit-log" element={<AuditLogPage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="/change-password" element={<ChangePasswordPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/accept-invite" element={<AcceptInvitePage />} />
        <Route path="/*" element={<ProtectedApp />} />
      </Routes>
    </AuthProvider>
  );
}

export default App;
