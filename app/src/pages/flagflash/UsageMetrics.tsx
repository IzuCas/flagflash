import { useState, useEffect, useMemo } from 'react';
import {
  BarChart3,
  Activity,
  Users,
  Layers,
  AlertCircle,
  Loader2,
  TrendingUp,
} from 'lucide-react';
import { usageMetricsApi } from '../../services/flagflash-api';
import { useAuth } from '../../contexts/AuthContext';
import type { UsageMetrics, UsageMetricsFilters, TimelinePoint, FlagMetric, EnvironmentMetric } from '../../types/flagflash';

// ─── Helpers ────────────────────────────────────────────────────────────────

const formatDate = (date: Date): string => date.toISOString();

const formatDateShort = (dateStr: string): string =>
  new Date(dateStr).toLocaleDateString('pt-BR', {
    day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit',
  });

const getDateRange = (period: 'day' | 'week' | 'month') => {
  const end = new Date();
  const start = new Date();
  if (period === 'day') start.setDate(start.getDate() - 1);
  else if (period === 'week') start.setDate(start.getDate() - 7);
  else start.setMonth(start.getMonth() - 1);
  return { start, end };
};

const PERIOD_OPTIONS = [
  { value: 'day' as const, label: 'Last 24 Hours' },
  { value: 'week' as const, label: 'Last 7 Days' },
  { value: 'month' as const, label: 'Last 30 Days' },
];

const GRANULARITY_OPTIONS = [
  { value: 'hour' as const, label: 'Hourly' },
  { value: 'day' as const, label: 'Daily' },
  { value: 'week' as const, label: 'Weekly' },
  { value: 'month' as const, label: 'Monthly' },
];

// ─── Bar Chart ───────────────────────────────────────────────────────────────

const SimpleBarChart = ({ data, maxValue }: { data: TimelinePoint[]; maxValue: number }) => {
  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-48 text-text-secondary text-sm">
        No data available for the selected period
      </div>
    );
  }

  return (
    <div className="flex items-end gap-0.5 h-48 w-full">
      {data.map((point, index) => {
        const height = maxValue > 0 ? (point.evaluations / maxValue) * 100 : 0;
        const trueH = point.evaluations > 0 ? (point.true_count / point.evaluations) * height : 0;
        const falseH = point.evaluations > 0 ? (point.false_count / point.evaluations) * height : 0;
        const otherH = height - trueH - falseH;

        return (
          <div key={index} className="flex-1 flex flex-col justify-end relative group" style={{ minWidth: '4px' }}>
            {/* Tooltip */}
            <div className="absolute bottom-full mb-2 left-1/2 -translate-x-1/2 bg-bg-secondary border border-border rounded-lg px-3 py-2 text-xs opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap z-10 pointer-events-none shadow-lg">
              <div className="font-semibold text-text-primary mb-1">{formatDateShort(point.timestamp)}</div>
              <div className="text-text-secondary">{point.evaluations.toLocaleString()} evals</div>
              <div className="text-green-400">✓ {point.true_count.toLocaleString()}</div>
              <div className="text-red-400">✗ {point.false_count.toLocaleString()}</div>
            </div>
            <div className="w-full bg-green-500/80 rounded-t-sm" style={{ height: `${trueH}%` }} />
            <div className="w-full bg-red-500/60" style={{ height: `${falseH}%` }} />
            {otherH > 0 && <div className="w-full bg-accent-purple/40 rounded-b-sm" style={{ height: `${otherH}%` }} />}
          </div>
        );
      })}
    </div>
  );
};

// ─── Stat Card ───────────────────────────────────────────────────────────────

interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon: React.ReactNode;
  iconBg: string;
}

const StatCard = ({ title, value, subtitle, icon, iconBg }: StatCardProps) => (
  <div className="bg-bg-secondary border border-border rounded-xl p-5">
    <div className="flex items-start justify-between mb-3">
      <span className="text-sm font-medium text-text-secondary">{title}</span>
      <div className={`p-2 rounded-lg ${iconBg}`}>{icon}</div>
    </div>
    <div className="text-3xl font-bold text-text-primary">
      {typeof value === 'number' ? value.toLocaleString() : value}
    </div>
    {subtitle && <div className="text-xs text-text-secondary mt-1.5">{subtitle}</div>}
  </div>
);

// ─── Page ────────────────────────────────────────────────────────────────────

