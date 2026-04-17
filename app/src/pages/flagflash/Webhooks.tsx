import React, { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  Webhook as WebhookIcon, 
  Plus, 
  Search, 
  Trash2, 
  Edit2,
  ChevronLeft,
  ChevronUp,
  Loader2,
  AlertCircle,
  CheckCircle,
  XCircle,
  ExternalLink,
  RefreshCw,
  Send,
  Clock,
  Activity,
} from 'lucide-react';
import { webhooksApi } from '../../services/flagflash-api';
import { useAuth } from '../../contexts/AuthContext';
import { usePermissions } from '../../hooks/usePermissions';
import { ConfirmDeleteModal, Modal } from '../../components';
import type { Webhook, CreateWebhookRequest, WebhookDelivery } from '../../types/flagflash';

const WEBHOOK_EVENTS = [
  { value: 'flag.created', label: 'Flag Created' },
  { value: 'flag.updated', label: 'Flag Updated' },
  { value: 'flag.deleted', label: 'Flag Deleted' },
  { value: 'flag.toggled', label: 'Flag Toggled' },
  { value: 'environment.created', label: 'Environment Created' },
  { value: 'environment.deleted', label: 'Environment Deleted' },
];

export default function WebhooksPage() {
  const { tenantId: urlTenantId } = useParams<{ tenantId: string }>();
  const { selectedTenant } = useAuth();
  const { isAtLeast } = usePermissions();
  const isAdmin = isAtLeast('admin');
  const activeTenantId = urlTenantId || selectedTenant?.id || '';
  
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [webhookToDelete, setWebhookToDelete] = useState<Webhook | null>(null);
  const [editingWebhook, setEditingWebhook] = useState<Webhook | null>(null);
  const [expandedDeliveries, setExpandedDeliveries] = useState<Record<string, boolean>>({});
  const [deliveries, setDeliveries] = useState<Record<string, WebhookDelivery[]>>({});
  const [loadingDeliveries, setLoadingDeliveries] = useState<Record<string, boolean>>({});
  const [testingWebhook, setTestingWebhook] = useState<Record<string, boolean>>({});
  const [retryingDelivery, setRetryingDelivery] = useState<Record<string, boolean>>({});

  const fetchData = useCallback(async () => {
    if (!activeTenantId) {
      setLoading(false);
      return;
    }
    
    try {
      setLoading(true);
      const response = await webhooksApi.list(activeTenantId);
      setWebhooks(response.webhooks || []);
      setError(null);
    } catch (err) {
      setError('Failed to load webhooks');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [activeTenantId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filteredWebhooks = webhooks.filter(webhook => 
    webhook.name.toLowerCase().includes(search.toLowerCase()) ||
    webhook.url.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = async () => {
    if (!webhookToDelete || !activeTenantId) return;
    
    await webhooksApi.delete(activeTenantId, webhookToDelete.id);
    setWebhooks(prev => prev.filter(w => w.id !== webhookToDelete.id));
    setWebhookToDelete(null);
  };

  const handleToggleEnabled = async (webhook: Webhook) => {
    if (!activeTenantId) return;
    
    const updated = await webhooksApi.update(activeTenantId, webhook.id, {
      enabled: !webhook.enabled,
    });
    setWebhooks(prev => prev.map(w => w.id === webhook.id ? updated : w));
  };

  const toggleDeliveries = async (webhookId: string) => {
    const isExpanded = expandedDeliveries[webhookId];
    setExpandedDeliveries(prev => ({ ...prev, [webhookId]: !isExpanded }));

    if (!isExpanded && !deliveries[webhookId]) {
      await loadDeliveries(webhookId);
    }
  };

  const loadDeliveries = async (webhookId: string) => {
    if (!activeTenantId) return;
    setLoadingDeliveries(prev => ({ ...prev, [webhookId]: true }));
    try {
      const res = await webhooksApi.listDeliveries(activeTenantId, webhookId);
      setDeliveries(prev => ({ ...prev, [webhookId]: res.deliveries || [] }));
    } catch {
      // ignore
    } finally {
      setLoadingDeliveries(prev => ({ ...prev, [webhookId]: false }));
    }
  };

  const handleTest = async (webhookId: string) => {
    if (!activeTenantId) return;
    setTestingWebhook(prev => ({ ...prev, [webhookId]: true }));
    try {
      await webhooksApi.test(activeTenantId, webhookId);
      // Refresh deliveries list
      await loadDeliveries(webhookId);
      setExpandedDeliveries(prev => ({ ...prev, [webhookId]: true }));
    } catch {
      // ignore
    } finally {
      setTestingWebhook(prev => ({ ...prev, [webhookId]: false }));
    }
  };

  const handleRetry = async (webhookId: string, deliveryId: string) => {
    if (!activeTenantId) return;
    setRetryingDelivery(prev => ({ ...prev, [deliveryId]: true }));
    try {
      await webhooksApi.retryDelivery(activeTenantId, webhookId, deliveryId);
      // Refresh deliveries list after a short delay to allow async processing
      setTimeout(() => loadDeliveries(webhookId), 1500);
    } catch {
      // ignore
    } finally {
      setRetryingDelivery(prev => ({ ...prev, [deliveryId]: false }));
    }
  };

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          {urlTenantId && (
            <Link
              to="/tenants"
              className="p-2 hover:bg-bg-secondary rounded-lg transition-colors"
            >
              <ChevronLeft size={20} className="text-text-secondary" />
            </Link>
          )}
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <WebhookIcon className="text-accent-purple" />
              Webhooks
            </h1>
            <p className="text-text-secondary mt-1">
              Configure webhooks to receive event notifications
            </p>
          </div>
          {isAdmin && (
            <button
              onClick={() => setShowCreateModal(true)}
              disabled={!activeTenantId}
              className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Plus size={20} />
              Create Webhook
            </button>
          )}
        </div>

        {/* Search */}
        <div className="relative mb-6">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary" size={20} />
          <input
            type="text"
            placeholder="Search webhooks..."
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
          /* Webhooks List */
          <div className="space-y-3">
            {filteredWebhooks.length === 0 ? (
              <div className="text-center py-12 text-text-secondary">
                <WebhookIcon size={48} className="mx-auto mb-4 opacity-50" />
                <p>No webhooks found</p>
                <p className="text-sm mt-1">Create a webhook to receive event notifications</p>
              </div>
            ) : (
              filteredWebhooks.map(webhook => (
                <div
                  key={webhook.id}
                  className={`bg-bg-secondary border rounded-xl overflow-hidden transition-colors ${
                    webhook.enabled ? 'border-border hover:border-accent-purple/50' : 'border-red-500/30 opacity-75'
                  }`}
                >
                  <div className="p-5">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-2">
                          <h3 className="font-semibold text-text-primary">{webhook.name}</h3>
                          <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                            webhook.enabled 
                              ? 'bg-green-500/20 text-green-400'
                              : 'bg-red-500/20 text-red-400'
                          }`}>
                            {webhook.enabled ? 'Enabled' : 'Disabled'}
                          </span>
                        </div>
                        
                        <div className="flex items-center gap-2 text-sm text-text-secondary mb-3">
                          <ExternalLink size={14} />
                          <span className="font-mono truncate max-w-md">{webhook.url}</span>
                        </div>

                        <div className="flex flex-wrap gap-2">
                          {webhook.events.map(event => (
                            <span
                              key={event}
                              className="px-2 py-1 bg-bg-tertiary rounded text-xs text-text-secondary"
                            >
                              {event}
                            </span>
                          ))}
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => handleTest(webhook.id)}
                          disabled={testingWebhook[webhook.id] || !webhook.enabled}
                          className="p-2 hover:bg-accent-purple/10 rounded-lg transition-colors text-text-secondary hover:text-accent-purple disabled:opacity-40 disabled:cursor-not-allowed"
                          title="Send test event"
                        >
                          {testingWebhook[webhook.id] ? <Loader2 size={18} className="animate-spin" /> : <Send size={18} />}
                        </button>
                        <button
                          onClick={() => toggleDeliveries(webhook.id)}
                          className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors text-text-secondary hover:text-text-primary"
                          title="View dispatched events"
                        >
                          {expandedDeliveries[webhook.id] ? <ChevronUp size={18} /> : <Activity size={18} />}
                        </button>
                        <button
                          onClick={() => handleToggleEnabled(webhook)}
                          className={`p-2 rounded-lg transition-colors ${
                            webhook.enabled
                              ? 'hover:bg-red-500/10 text-green-400 hover:text-red-400'
                              : 'hover:bg-green-500/10 text-red-400 hover:text-green-400'
                          }`}
                          title={webhook.enabled ? 'Disable' : 'Enable'}
                        >
                          {webhook.enabled ? <CheckCircle size={18} /> : <XCircle size={18} />}
                        </button>
                        <button
                          onClick={() => setEditingWebhook(webhook)}
                          className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors text-text-secondary hover:text-accent-purple"
                          title="Edit"
                        >
                          <Edit2 size={18} />
                        </button>
                        {isAdmin && (
                          <button
                            onClick={() => setWebhookToDelete(webhook)}
                            className="p-2 hover:bg-red-500/10 rounded-lg transition-colors text-text-secondary hover:text-red-400"
                            title="Delete"
                          >
                            <Trash2 size={18} />
                          </button>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Deliveries Panel */}
                  {expandedDeliveries[webhook.id] && (
                    <div className="border-t border-border bg-bg-primary/50">
                      <div className="px-5 py-3 flex items-center justify-between">
                        <h4 className="text-sm font-medium text-text-primary flex items-center gap-2">
                          <Activity size={14} className="text-accent-purple" />
                          Dispatched Events
                        </h4>
                        <button
                          onClick={() => loadDeliveries(webhook.id)}
                          className="p-1.5 hover:bg-bg-secondary rounded transition-colors text-text-secondary hover:text-text-primary"
                          title="Refresh"
                        >
                          <RefreshCw size={14} />
                        </button>
                      </div>

                      {loadingDeliveries[webhook.id] ? (
                        <div className="flex items-center justify-center py-6">
                          <Loader2 className="animate-spin text-accent-purple" size={20} />
                        </div>
                      ) : (deliveries[webhook.id] || []).length === 0 ? (
                        <div className="px-5 pb-5 text-center text-text-secondary text-sm">
                          <Clock size={24} className="mx-auto mb-2 opacity-40" />
                          No events dispatched yet
                        </div>
                      ) : (
                        <div className="divide-y divide-border/50">
                          {(deliveries[webhook.id] || []).map(delivery => (
                            <DeliveryRow
                              key={delivery.id}
                              delivery={delivery}
                              webhookId={webhook.id}
                              tenantId={activeTenantId}
                              retrying={retryingDelivery[delivery.id]}
                              onRetry={handleRetry}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              ))
            )}
          </div>
        )}

        {/* Create Modal */}
        {showCreateModal && (
          <CreateWebhookModal
            tenantId={activeTenantId}
            onClose={() => setShowCreateModal(false)}
            onCreate={(webhook) => {
              setWebhooks(prev => [...prev, webhook]);
              setShowCreateModal(false);
            }}
          />
        )}

        {/* Edit Modal */}
        {editingWebhook && (
          <EditWebhookModal
            tenantId={activeTenantId}
            webhook={editingWebhook}
            onClose={() => setEditingWebhook(null)}
            onUpdate={(webhook) => {
              setWebhooks(prev => prev.map(w => w.id === webhook.id ? webhook : w));
              setEditingWebhook(null);
            }}
          />
        )}

        {/* Delete Modal */}
        <ConfirmDeleteModal
          isOpen={!!webhookToDelete}
          onClose={() => setWebhookToDelete(null)}
          onConfirm={handleDelete}
          title="Delete Webhook"
          message={`Are you sure you want to delete "${webhookToDelete?.name}"? This action cannot be undone.`}
        />
      </div>
    </div>
  );
}

// ---- DeliveryRow Component ----
interface DeliveryRowProps {
  delivery: WebhookDelivery;
  webhookId: string;
  tenantId: string;
  retrying?: boolean;
  onRetry: (webhookId: string, deliveryId: string) => void;
}

function DeliveryRow({ delivery, webhookId, tenantId: _tenantId, retrying, onRetry }: DeliveryRowProps) {
  const statusConfig: Record<string, { color: string; icon: React.ReactNode; label: string }> = {
    success: { color: 'text-green-400', icon: <CheckCircle size={14} />, label: 'Success' },
    failed: { color: 'text-red-400', icon: <XCircle size={14} />, label: 'Failed' },
    retrying: { color: 'text-yellow-400', icon: <RefreshCw size={14} />, label: 'Retrying' },
    pending: { color: 'text-blue-400', icon: <Clock size={14} />, label: 'Pending' },
  };

  const st = statusConfig[delivery.status] ?? statusConfig.pending;
  const canRetry = delivery.status === 'failed' || delivery.status === 'retrying';

  return (
    <div className="px-5 py-3 flex items-center gap-4 hover:bg-bg-secondary/30 transition-colors">
      <div className={`flex items-center gap-1.5 ${st.color} text-xs font-medium min-w-[80px]`}>
        {st.icon}
        {st.label}
      </div>

      <span className="text-xs font-mono text-text-secondary bg-bg-tertiary px-2 py-0.5 rounded">
        {delivery.event_type}
      </span>

      {delivery.response_status && (
        <span className={`text-xs font-mono ${delivery.response_status >= 200 && delivery.response_status < 300 ? 'text-green-400' : 'text-red-400'}`}>
          HTTP {delivery.response_status}
        </span>
      )}

      {delivery.duration_ms > 0 && (
        <span className="text-xs text-text-secondary">{delivery.duration_ms}ms</span>
      )}

      {delivery.attempt > 1 && (
        <span className="text-xs text-text-secondary">Attempt #{delivery.attempt}</span>
      )}

      {delivery.error_message && (
        <span className="text-xs text-red-400 truncate max-w-xs" title={delivery.error_message}>
          {delivery.error_message}
        </span>
      )}

      <span className="text-xs text-text-secondary ml-auto">
        {new Date(delivery.created_at).toLocaleString()}
      </span>

      {canRetry && (
        <button
          onClick={() => onRetry(webhookId, delivery.id)}
          disabled={retrying}
          className="flex items-center gap-1 px-2 py-1 rounded text-xs bg-accent-purple/10 text-accent-purple hover:bg-accent-purple/20 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          title="Retry this delivery"
        >
          {retrying ? <Loader2 size={12} className="animate-spin" /> : <RefreshCw size={12} />}
          Retry
        </button>
      )}
    </div>
  );
}

interface CreateWebhookModalProps {
  tenantId: string;
  onClose: () => void;
  onCreate: (webhook: Webhook) => void;
}

function CreateWebhookModal({ tenantId, onClose, onCreate }: CreateWebhookModalProps) {
  const [name, setName] = useState('');
  const [url, setUrl] = useState('');
  const [secret, setSecret] = useState('');
  const [selectedEvents, setSelectedEvents] = useState<string[]>([]);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !url || selectedEvents.length === 0) return;

    try {
      setCreating(true);
      setError(null);
      const data: CreateWebhookRequest = {
        name,
        url,
        events: selectedEvents,
      };
      if (secret) data.secret = secret;
      
      const webhook = await webhooksApi.create(tenantId, data);
      onCreate(webhook);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setCreating(false);
    }
  };

  const toggleEvent = (event: string) => {
    setSelectedEvents(prev => 
      prev.includes(event) 
        ? prev.filter(e => e !== event)
        : [...prev, event]
    );
  };

  return (
    <Modal isOpen onClose={onClose} title="Create Webhook">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="My Webhook"
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">URL</label>
          <input
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com/webhook"
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">Secret (optional)</label>
          <input
            type="text"
            value={secret}
            onChange={(e) => setSecret(e.target.value)}
            placeholder="Optional secret for signature verification"
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-2">Events</label>
          <div className="grid grid-cols-2 gap-2">
            {WEBHOOK_EVENTS.map(event => (
              <label
                key={event.value}
                className={`flex items-center gap-2 p-2 border rounded-lg cursor-pointer transition-colors ${
                  selectedEvents.includes(event.value)
                    ? 'bg-accent-purple/10 border-accent-purple text-accent-purple'
                    : 'border-border text-text-secondary hover:border-accent-purple/50'
                }`}
              >
                <input
                  type="checkbox"
                  checked={selectedEvents.includes(event.value)}
                  onChange={() => toggleEvent(event.value)}
                  className="sr-only"
                />
                <span className="text-sm">{event.label}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={creating || !name || !url || selectedEvents.length === 0}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {creating && <Loader2 size={16} className="animate-spin" />}
            Create Webhook
          </button>
        </div>
      </form>
    </Modal>
  );
}

interface EditWebhookModalProps {
  tenantId: string;
  webhook: Webhook;
  onClose: () => void;
  onUpdate: (webhook: Webhook) => void;
}

function EditWebhookModal({ tenantId, webhook, onClose, onUpdate }: EditWebhookModalProps) {
  const [name, setName] = useState(webhook.name);
  const [url, setUrl] = useState(webhook.url);
  const [selectedEvents, setSelectedEvents] = useState<string[]>(webhook.events);
  const [updating, setUpdating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !url || selectedEvents.length === 0) return;

    try {
      setUpdating(true);
      setError(null);
      const updated = await webhooksApi.update(tenantId, webhook.id, {
        name,
        url,
        events: selectedEvents,
      });
      onUpdate(updated);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setUpdating(false);
    }
  };

  const toggleEvent = (event: string) => {
    setSelectedEvents(prev => 
      prev.includes(event) 
        ? prev.filter(e => e !== event)
        : [...prev, event]
    );
  };

  return (
    <Modal isOpen onClose={onClose} title="Edit Webhook">
      <form onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">URL</label>
          <input
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-2">Events</label>
          <div className="grid grid-cols-2 gap-2">
            {WEBHOOK_EVENTS.map(event => (
              <label
                key={event.value}
                className={`flex items-center gap-2 p-2 border rounded-lg cursor-pointer transition-colors ${
                  selectedEvents.includes(event.value)
                    ? 'bg-accent-purple/10 border-accent-purple text-accent-purple'
                    : 'border-border text-text-secondary hover:border-accent-purple/50'
                }`}
              >
                <input
                  type="checkbox"
                  checked={selectedEvents.includes(event.value)}
                  onChange={() => toggleEvent(event.value)}
                  className="sr-only"
                />
                <span className="text-sm">{event.label}</span>
              </label>
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-3 pt-4">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={updating || !name || !url || selectedEvents.length === 0}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {updating && <Loader2 size={16} className="animate-spin" />}
            Save Changes
          </button>
        </div>
      </form>
    </Modal>
  );
}
