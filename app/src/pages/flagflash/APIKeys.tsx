import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  Key, 
  Plus, 
  Search, 
  Trash2, 
  Ban,
  Copy,
  Eye,
  EyeOff,
  Check,
  ChevronLeft,
  Loader2,
  AlertCircle,
  Clock,
  Shield,
  Layers,
} from 'lucide-react';
import { apiKeysApi, environmentsApi, applicationsApi } from '../../services/flagflash-api';
import { useAuth } from '../../contexts/AuthContext';
import { usePermissions } from '../../hooks/usePermissions';
import { ConfirmDeleteModal, Modal } from '../../components';
import type { APIKey, APIKeyCreatedResponse, Environment, Application } from '../../types/flagflash';

export default function APIKeysPage() {
  const { tenantId: urlTenantId } = useParams<{ tenantId: string }>();
  const { selectedTenant } = useAuth();
  const { canCreateAPIKey, canRevokeAPIKey, canDeleteAPIKey } = usePermissions();
  const activeTenantId = urlTenantId || selectedTenant?.id || '';
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newKey, setNewKey] = useState<APIKeyCreatedResponse | null>(null);
  const [keyToDelete, setKeyToDelete] = useState<APIKey | null>(null);
  const [keyToRevoke, setKeyToRevoke] = useState<APIKey | null>(null);

  const fetchData = useCallback(async () => {
    if (!activeTenantId) {
      setLoading(false);
      return;
    }
    
    try {
      setLoading(true);
      const [keysResponse, appsResponse] = await Promise.all([
        apiKeysApi.list(activeTenantId),
        applicationsApi.list(activeTenantId)
      ]);
      setApiKeys(keysResponse.keys || []);
      setApplications(appsResponse.applications || []);

      // Fetch environments for all applications
      const allEnvs: Environment[] = [];
      for (const app of appsResponse.applications || []) {
        const envsResponse = await environmentsApi.list(activeTenantId, app.id);
        allEnvs.push(...(envsResponse.environments || []));
      }
      setEnvironments(allEnvs);
      setError(null);
    } catch (err) {
      setError('Failed to load API keys');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [activeTenantId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filteredKeys = apiKeys.filter(key => 
    key.name.toLowerCase().includes(search.toLowerCase()) ||
    key.key_prefix.toLowerCase().includes(search.toLowerCase())
  );

  const handleRevoke = async () => {
    if (!keyToRevoke || !activeTenantId) return;
    
    await apiKeysApi.revoke(activeTenantId, keyToRevoke.id);
    setApiKeys(prev => prev.map(k => k.id === keyToRevoke.id ? { ...k, active: false } : k));
    setKeyToRevoke(null);
  };

  const handleDelete = async () => {
    if (!keyToDelete || !activeTenantId) return;
    
    await apiKeysApi.delete(activeTenantId, keyToDelete.id);
    setApiKeys(prev => prev.filter(k => k.id !== keyToDelete.id));
    setKeyToDelete(null);
  };

  const getEnvironmentName = (envId: string) => {
    const env = environments.find(e => e.id === envId);
    return env?.name || 'Unknown';
  };

  const getEnvironmentColor = (envId: string) => {
    const env = environments.find(e => e.id === envId);
    return env?.color || '#8b5cf6';
  };

  const getApplicationName = (envId: string) => {
    const env = environments.find(e => e.id === envId);
    const app = applications.find(a => a.id === env?.application_id);
    return app?.name || null;
  };

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
              <Key className="text-accent-purple" />
              API Keys
            </h1>
            <p className="text-text-secondary mt-1">
              Manage API keys for SDK access
            </p>
          </div>
          {canCreateAPIKey && (
            <button
              onClick={() => setShowCreateModal(true)}
              disabled={!activeTenantId}
              className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Plus size={20} />
              Create API Key
            </button>
          )}
        </div>

        {/* Search */}
        <div className="relative mb-6">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary" size={20} />
          <input
            type="text"
            placeholder="Search API keys..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
          />
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
          /* API Keys List */
          <div className="space-y-3">
            {filteredKeys.map(key => (
              <div
                key={key.id}
                className={`bg-bg-secondary border rounded-xl p-5 transition-colors ${
                  key.active ? 'border-border hover:border-accent-purple/50' : 'border-red-500/30 opacity-75'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <h3 className="font-semibold text-text-primary">{key.name}</h3>
                      <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                        key.active 
                          ? 'bg-green-500/20 text-green-400'
                          : 'bg-red-500/20 text-red-400'
                      }`}>
                        {key.active ? 'Active' : 'Revoked'}
                      </span>
                    </div>
                    
                    <div className="flex items-center gap-4 text-sm mb-3">
                      <span className="text-text-secondary font-mono">
                        {key.key_prefix}...
                      </span>
                      {getApplicationName(key.environment_id) && (
                        <div className="flex items-center gap-1">
                          <Layers size={14} className="text-text-secondary" />
                          <span className="text-text-secondary">
                            {getApplicationName(key.environment_id)}
                          </span>
                        </div>
                      )}
                      <div className="flex items-center gap-1">
                        <div 
                          className="w-2 h-2 rounded-full"
                          style={{ backgroundColor: getEnvironmentColor(key.environment_id) }}
                        />
                        <span className="text-text-secondary">
                          {getEnvironmentName(key.environment_id)}
                        </span>
                      </div>
                    </div>

                    <div className="flex items-center gap-4 text-sm text-text-secondary">
                      <div className="flex items-center gap-1">
                        <Shield size={14} />
                        {key.permissions.join(', ')}
                      </div>
                      {key.last_used_at && (
                        <div className="flex items-center gap-1">
                          <Clock size={14} />
                          Last used: {new Date(key.last_used_at).toLocaleDateString()}
                        </div>
                      )}
                      {key.expires_at && (
                        <div className="flex items-center gap-1">
                          <Clock size={14} />
                          Expires: {new Date(key.expires_at).toLocaleDateString()}
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    {key.active && canRevokeAPIKey && (
                      <button
                        onClick={() => setKeyToRevoke(key)}
                        className="p-2 hover:bg-yellow-500/10 rounded-lg transition-colors"
                        title="Revoke"
                      >
                        <Ban size={18} className="text-yellow-400" />
                      </button>
                    )}
                    {canDeleteAPIKey && (
                      <button
                        onClick={() => setKeyToDelete(key)}
                        className="p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                        title="Delete"
                      >
                        <Trash2 size={18} className="text-red-400" />
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}

            {filteredKeys.length === 0 && !loading && (
              <div className="text-center py-12 text-text-secondary">
                {search ? 'No API keys found matching your search' : 'No API keys yet. Create your first API key!'}
              </div>
            )}
          </div>
        )}

        {/* Create Modal */}
        {showCreateModal && activeTenantId && (
          <CreateAPIKeyModal
            tenantId={activeTenantId}
            environments={environments}
            applications={applications}
            onClose={() => setShowCreateModal(false)}
            onCreated={(key) => {
              setNewKey(key);
              setShowCreateModal(false);
              fetchData();
            }}
          />
        )}

        {/* Show New Key Modal */}
        {newKey && (
          <NewKeyModal
            apiKey={newKey}
            environmentName={getEnvironmentName(newKey.environment_id)}
            onClose={() => setNewKey(null)}
          />
        )}

        {/* Delete Confirmation Modal */}
        <ConfirmDeleteModal
          isOpen={!!keyToDelete}
          onClose={() => setKeyToDelete(null)}
          onConfirm={handleDelete}
          title="Delete API Key"
          message={`Are you sure you want to delete the API key "${keyToDelete?.name}"? This action cannot be undone.`}
          itemName={keyToDelete?.name}
          confirmText="Delete"
        />

        {/* Revoke Confirmation Modal */}
        <ConfirmDeleteModal
          isOpen={!!keyToRevoke}
          onClose={() => setKeyToRevoke(null)}
          onConfirm={handleRevoke}
          title="Revoke API Key"
          message={`Are you sure you want to revoke the API key "${keyToRevoke?.name}"? The key will no longer be usable.`}
          itemName={keyToRevoke?.name}
          confirmText="Revoke"
        />
      </div>
    </div>
  );
}

interface CreateAPIKeyModalProps {
  tenantId: string;
  environments: Environment[];
  applications: Application[];
  onClose: () => void;
  onCreated: (key: APIKeyCreatedResponse) => void;
}

function CreateAPIKeyModal({ tenantId, environments, applications, onClose, onCreated }: CreateAPIKeyModalProps) {
  const [name, setName] = useState('');
  const [selectedAppId, setSelectedAppId] = useState('');
  const [environmentId, setEnvironmentId] = useState('');
  const [permissions, setPermissions] = useState<string[]>(['read']);
  const [expiresAt, setExpiresAt] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const filteredEnvironments = selectedAppId 
    ? environments.filter(e => e.application_id === selectedAppId)
    : [];

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const key = await apiKeysApi.create(tenantId, {
        name,
        environment_id: environmentId,
        permissions,
        expires_at: expiresAt || undefined,
      });
      onCreated(key);
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { error?: string } } };
      setError(axiosError.response?.data?.error || 'Failed to create API key');
    } finally {
      setLoading(false);
    }
  };

  const togglePermission = (perm: string) => {
    setPermissions(prev => 
      prev.includes(perm) 
        ? prev.filter(p => p !== perm)
        : [...prev, perm]
    );
  };

  return (
    <Modal
      title="Create API Key"
      onClose={onClose}
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            form="api-key-form"
            disabled={loading || permissions.length === 0}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            {loading && <Loader2 className="animate-spin" size={16} />}
            Create API Key
          </button>
        </>
      }
    >
      <form id="api-key-form" onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Production SDK Key"
            required
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Application
          </label>
          <select
            value={selectedAppId}
            onChange={(e) => {
              setSelectedAppId(e.target.value);
              setEnvironmentId('');
            }}
            required
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
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
            value={environmentId}
            onChange={(e) => setEnvironmentId(e.target.value)}
            required
            disabled={!selectedAppId}
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple disabled:opacity-50"
          >
            <option value="">Select environment...</option>
            {filteredEnvironments.map(env => (
              <option key={env.id} value={env.id}>{env.name}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Permissions
          </label>
          <div className="flex gap-2">
            {['read', 'write', 'admin'].map(perm => (
              <button
                key={perm}
                type="button"
                onClick={() => togglePermission(perm)}
                className={`px-3 py-2 rounded-lg text-sm transition-colors ${
                  permissions.includes(perm)
                    ? 'bg-accent-purple text-white'
                    : 'bg-bg-tertiary text-text-secondary hover:bg-bg-primary'
                }`}
              >
                {perm}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Expires At (optional)
          </label>
          <input
            type="date"
            value={expiresAt}
            onChange={(e) => setExpiresAt(e.target.value)}
            min={new Date().toISOString().split('T')[0]}
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
          />
        </div>
      </form>
    </Modal>
  );
}

interface NewKeyModalProps {
  apiKey: APIKeyCreatedResponse;
  environmentName: string;
  onClose: () => void;
}

function NewKeyModal({ apiKey, environmentName, onClose }: NewKeyModalProps) {
  const [showKey, setShowKey] = useState(false);
  const [copied, setCopied] = useState(false);

  const copyKey = () => {
    navigator.clipboard.writeText(apiKey.key);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Modal
      title={<span className="flex items-center gap-2"><Check className="text-green-400" />API Key Created</span>}
      maxWidth="lg"
      footer={
        <button
          onClick={onClose}
          className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors"
        >
          Done
        </button>
      }
    >
      <div className="space-y-4">
        <div className="p-4 bg-yellow-500/10 border border-yellow-500/30 rounded-lg">
          <p className="text-yellow-400 text-sm">
            Make sure to copy your API key now. You won't be able to see it again!
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            API Key
          </label>
          <div className="flex items-center gap-2">
            <input
              type={showKey ? 'text' : 'password'}
              value={apiKey.key}
              readOnly
              className="flex-1 px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary font-mono text-sm"
            />
            <button
              onClick={() => setShowKey(!showKey)}
              className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
            >
              {showKey ? <EyeOff size={20} className="text-text-secondary" /> : <Eye size={20} className="text-text-secondary" />}
            </button>
            <button
              onClick={copyKey}
              className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
            >
              {copied ? <Check size={20} className="text-green-400" /> : <Copy size={20} className="text-text-secondary" />}
            </button>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-text-secondary">Name:</span>
            <span className="ml-2 text-text-primary">{apiKey.name}</span>
          </div>
          <div>
            <span className="text-text-secondary">Environment:</span>
            <span className="ml-2 text-text-primary">{environmentName}</span>
          </div>
          <div>
            <span className="text-text-secondary">Permissions:</span>
            <span className="ml-2 text-text-primary">{apiKey.permissions.join(', ')}</span>
          </div>
          {apiKey.expires_at && (
            <div>
              <span className="text-text-secondary">Expires:</span>
              <span className="ml-2 text-text-primary">
                {new Date(apiKey.expires_at).toLocaleDateString()}
              </span>
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
}
