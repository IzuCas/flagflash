import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  Server, 
  Plus, 
  Search, 
  Edit2, 
  Trash2, 
  ChevronRight,
  ChevronLeft,
  Flag,
  Loader2,
  AlertCircle
} from 'lucide-react';
import { environmentsApi, applicationsApi } from '../../services/flagflash-api';
import { ConfirmDeleteModal, Modal } from '../../components';
import type { Environment, Application } from '../../types/flagflash';

const DEFAULT_COLORS = [
  '#22c55e', // Green (Production)
  '#f59e0b', // Yellow (Staging)
  '#3b82f6', // Blue (Development)
  '#8b5cf6', // Purple
  '#ec4899', // Pink
  '#ef4444', // Red
];

export default function EnvironmentsPage() {
  const { tenantId, appId } = useParams<{ tenantId: string; appId: string }>();
  const [environments, setEnvironments] = useState<Environment[]>([]);
  const [application, setApplication] = useState<Application | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingEnv, setEditingEnv] = useState<Environment | null>(null);
  const [deletingEnv, setDeletingEnv] = useState<Environment | null>(null);

  const fetchData = useCallback(async () => {
    if (!tenantId || !appId) return;
    
    try {
      setLoading(true);
      const [envsResponse, appData] = await Promise.all([
        environmentsApi.list(tenantId, appId),
        applicationsApi.get(tenantId, appId)
      ]);
      setEnvironments(envsResponse.environments || []);
      setApplication(appData);
      setError(null);
    } catch (err) {
      setError('Failed to load environments');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [tenantId, appId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filteredEnvs = environments.filter(env => 
    env.name.toLowerCase().includes(search.toLowerCase()) ||
    env.description?.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = async () => {
    if (!deletingEnv || !tenantId || !appId) return;
    await environmentsApi.delete(tenantId, appId, deletingEnv.id);
    setEnvironments(prev => prev.filter(e => e.id !== deletingEnv.id));
    setDeletingEnv(null);
  };

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          <Link
            to={`/tenants/${tenantId}/applications`}
            className="p-2 hover:bg-bg-secondary rounded-lg transition-colors"
          >
            <ChevronLeft size={20} className="text-text-secondary" />
          </Link>
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <Server className="text-accent-purple" />
              Environments
            </h1>
            {application && (
              <p className="text-text-secondary mt-1">
                Application: {application.name}
              </p>
            )}
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors"
          >
            <Plus size={20} />
            Create Environment
          </button>
        </div>

        {/* Search */}
        <div className="relative mb-6">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary" size={20} />
          <input
            type="text"
            placeholder="Search environments..."
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
          /* Environments Grid */
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredEnvs.map(env => (
              <div
                key={env.id}
                className="bg-bg-secondary border border-border rounded-xl p-5 hover:border-accent-purple/50 transition-colors group"
              >
                <div className="flex items-start justify-between mb-4">
                  <div 
                    className="w-10 h-10 rounded-lg flex items-center justify-center"
                    style={{ backgroundColor: `${env.color || '#8b5cf6'}20` }}
                  >
                    <Server style={{ color: env.color || '#8b5cf6' }} size={20} />
                  </div>
                  <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={() => setEditingEnv(env)}
                      className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
                    >
                      <Edit2 size={16} className="text-text-secondary" />
                    </button>
                    <button
                      onClick={() => setDeletingEnv(env)}
                      className="p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                    >
                      <Trash2 size={16} className="text-red-400" />
                    </button>
                  </div>
                </div>
                
                <div className="flex items-center gap-2 mb-1">
                  <div 
                    className="w-3 h-3 rounded-full" 
                    style={{ backgroundColor: env.color || '#8b5cf6' }}
                  />
                  <h3 className="font-semibold text-text-primary">{env.name}</h3>
                </div>
                <p className="text-sm text-text-secondary mb-4 line-clamp-2">
                  {env.description || 'No description'}
                </p>
                
                <Link
                  to={`/tenants/${tenantId}/applications/${appId}/environments/${env.id}/flags`}
                  className="flex items-center justify-between text-sm text-accent-purple hover:text-accent-purple/80 transition-colors"
                >
                  <span className="flex items-center gap-2">
                    <Flag size={14} />
                    Manage Flags
                  </span>
                  <ChevronRight size={16} />
                </Link>
              </div>
            ))}

            {filteredEnvs.length === 0 && !loading && (
              <div className="col-span-full text-center py-12 text-text-secondary">
                {search ? 'No environments found matching your search' : 'No environments yet. Create your first environment!'}
              </div>
            )}
          </div>
        )}

        {/* Create/Edit Modal */}
        {(showCreateModal || editingEnv) && tenantId && appId && (
          <EnvironmentModal
            env={editingEnv}
            tenantId={tenantId}
            appId={appId}
            onClose={() => {
              setShowCreateModal(false);
              setEditingEnv(null);
            }}
            onSave={fetchData}
          />
        )}

        {/* Delete Confirmation Modal */}
        <ConfirmDeleteModal
          isOpen={!!deletingEnv}
          onClose={() => setDeletingEnv(null)}
          onConfirm={handleDelete}
          title="Delete Environment"
          itemName={deletingEnv?.name}
          message="This will permanently delete the environment and all its feature flags. This action cannot be undone."
        />
      </div>
    </div>
  );
}

