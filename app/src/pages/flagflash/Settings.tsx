import { useState, useEffect } from 'react';
import {
  User,
  Building2,
  KeyRound,
  Eye,
  EyeOff,
  AlertCircle,
  CheckCircle2,
  Loader2,
  ShieldCheck,
  Shield,
  Lock,
  Save,
} from 'lucide-react';
import { authApi } from '../../services/api';
import { tenantsApi, flagflashAuthApi } from '../../services/flagflash-api';
import { useAuth } from '../../contexts/AuthContext';

type Section = 'profile' | 'password' | 'tenant';

export default function SettingsPage() {
  const [section, setSection] = useState<Section>('profile');

  const navItem = (id: Section, label: string, icon: React.ReactNode) => (
    <button
      onClick={() => setSection(id)}
      className={`flex items-center gap-3 w-full px-4 py-3 rounded-lg text-sm font-medium transition-all duration-200 text-left ${
        section === id
          ? 'bg-accent-purple/10 text-accent-purple border-l-2 border-accent-purple'
          : 'text-text-secondary hover:bg-bg-tertiary hover:text-text-primary'
      }`}
    >
      {icon}
      {label}
    </button>
  );

  return (
    <div className="p-6 min-w-0 overflow-hidden">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-2xl font-bold text-text-primary mb-6">Configurações</h1>

        <div className="flex gap-6 items-start">
          {/* Sidebar menu */}
          <div className="w-48 shrink-0 bg-bg-secondary border border-border rounded-xl p-2 space-y-1">
            {navItem('profile', 'Dados Pessoais', <User size={18} />)}
            {navItem('password', 'Senha', <Lock size={18} />)}
            {navItem('tenant', 'Tenant', <Building2 size={18} />)}
          </div>

          {/* Content */}
          <div className="flex-1 min-w-0">
            {section === 'profile' && <ProfileSection />}
            {section === 'password' && <PasswordSection />}
            {section === 'tenant' && <TenantSection />}
          </div>
        </div>
      </div>
    </div>
  );
}

