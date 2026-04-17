import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  ShieldAlert, 
  Plus, 
  ChevronLeft,
  Loader2,
  AlertCircle,
  AlertTriangle,
  Power,
  Clock,
  Shield,
  Wrench,
} from 'lucide-react';
import { emergencyControlsApi, environmentsApi, applicationsApi } from '../../services/flagflash-api';
import { useAuth } from '../../contexts/AuthContext';
import { usePermissions } from '../../hooks/usePermissions';
import { ConfirmDeleteModal, Modal } from '../../components';
import type { EmergencyControl, EmergencyControlType, Environment } from '../../types/flagflash';

export default function EmergencyControlsPage() {
  const { tenantId: urlTenantId } = useParams<{ tenantId: string }>();
  const { selectedTenant } = useAuth();
  const { isAtLeast } = usePermissions();
  const isAdmin = isAtLeast('admin');
  const activeTenantId = urlTenantId || selectedTenant?.id || '';
  
  const [controls, setControls] = useState<EmergencyControl[]>([]);
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showActivateModal, setShowActivateModal] = useState(false);
  const [controlToDeactivate, setControlToDeactivate] = useState<EmergencyControl | null>(null);

  const fetchData = useCallback(async () => {
    if (!activeTenantId) {
      setLoading(false);
      return;
    }
    
    try {
      setLoading(true);
      const [controlsResponse, appsResponse] = await Promise.all([
        emergencyControlsApi.list(activeTenantId),
        applicationsApi.list(activeTenantId),
      ]);
      
      setControls(controlsResponse.controls || []);
      
      // Fetch all environments
      const allEnvs: Environment[] = [];
      for (const app of appsResponse.applications || []) {
        const envsResponse = await environmentsApi.list(activeTenantId, app.id);
        allEnvs.push(...(envsResponse.environments || []));
      }
      setEnvironments(allEnvs);
      setError(null);
    } catch (err) {
      setError('Failed to load emergency controls');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [activeTenantId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleDeactivate = async () => {
    if (!controlToDeactivate || !activeTenantId) return;
    
    await emergencyControlsApi.deactivate(activeTenantId, controlToDeactivate.id);
    setControls(prev => prev.filter(c => c.id !== controlToDeactivate.id));
    setControlToDeactivate(null);
  };

  const getEnvironmentName = (envId?: string) => {
    if (!envId) return 'All Environments';
    const env = environments.find(e => e.id === envId);
    return env?.name || 'Unknown';
  };

  const activeControls = controls.filter(c => c.enabled);
  const hasActiveKillSwitch = activeControls.some(c => c.control_type === 'kill_switch');
  const hasActiveMaintenance = activeControls.some(c => c.control_type === 'maintenance');

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          {urlTenantId && (
            <Link
              to="/tenants"
              className="p-2 hover:bg-bg-secondary rounded-lg transition-colors"
            >
              <ChevronLeft size={20} className="text-text-secondary" />
            </Link>
          )}
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <ShieldAlert className="text-red-400" />
              Emergency Controls
            </h1>
            <p className="text-text-secondary mt-1">
              Kill switches and maintenance mode controls
            </p>
          </div>
          {isAdmin && (
            <button
              onClick={() => setShowActivateModal(true)}
              disabled={!activeTenantId}
              className="flex items-center gap-2 px-4 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Plus size={20} />
              Activate Control
            </button>
          )}
        </div>

        {/* Status Overview */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
          <div className={`p-6 rounded-xl border ${
            hasActiveKillSwitch 
              ? 'bg-red-500/10 border-red-500/50' 
              : 'bg-bg-secondary border-border'
          }`}>
            <div className="flex items-center gap-3 mb-2">
              <div className={`p-2 rounded-lg ${
                hasActiveKillSwitch ? 'bg-red-500/20' : 'bg-bg-tertiary'
              }`}>
                <Power size={20} className={hasActiveKillSwitch ? 'text-red-400' : 'text-text-secondary'} />
              </div>
              <div>
                <h3 className="font-semibold text-text-primary">Kill Switch</h3>
                <p className={`text-sm ${hasActiveKillSwitch ? 'text-red-400' : 'text-green-400'}`}>
                  {hasActiveKillSwitch ? 'ACTIVE - All flags disabled' : 'Inactive'}
                </p>
              </div>
            </div>
            {hasActiveKillSwitch && (
              <p className="text-sm text-text-secondary mt-2">
                All feature flag evaluations are returning default values.
              </p>
            )}
          </div>

          <div className={`p-6 rounded-xl border ${
            hasActiveMaintenance 
              ? 'bg-yellow-500/10 border-yellow-500/50' 
              : 'bg-bg-secondary border-border'
          }`}>
            <div className="flex items-center gap-3 mb-2">
              <div className={`p-2 rounded-lg ${
                hasActiveMaintenance ? 'bg-yellow-500/20' : 'bg-bg-tertiary'
              }`}>
                <Wrench size={20} className={hasActiveMaintenance ? 'text-yellow-400' : 'text-text-secondary'} />
              </div>
              <div>
                <h3 className="font-semibold text-text-primary">Maintenance Mode</h3>
                <p className={`text-sm ${hasActiveMaintenance ? 'text-yellow-400' : 'text-green-400'}`}>
                  {hasActiveMaintenance ? 'ACTIVE - Read-only mode' : 'Inactive'}
                </p>
              </div>
            </div>
            {hasActiveMaintenance && (
              <p className="text-sm text-text-secondary mt-2">
                Flag modifications are temporarily disabled.
              </p>
            )}
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="mb-6 p-4 bg-red-500/10 border border-red-500/50 rounded-lg flex items-center gap-3 text-red-400">
            <AlertCircle size={20} />
            {error}
          </div>
        )}

        {/* Loading */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="animate-spin text-accent-purple" size={32} />
          </div>
        ) : (
          /* Active Controls */
          <div className="space-y-4">
            <h2 className="text-lg font-semibold text-text-primary">Active Controls</h2>
            
            {activeControls.length === 0 ? (
              <div className="text-center py-12 bg-bg-secondary border border-border rounded-xl">
                <Shield size={48} className="mx-auto mb-4 text-green-400" />
                <p className="text-text-primary font-medium">All Systems Normal</p>
                <p className="text-sm text-text-secondary mt-1">No emergency controls are currently active</p>
              </div>
            ) : (
              activeControls.map(control => (
                <div
                  key={control.id}
                  className={`bg-bg-secondary border rounded-xl p-5 ${
                    control.control_type === 'kill_switch' 
                      ? 'border-red-500/50' 
                      : 'border-yellow-500/50'
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        {control.control_type === 'kill_switch' ? (
                          <Power className="text-red-400" size={20} />
                        ) : (
                          <Wrench className="text-yellow-400" size={20} />
                        )}
                        <h3 className="font-semibold text-text-primary">
                          {control.control_type === 'kill_switch' ? 'Kill Switch' : 'Maintenance Mode'}
                        </h3>
                        <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                          control.control_type === 'kill_switch' 
                            ? 'bg-red-500/20 text-red-400'
                            : 'bg-yellow-500/20 text-yellow-400'
                        }`}>
                          Active
                        </span>
                      </div>
                      
                      <p className="text-text-secondary mb-3">{control.reason}</p>

                      <div className="flex items-center gap-4 text-sm text-text-secondary">
                        <span className="flex items-center gap-1">
                          <Shield size={14} />
                          {getEnvironmentName(control.environment_id)}
                        </span>
                        {control.enabled_at && (
                          <span className="flex items-center gap-1">
                            <Clock size={14} />
                            Activated {new Date(control.enabled_at).toLocaleString()}
                          </span>
                        )}
                        {control.expires_at && (
                          <span className="flex items-center gap-1">
                            <AlertTriangle size={14} />
                            Expires {new Date(control.expires_at).toLocaleString()}
                          </span>
                        )}
                      </div>
                    </div>

                    {isAdmin && (
                      <button
                        onClick={() => setControlToDeactivate(control)}
                        className="px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                      >
                        Deactivate
                      </button>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {/* Activate Modal */}
        {showActivateModal && (
          <ActivateControlModal
            tenantId={activeTenantId}
            environments={environments}
            onClose={() => setShowActivateModal(false)}
            onActivate={(control) => {
              setControls(prev => [...prev, control]);
              setShowActivateModal(false);
            }}
          />
        )}

        {/* Deactivate Modal */}
        <ConfirmDeleteModal
          isOpen={!!controlToDeactivate}
          onClose={() => setControlToDeactivate(null)}
          onConfirm={handleDeactivate}
          title="Deactivate Emergency Control"
          message={`Are you sure you want to deactivate this ${controlToDeactivate?.control_type === 'kill_switch' ? 'kill switch' : 'maintenance mode'}? Normal operations will resume immediately.`}
          confirmText="Deactivate"
        />
      </div>
    </div>
  );
}

interface ActivateControlModalProps {
  tenantId: string;
  environments: Environment[];
  onClose: () => void;
  onActivate: (control: EmergencyControl) => void;
}

function ActivateControlModal({ tenantId, environments, onClose, onActivate }: ActivateControlModalProps) {
  const [controlType, setControlType] = useState<EmergencyControlType>('kill_switch');
  const [environmentId, setEnvironmentId] = useState<string>('');
  const [reason, setReason] = useState('');
  const [expiresInMinutes, setExpiresInMinutes] = useState<number | undefined>();
  const [activating, setActivating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!reason) return;

    try {
      setActivating(true);
      setError(null);
      
      const apiMethod = controlType === 'kill_switch' 
        ? emergencyControlsApi.activateKillSwitch 
        : emergencyControlsApi.activateMaintenance;
      
      const control = await apiMethod(tenantId, {
        control_type: controlType,
        environment_id: environmentId || undefined,
        reason,
        expires_in_minutes: expiresInMinutes,
      });
      
      onActivate(control);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setActivating(false);
    }
  };

  return (
    <Modal isOpen onClose={onClose} title="Activate Emergency Control">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div className="p-3 bg-yellow-500/10 border border-yellow-500/50 rounded-lg text-yellow-400 text-sm flex items-start gap-2">
          <AlertTriangle size={16} className="mt-0.5 shrink-0" />
          <span>This will immediately affect all feature flag evaluations for the selected scope.</span>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-2">Control Type</label>
          <div className="grid grid-cols-2 gap-3">
            <button
              type="button"
              onClick={() => setControlType('kill_switch')}
              className={`p-4 border rounded-lg text-left transition-colors ${
                controlType === 'kill_switch'
                  ? 'bg-red-500/10 border-red-500 text-red-400'
                  : 'border-border text-text-secondary hover:border-red-500/50'
              }`}
            >
              <Power size={20} className="mb-2" />
              <div className="font-medium">Kill Switch</div>
              <div className="text-xs mt-1 opacity-75">Disable all flags</div>
            </button>
            <button
              type="button"
              onClick={() => setControlType('maintenance')}
              className={`p-4 border rounded-lg text-left transition-colors ${
                controlType === 'maintenance'
                  ? 'bg-yellow-500/10 border-yellow-500 text-yellow-400'
                  : 'border-border text-text-secondary hover:border-yellow-500/50'
              }`}
            >
              <Wrench size={20} className="mb-2" />
              <div className="font-medium">Maintenance</div>
              <div className="text-xs mt-1 opacity-75">Read-only mode</div>
            </button>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">Scope</label>
          <select
            value={environmentId}
            onChange={(e) => setEnvironmentId(e.target.value)}
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
          >
            <option value="">All Environments</option>
            {environments.map(env => (
              <option key={env.id} value={env.id}>{env.name}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">Reason</label>
          <textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Describe why this control is being activated..."
            rows={3}
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple resize-none"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">Auto-expire (optional)</label>
          <select
            value={expiresInMinutes || ''}
            onChange={(e) => setExpiresInMinutes(e.target.value ? Number(e.target.value) : undefined)}
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
          >
            <option value="">Never (manual deactivation)</option>
            <option value="15">15 minutes</option>
            <option value="30">30 minutes</option>
            <option value="60">1 hour</option>
            <option value="120">2 hours</option>
            <option value="240">4 hours</option>
            <option value="480">8 hours</option>
            <option value="1440">24 hours</option>
          </select>
        </div>

        <div className="flex justify-end gap-3 pt-4">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={activating || !reason}
            className={`px-4 py-2 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 ${
              controlType === 'kill_switch' 
                ? 'bg-red-500 hover:bg-red-600' 
                : 'bg-yellow-500 hover:bg-yellow-600'
            }`}
          >
            {activating && <Loader2 size={16} className="animate-spin" />}
            Activate {controlType === 'kill_switch' ? 'Kill Switch' : 'Maintenance Mode'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
