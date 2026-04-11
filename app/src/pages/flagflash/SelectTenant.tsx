import { useState, useEffect, useRef, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Building2, Zap, ChevronRight, Loader2, AlertCircle, LogOut, Shield, Plus, X } from 'lucide-react';
import { tenantsApi, flagflashAuthApi } from '../../services/flagflash-api';
import { useAuth, TenantWithRole } from '../../contexts/AuthContext';

export default function SelectTenantPage() {
  const { user, selectTenant, logout, setTenants } = useAuth();
  const navigate = useNavigate();
  const [tenants, setLocalTenants] = useState<TenantWithRole[]>([]);
  const [loading, setLoading] = useState(true);
  const [switching, setSwitching] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const fetchedRef = useRef(false);
  const [showCreateModal, setShowCreateModal] = useState(false);

  const loadTenants = useCallback(() => {
    setLoading(true);
    setError(null);
    tenantsApi.listMyTenants()
      .then(r => {
        const tenantList = r.tenants || [];
        setLocalTenants(tenantList);
        setTenants(tenantList);
      })
      .catch(() => setError('Falha ao carregar tenants'))
      .finally(() => setLoading(false));
  }, [setTenants]);

  useEffect(() => {
    // Fetch fresh tenants on mount only
    if (!fetchedRef.current) {
      fetchedRef.current = true;
      loadTenants();
    }
  }, [loadTenants]);

  const handleSelect = async (tenant: TenantWithRole) => {
    setSwitching(tenant.id);
    setError(null);
    
    try {
      // Call switch tenant to get a new JWT with the selected tenant
      const response = await flagflashAuthApi.switchTenant(tenant.id);
      
      // Update token in localStorage
      localStorage.setItem('flagflash_token', response.token);
      
      // Update context with selected tenant
      selectTenant({ 
        id: tenant.id, 
        name: tenant.name, 
        slug: tenant.slug,
        role: tenant.role 
      });
      
      navigate('/');
    } catch (err) {
      setError('Falha ao selecionar tenant');
      setSwitching(null);
    }
  };

  const getRoleBadge = (role: string) => {
    const colors: Record<string, string> = {
      owner: 'bg-amber-500/20 text-amber-400',
      admin: 'bg-blue-500/20 text-blue-400',
      editor: 'bg-green-500/20 text-green-400',
      viewer: 'bg-gray-500/20 text-gray-400',
    };
    return colors[role] || colors.viewer;
  };

  return (
    <div className="min-h-screen bg-bg-primary flex flex-col">
      {/* Header */}
      <header className="border-b border-border px-8 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-accent-purple to-pink-500 flex items-center justify-center">
            <Zap size={18} className="text-white" />
          </div>
          <span className="text-lg font-bold text-text-primary">FlagFlash</span>
        </div>
        <button
          onClick={logout}
          className="flex items-center gap-2 text-sm text-text-secondary hover:text-red-400 transition-colors"
        >
          <LogOut size={16} />
          Sair
        </button>
      </header>

      {/* Content */}
      <div className="flex-1 flex flex-col items-center justify-center p-8">
        <div className="w-full max-w-2xl">
          <h2 className="text-3xl font-bold text-text-primary mb-2">
            Bem-vindo, {user?.email}!
          </h2>
          <p className="text-text-secondary mb-8 text-lg">
            Selecione um tenant para continuar.
          </p>

          {error && (
            <div className="p-4 bg-red-500/10 border border-red-500/50 rounded-lg mb-6 flex items-center gap-3">
              <AlertCircle className="text-red-500 shrink-0" size={20} />
              <span className="text-red-400">{error}</span>
            </div>
          )}

          {loading ? (
            <div className="flex justify-center py-20">
              <Loader2 className="animate-spin text-accent-purple" size={32} />
            </div>
          ) : tenants.length === 0 ? (
            <div className="text-center py-20">
              <div className="w-16 h-16 rounded-full bg-bg-tertiary flex items-center justify-center mx-auto mb-4">
                <Building2 size={32} className="text-text-secondary" />
              </div>
              <h3 className="text-lg font-semibold text-text-primary mb-2">Nenhum tenant disponível</h3>
              <p className="text-text-secondary">Você não tem acesso a nenhum tenant. Contate um administrador.</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {tenants.map(tenant => (
                <button
                  key={tenant.id}
                  onClick={() => handleSelect(tenant)}
                  disabled={switching !== null}
                  className="w-full text-left p-5 bg-bg-secondary border border-border rounded-xl hover:border-accent-purple hover:bg-accent-purple/5 transition-all duration-200 group disabled:opacity-50 disabled:cursor-wait"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3 min-w-0">
                      <div className="w-10 h-10 rounded-lg bg-accent-purple/10 flex items-center justify-center shrink-0">
                        {switching === tenant.id ? (
                          <Loader2 size={20} className="text-accent-purple animate-spin" />
                        ) : (
                          <Building2 size={20} className="text-accent-purple" />
                        )}
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <p className="font-semibold text-text-primary truncate">{tenant.name}</p>
                          <span className={`px-2 py-0.5 rounded-full text-xs font-medium flex items-center gap-1 ${getRoleBadge(tenant.role)}`}>
                            <Shield size={10} />
                            {tenant.role}
                          </span>
                        </div>
                        <p className="text-sm text-text-secondary truncate">{tenant.slug}</p>
                      </div>
                    </div>
                    <ChevronRight
                      size={20}
                      className="text-text-secondary group-hover:text-accent-purple transition-colors shrink-0 ml-2"
                    />
                  </div>
                </button>
              ))}
              
              {/* Create New Tenant Button */}
              <button
                onClick={() => setShowCreateModal(true)}
                disabled={switching !== null}
                className="w-full text-left p-5 bg-bg-secondary border border-dashed border-border rounded-xl hover:border-accent-purple hover:bg-accent-purple/5 transition-all duration-200 group disabled:opacity-50"
              >
                <div className="flex items-center justify-center gap-3 py-2">
                  <div className="w-10 h-10 rounded-lg bg-accent-purple/10 flex items-center justify-center">
                    <Plus size={20} className="text-accent-purple" />
                  </div>
                  <p className="font-semibold text-text-secondary group-hover:text-accent-purple transition-colors">
                    Criar novo tenant
                  </p>
                </div>
              </button>
            </div>
          )}
        </div>
      </div>
      
      {/* Create Tenant Modal */}
      {showCreateModal && (
        <CreateTenantModal
          onClose={() => setShowCreateModal(false)}
          onCreated={() => {
            setShowCreateModal(false);
            fetchedRef.current = false;
            loadTenants();
          }}
        />
      )}
    </div>
  );
}