interface EnvironmentModalProps {
  env: Environment | null;
  tenantId: string;
  appId: string;
  onClose: () => void;
  onSave: () => void;
}

function EnvironmentModal({ env, tenantId, appId, onClose, onSave }: EnvironmentModalProps) {
  const [name, setName] = useState(env?.name || '');
  const [slug, setSlug] = useState(env?.slug || '');
  const [description, setDescription] = useState(env?.description || '');
  const [color, setColor] = useState(env?.color || DEFAULT_COLORS[0]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditing = !!env;

  // Auto-generate slug from name
  const handleNameChange = (value: string) => {
    setName(value);
    if (!isEditing && !slug) {
      setSlug(value.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, ''));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      if (isEditing) {
        await environmentsApi.update(tenantId, appId, env.id, { name, slug, description, color });
      } else {
        await environmentsApi.create(tenantId, appId, { name, slug, description, color });
      }
      onSave();
      onClose();
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { error?: string } } };
      setError(axiosError.response?.data?.error || 'Failed to save environment');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={isEditing ? 'Edit Environment' : 'Create Environment'}
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
            form="environment-form"
            disabled={loading}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            {loading && <Loader2 className="animate-spin" size={16} />}
            {isEditing ? 'Save Changes' : 'Create Environment'}
          </button>
        </>
      }
    >
      <form id="environment-form" onSubmit={handleSubmit} className="space-y-4">
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
            onChange={(e) => handleNameChange(e.target.value)}
            placeholder="Production"
            required
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Slug
          </label>
          <input
            type="text"
            value={slug}
            onChange={(e) => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
            placeholder="production"
            required
            pattern="^[a-z0-9-]+$"
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple font-mono"
          />
          <p className="text-xs text-text-secondary mt-1">
            Only lowercase letters, numbers and hyphens
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Description (optional)
          </label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Production environment for live users"
            rows={3}
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple resize-none"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Color
          </label>
          <div className="flex items-center gap-3">
            <div className="flex gap-2">
              {DEFAULT_COLORS.map(c => (
                <button
                  key={c}
                  type="button"
                  onClick={() => setColor(c)}
                  className={`w-8 h-8 rounded-full transition-transform ${
                    color === c ? 'ring-2 ring-white ring-offset-2 ring-offset-bg-secondary scale-110' : ''
                  }`}
                  style={{ backgroundColor: c }}
                />
              ))}
            </div>
            <input
              type="color"
              value={color}
              onChange={(e) => setColor(e.target.value)}
              className="w-8 h-8 rounded cursor-pointer"
            />
          </div>
        </div>
      </form>
    </Modal>
  );
}