/* ─── Profile Section ─────────────────────────────────────────── */
function ProfileSection() {
  const { user, updateUser } = useAuth();
  const [email, setEmail] = useState(user?.email ?? '');
  const [name, setName] = useState(user?.name ?? '');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(true);

  // Fetch fresh profile data on mount
  useEffect(() => {
    flagflashAuthApi.getProfile()
      .then(profile => {
        setEmail(profile.email);
        setName(profile.name);
        updateUser({ id: profile.id, email: profile.email, name: profile.name });
      })
      .catch(() => {
        // Use cached data if fetch fails
        setEmail(user?.email ?? '');
        setName(user?.name ?? '');
      })
      .finally(() => setFetching(false));
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    const trimmed = name.trim();
    if (trimmed.length === 0) {
      setError('O nome não pode ser vazio');
      return;
    }

    setLoading(true);
    try {
      const updated = await flagflashAuthApi.updateProfile({ name: trimmed });
      updateUser({ name: updated.name });
      setSuccess('Dados atualizados com sucesso!');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao atualizar dados');
    } finally {
      setLoading(false);
    }
  };

  if (fetching) {
    return (
      <div className="bg-bg-secondary border border-border rounded-xl p-6 flex items-center justify-center">
        <Loader2 size={24} className="animate-spin text-accent-purple" />
      </div>
    );
  }

  return (
    <div className="bg-bg-secondary border border-border rounded-xl">
      <div className="px-6 py-4 border-b border-border flex items-center gap-3">
        <User size={20} className="text-accent-purple" />
        <h2 className="text-lg font-semibold text-text-primary">Dados Pessoais</h2>
      </div>

      <form onSubmit={handleSubmit} className="p-6 space-y-5">
        {error && (
          <div className="flex items-center gap-2 bg-red-500/10 border border-red-500/30 text-red-400 rounded-lg px-4 py-3 text-sm">
            <AlertCircle size={16} className="shrink-0" />
            {error}
          </div>
        )}
        {success && (
          <div className="flex items-center gap-2 bg-green-500/10 border border-green-500/30 text-green-400 rounded-lg px-4 py-3 text-sm">
            <CheckCircle2 size={16} className="shrink-0" />
            {success}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Email</label>
          <div className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm opacity-75">
            {email || '-'}
          </div>
          <p className="text-xs text-text-secondary mt-1.5">
            O email não pode ser alterado.
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Nome</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            disabled={loading}
            className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors disabled:opacity-50"
            placeholder="Seu nome"
          />
        </div>

        <div className="flex justify-end pt-2">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-5 py-2.5 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 text-sm font-medium"
          >
            {loading && <Loader2 size={16} className="animate-spin" />}
            <Save size={16} />
            Salvar
          </button>
        </div>
      </form>
    </div>
  );
}

/* ─── Password Section ─────────────────────────────────────────── */
function PasswordSection() {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showCurrent, setShowCurrent] = useState(false);
  const [showNew, setShowNew] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  const passwordStrength = (p: string) => {
    let score = 0;
    if (p.length >= 8) score++;
    if (p.length >= 12) score++;
    if (/[A-Z]/.test(p)) score++;
    if (/[0-9]/.test(p)) score++;
    if (/[^A-Za-z0-9]/.test(p)) score++;
    return score;
  };

  const strengthLabel = ['Muito fraca', 'Fraca', 'Razoável', 'Boa', 'Forte', 'Muito forte'];
  const strengthColor = ['bg-red-500', 'bg-orange-500', 'bg-yellow-500', 'bg-lime-500', 'bg-green-500', 'bg-emerald-500'];
  const score = passwordStrength(newPassword);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (newPassword !== confirmPassword) { 
      setError('As novas senhas não coincidem'); 
      return; 
    }
    if (newPassword.length < 8) { 
      setError('A nova senha precisa ter pelo menos 8 caracteres'); 
      return; 
    }

    setLoading(true);
    try {
      await authApi.changePassword(currentPassword, newPassword);
      setSuccess('Senha alterada com sucesso!');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao alterar senha');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-bg-secondary border border-border rounded-xl">
      <div className="px-6 py-4 border-b border-border flex items-center gap-3">
        <ShieldCheck size={20} className="text-accent-purple" />
        <h2 className="text-lg font-semibold text-text-primary">Alterar Senha</h2>
      </div>

      <form onSubmit={handleSubmit} className="p-6 space-y-5">
        {error && (
          <div className="flex items-center gap-2 bg-red-500/10 border border-red-500/30 text-red-400 rounded-lg px-4 py-3 text-sm">
            <AlertCircle size={16} className="shrink-0" />
            {error}
          </div>
        )}
        {success && (
          <div className="flex items-center gap-2 bg-green-500/10 border border-green-500/30 text-green-400 rounded-lg px-4 py-3 text-sm">
            <CheckCircle2 size={16} className="shrink-0" />
            {success}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Senha atual</label>
          <div className="relative">
            <input
              type={showCurrent ? 'text' : 'password'}
              autoComplete="current-password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              required
              disabled={loading}
              className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 pr-10 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors disabled:opacity-50"
              placeholder="Digite sua senha atual"
            />
            <button
              type="button"
              onClick={() => setShowCurrent(v => !v)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary hover:text-text-primary"
            >
              {showCurrent ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Nova senha</label>
          <div className="relative">
            <input
              type={showNew ? 'text' : 'password'}
              autoComplete="new-password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              required
              disabled={loading}
              className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 pr-10 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors disabled:opacity-50"
              placeholder="Mínimo 8 caracteres"
            />
            <button
              type="button"
              onClick={() => setShowNew(v => !v)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-text-secondary hover:text-text-primary"
            >
              {showNew ? <EyeOff size={16} /> : <Eye size={16} />}
            </button>
          </div>
          {newPassword && (
            <div className="mt-2 space-y-1">
              <div className="flex gap-1">
                {[...Array(5)].map((_, i) => (
                  <div
                    key={i}
                    className={`h-1 flex-1 rounded-full transition-colors ${i < score ? strengthColor[score] : 'bg-bg-tertiary'}`}
                  />
                ))}
              </div>
              <p className="text-xs text-text-secondary">{strengthLabel[score]}</p>
            </div>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Confirmar nova senha</label>
          <input
            type="password"
            autoComplete="new-password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            required
            disabled={loading}
            className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors disabled:opacity-50"
            placeholder="Repita a nova senha"
          />
        </div>

        <div className="flex justify-end pt-2">
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-5 py-2.5 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 text-sm font-medium"
          >
            {loading && <Loader2 size={16} className="animate-spin" />}
            <KeyRound size={16} />
            Alterar senha
          </button>
        </div>
      </form>
    </div>
  );
}

/* ─── Tenant Section ───────────────────────────────────────────── */
function TenantSection() {
  const { selectedTenant, selectTenant } = useAuth();

  const [name, setName] = useState(selectedTenant?.name ?? '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const isOwner = selectedTenant?.role === 'owner';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTenant || !isOwner) return;
    setError('');
    setSuccess('');

    const trimmed = name.trim();
    if (!trimmed) { setError('O nome do tenant não pode ser vazio'); return; }

    setLoading(true);
    try {
      const updated = await tenantsApi.update(selectedTenant.id, { name: trimmed });
      selectTenant({ id: updated.id, name: updated.name, slug: updated.slug, role: selectedTenant.role });
      setSuccess('Tenant atualizado com sucesso!');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao atualizar tenant');
    } finally {
      setLoading(false);
    }
  };

  if (!selectedTenant) {
    return (
      <div className="bg-bg-secondary border border-border rounded-xl p-6 text-text-secondary text-sm">
        Nenhum tenant selecionado.
      </div>
    );
  }

  return (
    <div className="bg-bg-secondary border border-border rounded-xl">
      <div className="px-6 py-4 border-b border-border flex items-center gap-3">
        <Building2 size={20} className="text-accent-purple" />
        <h2 className="text-lg font-semibold text-text-primary">Configurações do Tenant</h2>
      </div>

      <form onSubmit={handleSubmit} className="p-6 space-y-5">
        {!isOwner && (
          <div className="flex items-center gap-2 bg-yellow-500/10 border border-yellow-500/30 text-yellow-400 rounded-lg px-4 py-3 text-sm">
            <Shield size={16} className="shrink-0" />
            Apenas owners podem editar as configurações do tenant.
          </div>
        )}
        {error && (
          <div className="flex items-center gap-2 bg-red-500/10 border border-red-500/30 text-red-400 rounded-lg px-4 py-3 text-sm">
            <AlertCircle size={16} className="shrink-0" />
            {error}
          </div>
        )}
        {success && (
          <div className="flex items-center gap-2 bg-green-500/10 border border-green-500/30 text-green-400 rounded-lg px-4 py-3 text-sm">
            <CheckCircle2 size={16} className="shrink-0" />
            {success}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Nome do Tenant</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            disabled={loading || !isOwner}
            className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            placeholder="Nome do tenant"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Slug</label>
          <input
            type="text"
            value={selectedTenant.slug}
            disabled
            className="w-full bg-bg-primary border border-border rounded-lg px-3 py-2.5 text-text-secondary text-sm opacity-60 cursor-not-allowed"
          />
          <p className="text-xs text-text-secondary mt-1">O slug não pode ser alterado após a criação.</p>
        </div>

        {isOwner && (
          <div className="flex justify-end pt-2">
            <button
              type="submit"
              disabled={loading}
              className="flex items-center gap-2 px-5 py-2.5 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 text-sm font-medium"
            >
              {loading && <Loader2 size={16} className="animate-spin" />}
              <Building2 size={16} />
              Salvar tenant
            </button>
          </div>
        )}
      </form>
    </div>
  );
}
