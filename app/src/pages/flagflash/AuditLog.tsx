import { useState, useEffect, useCallback } from 'react';
import { 
  FileText, 
  Filter,
  Loader2,
  AlertCircle,
  Search,
  ChevronLeft,
  ChevronRight,
  Clock,
  User,
  Hash,
  Activity,
  Eye
} from 'lucide-react';
import { auditLogsApi } from '../../services/flagflash-api';
import { useAuth } from '../../contexts/AuthContext';
import type { AuditLog, EntityType, AuditAction, AuditLogFilters } from '../../types/flagflash';

const ENTITY_TYPE_LABELS: Record<EntityType, string> = {
  tenant: 'Tenant',
  application: 'Application',
  environment: 'Environment',
  feature_flag: 'Feature Flag',
  targeting_rule: 'Targeting Rule',
  api_key: 'API Key',
  user: 'User',
};

const ACTION_LABELS: Record<AuditAction, string> = {
  create: 'Created',
  update: 'Updated',
  delete: 'Deleted',
  enable: 'Enabled',
  disable: 'Disabled',
  toggle: 'Toggled',
  revoke: 'Revoked',
  rotate: 'Rotated',
};

const ACTION_COLORS: Record<AuditAction, string> = {
  create: 'bg-green-500/20 text-green-400',
  update: 'bg-blue-500/20 text-blue-400',
  delete: 'bg-red-500/20 text-red-400',
  enable: 'bg-emerald-500/20 text-emerald-400',
  disable: 'bg-amber-500/20 text-amber-400',
  toggle: 'bg-purple-500/20 text-purple-400',
  revoke: 'bg-orange-500/20 text-orange-400',
  rotate: 'bg-cyan-500/20 text-cyan-400',
};

