import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { Zap, Loader2, AlertCircle, CheckCircle2, UserPlus } from 'lucide-react';
import { inviteApi } from '../services/flagflash-api';
import type { InviteDetails } from '../types/flagflash';

export default function AcceptInvitePage() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token') || '';

  const [invite, setInvite] = useState<InviteDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!token) {
      setError('No invite token provided');
      setLoading(false);
      return;
    }

    inviteApi.validate(token)
      .then(data => {
        setInvite(data);
        setError('');
      })
      .catch(err => {
        setError(err.message || 'Invalid or expired invitation');
      })
      .finally(() => setLoading(false));
  }, [token]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!invite?.user_exists) {
      if (!name.trim()) {
        setError('Name is required');
        return;
      }
      if (password.length < 8) {
        setError('Password must be at least 8 characters');
        return;
      }
      if (password !== confirmPassword) {
        setError('Passwords do not match');
        return;
      }
    }

    setSubmitting(true);
    try {
      await inviteApi.accept({
        token,
        name: invite?.user_exists ? undefined : name,
        password: invite?.user_exists ? undefined : password,
      });
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to accept invitation');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-bg-primary flex items-center justify-center">
        <Loader2 className="animate-spin text-accent-purple" size={40} />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-bg-primary flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="flex flex-col items-center mb-8">
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-fuchsia-500 to-violet-600 flex items-center justify-center mb-4 shadow-lg shadow-fuchsia-500/30">
            <Zap size={32} className="text-white" fill="white" />
          </div>
          <h1 className="text-2xl font-bold text-text-primary">FlagFlash</h1>
          <p className="text-text-secondary text-sm mt-1">Feature Flags Platform</p>
        </div>

        <div className="bg-bg-secondary border border-border rounded-xl p-6 shadow-xl">
          {success ? (
            <div className="text-center space-y-4">
              <div className="w-16 h-16 rounded-full bg-green-500/20 flex items-center justify-center mx-auto">
                <CheckCircle2 size={32} className="text-green-400" />
              </div>
              <h2 className="text-xl font-bold text-text-primary">Invitation Accepted!</h2>
              <p className="text-text-secondary">
                You now have access to <strong className="text-text-primary">{invite?.tenant_name}</strong>.
              </p>
              <button
                onClick={() => navigate('/login')}
                className="w-full mt-4 bg-gradient-to-r from-fuchsia-500 to-violet-600 hover:from-fuchsia-400 hover:to-violet-500 text-white font-medium rounded-lg px-4 py-2.5 text-sm transition-all duration-200 shadow-md shadow-fuchsia-500/20"
              >
                Go to Login
              </button>
            </div>
          ) : error && !invite ? (
            <div className="text-center space-y-4">
              <div className="w-16 h-16 rounded-full bg-red-500/20 flex items-center justify-center mx-auto">
                <AlertCircle size={32} className="text-red-400" />
              </div>
              <h2 className="text-xl font-bold text-text-primary">Invalid Invitation</h2>
              <p className="text-text-secondary">{error}</p>
              <button
                onClick={() => navigate('/login')}
                className="w-full mt-4 bg-bg-tertiary text-text-primary font-medium rounded-lg px-4 py-2.5 text-sm transition-colors hover:bg-bg-tertiary/80"
              >
                Go to Login
              </button>
            </div>
          ) : invite ? (
            <>
              <div className="flex items-center gap-3 mb-6">
                <UserPlus size={24} className="text-accent-purple" />
                <div>
                  <h2 className="text-lg font-bold text-text-primary">Accept Invitation</h2>
                  <p className="text-sm text-text-secondary">
                    Join <strong className="text-text-primary">{invite.tenant_name}</strong> as <strong className="text-accent-purple">{invite.role}</strong>
                  </p>
                </div>
              </div>

              <div className="mb-4 p-3 bg-accent-purple/10 border border-accent-purple/20 rounded-lg">
                <p className="text-sm text-text-secondary">
                  Invited: <strong className="text-text-primary">{invite.email}</strong>
                </p>
              </div>

              {error && (
                <div className="mb-4 flex items-center gap-2 bg-red-500/10 border border-red-500/30 text-red-400 rounded-lg px-4 py-3 text-sm">
                  <AlertCircle size={16} className="shrink-0" />
                  <span>{error}</span>
                </div>
              )}

              <form onSubmit={handleSubmit} className="space-y-4">
                {invite.user_exists ? (
                  <p className="text-sm text-text-secondary">
                    Your account already exists. Click below to join this tenant.
                  </p>
                ) : (
                  <>
                    <div>
                      <label className="block text-sm font-medium text-text-secondary mb-1.5">
                        Name
                      </label>
                      <input
                        type="text"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        required
                        placeholder="Your full name"
                        className="w-full bg-bg-tertiary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-text-secondary mb-1.5">
                        Password
                      </label>
                      <input
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                        minLength={8}
                        placeholder="Minimum 8 characters"
                        className="w-full bg-bg-tertiary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-text-secondary mb-1.5">
                        Confirm Password
                      </label>
                      <input
                        type="password"
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        required
                        minLength={8}
                        placeholder="Repeat your password"
                        className="w-full bg-bg-tertiary border border-border rounded-lg px-3 py-2.5 text-text-primary text-sm placeholder-text-secondary focus:outline-none focus:ring-2 focus:ring-accent-purple/50 focus:border-accent-purple transition-colors"
                      />
                    </div>
                  </>
                )}

                <button
                  type="submit"
                  disabled={submitting}
                  className="w-full flex items-center justify-center gap-2 bg-gradient-to-r from-fuchsia-500 to-violet-600 hover:from-fuchsia-400 hover:to-violet-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium rounded-lg px-4 py-2.5 text-sm transition-all duration-200 shadow-md shadow-fuchsia-500/20"
                >
                  {submitting ? (
                    <Loader2 className="animate-spin" size={18} />
                  ) : (
                    <>
                      <UserPlus size={18} />
                      {invite.user_exists ? 'Join Tenant' : 'Create Account & Join'}
                    </>
                  )}
                </button>
              </form>
            </>
          ) : null}
        </div>
      </div>
    </div>
  );
}
