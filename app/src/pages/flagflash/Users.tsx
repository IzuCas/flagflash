import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  Users as UsersIcon, 
  Plus, 
  Search, 
  Edit2, 
  Trash2, 
  Loader2,
  AlertCircle,
  Shield,
  Mail,
  ArrowLeft,
  UserPlus,
  Copy,
  Check,
  Link as LinkIcon
} from 'lucide-react';
import { usersApi, tenantsApi } from '../../services/flagflash-api';
import { ConfirmDeleteModal, Modal } from '../../components';
import { useAuth } from '../../contexts/AuthContext';
import { usePermissions, ROLE_BADGES } from '../../hooks/usePermissions';
import type { UserWithMembership, Tenant, UserRole, InviteResponse } from '../../types/flagflash';

const ROLES: { value: UserRole; label: string; color: string }[] = [
  { value: 'owner', label: 'Owner', color: 'text-purple-400' },
  { value: 'admin', label: 'Admin', color: 'text-blue-400' },
  { value: 'member', label: 'Member', color: 'text-green-400' },
  { value: 'viewer', label: 'Viewer', color: 'text-gray-400' },
];

export default function UsersPage() {
  const { tenantId } = useParams<{ tenantId: string }>();
  const { user: authUser } = useAuth();
  const { canCreateUser, canUpdateUser, canManageRole } = usePermissions();
  const [users, setUsers] = useState<UserWithMembership[]>([]);
  const [tenant, setTenant] = useState<Tenant | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingUser, setEditingUser] = useState<UserWithMembership | null>(null);
  const [deletingUser, setDeletingUser] = useState<UserWithMembership | null>(null);

  const fetchData = useCallback(async () => {
    if (!tenantId) return;
    
    try {
      setLoading(true);
      const [usersResponse, tenantResponse] = await Promise.all([
        usersApi.list(tenantId),
        tenantsApi.get(tenantId),
      ]);
      setUsers(usersResponse.users || []);
      setTenant(tenantResponse);
      setError(null);
    } catch (err) {
      setError('Failed to load users');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [tenantId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filteredUsers = users.filter(u => 
    u.name.toLowerCase().includes(search.toLowerCase()) ||
    u.email.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = async () => {
    if (!deletingUser || !tenantId) return;
    try {
      await usersApi.delete(tenantId, deletingUser.id);
      setUsers(prev => prev.filter(u => u.id !== deletingUser.id));
      setDeletingUser(null);
    } catch (err) {
      console.error('Failed to delete user:', err);
    }
  };

  const getRoleInfo = (role: string) => {
    return ROLE_BADGES[role as UserRole] || ROLE_BADGES.member;
  };

  // Returns true only when the current user has permission to edit/delete the target
  const canManageUser = (targetId: string, targetRole: UserRole): boolean => {
    if (authUser?.id === targetId) return false; // cannot manage yourself — use Settings instead
    if (!canUpdateUser) return false; // must have update permission
    return canManageRole(targetRole);
  };

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto">
        {/* Breadcrumb */}
        <div className="mb-4">
          <Link 
            to="/tenants" 
            className="text-text-secondary hover:text-text-primary flex items-center gap-2 text-sm"
          >
            <ArrowLeft size={16} />
            Back to Tenants
          </Link>
        </div>

        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <UsersIcon className="text-accent-purple" />
              Users
            </h1>
            <p className="text-text-secondary mt-1">
              Manage users for {tenant?.name || 'tenant'}
            </p>
          </div>
          {canCreateUser && (
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors"
            >
              <Plus size={20} />
              Add User
            </button>
          )}
        </div>

        {/* Search */}
        <div className="relative mb-6">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary" size={20} />
          <input
            type="text"
            placeholder="Search users by name or email..."
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
          /* Users Table */
          <div className="bg-bg-secondary border border-border rounded-xl overflow-hidden">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border">
                  <th className="text-left px-6 py-4 text-text-secondary font-medium">User</th>
                  <th className="text-left px-6 py-4 text-text-secondary font-medium">Email</th>
                  <th className="text-left px-6 py-4 text-text-secondary font-medium">Role</th>
                  <th className="text-left px-6 py-4 text-text-secondary font-medium">Status</th>
                  <th className="text-right px-6 py-4 text-text-secondary font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredUsers.map(user => {
                  const roleInfo = getRoleInfo(user.role);
                  return (
                    <tr 
                      key={user.id}
                      className="border-b border-border/50 hover:bg-bg-tertiary/50 transition-colors"
                    >
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 rounded-full bg-accent-purple/20 flex items-center justify-center">
                            <span className="text-accent-purple font-medium">
                              {user.name.charAt(0).toUpperCase()}
                            </span>
                          </div>
                          <span className="font-medium text-text-primary">{user.name}</span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2 text-text-secondary">
                          <Mail size={16} />
                          {user.email}
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2">
                          <Shield size={16} className={roleInfo.color} />
                          <span className={`text-sm font-medium ${roleInfo.color}`}>
                            {roleInfo.label}
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                          user.active 
                            ? 'bg-green-500/20 text-green-400' 
                            : 'bg-gray-500/20 text-gray-400'
                        }`}>
                          {user.active ? 'Active' : 'Inactive'}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        {canManageUser(user.id, user.role) ? (
                          <div className="flex items-center justify-end gap-2">
                            <button
                              onClick={() => setEditingUser(user)}
                              className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
                              title="Edit user"
                            >
                              <Edit2 size={16} className="text-text-secondary" />
                            </button>
                            <button
                              onClick={() => setDeletingUser(user)}
                              className="p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                              title="Remove user"
                            >
                              <Trash2 size={16} className="text-red-400" />
                            </button>
                          </div>
                        ) : (
                          <div className="flex justify-end pr-2">
                            <span className="text-xs text-text-secondary italic" title="Protected role">—</span>
                          </div>
                        )}
                      </td>
                    </tr>
                  );
                })}

                {filteredUsers.length === 0 && !loading && (
                  <tr>
                    <td colSpan={5} className="text-center py-12 text-text-secondary">
                      {search 
                        ? 'No users found matching your search' 
                        : 'No users yet. Add your first user!'}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}

        {/* Create/Edit Modal */}
        {(showCreateModal || editingUser) && tenantId && (
          <UserModal
            tenantId={tenantId}
            user={editingUser}
            onClose={() => {
              setShowCreateModal(false);
              setEditingUser(null);
            }}
            onSave={fetchData}
          />
        )}

        {/* Delete Confirmation Modal */}
        <ConfirmDeleteModal
          isOpen={!!deletingUser}
          onClose={() => setDeletingUser(null)}
          onConfirm={handleDelete}
          title="Remove User"
          itemName={deletingUser?.name}
          message="This will remove the user from this tenant. They will lose access to this tenant's resources."
        />
      </div>
    </div>
  );
}

interface UserModalProps {
  tenantId: string;
  user: UserWithMembership | null;
  onClose: () => void;
  onSave: () => void;
}

function UserModal({ tenantId, user, onClose, onSave }: UserModalProps) {
  const [name, setName] = useState(user?.name || '');
  const [email, setEmail] = useState(user?.email || '');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<UserRole>(user?.role || 'member');
  const [active, setActive] = useState(user?.active ?? true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [mode, setMode] = useState<'create' | 'invite'>('create');
  const [inviteResult, setInviteResult] = useState<InviteResponse | null>(null);
  const [copied, setCopied] = useState(false);

  const isEditing = !!user;

  const handleCopyLink = async () => {
    if (!inviteResult?.invite_link) return;
    await navigator.clipboard.writeText(inviteResult.invite_link);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      if (isEditing) {
        await usersApi.update(tenantId, user.id, { name, role, active });
        onSave();
        onClose();
      } else if (mode === 'invite') {
        const result = await usersApi.invite(tenantId, { email, role });
        setInviteResult(result);
        onSave();
      } else {
        await usersApi.create(tenantId, { email, password, name, role });
        onSave();
        onClose();
      }
    } catch (err: unknown) {
      const error = err as Error;
      setError(error.message || 'Failed to save user');
    } finally {
      setLoading(false);
    }
  };

  // Show invite success view
  if (inviteResult) {
    return (
      <Modal
        title={
          <span className="flex items-center gap-2">
            <Mail size={20} className="text-green-400" />
            Convite Enviado
          </span>
        }
        onClose={onClose}
        footer={
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors"
          >
            Fechar
          </button>
        }
      >
        <div className="space-y-4">
          <div className="p-4 bg-green-500/10 border border-green-500/30 rounded-lg">
            <p className="text-green-400 text-sm">
              {inviteResult.email_sent
                ? `Email de convite enviado para ${inviteResult.email}`
                : `Convite gerado para ${inviteResult.email} (email não configurado)`}
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2 flex items-center gap-2">
              <LinkIcon size={14} />
              Link do convite
            </label>
            <div className="flex gap-2">
              <input
                type="text"
                readOnly
                value={inviteResult.invite_link}
                className="flex-1 px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary text-sm font-mono truncate focus:outline-none"
              />
              <button
                type="button"
                onClick={handleCopyLink}
                className={`px-3 py-2 rounded-lg border transition-colors flex items-center gap-1.5 text-sm ${
                  copied
                    ? 'bg-green-500/10 border-green-500/30 text-green-400'
                    : 'bg-bg-primary border-border text-text-secondary hover:text-text-primary hover:border-accent-purple'
                }`}
              >
                {copied ? <Check size={16} /> : <Copy size={16} />}
                {copied ? 'Copiado!' : 'Copiar'}
              </button>
            </div>
            <p className="mt-2 text-xs text-text-secondary">
              Compartilhe este link com o usuário. Expira em {new Date(inviteResult.expires_at).toLocaleDateString('pt-BR')}.
            </p>
          </div>
        </div>
      </Modal>
    );
  }

  return (
    <Modal
      title={
        isEditing ? (
          <span className="flex items-center gap-2">
            <Edit2 size={20} className="text-accent-purple" />
            Edit User
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <UserPlus size={20} className="text-accent-purple" />
            Add User
          </span>
        )
      }
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
            form="user-form"
            disabled={loading}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            {loading && <Loader2 className="animate-spin" size={16} />}
            {isEditing ? 'Save Changes' : (mode === 'invite' ? 'Invite User' : 'Create User')}
          </button>
        </>
      }
    >
      {!isEditing && (
        <div className="flex rounded-lg bg-bg-primary p-1 mb-4">
          <button
            type="button"
            onClick={() => setMode('create')}
            className={`flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-colors ${
              mode === 'create'
                ? 'bg-accent-purple text-white'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Create New User
          </button>
          <button
            type="button"
            onClick={() => setMode('invite')}
            className={`flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-colors ${
              mode === 'invite'
                ? 'bg-accent-purple text-white'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            Invite Existing
          </button>
        </div>
      )}
      <form id="user-form" onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

          {/* Name - only for create or edit */}
          {(mode === 'create' || isEditing) && (
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Name
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="John Doe"
                required
                className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
              />
            </div>
          )}

          {/* Email - only for create/invite */}
          {!isEditing && (
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="john@example.com"
                required
                className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
              />
            </div>
          )}

          {/* Password - only for create new user */}
          {!isEditing && mode === 'create' && (
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                minLength={8}
                className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
              />
              <p className="mt-1 text-xs text-text-secondary">
                Minimum 8 characters
              </p>
            </div>
          )}

          {/* Role */}
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Role
            </label>
            <select
              value={role}
              onChange={(e) => setRole(e.target.value as UserRole)}
              className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            >
              {ROLES.map(r => (
                <option key={r.value} value={r.value}>{r.label}</option>
              ))}
            </select>
          </div>

          {/* Active Toggle - only for edit */}
          {isEditing && (
            <div className="flex items-center justify-between">
              <label className="text-sm font-medium text-text-secondary">
                Active
              </label>
              <button
                type="button"
                onClick={() => setActive(!active)}
                className={`relative w-12 h-6 rounded-full transition-colors ${
                  active ? 'bg-green-500' : 'bg-bg-tertiary'
                }`}
              >
                <span
                  className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-transform ${
                    active ? 'left-7' : 'left-1'
                  }`}
                />
              </button>
            </div>
          )}

      </form>
    </Modal>
  );
}