export default function AuditLogPage() {
  const { selectedTenant } = useAuth();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  
  // Filters
  const [filters, setFilters] = useState<AuditLogFilters>({});
  const [showFilters, setShowFilters] = useState(false);
  const [searchEntityId, setSearchEntityId] = useState('');

  const fetchLogs = useCallback(async () => {
    if (!selectedTenant?.id) return;
    
    try {
      setLoading(true);
      setError(null);
      const response = await auditLogsApi.list(selectedTenant.id, filters, page, 20);
      setLogs(response.logs);
      setTotalPages(response.pagination.total_pages);
    } catch (err) {
      setError('Failed to load audit logs');
      setLogs([]);
    } finally {
      setLoading(false);
    }
  }, [selectedTenant?.id, filters, page]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const handleFilterChange = (key: keyof AuditLogFilters, value: string) => {
    setFilters(prev => ({
      ...prev,
      [key]: value || undefined,
    }));
    setPage(1);
  };

  const clearFilters = () => {
    setFilters({});
    setSearchEntityId('');
    setPage(1);
  };

  const handleSearchEntityId = () => {
    if (searchEntityId) {
      setFilters(prev => ({ ...prev, entity_id: searchEntityId }));
    }
  };

  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="p-6 min-w-0 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
            <FileText className="text-accent-purple" />
            Audit Log
          </h1>
          <p className="text-text-secondary mt-1">Track all changes and actions in your feature flag system</p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`px-4 py-2 rounded-lg flex items-center gap-2 transition-colors ${
              showFilters || Object.keys(filters).length > 0
                ? 'bg-accent-purple text-white'
                : 'bg-bg-secondary border border-border text-text-primary hover:bg-bg-tertiary'
            }`}
          >
            <Filter size={18} />
            Filters
            {Object.keys(filters).length > 0 && (
              <span className="ml-1 px-2 py-0.5 bg-white/20 rounded-full text-xs">
                {Object.keys(filters).length}
              </span>
            )}
          </button>
        </div>
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <div className="bg-bg-secondary border border-border rounded-xl p-4 mb-6 modal-content">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Entity Type Filter */}
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Entity Type
              </label>
              <select
                value={filters.entity_type || ''}
                onChange={(e) => handleFilterChange('entity_type', e.target.value as EntityType)}
                className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
              >
                <option value="">All Types</option>
                {Object.entries(ENTITY_TYPE_LABELS).map(([value, label]) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </select>
            </div>

            {/* Action Filter */}
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Action
              </label>
              <select
                value={filters.action || ''}
                onChange={(e) => handleFilterChange('action', e.target.value as AuditAction)}
                className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
              >
                <option value="">All Actions</option>
                {Object.entries(ACTION_LABELS).map(([value, label]) => (
                  <option key={value} value={value}>{label}</option>
                ))}
              </select>
            </div>

            {/* Entity ID Search */}
            <div>
              <label className="block text-sm font-medium text-text-secondary mb-2">
                Entity ID
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  value={searchEntityId}
                  onChange={(e) => setSearchEntityId(e.target.value)}
                  placeholder="UUID..."
                  className="flex-1 px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
                />
                <button
                  onClick={handleSearchEntityId}
                  className="px-3 py-2 bg-bg-tertiary border border-border rounded-lg hover:bg-border transition-colors"
                >
                  <Search size={18} />
                </button>
              </div>
            </div>

            {/* Clear Filters */}
            <div className="flex items-end">
              <button
                onClick={clearFilters}
                className="w-full px-4 py-2 bg-bg-tertiary border border-border rounded-lg text-text-primary hover:bg-border transition-colors"
              >
                Clear Filters
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="p-4 bg-red-500/10 border border-red-500/50 rounded-lg mb-6 flex items-center gap-3">
          <AlertCircle className="text-red-500" size={20} />
          <span className="text-red-400">{error}</span>
        </div>
      )}

      {/* Main content: list + detail side panel */}
      <div className="flex gap-4 items-start min-w-0">
        {/* List column */}
        <div className="w-0 flex-[55] min-w-0">
          {loading ? (
            <div className="flex items-center justify-center py-20">
              <Loader2 className="animate-spin text-accent-purple" size={32} />
            </div>
          ) : logs.length === 0 ? (
            <div className="text-center py-20">
              <FileText className="mx-auto text-text-secondary mb-4" size={48} />
              <p className="text-text-secondary">No audit logs found</p>
              {Object.keys(filters).length > 0 && (
                <button
                  onClick={clearFilters}
                  className="mt-4 text-accent-purple hover:underline"
                >
                  Clear filters
                </button>
              )}
            </div>
          ) : (
            <>
              {/* Logs List */}
              <div className="space-y-3">
                {logs.map(log => (
                  <div
                    key={log.id}
                    className={`border rounded-xl p-4 hover:border-accent-purple/50 transition-colors cursor-pointer ${
                      selectedLog?.id === log.id
                        ? 'bg-accent-purple/5 border-accent-purple'
                        : 'bg-bg-secondary border-border'
                    }`}
                    onClick={() => setSelectedLog(log)}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex items-start gap-4">
                        {/* Icon */}
                        <div className="p-2 bg-bg-tertiary rounded-lg">
                          <Activity size={20} className="text-accent-purple" />
                        </div>

                        {/* Content */}
                        <div>
                          <div className="flex items-center gap-2 mb-1">
                            <span className={`px-2 py-0.5 rounded text-xs font-medium ${ACTION_COLORS[log.action]}`}>
                              {ACTION_LABELS[log.action]}
                            </span>
                            <span className="text-text-primary font-medium">
                              {ENTITY_TYPE_LABELS[log.entity_type]}
                            </span>
                          </div>
                          
                          <div className="flex items-center gap-4 text-sm text-text-secondary">
                            <span className="flex items-center gap-1">
                              <Hash size={14} />
                              {log.entity_id.slice(0, 8)}...
                            </span>
                            <span className="flex items-center gap-1">
                              <User size={14} />
                              {log.actor_type === 'system' ? 'System' : (log.actor_name || log.actor_id.slice(0, 8) + '...')}
                            </span>
                            <span className="flex items-center gap-1">
                              <Clock size={14} />
                              {formatRelativeTime(log.created_at)}
                            </span>
                          </div>
                        </div>
                      </div>

                      <button className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors">
                        <Eye size={18} className={selectedLog?.id === log.id ? 'text-accent-purple' : 'text-text-secondary'} />
                      </button>
                    </div>
                  </div>
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-2 mt-6">
                  <button
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page === 1}
                    className="p-2 bg-bg-secondary border border-border rounded-lg hover:bg-bg-tertiary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    <ChevronLeft size={18} />
                  </button>
                  
                  <span className="px-4 py-2 text-text-secondary">
                    Page {page} of {totalPages}
                  </span>
                  
                  <button
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                    disabled={page === totalPages}
                    className="p-2 bg-bg-secondary border border-border rounded-lg hover:bg-bg-tertiary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                  >
                    <ChevronRight size={18} />
                  </button>
                </div>
              )}
            </>
          )}
        </div>

        {/* Detail Side Panel */}
        <div className="w-0 flex-[45] min-w-0 sticky top-6">
          {selectedLog ? (
            <AuditLogDetailPanel log={selectedLog} onClose={() => setSelectedLog(null)} />
          ) : (
            <div className="bg-bg-secondary border border-border border-dashed rounded-xl flex flex-col items-center justify-center py-20 px-6 text-center">
              <Eye size={40} className="text-text-secondary mb-4 opacity-40" />
              <p className="text-text-primary font-medium mb-1">Nenhum registro selecionado</p>
              <p className="text-text-secondary text-sm">Selecione um item da lista para visualizar os detalhes</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function AuditLogDetailPanel({ log, onClose }: { log: AuditLog; onClose: () => void }) {
  return (
    <div className="bg-bg-secondary border border-border rounded-xl flex flex-col overflow-hidden" style={{maxHeight: 'calc(100vh - 180px)'}}>
      {/* Header */}
      <div className="p-4 border-b border-border shrink-0 flex items-center justify-between">
        <div className="flex items-center gap-3 min-w-0">
          <span className={`px-2 py-1 rounded text-sm font-medium shrink-0 ${ACTION_COLORS[log.action]}`}>
            {ACTION_LABELS[log.action]}
          </span>
          <span className="text-base font-semibold text-text-primary truncate">
            {ENTITY_TYPE_LABELS[log.entity_type]}
          </span>
        </div>
        <button
          onClick={onClose}
          className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors text-text-secondary shrink-0 ml-2"
        >
          ✕
        </button>
      </div>

      {/* Body */}
      <div className="p-4 overflow-y-auto space-y-5">
        {/* Info Grid */}
        <div className="grid grid-cols-2 gap-3">
          <div>
            <p className="text-xs text-text-secondary uppercase tracking-wide mb-1">Log ID</p>
            <p className="text-xs font-mono text-text-primary break-all">{String(log.id)}</p>
          </div>
          <div>
            <p className="text-xs text-text-secondary uppercase tracking-wide mb-1">Entity ID</p>
            <p className="text-xs font-mono text-text-primary break-all">{String(log.entity_id)}</p>
          </div>
          <div>
            <p className="text-xs text-text-secondary uppercase tracking-wide mb-1">Actor</p>
            <p className="text-sm text-text-primary capitalize">
              {log.actor_type}: {log.actor_type === 'system' ? 'System' : (log.actor_name || log.actor_id)}
            </p>
          </div>
          <div>
            <p className="text-xs text-text-secondary uppercase tracking-wide mb-1">Timestamp</p>
            <p className="text-sm text-text-primary">{new Date(log.created_at).toLocaleString()}</p>
          </div>
        </div>

        {/* Old Value */}
        {log.old_value != null && (
          <div>
            <p className="text-xs text-text-secondary uppercase tracking-wide mb-2">Previous Value</p>
            <pre className="p-3 bg-bg-primary border border-border rounded-lg text-xs overflow-x-auto">
              <code className="text-red-400">{JSON.stringify(log.old_value, null, 2)}</code>
            </pre>
          </div>
        )}

        {/* New Value */}
        {log.new_value != null && (
          <div>
            <p className="text-xs text-text-secondary uppercase tracking-wide mb-2">New Value</p>
            <pre className="p-3 bg-bg-primary border border-border rounded-lg text-xs overflow-x-auto">
              <code className="text-green-400">{JSON.stringify(log.new_value, null, 2)}</code>
            </pre>
          </div>
        )}

        {/* Metadata */}
        {log.metadata != null && Object.keys(log.metadata as Record<string, unknown>).length > 0 && (
          <div>
            <p className="text-xs text-text-secondary uppercase tracking-wide mb-2">Metadata</p>
            <pre className="p-3 bg-bg-primary border border-border rounded-lg text-xs overflow-x-auto">
              <code className="text-text-primary">{JSON.stringify(log.metadata, null, 2)}</code>
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
