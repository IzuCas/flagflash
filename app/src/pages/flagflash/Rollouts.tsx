import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  GitBranch, 
  Play,
  Pause,
  RotateCcw,
  Trash2,
  ChevronLeft,
  Loader2,
  AlertCircle,
  Plus,
  Clock,
  TrendingUp,
  CheckCircle,
  XCircle,
  Flag,
} from 'lucide-react';
import { rolloutsApi, applicationsApi, environmentsApi, featureFlagsApi } from '../../services/flagflash-api';
import { useAuth } from '../../contexts/AuthContext';
import { usePermissions } from '../../hooks/usePermissions';
import { ConfirmDeleteModal, Modal } from '../../components';
import type { RolloutPlan, RolloutStatus, Application, Environment, FeatureFlag, CreateRolloutRequest } from '../../types/flagflash';

const STATUS_COLORS: Record<RolloutStatus, string> = {
  draft: 'bg-text-tertiary',
  active: 'bg-accent-green',
  paused: 'bg-accent-amber',
  completed: 'bg-accent-blue',
  failed: 'bg-accent-red',
};

const STATUS_ICONS: Record<RolloutStatus, React.ReactNode> = {
  draft: <Clock size={14} />,
  active: <Play size={14} />,
  paused: <Pause size={14} />,
  completed: <CheckCircle size={14} />,
  failed: <XCircle size={14} />,
};