export default function UsageMetricsPage() {
  const { tenants, selectedTenant } = useAuth();

  const [selectedTenantId, setSelectedTenantId] = useState<string>('');
  const [metrics, setMetrics] = useState<UsageMetrics | null>(null);
  const [flagMetrics, setFlagMetrics] = useState<FlagMetric[]>([]);
  const [envMetrics, setEnvMetrics] = useState<EnvironmentMetric[]>([]);
  const [timeline, setTimeline] = useState<TimelinePoint[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [period, setPeriod] = useState<'day' | 'week' | 'month'>('week');
  const [granularity, setGranularity] = useState<UsageMetricsFilters['granularity']>('hour');

  // Initialise tenant from auth context
  useEffect(() => {
    if (selectedTenantId) return;
    if (selectedTenant) setSelectedTenantId(selectedTenant.id);
    else if (tenants.length > 0) setSelectedTenantId(tenants[0].id);
  }, [selectedTenant, tenants, selectedTenantId]);

  // Load metrics
  useEffect(() => {
    if (!selectedTenantId) return;

    const load = async () => {
      setLoading(true);
      setError(null);
      try {
        const { start, end } = getDateRange(period);
        const filters: UsageMetricsFilters = {
          start_date: formatDate(start),
          end_date: formatDate(end),
          granularity,
        };

        const [metricsData, timelineData, flagData, envData] = await Promise.all([
          usageMetricsApi.getSummary(selectedTenantId, filters),
          usageMetricsApi.getTimeline(selectedTenantId, filters),
          usageMetricsApi.getFlagMetrics(selectedTenantId, filters),
          usageMetricsApi.getEnvironmentMetrics(selectedTenantId, formatDate(start), formatDate(end)),
        ]);

        setMetrics(metricsData);
        setTimeline(timelineData.timeline || []);
        setFlagMetrics(flagData.flags || []);
        setEnvMetrics(envData.environments || []);
      } catch (err) {
        console.error('Failed to load metrics:', err);
        setError('Failed to load usage metrics. Please try again.');
      } finally {
        setLoading(false);
      }
    };

    load();
  }, [selectedTenantId, period, granularity]);

  const maxEvaluations = useMemo(
    () => (timeline.length === 0 ? 0 : Math.max(...timeline.map(p => p.evaluations))),
    [timeline],
  );

  const avgPerGranularity = useMemo(
    () => (timeline.length === 0 || !metrics ? 0 : Math.round(metrics.total_evaluations / timeline.length)),
    [timeline, metrics],
  );

  const currentTenantName = tenants.find(t => t.id === selectedTenantId)?.name ?? selectedTenant?.name ?? '';

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto space-y-6">

        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <BarChart3 className="text-accent-purple" />
              Usage Analytics
            </h1>
            <p className="text-text-secondary mt-1 text-sm">
              Monitor feature flag evaluation metrics{currentTenantName ? ` for ${currentTenantName}` : ''}
            </p>
          </div>

          {tenants.length > 1 && (
            <select
              value={selectedTenantId}
              onChange={(e) => setSelectedTenantId(e.target.value)}
              className="bg-bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-text-primary focus:outline-none focus:border-accent-purple"
            >
              {tenants.map((t) => (
                <option key={t.id} value={t.id}>{t.name}</option>
              ))}
            </select>
          )}
        </div>

        {/* Filters */}
        <div className="bg-bg-secondary border border-border rounded-xl p-4">
          <div className="flex flex-wrap items-center gap-6">
            <div className="flex items-center gap-3">
              <span className="text-sm font-medium text-text-secondary">Period:</span>
              <div className="flex rounded-lg border border-border overflow-hidden">
                {PERIOD_OPTIONS.map((opt) => (
                  <button
                    key={opt.value}
                    onClick={() => setPeriod(opt.value)}
                    className={`px-4 py-2 text-sm font-medium transition-colors ${
                      period === opt.value
                        ? 'bg-accent-purple text-white'
                        : 'text-text-secondary hover:text-text-primary hover:bg-bg-tertiary'
                    }`}
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
            </div>

            <div className="flex items-center gap-3">
              <span className="text-sm font-medium text-text-secondary">Granularity:</span>
              <select
                value={granularity}
                onChange={(e) => setGranularity(e.target.value as UsageMetricsFilters['granularity'])}
                className="bg-bg-primary border border-border rounded-lg px-3 py-2 text-sm text-text-primary focus:outline-none focus:border-accent-purple"
              >
                {GRANULARITY_OPTIONS.map((opt) => (
                  <option key={opt.value} value={opt.value}>{opt.label}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Error */}
        {error && (
          <div className="p-4 bg-red-500/10 border border-red-500/50 rounded-xl flex items-center gap-3 text-red-400">
            <AlertCircle size={20} />
            {error}
          </div>
        )}

        {/* Loading */}
        {loading ? (
          <div className="flex items-center justify-center py-16 gap-3">
            <Loader2 className="animate-spin text-accent-purple" size={28} />
            <span className="text-text-secondary">Loading metrics...</span>
          </div>
        ) : (
          <>
            {/* Stat Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              <StatCard
                title="Total Evaluations"
                value={metrics?.total_evaluations ?? 0}
                subtitle={`Avg ${avgPerGranularity.toLocaleString()} per ${granularity}`}
                icon={<Activity size={18} className="text-accent-purple" />}
                iconBg="bg-accent-purple/10"
              />
              <StatCard
                title="Unique Flags"
                value={metrics?.unique_flags ?? 0}
                subtitle="Flags evaluated"
                icon={<TrendingUp size={18} className="text-blue-400" />}
                iconBg="bg-blue-400/10"
              />
              <StatCard
                title="Unique Users"
                value={metrics?.unique_users ?? 0}
                subtitle="Distinct user IDs"
                icon={<Users size={18} className="text-green-400" />}
                iconBg="bg-green-400/10"
              />
              <StatCard
                title="Environments"
                value={envMetrics.length}
                subtitle="Active environments"
                icon={<Layers size={18} className="text-orange-400" />}
                iconBg="bg-orange-400/10"
              />
            </div>

            {/* Timeline */}
            <div className="bg-bg-secondary border border-border rounded-xl p-6">
              <h2 className="text-lg font-semibold text-text-primary mb-5">Evaluations Over Time</h2>
              <SimpleBarChart data={timeline} maxValue={maxEvaluations} />
              <div className="flex items-center justify-center gap-6 mt-4 text-sm text-text-secondary">
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 bg-green-500/80 rounded-sm" />
                  <span>True</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 bg-red-500/60 rounded-sm" />
                  <span>False</span>
                </div>
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 bg-accent-purple/40 rounded-sm" />
                  <span>Other</span>
                </div>
              </div>
            </div>

            {/* Environments + Top Flags */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">

              {/* Environments */}
              <div className="bg-bg-secondary border border-border rounded-xl p-6">
                <h2 className="text-lg font-semibold text-text-primary mb-5 flex items-center gap-2">
                  <Layers size={18} className="text-orange-400" />
                  By Environment
                </h2>
                {envMetrics.length === 0 ? (
                  <div className="text-center py-10">
                    <Layers className="mx-auto mb-3 text-text-secondary opacity-30" size={36} />
                    <p className="text-text-secondary text-sm">No environment data available</p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {envMetrics.map((env) => (
                      <div
                        key={env.environment_id}
                        className="flex items-center justify-between p-4 bg-bg-tertiary/50 hover:bg-bg-tertiary rounded-xl transition-colors"
                      >
                        <div>
                          <div className="font-medium text-text-primary">{env.environment_name}</div>
                          <div className="text-xs text-text-secondary mt-0.5">
                            {env.unique_flags} flags · {env.unique_users} users
                          </div>
                        </div>
                        <div className="text-right">
                          <div className="text-xl font-bold text-text-primary">
                            {env.evaluations.toLocaleString()}
                          </div>
                          <div className="text-xs text-text-secondary">evaluations</div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Top Flags */}
              <div className="bg-bg-secondary border border-border rounded-xl p-6">
                <h2 className="text-lg font-semibold text-text-primary mb-5 flex items-center gap-2">
                  <TrendingUp size={18} className="text-blue-400" />
                  Top Feature Flags
                </h2>
                {flagMetrics.length === 0 ? (
                  <div className="text-center py-10">
                    <BarChart3 className="mx-auto mb-3 text-text-secondary opacity-30" size={36} />
                    <p className="text-text-secondary text-sm">No flag data available</p>
                  </div>
                ) : (
                  <div className="space-y-3 max-h-96 overflow-y-auto pr-1">
                    {flagMetrics.slice(0, 10).map((flag) => {
                      const truePercent = flag.evaluations > 0
                        ? Math.round((flag.true_count / flag.evaluations) * 100)
                        : 0;
                      return (
                        <div
                          key={flag.flag_id}
                          className="p-4 bg-bg-tertiary/50 hover:bg-bg-tertiary rounded-xl transition-colors"
                        >
                          <div className="flex items-center justify-between mb-2">
                            <div className="min-w-0 flex-1 mr-3">
                              <div className="font-medium text-text-primary truncate">{flag.flag_name}</div>
                              <div className="text-xs text-text-secondary mt-0.5 flex items-center gap-2">
                                <code className="bg-bg-primary px-1.5 py-0.5 rounded text-accent-purple/80 font-mono">
                                  {flag.flag_key}
                                </code>
                                <span>· {flag.environment_name}</span>
                              </div>
                            </div>
                            <div className="text-xl font-bold text-text-primary flex-shrink-0">
                              {flag.evaluations.toLocaleString()}
                            </div>
                          </div>
                          <div className="flex items-center justify-between text-xs mb-1.5">
                            <span className="text-green-400">✓ {flag.true_count.toLocaleString()} ({truePercent}%)</span>
                            <span className="text-red-400">✗ {flag.false_count.toLocaleString()} ({100 - truePercent}%)</span>
                          </div>
                          <div className="w-full h-1.5 bg-red-500/20 rounded-full overflow-hidden">
                            <div
                              className="h-full bg-green-500 transition-all"
                              style={{ width: `${truePercent}%` }}
                            />
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>

            {/* Empty state */}
            {metrics && metrics.total_evaluations === 0 && (
              <div className="bg-bg-secondary border border-border rounded-xl p-12 text-center">
                <BarChart3 className="mx-auto mb-4 text-text-secondary opacity-20" size={52} />
                <h3 className="text-lg font-semibold text-text-primary mb-2">No Evaluation Data Yet</h3>
                <p className="text-text-secondary max-w-md mx-auto text-sm">
                  Once your applications start evaluating feature flags via the SDK,
                  usage metrics will appear here.
                </p>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
