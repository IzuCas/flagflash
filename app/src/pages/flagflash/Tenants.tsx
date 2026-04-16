import { useState, useEffect, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { 
  Building2, 
  Plus, 
  Search, 
  Edit2, 
  Trash2, 
  ChevronRight,
  Loader2,
  AlertCircle,
  Shield
} from 'lucide-react';
import { tenantsApi, TenantWithRole } from '../../services/flagflash-api';
import { ConfirmDeleteModal, Modal } from '../../components';
import { useAuth } from '../../contexts/AuthContext';

export default function TenantsPage() {
  const { selectedTenant, clearTenant } = useAuth();
  const navigate = useNavigate();
  const [tenants, setTenants] = useState<TenantWithRole[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingTenant, setEditingTenant] = useState<TenantWithRole | null>(null);
  const [deletingTenant, setDeletingTenant] = useState<TenantWithRole | null>(null);

  const fetchTenants = useCallback(async () => {
    try {
      setLoading(true);
      const response = await tenantsApi.listMyTenants();
      setTenants(response.tenants || []);
      setError(null);
    } catch (err) {
      setError('Failed to load tenants');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTenants();
  }, [fetchTenants]);

  const filteredTenants = tenants.filter(t => 
    t.name.toLowerCase().includes(search.toLowerCase()) ||
    t.slug.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = async () => {
    if (!deletingTenant) return;
    await tenantsApi.delete(deletingTenant.id);
    const remaining = tenants.filter(t => t.id !== deletingTenant.id);
    setTenants(remaining);
    setDeletingTenant(null);

    // If deleted the current tenant or no tenants left, redirect to select tenant
    if (remaining.length === 0 || deletingTenant.id === selectedTenant?.id) {
      clearTenant();
      navigate('/select-tenant');
    }
  };

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <Building2 className="text-accent-purple" />
              Tenants
            </h1>
            <p className="text-text-secondary mt-1">
              Manage your organization tenants
            </p>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors"
          >
            <Plus size={20} />
            Create Tenant
          </button>
        </div>

        {/* Search */}
        <div className="relative mb-6">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary" size={20} />
          <input
            type="text"
            placeholder="Search tenants..."
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
          /* Tenants Grid */
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredTenants.map(tenant => (
              <div
                key={tenant.id}
                className="bg-bg-secondary border border-border rounded-xl p-5 hover:border-accent-purple/50 transition-colors group"
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="w-10 h-10 rounded-lg bg-accent-purple/20 flex items-center justify-center">
                    <Building2 className="text-accent-purple" size={20} />
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`px-2 py-1 rounded text-xs flex items-center gap-1 ${
                      tenant.role === 'owner'
                        ? 'bg-purple-500/20 text-purple-400'
                        : tenant.role === 'admin'
                        ? 'bg-blue-500/20 text-blue-400'
                        : 'bg-gray-500/20 text-gray-400'
                    }`}>
                      <Shield size={12} />
                      {tenant.role}
                    </span>
                  </div>
                </div>
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h3 className="font-semibold text-text-primary mb-1">{tenant.name}</h3>
                    <p className="text-sm text-text-secondary font-mono">{tenant.slug}</p>
                  </div>
                  {tenant.role === 'owner' && (
                    <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={() => setEditingTenant(tenant)}
                        className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
                      >
                        <Edit2 size={16} className="text-text-secondary" />
                      </button>
                      <button
                        onClick={() => setDeletingTenant(tenant)}
                        className="p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                      >
                        <Trash2 size={16} className="text-red-400" />
                      </button>
                    </div>
                  )}
                </div>
                
                <Link
                  to={`/tenants/${tenant.id}/applications`}
                  className="flex items-center justify-between text-sm text-accent-purple hover:text-accent-purple/80 transition-colors"
                >
                  View Applications
                  <ChevronRight size={16} />
                </Link>
              </div>
            ))}

            {filteredTenants.length === 0 && !loading && (
              <div className="col-span-full text-center py-12 text-text-secondary">
                {search ? 'No tenants found matching your search' : 'No tenants yet. Create your first tenant!'}
              </div>
            )}
          </div>
        )}

        {/* Create/Edit Modal */}
        {(showCreateModal || editingTenant) && (
          <TenantModal
            tenant={editingTenant}
            onClose={() => {
              setShowCreateModal(false);
              setEditingTenant(null);
            }}
            onSave={fetchTenants}
          />
        )}

        {/* Delete Confirmation Modal */}
        <ConfirmDeleteModal
          isOpen={!!deletingTenant}
          onClose={() => setDeletingTenant(null)}
          onConfirm={handleDelete}
          title="Delete Tenant"
          itemName={deletingTenant?.name}
          message={`This will permanently delete the tenant and all associated data. This action cannot be undone.`}
        />
      </div>
    </div>
  );
}

interface TenantModalProps {
  tenant: TenantWithRole | null;
  onClose: () => void;
  onSave: () => void;
}

function TenantModal({ tenant, onClose, onSave }: TenantModalProps) {
  const [name, setName] = useState(tenant?.name || '');
  const [slug, setSlug] = useState(tenant?.slug || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditing = !!tenant;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      if (isEditing) {
        await tenantsApi.update(tenant.id, { name });
      } else {
        await tenantsApi.create({ name, slug });
      }
      onSave();
      onClose();
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { error?: string } } };
      setError(axiosError.response?.data?.error || 'Failed to save tenant');
    } finally {
      setLoading(false);
    }
  };

  // Auto-generate slug from name
  useEffect(() => {
    if (!isEditing && name) {
      setSlug(name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, ''));
    }
  }, [name, isEditing]);

  return (
    <Modal
      title={isEditing ? 'Edit Tenant' : 'Create Tenant'}
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
            form="tenant-form"
            disabled={loading}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            {loading && <Loader2 className="animate-spin" size={16} />}
            {isEditing ? 'Save Changes' : 'Create Tenant'}
          </button>
        </>
      }
    >
      <form id="tenant-form" onSubmit={handleSubmit} className="space-y-4">
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
            placeholder="My Organization"
            required
            className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
          />
        </div>

        {!isEditing && (
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Slug
            </label>
            <input
              type="text"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder="my-organization"
              required
              pattern="^[a-z0-9-]+$"
              className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
            />
            <p className="mt-1 text-xs text-text-secondary">
              Lowercase letters, numbers, and hyphens only
            </p>
          </div>
        )}
      </form>
    </Modal>
  );
}