export default function RolloutsPage() {
  const { tenantId: urlTenantId } = useParams<{ tenantId: string }>();
  const { selectedTenant } = useAuth();
  const { isAtLeast } = usePermissions();
  const isAdmin = isAtLeast('admin');
  const activeTenantId = urlTenantId || selectedTenant?.id || '';
  
  const [applications, setApplications] = useState<Application[]>([]);
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [flags, setFlags] = useState<FeatureFlag[]>([]);
  const [rollouts, setRollouts] = useState<RolloutPlan[]>([]);
  
  const [selectedAppId, setSelectedAppId] = useState<string>('');
  const [selectedEnvId, setSelectedEnvId] = useState<string>('');
  const [selectedFlagId, setSelectedFlagId] = useState<string>('');
  
  const [loading, setLoading] = useState(true);
  const [loadingRollouts, setLoadingRollouts] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [rolloutToDelete, setRolloutToDelete] = useState<RolloutPlan | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  // Load applications on mount
  useEffect(() => {
    if (!activeTenantId) {
      setLoading(false);
      return;
    }

    const loadApps = async () => {
      try {
        setLoading(true);
        const response = await applicationsApi.list(activeTenantId);
        setApplications(response.applications || []);
        setError(null);
      } catch (err) {
        setError('Failed to load applications');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    loadApps();
  }, [activeTenantId]);

  // Load environments when app changes
  useEffect(() => {
    if (!selectedAppId || !activeTenantId) {
      setEnvironments([]);
      setSelectedEnvId('');
      return;
    }

    const loadEnvs = async () => {
      try {
        const response = await environmentsApi.list(activeTenantId, selectedAppId);
        setEnvironments(response.environments || []);
      } catch (err) {
        console.error('Failed to load environments:', err);
      }
    };

    loadEnvs();
  }, [activeTenantId, selectedAppId]);

  // Load flags when env changes
  useEffect(() => {
    if (!selectedEnvId || !selectedAppId || !activeTenantId) {
      setFlags([]);
      setSelectedFlagId('');
      return;
    }

    const loadFlags = async () => {
      try {
        const response = await featureFlagsApi.list(activeTenantId, selectedAppId, selectedEnvId);
        setFlags(response.flags || []);
      } catch (err) {
        console.error('Failed to load flags:', err);
      }
    };

    loadFlags();
  }, [activeTenantId, selectedAppId, selectedEnvId]);

  // Load rollouts when flag changes
  useEffect(() => {
    if (!selectedFlagId || !selectedEnvId || !selectedAppId || !activeTenantId) {
      setRollouts([]);
      return;
    }

    const loadRollouts = async () => {
      try {
        setLoadingRollouts(true);
        const response = await rolloutsApi.list(activeTenantId, selectedAppId, selectedEnvId, selectedFlagId);
        setRollouts(response.plans || []);
      } catch (err) {
        console.error('Failed to load rollouts:', err);
        setRollouts([]);
      } finally {
        setLoadingRollouts(false);
      }
    };

    loadRollouts();
  }, [activeTenantId, selectedAppId, selectedEnvId, selectedFlagId]);

  const handleAction = async (rollout: RolloutPlan, action: 'start' | 'pause' | 'resume' | 'rollback') => {
    setActionLoading(rollout.id);
    try {
      let updated: RolloutPlan;
      switch (action) {
        case 'start':
          updated = await rolloutsApi.start(activeTenantId, rollout.id);
          break;
        case 'pause':
          updated = await rolloutsApi.pause(activeTenantId, rollout.id);
          break;
        case 'resume':
          updated = await rolloutsApi.resume(activeTenantId, rollout.id);
          break;
        case 'rollback':
          updated = await rolloutsApi.rollback(activeTenantId, rollout.id, 'Manual rollback');
          break;
      }
      setRollouts(prev => prev.map(r => r.id === rollout.id ? updated : r));
    } catch (err) {
      console.error(`Failed to ${action} rollout:`, err);
    } finally {
      setActionLoading(null);
    }
  };

  const handleDelete = async () => {
    if (!rolloutToDelete) return;
    
    try {
      await rolloutsApi.delete(activeTenantId, rolloutToDelete.id);
      setRollouts(prev => prev.filter(r => r.id !== rolloutToDelete.id));
      setRolloutToDelete(null);
    } catch (err) {
      console.error('Failed to delete rollout:', err);
    }
  };

  const handleCreateRollout = async (data: CreateRolloutRequest) => {
    try {
      const created = await rolloutsApi.create(activeTenantId, selectedAppId, selectedEnvId, selectedFlagId, data);
      setRollouts(prev => [...prev, created]);
      setShowCreateModal(false);
    } catch (err) {
      console.error('Failed to create rollout:', err);
    }
  };

  const selectedFlag = flags.find(f => f.id === selectedFlagId);

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
              <GitBranch className="text-accent-purple" />
              Progressive Rollouts
            </h1>
            <p className="text-text-secondary mt-1">
              Gradually release features to users with automatic percentage increases
            </p>
          </div>
        </div>

        {/* Error State */}
        {error && (
          <div className="bg-accent-red/10 border border-accent-red/20 rounded-lg p-4 mb-6 flex items-center gap-3">
            <AlertCircle className="text-accent-red" size={20} />
            <span className="text-accent-red">{error}</span>
          </div>
        )}

        {/* Loading State */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="animate-spin text-accent-purple" size={32} />
          </div>
        ) : (
          <>
            {/* Selectors */}
            <div className="bg-bg-secondary rounded-lg p-6 mb-6">
              <h2 className="text-lg font-medium text-text-primary mb-4">Select Flag</h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    Application
                  </label>
                  <select
                    value={selectedAppId}
                    onChange={(e) => setSelectedAppId(e.target.value)}
                    className="w-full px-3 py-2 bg-bg-primary border border-border-primary rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
                  >
                    <option value="">Select application...</option>
                    {applications.map(app => (
                      <option key={app.id} value={app.id}>{app.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    Environment
                  </label>
                  <select
                    value={selectedEnvId}
                    onChange={(e) => setSelectedEnvId(e.target.value)}
                    disabled={!selectedAppId}
                    className="w-full px-3 py-2 bg-bg-primary border border-border-primary rounded-lg text-text-primary focus:outline-none focus:border-accent-purple disabled:opacity-50"
                  >
                    <option value="">Select environment...</option>
                    {environments.map(env => (
                      <option key={env.id} value={env.id}>{env.name}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    Feature Flag
                  </label>
                  <select
                    value={selectedFlagId}
                    onChange={(e) => setSelectedFlagId(e.target.value)}
                    disabled={!selectedEnvId}
                    className="w-full px-3 py-2 bg-bg-primary border border-border-primary rounded-lg text-text-primary focus:outline-none focus:border-accent-purple disabled:opacity-50"
                  >
                    <option value="">Select flag...</option>
                    {flags.map(flag => (
                      <option key={flag.id} value={flag.id}>{flag.key} - {flag.name}</option>
                    ))}
                  </select>
                </div>
              </div>
            </div>

            {/* Rollouts Section */}
            {selectedFlagId && (
              <div className="bg-bg-secondary rounded-lg p-6">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <Flag size={20} className="text-accent-blue" />
                    <h2 className="text-lg font-medium text-text-primary">
                      Rollouts for {selectedFlag?.name}
                    </h2>
                  </div>
                  {isAdmin && (
                    <button
                      onClick={() => setShowCreateModal(true)}
                      className="flex items-center gap-2 px-4 py-2 bg-accent-purple hover:bg-accent-purple/80 text-white rounded-lg transition-colors"
                    >
                      <Plus size={16} />
                      New Rollout
                    </button>
                  )}
                </div>

                {loadingRollouts ? (
                  <div className="flex items-center justify-center py-8">
                    <Loader2 className="animate-spin text-accent-purple" size={24} />
                  </div>
                ) : rollouts.length === 0 ? (
                  <div className="text-center py-8 text-text-secondary">
                    <TrendingUp size={32} className="mx-auto mb-2 opacity-50" />
                    <p>No rollouts configured for this flag</p>
                    <p className="text-sm mt-1">Create a rollout to gradually release this feature</p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {rollouts.map(rollout => (
                      <div
                        key={rollout.id}
                        className="bg-bg-primary rounded-lg p-4 border border-border-primary"
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <div className={`w-3 h-3 rounded-full ${STATUS_COLORS[rollout.status]}`} />
                            <div>
                              <h3 className="font-medium text-text-primary">{rollout.name}</h3>
                              <div className="flex items-center gap-4 text-sm text-text-secondary mt-1">
                                <span className="flex items-center gap-1">
                                  {STATUS_ICONS[rollout.status]}
                                  {rollout.status}
                                </span>
                                <span>
                                  {rollout.current_percentage}% → {rollout.target_percentage}%
                                </span>
                                <span>
                                  +{rollout.increment_percentage}% every {rollout.increment_interval_minutes}min
                                </span>
                              </div>
                            </div>
                          </div>
                          
                          <div className="flex items-center gap-2">
                            {/* Progress bar */}
                            <div className="w-32 h-2 bg-bg-tertiary rounded-full overflow-hidden mr-4">
                              <div
                                className="h-full bg-accent-green transition-all"
                                style={{ width: `${(rollout.current_percentage / rollout.target_percentage) * 100}%` }}
                              />
                            </div>

                            {/* Action buttons */}
                            {rollout.status === 'draft' && (
                              <button
                                onClick={() => handleAction(rollout, 'start')}
                                disabled={actionLoading === rollout.id}
                                className="p-2 hover:bg-bg-secondary rounded-lg text-accent-green"
                                title="Start"
                              >
                                {actionLoading === rollout.id ? (
                                  <Loader2 size={16} className="animate-spin" />
                                ) : (
                                  <Play size={16} />
                                )}
                              </button>
                            )}
                            {rollout.status === 'active' && (
                              <button
                                onClick={() => handleAction(rollout, 'pause')}
                                disabled={actionLoading === rollout.id}
                                className="p-2 hover:bg-bg-secondary rounded-lg text-accent-amber"
                                title="Pause"
                              >
                                {actionLoading === rollout.id ? (
                                  <Loader2 size={16} className="animate-spin" />
                                ) : (
                                  <Pause size={16} />
                                )}
                              </button>
                            )}
                            {rollout.status === 'paused' && (
                              <button
                                onClick={() => handleAction(rollout, 'resume')}
                                disabled={actionLoading === rollout.id}
                                className="p-2 hover:bg-bg-secondary rounded-lg text-accent-green"
                                title="Resume"
                              >
                                {actionLoading === rollout.id ? (
                                  <Loader2 size={16} className="animate-spin" />
                                ) : (
                                  <Play size={16} />
                                )}
                              </button>
                            )}
                            {(rollout.status === 'active' || rollout.status === 'paused') && (
                              <button
                                onClick={() => handleAction(rollout, 'rollback')}
                                disabled={actionLoading === rollout.id}
                                className="p-2 hover:bg-bg-secondary rounded-lg text-accent-red"
                                title="Rollback"
                              >
                                <RotateCcw size={16} />
                              </button>
                            )}
                            {isAdmin && (
                              <button
                                onClick={() => setRolloutToDelete(rollout)}
                                className="p-2 hover:bg-bg-secondary rounded-lg text-text-secondary hover:text-accent-red"
                                title="Delete"
                              >
                                <Trash2 size={16} />
                              </button>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Instructions when nothing selected */}
            {!selectedFlagId && (
              <div className="bg-bg-secondary rounded-lg p-8 text-center">
                <TrendingUp size={48} className="mx-auto text-text-tertiary mb-4" />
                <h3 className="text-lg font-medium text-text-primary">Select a Feature Flag</h3>
                <p className="text-text-secondary mt-2 max-w-md mx-auto">
                  Choose an application, environment, and feature flag above to view and manage its rollout plans.
                </p>
              </div>
            )}
          </>
        )}
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <CreateRolloutModal
          onClose={() => setShowCreateModal(false)}
          onSubmit={handleCreateRollout}
        />
      )}

      {/* Delete Confirmation */}
      <ConfirmDeleteModal
        isOpen={!!rolloutToDelete}
        onClose={() => setRolloutToDelete(null)}
        onConfirm={handleDelete}
        title="Delete Rollout Plan"
        message={`Are you sure you want to delete the rollout plan "${rolloutToDelete?.name}"? This action cannot be undone.`}
      />
    </div>
  );
}

// Create Rollout Modal Component
function CreateRolloutModal({ 
  onClose, 
  onSubmit 
}: { 
  onClose: () => void;
  onSubmit: (data: CreateRolloutRequest) => Promise<void>;
}) {
  const [name, setName] = useState('');
  const [targetPercentage, setTargetPercentage] = useState(100);
  const [incrementPercentage, setIncrementPercentage] = useState(10);
  const [intervalMinutes, setIntervalMinutes] = useState(60);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await onSubmit({
        name,
        target_percentage: targetPercentage,
        increment_percentage: incrementPercentage,
        increment_interval_minutes: intervalMinutes,
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal title="Create Rollout Plan" onClose={onClose}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1">
            Name *
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., Gradual rollout to production"
            required
            className="w-full px-3 py-2 bg-bg-primary border border-border-primary rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
          />
        </div>

        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-1">
              Target % *
            </label>
            <input
              type="number"
              value={targetPercentage}
              onChange={(e) => setTargetPercentage(Number(e.target.value))}
              min={1}
              max={100}
              required
              className="w-full px-3 py-2 bg-bg-primary border border-border-primary rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-1">
              Increment % *
            </label>
            <input
              type="number"
              value={incrementPercentage}
              onChange={(e) => setIncrementPercentage(Number(e.target.value))}
              min={1}
              max={100}
              required
              className="w-full px-3 py-2 bg-bg-primary border border-border-primary rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-1">
              Interval (min) *
            </label>
            <input
              type="number"
              value={intervalMinutes}
              onChange={(e) => setIntervalMinutes(Number(e.target.value))}
              min={1}
              required
              className="w-full px-3 py-2 bg-bg-primary border border-border-primary rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            />
          </div>
        </div>

        <p className="text-sm text-text-tertiary">
          This will increase the flag's rollout percentage by {incrementPercentage}% every {intervalMinutes} minutes until it reaches {targetPercentage}%.
        </p>

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
            disabled={submitting || !name}
            className="px-4 py-2 bg-accent-purple hover:bg-accent-purple/80 text-white rounded-lg transition-colors disabled:opacity-50"
          >
            {submitting ? 'Creating...' : 'Create Rollout'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