/* ─── Create Tenant Modal ─────────────────────────────────────── */
interface CreateTenantModalProps {
  onClose: () => void;
  onCreated: () => void;
}

function CreateTenantModal({ onClose, onCreated }: CreateTenantModalProps) {
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const generateSlug = (text: string) => {
    return text
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '');
  };

  const handleNameChange = (value: string) => {
    setName(value);
    setSlug(generateSlug(value));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!name.trim()) {
      setError('O nome é obrigatório');
      return;
    }

    if (!slug.trim()) {
      setError('O slug é obrigatório');
      return;
    }

    setLoading(true);
    try {
      await tenantsApi.create({ name: name.trim(), slug: slug.trim() });
      onCreated();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao criar tenant');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative bg-bg-secondary border border-border rounded-xl w-full max-w-md mx-4 shadow-2xl">
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <h3 className="text-lg font-semibold text-text-primary flex items-center gap-2">
            <Building2 size={20} className="text-accent-purple" />
            Criar Novo Tenant
          </h3>
          <button
            onClick={onClose}
            className="text-text-secondary hover:text-text-primary transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="flex items-center gap-2 bg-red-500/10 border border-red-500/30 text-red-400 rounded-lg px-4 py-3 text-sm">
              <AlertCircle size={16} className="shrink-0" />
              {error}
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-1.5">
              Nome do Tenant
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder="Minha Empresa"
              disabled={loading}
              className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors disabled:opacity-50"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-1.5">
              Slug (URL)
            </label>
            <input
              type="text"
              value={slug}
              onChange={(e) => setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
              placeholder="minha-empresa"
              disabled={loading}
              className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors disabled:opacity-50"
            />
            <p className="text-xs text-text-secondary mt-1">
              Apenas letras minúsculas, números e hífens
            </p>
          </div>

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="flex-1 px-4 py-2.5 border border-border rounded-lg text-text-secondary hover:bg-bg-tertiary transition-colors text-sm font-medium disabled:opacity-50"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={loading || !name.trim() || !slug.trim()}
              className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors text-sm font-medium disabled:opacity-50"
            >
              {loading && <Loader2 size={16} className="animate-spin" />}
              <Plus size={16} />
              Criar Tenant
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
