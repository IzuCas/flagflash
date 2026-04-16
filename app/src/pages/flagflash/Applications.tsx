import { useState, useEffect, useCallback } from 'react';
import { useParams, Link, useSearchParams } from 'react-router-dom';
import { 
  Layers, 
  Plus, 
  Search, 
  Edit2, 
  Trash2, 
  ChevronRight,
  ChevronLeft,
  Loader2,
  AlertCircle
} from 'lucide-react';
import { applicationsApi, tenantsApi } from '../../services/flagflash-api';
import { ConfirmDeleteModal, Modal } from '../../components';
import type { Application, Tenant } from '../../types/flagflash';

export default function ApplicationsPage() {
  const { tenantId } = useParams<{ tenantId: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const [applications, setApplications] = useState<Application[]>([]);
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(() => searchParams.get('new') === 'true');
  const [editingApp, setEditingApp] = useState<Application | null>(null);
  const [deletingApp, setDeletingApp] = useState<Application | null>(null);

  // Clear ?new param after modal opens
  useEffect(() => {
    if (searchParams.get('new') === 'true') {
      setShowCreateModal(true);
      setSearchParams({}, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  const fetchData = useCallback(async () => {
    if (!tenantId) return;
    
    try {
      setLoading(true);
      const [appsResponse, tenantData] = await Promise.all([
        applicationsApi.list(tenantId),
        tenantsApi.get(tenantId)
      ]);
      setApplications(appsResponse.applications || []);
      setTenant(tenantData);
      setError(null);
    } catch (err) {
      setError('Failed to load applications');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [tenantId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filteredApps = applications.filter(app => 
    app.name.toLowerCase().includes(search.toLowerCase()) ||
    app.description?.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = async () => {
    if (!deletingApp || !tenantId) return;
    await applicationsApi.delete(tenantId, deletingApp.id);
    setApplications(prev => prev.filter(a => a.id !== deletingApp.id));
    setDeletingApp(null);
    window.dispatchEvent(new CustomEvent('appschanged'));
  };

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          <Link
            to="/tenants"
            className="p-2 hover:bg-bg-secondary rounded-lg transition-colors"
          >
            <ChevronLeft size={20} className="text-text-secondary" />
          </Link>
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <Layers className="text-accent-purple" />
              Applications
            </h1>
            {tenant && (
              <p className="text-text-secondary mt-1">
                Tenant: {tenant.name}
              </p>
            )}
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors"
          >
            <Plus size={20} />
            Create Application
          </button>
        </div>

        {/* Search */}
        <div className="relative mb-6">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary" size={20} />
          <input
            type="text"
            placeholder="Search applications..."
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
          /* Applications Grid */
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredApps.map(app => (
              <div
                key={app.id}
                className="bg-bg-secondary border border-border rounded-xl p-5 hover:border-accent-purple/50 transition-colors group"
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="w-10 h-10 rounded-lg bg-accent-blue/20 flex items-center justify-center">
                    <Layers className="text-accent-blue" size={20} />
                  </div>
                  <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button
                      onClick={() => setEditingApp(app)}
                      className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
                    >
                      <Edit2 size={16} className="text-text-secondary" />
                    </button>
                    <button
                      onClick={() => setDeletingApp(app)}
                      className="p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                    >
                      <Trash2 size={16} className="text-red-400" />
                    </button>
                  </div>
                </div>
                
                <h3 className="font-semibold text-text-primary mb-1">{app.name}</h3>
                <p className="text-sm text-text-secondary mb-4 line-clamp-2">
                  {app.description || 'No description'}
                </p>
                
                <Link
                  to={`/tenants/${tenantId}/applications/${app.id}/environments`}
                  className="flex items-center justify-between text-sm text-accent-purple hover:text-accent-purple/80 transition-colors"
                >
                  View Environments
                  <ChevronRight size={16} />
                </Link>
              </div>
            ))}

            {filteredApps.length === 0 && !loading && (
              <div className="col-span-full text-center py-12 text-text-secondary">
                {search ? 'No applications found matching your search' : 'No applications yet. Create your first application!'}
              </div>
            )}
          </div>
        )}

        {/* Create/Edit Modal */}
        {(showCreateModal || editingApp) && tenantId && (
          <ApplicationModal
            app={editingApp}
            tenantId={tenantId}
            onClose={() => {
              setShowCreateModal(false);
              setEditingApp(null);
            }}
            onSave={fetchData}
          />
        )}

        {/* Delete Confirmation Modal */}
        <ConfirmDeleteModal
          isOpen={!!deletingApp}
          onClose={() => setDeletingApp(null)}
          onConfirm={handleDelete}
          title="Delete Application"
          itemName={deletingApp?.name}
          message="This will permanently delete the application and all its environments and feature flags. This action cannot be undone."
        />
      </div>
    </div>
  );
}

interface ApplicationModalProps {
  app: Application | null;
  tenantId: string;
  onClose: () => void;
  onSave: () => void;
}

function ApplicationModal({ app, tenantId, onClose, onSave }: ApplicationModalProps) {
  const [name, setName] = useState(app?.name || '');
  const [slug, setSlug] = useState(app?.slug || '');
  const [description, setDescription] = useState(app?.description || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditing = !!app;

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
        await applicationsApi.update(tenantId, app.id, { name, slug, description });
      } else {
        await applicationsApi.create(tenantId, { name, slug, description });
      }
      window.dispatchEvent(new CustomEvent('appschanged'));
      onSave();
      onClose();
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { error?: string } } };
      setError(axiosError.response?.data?.error || 'Failed to save application');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={isEditing ? 'Edit Application' : 'Create Application'}
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
            form="application-form"
            disabled={loading}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            {loading && <Loader2 className="animate-spin" size={16} />}
            {isEditing ? 'Save Changes' : 'Create Application'}
          </button>
        </>
      }
    >
      <form id="application-form" onSubmit={handleSubmit} className="space-y-4">
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
            placeholder="My Web App"
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
            placeholder="my-web-app"
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
            placeholder="Main web application for customers"
            rows={3}
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple resize-none"
          />
        </div>
      </form>
    </Modal>
  );
}
