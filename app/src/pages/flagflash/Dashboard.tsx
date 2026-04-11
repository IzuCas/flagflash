import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import {
  Flag,
  Building2,
  Layers,
  Key,
  Activity,
  Users,
  Zap,
  ArrowRight,
  Loader2,
  AlertCircle,
} from 'lucide-react';
import { tenantsApi } from '../../services/flagflash-api';
import type { TenantWithRole } from '../../services/flagflash-api';

interface DashboardStats {
  totalTenants: number;
  totalApplications: number;
  totalEnvironments: number;
  totalFlags: number;
  enabledFlags: number;
  disabledFlags: number;
  totalApiKeys: number;
  activeApiKeys: number;
}

export default function DashboardPage() {
  const [tenants, setTenants] = useState<TenantWithRole[]>([]);
  const [stats, setStats] = useState<DashboardStats>({
    totalTenants: 0,
    totalApplications: 0,
    totalEnvironments: 0,
    totalFlags: 0,
    enabledFlags: 0,
    disabledFlags: 0,
    totalApiKeys: 0,
    activeApiKeys: 0,
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await tenantsApi.listMyTenants();
        const tenantList = response.tenants || [];
        setTenants(tenantList);
        
        // Calculate stats from tenants (in a real app, you'd have a dedicated stats endpoint)
        setStats({
          totalTenants: tenantList.length,
          totalApplications: 0, // Would need to fetch from apps endpoint
          totalEnvironments: 0,
          totalFlags: 0,
          enabledFlags: 0,
          disabledFlags: 0,
          totalApiKeys: 0,
          activeApiKeys: 0,
        });
        setError(null);
      } catch (err) {
        setError('Failed to load dashboard data');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const statCards = [
    {
      title: 'Tenants',
      value: stats.totalTenants,
      icon: Building2,
      color: 'blue',
      link: '/tenants'
    },
    {
      title: 'Feature Flags',
      value: stats.totalFlags,
      icon: Flag,
      color: 'purple',
      subtext: `${stats.enabledFlags} enabled`
    },
    {
      title: 'API Keys',
      value: stats.totalApiKeys,
      icon: Key,
      color: 'green',
      subtext: `${stats.activeApiKeys} active`
    },
    {
      title: 'Environments',
      value: stats.totalEnvironments,
      icon: Layers,
      color: 'orange'
    }
  ];

  const colorClasses = {
    blue: 'bg-blue-500/20 text-blue-400',
    purple: 'bg-purple-500/20 text-purple-400',
    green: 'bg-green-500/20 text-green-400',
    orange: 'bg-orange-500/20 text-orange-400'
  };

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-text-primary flex items-center gap-3">
            <Zap className="text-accent-purple" />
            FlagFlash Dashboard
          </h1>
          <p className="text-text-secondary mt-2">
            Manage your feature flags, configurations, and deployments
          </p>
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
          <>
            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              {statCards.map((stat, index) => (
                <div
                  key={index}
                  className="bg-bg-secondary border border-border rounded-xl p-6 hover:border-accent-purple/50 transition-colors"
                >
                  <div className="flex items-center justify-between mb-4">
                    <div className={`p-3 rounded-lg ${colorClasses[stat.color as keyof typeof colorClasses]}`}>
                      <stat.icon size={24} />
                    </div>
                    {stat.link && (
                      <Link
                        to={stat.link}
                        className="text-text-secondary hover:text-accent-purple transition-colors"
                      >
                        <ArrowRight size={20} />
                      </Link>
                    )}
                  </div>
                  <h3 className="text-2xl font-bold text-text-primary">{stat.value}</h3>
                  <p className="text-text-secondary text-sm mt-1">{stat.title}</p>
                  {stat.subtext && (
                    <p className="text-text-secondary text-xs mt-1">{stat.subtext}</p>
                  )}
                </div>
              ))}
            </div>

            {/* Main Content Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Tenants List */}
              <div className="lg:col-span-2 bg-bg-secondary border border-border rounded-xl">
                <div className="p-6 border-b border-border flex items-center justify-between">
                  <h2 className="text-lg font-semibold text-text-primary flex items-center gap-2">
                    <Building2 size={20} className="text-blue-400" />
                    Your Tenants
                  </h2>
                  <Link
                    to="/tenants"
                    className="text-sm text-accent-purple hover:underline"
                  >
                    View All
                  </Link>
                </div>
                <div className="p-6">
                  {tenants.length > 0 ? (
                    <div className="space-y-4">
                      {tenants.slice(0, 5).map(tenant => (
                        <Link
                          key={tenant.id}
                          to={`/tenants/${tenant.id}/applications`}
                          className="flex items-center justify-between p-4 bg-bg-tertiary rounded-lg hover:bg-bg-primary transition-colors"
                        >
                          <div className="flex items-center gap-4">
                            <div className="w-10 h-10 bg-accent-purple/20 rounded-lg flex items-center justify-center">
                              <Building2 size={20} className="text-accent-purple" />
                            </div>
                            <div>
                              <h3 className="font-medium text-text-primary">{tenant.name}</h3>
                              <p className="text-sm text-text-secondary">
                                {tenant.slug}
                              </p>
                            </div>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className={`px-2 py-1 rounded text-xs ${
                              tenant.role === 'owner'
                                ? 'bg-purple-500/20 text-purple-400'
                                : tenant.role === 'admin'
                                ? 'bg-blue-500/20 text-blue-400'
                                : 'bg-gray-500/20 text-gray-400'
                            }`}>
                              {tenant.role}
                            </span>
                            <ArrowRight size={16} className="text-text-secondary" />
                          </div>
                        </Link>
                      ))}
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <Building2 className="mx-auto text-text-secondary mb-3" size={40} />
                      <p className="text-text-secondary">No tenants yet</p>
                      <Link
                        to="/tenants"
                        className="text-accent-purple hover:underline text-sm mt-2 inline-block"
                      >
                        Create your first tenant
                      </Link>
                    </div>
                  )}
                </div>
              </div>

              {/* Quick Actions */}
              <div className="bg-bg-secondary border border-border rounded-xl">
                <div className="p-6 border-b border-border">
                  <h2 className="text-lg font-semibold text-text-primary flex items-center gap-2">
                    <Zap size={20} className="text-yellow-400" />
                    Quick Actions
                  </h2>
                </div>
                <div className="p-6 space-y-3">
                  <Link
                    to="/tenants"
                    className="flex items-center gap-3 p-3 bg-bg-tertiary rounded-lg hover:bg-bg-primary transition-colors"
                  >
                    <div className="p-2 bg-blue-500/20 rounded-lg">
                      <Building2 size={18} className="text-blue-400" />
                    </div>
                    <div>
                      <p className="font-medium text-text-primary">Create Tenant</p>
                      <p className="text-xs text-text-secondary">Add a new organization</p>
                    </div>
                  </Link>
                  
                  <button
                    disabled
                    className="flex items-center gap-3 p-3 bg-bg-tertiary rounded-lg opacity-50 cursor-not-allowed w-full text-left"
                  >
                    <div className="p-2 bg-purple-500/20 rounded-lg">
                      <Flag size={18} className="text-purple-400" />
                    </div>
                    <div>
                      <p className="font-medium text-text-primary">Create Flag</p>
                      <p className="text-xs text-text-secondary">Select tenant first</p>
                    </div>
                  </button>
                  
                  <button
                    disabled
                    className="flex items-center gap-3 p-3 bg-bg-tertiary rounded-lg opacity-50 cursor-not-allowed w-full text-left"
                  >
                    <div className="p-2 bg-green-500/20 rounded-lg">
                      <Key size={18} className="text-green-400" />
                    </div>
                    <div>
                      <p className="font-medium text-text-primary">Generate API Key</p>
                      <p className="text-xs text-text-secondary">Select environment first</p>
                    </div>
                  </button>
                </div>
              </div>
            </div>

            {/* Feature Highlights */}
            <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-gradient-to-br from-purple-500/10 to-purple-500/5 border border-purple-500/30 rounded-xl p-6">
                <Flag className="text-purple-400 mb-3" size={28} />
                <h3 className="font-semibold text-text-primary mb-2">Feature Flags</h3>
                <p className="text-sm text-text-secondary">
                  Create boolean, string, number, or JSON flags with targeting rules and rollout strategies.
                </p>
              </div>
              
              <div className="bg-gradient-to-br from-blue-500/10 to-blue-500/5 border border-blue-500/30 rounded-xl p-6">
                <Activity className="text-blue-400 mb-3" size={28} />
                <h3 className="font-semibold text-text-primary mb-2">Real-time Updates</h3>
                <p className="text-sm text-text-secondary">
                  Flag changes are propagated instantly via WebSocket to all connected SDKs.
                </p>
              </div>
              
              <div className="bg-gradient-to-br from-green-500/10 to-green-500/5 border border-green-500/30 rounded-xl p-6">
                <Users className="text-green-400 mb-3" size={28} />
                <h3 className="font-semibold text-text-primary mb-2">Multi-Tenant</h3>
                <p className="text-sm text-text-secondary">
                  Isolate configurations per tenant with application and environment scoping.
                </p>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
