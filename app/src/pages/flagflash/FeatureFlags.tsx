import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  Flag, 
  Plus, 
  Search, 
  Trash2, 
  ToggleLeft,
  ToggleRight,
  Copy,
  Target,
  Loader2,
  AlertCircle,
  ChevronLeft,
  Tag,
  X,
  GripVertical,
  ChevronDown,
  ChevronUp,
  Edit2,
  Settings,
  Check,
  AlertTriangle,
} from 'lucide-react';
import { featureFlagsApi, environmentsApi, targetingRulesApi } from '../../services/flagflash-api';
import type { FeatureFlag, Environment, FlagType, TargetingRule, Condition, Operator } from '../../types/flagflash';
import { ConfirmDeleteModal, Modal } from '../../components';

// ============================================
// JSON EDITOR COMPONENT
// ============================================
interface JsonEditorProps {
  value: string;
  onChange: (value: string) => void;
  className?: string;
}

function JsonEditor({ value, onChange, className = '' }: JsonEditorProps) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const highlightRef = useRef<HTMLDivElement>(null);
  const [isValid, setIsValid] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Validate JSON on change
  useEffect(() => {
    try {
      if (value.trim()) {
        JSON.parse(value);
      }
      setIsValid(true);
      setErrorMessage(null);
    } catch (err) {
      setIsValid(false);
      setErrorMessage((err as Error).message);
    }
  }, [value]);

  // Sync scroll between textarea and highlight overlay
  const handleScroll = () => {
    if (textareaRef.current && highlightRef.current) {
      highlightRef.current.scrollTop = textareaRef.current.scrollTop;
      highlightRef.current.scrollLeft = textareaRef.current.scrollLeft;
    }
  };

  // Handle Alt+Shift+F to format
  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.altKey && e.shiftKey && e.key === 'F') {
      e.preventDefault();
      formatJson();
    }
    // Handle Tab for indentation
    if (e.key === 'Tab') {
      e.preventDefault();
      const textarea = textareaRef.current;
      if (textarea) {
        const start = textarea.selectionStart;
        const end = textarea.selectionEnd;
        const newValue = value.substring(0, start) + '  ' + value.substring(end);
        onChange(newValue);
        setTimeout(() => {
          textarea.selectionStart = textarea.selectionEnd = start + 2;
        }, 0);
      }
    }
  };

  const formatJson = () => {
    try {
      const parsed = JSON.parse(value);
      const formatted = JSON.stringify(parsed, null, 2);
      onChange(formatted);
    } catch {
      // Can't format invalid JSON
    }
  };

  // Syntax highlighting - returns HTML string with colored spans
  const highlightSyntax = (text: string): string => {
    if (!text) return '';
    
    // Escape HTML first
    let html = text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');
    
    // Apply syntax highlighting with regex
    // Order matters - do strings first to avoid highlighting inside strings
    
    // Strings (keys and values)
    html = html.replace(
      /("(?:[^"\\]|\\.)*")\s*:/g,
      '<span class="text-accent-purple">$1</span>:'
    );
    html = html.replace(
      /:\s*("(?:[^"\\]|\\.)*")/g,
      ': <span class="text-green-400">$1</span>'
    );
    // Standalone strings in arrays
    html = html.replace(
      /(\[|,)\s*("(?:[^"\\]|\\.)*")/g,
      '$1 <span class="text-green-400">$2</span>'
    );
    
    // Numbers
    html = html.replace(
      /:\s*(-?\d+\.?\d*(?:[eE][+-]?\d+)?)\b/g,
      ': <span class="text-yellow-400">$1</span>'
    );
    html = html.replace(
      /(\[|,)\s*(-?\d+\.?\d*(?:[eE][+-]?\d+)?)\b/g,
      '$1 <span class="text-yellow-400">$2</span>'
    );
    
    // Booleans
    html = html.replace(
      /:\s*(true|false)\b/g,
      ': <span class="text-blue-400">$1</span>'
    );
    html = html.replace(
      /(\[|,)\s*(true|false)\b/g,
      '$1 <span class="text-blue-400">$2</span>'
    );
    
    // Null
    html = html.replace(
      /:\s*(null)\b/g,
      ': <span class="text-red-400">$1</span>'
    );
    html = html.replace(
      /(\[|,)\s*(null)\b/g,
      '$1 <span class="text-red-400">$2</span>'
    );
    
    // Brackets and braces
    html = html.replace(
      /([{}\[\]])/g,
      '<span class="text-text-secondary">$1</span>'
    );
    
    return html;
  };

  return (
    <div className={className}>
      {/* Editor container with overlay */}
      <div className="relative">
        {/* Syntax highlighted overlay */}
        <div
          ref={highlightRef}
          className={`absolute inset-0 p-3 font-mono text-sm whitespace-pre-wrap break-words overflow-hidden pointer-events-none bg-bg-primary rounded-lg border ${
            isValid ? 'border-border' : 'border-red-500'
          }`}
          aria-hidden="true"
          dangerouslySetInnerHTML={{ __html: highlightSyntax(value) + '<br/>' }}
        />
        
        {/* Actual textarea (transparent text, visible caret) */}
        <textarea
          ref={textareaRef}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onKeyDown={handleKeyDown}
          onScroll={handleScroll}
          rows={6}
          className={`relative w-full p-3 font-mono text-sm bg-transparent border rounded-lg resize-y focus:outline-none transition-colors ${
            isValid ? 'border-border focus:border-accent-purple' : 'border-red-500 focus:border-red-500'
          }`}
          style={{ 
            color: 'transparent',
            caretColor: 'var(--color-text-primary)',
            WebkitTextFillColor: 'transparent',
          }}
          spellCheck={false}
          placeholder='{ "key": "value" }'
        />
      </div>

      {/* Status bar */}
      <div className="flex items-center justify-between mt-2">
        <div className="flex items-center gap-3">
          {isValid ? (
            <span className="flex items-center gap-1 text-xs text-green-400">
              <Check size={12} />
              Valid JSON
            </span>
          ) : (
            <span className="flex items-center gap-1 text-xs text-red-400 max-w-[250px] truncate" title={errorMessage || undefined}>
              <AlertTriangle size={12} />
              {errorMessage || 'Invalid JSON'}
            </span>
          )}
        </div>
        <button
          type="button"
          onClick={formatJson}
          disabled={!isValid}
          className="text-xs text-text-secondary hover:text-accent-purple disabled:opacity-50 disabled:hover:text-text-secondary transition-colors"
        >
          Format (Alt+Shift+F)
        </button>
      </div>
    </div>
  );
}

const OPERATORS: { value: Operator; label: string; description: string }[] = [
  { value: 'eq', label: '=', description: 'equals' },
  { value: 'neq', label: '≠', description: 'not equals' },
  { value: 'gt', label: '>', description: 'greater than' },
  { value: 'gte', label: '≥', description: 'greater or equal' },
  { value: 'lt', label: '<', description: 'less than' },
  { value: 'lte', label: '≤', description: 'less or equal' },
  { value: 'contains', label: 'contains', description: 'contains substring' },
  { value: 'not_contains', label: 'not contains', description: 'does not contain' },
  { value: 'starts_with', label: 'starts with', description: 'starts with' },
  { value: 'ends_with', label: 'ends with', description: 'ends with' },
  { value: 'in', label: 'in', description: 'in list' },
  { value: 'not_in', label: 'not in', description: 'not in list' },
  { value: 'matches', label: 'matches', description: 'regex match' },
  { value: 'exists', label: 'exists', description: 'attribute exists' },
];

type PanelTab = 'details' | 'rules' | 'settings';

export default function FeatureFlagsPage() {
  const { tenantId, appId, envId } = useParams<{ tenantId: string; appId: string; envId: string }>();
  const [flags, setFlags] = useState<FeatureFlag[]>([]);
  const [environment, setEnvironment] = useState<Environment | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedFlag, setSelectedFlag] = useState<FeatureFlag | null>(null);
  const [deletingFlag, setDeletingFlag] = useState<FeatureFlag | null>(null);
  const [filterType, setFilterType] = useState<FlagType | 'all'>('all');
  const [filterEnabled, setFilterEnabled] = useState<'all' | 'enabled' | 'disabled'>('all');

  const fetchData = useCallback(async () => {
    if (!tenantId || !appId || !envId) return;
    
    try {
      setLoading(true);
      const [flagsResponse, envData] = await Promise.all([
        featureFlagsApi.list(tenantId, appId, envId),
        environmentsApi.get(tenantId, appId, envId)
      ]);
      setFlags(flagsResponse.flags || []);
      setEnvironment(envData);
      setError(null);
    } catch (err) {
      setError('Failed to load feature flags');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [tenantId, appId, envId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filteredFlags = flags.filter(flag => {
    const matchesSearch = !search || 
      flag.name.toLowerCase().includes(search.toLowerCase()) ||
      flag.key.toLowerCase().includes(search.toLowerCase()) ||
      flag.tags?.some(t => t.toLowerCase().includes(search.toLowerCase()));
    const matchesType = filterType === 'all' || flag.type === filterType;
    const matchesEnabled = filterEnabled === 'all' || 
      (filterEnabled === 'enabled' && flag.enabled) ||
      (filterEnabled === 'disabled' && !flag.enabled);
    return matchesSearch && matchesType && matchesEnabled;
  });

  const handleToggle = async (flag: FeatureFlag) => {
    if (!tenantId || !appId || !envId) return;
    try {
      const updated = await featureFlagsApi.toggle(tenantId, appId, envId, flag.id, !flag.enabled);
      setFlags(prev => prev.map(f => f.id === flag.id ? updated : f));
      if (selectedFlag?.id === flag.id) {
        setSelectedFlag(updated);
      }
    } catch (err) {
      setError('Failed to toggle flag');
    }
  };

  const handleDelete = async () => {
    if (!deletingFlag || !tenantId || !appId || !envId) return;
    await featureFlagsApi.delete(tenantId, appId, envId, deletingFlag.id);
    setFlags(prev => prev.filter(f => f.id !== deletingFlag.id));
    if (selectedFlag?.id === deletingFlag.id) {
      setSelectedFlag(null);
    }
  };

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key);
  };

  const getTypeIcon = (type: FlagType) => {
    switch (type) {
      case 'boolean': return '🔘';
      case 'string': return '📝';
      case 'number': return '🔢';
      case 'json': return '📋';
    }
  };

  const getTypeColor = (type: FlagType) => {
    switch (type) {
      case 'boolean': return 'text-green-400 bg-green-400/10';
      case 'string': return 'text-blue-400 bg-blue-400/10';
      case 'number': return 'text-yellow-400 bg-yellow-400/10';
      case 'json': return 'text-purple-400 bg-purple-400/10';
    }
  };

  const handleFlagUpdated = (updatedFlag: FeatureFlag) => {
    setFlags(prev => prev.map(f => f.id === updatedFlag.id ? updatedFlag : f));
    setSelectedFlag(updatedFlag);
  };

  return (
    <div className="min-h-screen bg-bg-primary">
      {/* Main Content */}
      <div className="p-6">
        <div className="max-w-7xl mx-auto">
          {/* Header */}
          <div className="flex items-center gap-4 mb-6">
            <Link
              to={`/tenants/${tenantId}/applications/${appId}/environments`}
              className="p-2 hover:bg-bg-secondary rounded-lg transition-colors"
            >
              <ChevronLeft size={20} className="text-text-secondary" />
            </Link>
            <div className="flex-1">
              <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
                <Flag className="text-accent-purple" />
                Feature Flags
              </h1>
              {environment && (
                <div className="flex items-center gap-2 mt-1">
                  <div 
                    className="w-3 h-3 rounded-full" 
                    style={{ backgroundColor: environment.color || '#8b5cf6' }}
                  />
                  <span className="text-text-secondary">{environment.name}</span>
                </div>
              )}
            </div>
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors"
            >
              <Plus size={20} />
              Create Flag
            </button>
          </div>

          {/* Filters */}
          <div className="flex flex-wrap gap-4 mb-6">
            <div className="relative flex-1 min-w-[200px]">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary" size={20} />
              <input
                type="text"
                placeholder="Search flags by name, key, or tag..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full pl-10 pr-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
              />
            </div>
            
            <select
              value={filterType}
              onChange={(e) => setFilterType(e.target.value as FlagType | 'all')}
              className="px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            >
              <option value="all">All Types</option>
              <option value="boolean">Boolean</option>
              <option value="string">String</option>
              <option value="number">Number</option>
              <option value="json">JSON</option>
            </select>

            <select
              value={filterEnabled}
              onChange={(e) => setFilterEnabled(e.target.value as 'all' | 'enabled' | 'disabled')}
              className="px-4 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            >
              <option value="all">All States</option>
              <option value="enabled">Enabled</option>
              <option value="disabled">Disabled</option>
            </select>
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
            /* Flags List */
            <div className="space-y-3">
              {filteredFlags.map(flag => (
                <div
                  key={flag.id}
                  onClick={() => setSelectedFlag(flag)}
                  className={`bg-bg-secondary border rounded-xl p-5 cursor-pointer transition-colors ${
                    selectedFlag?.id === flag.id 
                      ? 'border-accent-purple' 
                      : 'border-border hover:border-accent-purple/50'
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="font-semibold text-text-primary">{flag.name}</h3>
                        <span className={`px-2 py-0.5 rounded text-xs font-medium ${getTypeColor(flag.type)}`}>
                          {getTypeIcon(flag.type)} {flag.type}
                        </span>
                        {flag.version > 1 && (
                          <span className="px-2 py-0.5 rounded text-xs bg-bg-tertiary text-text-secondary">
                            v{flag.version}
                          </span>
                        )}
                      </div>
                      
                      <div className="flex items-center gap-2 mb-3">
                        <code className="text-sm text-text-secondary font-mono bg-bg-primary px-2 py-1 rounded">
                          {flag.key}
                        </code>
                        <button
                          onClick={(e) => { e.stopPropagation(); copyKey(flag.key); }}
                          className="p-1 hover:bg-bg-tertiary rounded transition-colors"
                          title="Copy key"
                        >
                          <Copy size={14} className="text-text-secondary" />
                        </button>
                      </div>

                      {flag.description && (
                        <p className="text-sm text-text-secondary mb-3">{flag.description}</p>
                      )}

                      <div className="flex items-center gap-4 text-sm">
                        <div className="text-text-secondary">
                          Default: <span className="font-mono text-text-primary">{JSON.stringify(flag.default_value)}</span>
                        </div>
                        {flag.tags && flag.tags.length > 0 && (
                          <div className="flex items-center gap-1">
                            <Tag size={12} className="text-text-secondary" />
                            {flag.tags.map(tag => (
                              <span key={tag} className="px-2 py-0.5 bg-bg-tertiary rounded text-xs text-text-secondary">
                                {tag}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center gap-3">
                      <button
                        onClick={(e) => { e.stopPropagation(); handleToggle(flag); }}
                        className={`flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors ${
                          flag.enabled 
                            ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30' 
                            : 'bg-bg-tertiary text-text-secondary hover:bg-bg-primary'
                        }`}
                      >
                        {flag.enabled ? (
                          <>
                            <ToggleRight size={20} />
                            Enabled
                          </>
                        ) : (
                          <>
                            <ToggleLeft size={20} />
                            Disabled
                          </>
                        )}
                      </button>

                      <button
                        onClick={(e) => { e.stopPropagation(); setDeletingFlag(flag); }}
                        className="p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                      >
                        <Trash2 size={18} className="text-red-400" />
                      </button>
                    </div>
                  </div>
                </div>
              ))}

              {filteredFlags.length === 0 && !loading && (
                <div className="text-center py-12 text-text-secondary">
                  {search || filterType !== 'all' || filterEnabled !== 'all'
                    ? 'No flags found matching your filters'
                    : 'No feature flags yet. Create your first flag!'}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Lateral Slide Panel */}
      {selectedFlag && tenantId && appId && envId && (
        <SidePanel
          flag={selectedFlag}
          tenantId={tenantId}
          appId={appId}
          envId={envId}
          onClose={() => setSelectedFlag(null)}
          onFlagUpdated={handleFlagUpdated}
          onDelete={() => setDeletingFlag(selectedFlag)}
        />
      )}

      {/* Create Modal */}
      {showCreateModal && tenantId && appId && envId && (
        <CreateFlagModal
          tenantId={tenantId}
          appId={appId}
          envId={envId}
          onClose={() => setShowCreateModal(false)}
          onSave={fetchData}
        />
      )}

      {/* Delete Confirmation Modal */}
      <ConfirmDeleteModal
        isOpen={!!deletingFlag}
        onClose={() => setDeletingFlag(null)}
        onConfirm={handleDelete}
        title="Delete Feature Flag"
        message="Are you sure you want to delete this feature flag? This action cannot be undone and may affect your applications."
        itemName={deletingFlag?.name}
      />
    </div>
  );
}

// ============================================
// SIDE PANEL COMPONENT
// ============================================
interface SidePanelProps {
  flag: FeatureFlag;
  tenantId: string;
  appId: string;
  envId: string;
  onClose: () => void;
  onFlagUpdated: (flag: FeatureFlag) => void;
  onDelete: () => void;
}

function SidePanel({ flag, tenantId, appId, envId, onClose, onFlagUpdated, onDelete }: SidePanelProps) {
  const [activeTab, setActiveTab] = useState<PanelTab>('details');

  const tabs: { id: PanelTab; label: string; icon: React.ReactNode }[] = [
    { id: 'details', label: 'Value', icon: <Flag size={16} /> },
    { id: 'rules', label: 'Targeting Rules', icon: <Target size={16} /> },
    { id: 'settings', label: 'Settings', icon: <Settings size={16} /> },
  ];

  return (
    <div className="fixed inset-0 z-50">
      {/* Backdrop - semi-transparent with soft blur */}
      <div 
        className="absolute inset-0 bg-black/30 backdrop-blur-[2px]"
        onClick={onClose}
      />
      {/* Panel */}
      <div className="absolute top-0 right-0 h-full w-[500px] bg-bg-secondary border-l border-border shadow-2xl flex flex-col animate-slide-in">
      {/* Panel Header */}
      <div className="p-4 border-b border-border flex items-center justify-between shrink-0">
        <div>
          <h2 className="text-lg font-semibold text-text-primary">Edit Feature: {flag.key}</h2>
          <p className="text-sm text-text-secondary">{flag.name}</p>
        </div>
        <button
          onClick={onClose}
          className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
        >
          <X size={20} className="text-text-secondary" />
        </button>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-border shrink-0">
        {tabs.map(tab => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`flex items-center gap-2 px-4 py-3 text-sm font-medium transition-colors border-b-2 ${
              activeTab === tab.id
                ? 'text-accent-purple border-accent-purple'
                : 'text-text-secondary border-transparent hover:text-text-primary hover:border-border'
            }`}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === 'details' && (
          <DetailsTab
            flag={flag}
            tenantId={tenantId}
            appId={appId}
            envId={envId}
            onFlagUpdated={onFlagUpdated}
          />
        )}
        {activeTab === 'rules' && (
          <RulesTab
            flag={flag}
            tenantId={tenantId}
            appId={appId}
            envId={envId}
          />
        )}
        {activeTab === 'settings' && (
          <SettingsTab
            flag={flag}
            onDelete={onDelete}
          />
        )}
      </div>
      </div>
    </div>
  );
}

// ============================================
// DETAILS TAB
// ============================================
interface DetailsTabProps {
  flag: FeatureFlag;
  tenantId: string;
  appId: string;
  envId: string;
  onFlagUpdated: (flag: FeatureFlag) => void;
}

function DetailsTab({ flag, tenantId, appId, envId, onFlagUpdated }: DetailsTabProps) {
  const [name, setName] = useState(flag.name);
  const [description, setDescription] = useState(flag.description);
  const [defaultValue, setDefaultValue] = useState(JSON.stringify(flag.default_value));
  const [enabled, setEnabled] = useState(flag.enabled);
  const [tags, setTags] = useState(flag.tags?.join(', ') || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    setName(flag.name);
    setDescription(flag.description);
    setDefaultValue(JSON.stringify(flag.default_value));
    setEnabled(flag.enabled);
    setTags(flag.tags?.join(', ') || '');
  }, [flag]);

  const handleSave = async () => {
    setLoading(true);
    setError(null);
    setSuccess(false);

    try {
      let parsedValue: unknown;
      try {
        parsedValue = JSON.parse(defaultValue);
      } catch {
        parsedValue = defaultValue;
      }

      const parsedTags = tags.split(',').map(t => t.trim()).filter(Boolean);

      const updated = await featureFlagsApi.update(tenantId, appId, envId, flag.id, {
        name,
        description,
        default_value: parsedValue,
        enabled,
        tags: parsedTags,
      });
      
      onFlagUpdated(updated);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { error?: string } } };
      setError(axiosError.response?.data?.error || 'Failed to save flag');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-4 space-y-5">
      {/* Enabled Toggle */}
      <div className="flex items-center justify-between p-4 bg-bg-primary rounded-lg">
        <div>
          <span className="text-text-primary font-medium">Enabled</span>
          <p className="text-sm text-text-secondary">Toggle the feature flag on/off</p>
        </div>
        <button
          type="button"
          onClick={() => setEnabled(!enabled)}
          className={`relative w-14 h-7 rounded-full transition-colors ${
            enabled ? 'bg-green-500' : 'bg-bg-tertiary'
          }`}
        >
          <div
            className={`absolute top-1 w-5 h-5 bg-white rounded-full transition-transform ${
              enabled ? 'translate-x-8' : 'translate-x-1'
            }`}
          />
        </button>
      </div>

      {/* Flag Info */}
      <div className="p-4 bg-bg-primary rounded-lg">
        <div className="flex items-center gap-2 mb-2">
          <span className="text-xs font-medium text-text-secondary">KEY</span>
          <code className="text-sm font-mono text-accent-purple">{flag.key}</code>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-text-secondary">TYPE</span>
          <span className="text-sm text-text-primary">{flag.type}</span>
        </div>
      </div>

      {error && (
        <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
          {error}
        </div>
      )}

      {success && (
        <div className="p-3 bg-green-500/10 border border-green-500/50 rounded-lg text-green-400 text-sm">
          Flag updated successfully!
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
          className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-text-secondary mb-2">
          Description
        </label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={2}
          className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple resize-none"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-text-secondary mb-2">
          Default Value
        </label>
        {flag.type === 'json' ? (
          <JsonEditor
            value={defaultValue}
            onChange={setDefaultValue}
          />
        ) : (
          <>
            <input
              type="text"
              value={defaultValue}
              onChange={(e) => setDefaultValue(e.target.value)}
              className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
            />
            <p className="mt-1 text-xs text-text-secondary">
              Enter a valid JSON value (e.g., true, "string", 123)
            </p>
          </>
        )}
      </div>

      <div>
        <label className="block text-sm font-medium text-text-secondary mb-2">
          Tags
        </label>
        <input
          type="text"
          value={tags}
          onChange={(e) => setTags(e.target.value)}
          placeholder="beta, feature, ui"
          className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
        />
        <p className="mt-1 text-xs text-text-secondary">Comma-separated tags</p>
      </div>

      {/* Save Button */}
      <div className="pt-4 border-t border-border">
        <button
          onClick={handleSave}
          disabled={loading}
          className="w-full px-4 py-2.5 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
        >
          {loading && <Loader2 className="animate-spin" size={16} />}
          Update Feature Value
        </button>
      </div>
    </div>
  );
}

// ============================================
// RULES TAB
// ============================================
interface RulesTabProps {
  flag: FeatureFlag;
  tenantId: string;
  appId: string;
  envId: string;
}

function RulesTab({ flag, tenantId, appId, envId }: RulesTabProps) {
  const [rules, setRules] = useState<TargetingRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedRules, setExpandedRules] = useState<Set<string>>(new Set());
  const [editingRule, setEditingRule] = useState<TargetingRule | null>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [deletingRule, setDeletingRule] = useState<TargetingRule | null>(null);

  const fetchRules = useCallback(async () => {
    try {
      setLoading(true);
      const response = await targetingRulesApi.list(tenantId, appId, envId, flag.id);
      setRules(response.rules || []);
      setError(null);
    } catch (err) {
      setError('Failed to load targeting rules');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [tenantId, appId, envId, flag.id]);

  useEffect(() => {
    fetchRules();
  }, [fetchRules]);

  const handleToggle = async (rule: TargetingRule) => {
    try {
      const updated = await targetingRulesApi.update(tenantId, appId, envId, flag.id, rule.id, {
        enabled: !rule.enabled
      });
      setRules(prev => prev.map(r => r.id === rule.id ? updated : r));
    } catch (err) {
      setError('Failed to toggle rule');
    }
  };

  const handleDelete = async () => {
    if (!deletingRule) return;
    await targetingRulesApi.delete(tenantId, appId, envId, flag.id, deletingRule.id);
    setRules(prev => prev.filter(r => r.id !== deletingRule.id));
  };

  const toggleExpand = (ruleId: string) => {
    setExpandedRules(prev => {
      const next = new Set(prev);
      if (next.has(ruleId)) {
        next.delete(ruleId);
      } else {
        next.add(ruleId);
      }
      return next;
    });
  };

  const formatCondition = (condition: Condition) => {
    const op = OPERATORS.find(o => o.value === condition.operator);
    const valueStr = typeof condition.value === 'object' 
      ? JSON.stringify(condition.value)
      : String(condition.value);
    return `${condition.attribute} ${op?.label || condition.operator} ${valueStr}`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="animate-spin text-accent-purple" size={24} />
      </div>
    );
  }

  return (
    <div className="p-4">
      {/* Info Box */}
      <div className="mb-4 p-3 bg-amber-500/10 border border-amber-500/30 rounded-lg">
        <p className="text-sm text-amber-200">
          <strong>NOTE:</strong> Rules are evaluated in priority order. The first matching rule determines the flag value.
        </p>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
          {error}
        </div>
      )}

      {/* Rules List */}
      <div className="space-y-3">
        {rules.length === 0 && !showCreateForm ? (
          <div className="text-center py-8 text-text-secondary">
            <Target size={32} className="mx-auto mb-2 opacity-50" />
            <p>No targeting rules yet.</p>
          </div>
        ) : (
          rules
            .sort((a, b) => a.priority - b.priority)
            .map((rule) => (
              <div
                key={rule.id}
                className={`bg-bg-primary border rounded-lg transition-colors ${
                  rule.enabled ? 'border-border' : 'border-border/50 opacity-60'
                }`}
              >
                {/* Rule Header */}
                <div className="p-3 flex items-center gap-3">
                  <div className="text-text-secondary cursor-move">
                    <GripVertical size={16} />
                  </div>
                  
                  <div className="w-6 h-6 flex items-center justify-center bg-bg-tertiary rounded text-text-secondary font-mono text-xs">
                    {rule.priority}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <h4 className="font-medium text-text-primary text-sm truncate">{rule.name}</h4>
                      <span className={`px-1.5 py-0.5 rounded text-xs ${
                        rule.enabled 
                          ? 'bg-green-500/20 text-green-400'
                          : 'bg-gray-500/20 text-gray-400'
                      }`}>
                        {rule.enabled ? 'On' : 'Off'}
                      </span>
                    </div>
                    <div className="text-xs text-text-secondary mt-0.5">
                      {rule.conditions.length} condition{rule.conditions.length !== 1 ? 's' : ''} → 
                      <span className="font-mono text-accent-purple ml-1">
                        {JSON.stringify(rule.value)}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => toggleExpand(rule.id)}
                      className="p-1.5 hover:bg-bg-tertiary rounded transition-colors"
                    >
                      {expandedRules.has(rule.id) ? (
                        <ChevronUp size={14} className="text-text-secondary" />
                      ) : (
                        <ChevronDown size={14} className="text-text-secondary" />
                      )}
                    </button>

                    <button
                      onClick={() => handleToggle(rule)}
                      className={`p-1.5 rounded transition-colors ${
                        rule.enabled ? 'hover:bg-green-500/10 text-green-400' : 'hover:bg-bg-tertiary text-text-secondary'
                      }`}
                    >
                      {rule.enabled ? <ToggleRight size={14} /> : <ToggleLeft size={14} />}
                    </button>

                    <button
                      onClick={() => setEditingRule(rule)}
                      className="p-1.5 hover:bg-bg-tertiary rounded transition-colors"
                    >
                      <Edit2 size={14} className="text-text-secondary" />
                    </button>

                    <button
                      onClick={() => setDeletingRule(rule)}
                      className="p-1.5 hover:bg-red-500/10 rounded transition-colors"
                    >
                      <Trash2 size={14} className="text-red-400" />
                    </button>
                  </div>
                </div>

                {/* Expanded Conditions */}
                {expandedRules.has(rule.id) && (
                  <div className="px-3 pb-3 pt-0">
                    <div className="border-t border-border pt-3">
                      <h5 className="text-xs font-medium text-text-secondary mb-2">Conditions (ALL must match):</h5>
                      <div className="space-y-1.5">
                        {rule.conditions.map((condition, idx) => (
                          <div key={idx} className="text-xs">
                            <span className="px-2 py-1 bg-bg-tertiary rounded font-mono text-text-primary">
                              {formatCondition(condition)}
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))
        )}
      </div>

      {/* Add Rule Button */}
      <button
        onClick={() => setShowCreateForm(true)}
        className="w-full mt-4 px-4 py-2.5 border-2 border-dashed border-border rounded-lg text-text-secondary hover:border-accent-purple hover:text-accent-purple transition-colors flex items-center justify-center gap-2"
      >
        <Plus size={18} />
        Add Rule
      </button>

      {/* Create/Edit Rule Modal */}
      {(showCreateForm || editingRule) && (
        <RuleFormModal
          rule={editingRule}
          flagType={flag.type}
          tenantId={tenantId}
          appId={appId}
          envId={envId}
          flagId={flag.id}
          onClose={() => {
            setShowCreateForm(false);
            setEditingRule(null);
          }}
          onSave={fetchRules}
        />
      )}

      {/* Delete Modal */}
      <ConfirmDeleteModal
        isOpen={!!deletingRule}
        onClose={() => setDeletingRule(null)}
        onConfirm={handleDelete}
        title="Delete Targeting Rule"
        message={`Are you sure you want to delete the rule "${deletingRule?.name}"?`}
        itemName={deletingRule?.name}
        confirmText="Delete"
      />
    </div>
  );
}

// ============================================
// SETTINGS TAB
// ============================================
interface SettingsTabProps {
  flag: FeatureFlag;
  onDelete: () => void;
}

function SettingsTab({ flag, onDelete }: SettingsTabProps) {
  return (
    <div className="p-4 space-y-6">
      {/* Flag Info */}
      <div className="p-4 bg-bg-primary rounded-lg space-y-3">
        <h3 className="text-sm font-medium text-text-secondary uppercase tracking-wide">Flag Information</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-text-secondary">ID</span>
            <p className="font-mono text-text-primary">{flag.id}</p>
          </div>
          <div>
            <span className="text-text-secondary">Version</span>
            <p className="text-text-primary">v{flag.version}</p>
          </div>
          <div>
            <span className="text-text-secondary">Created</span>
            <p className="text-text-primary">{new Date(flag.created_at).toLocaleDateString()}</p>
          </div>
          <div>
            <span className="text-text-secondary">Updated</span>
            <p className="text-text-primary">{new Date(flag.updated_at).toLocaleDateString()}</p>
          </div>
        </div>
      </div>

      {/* Danger Zone */}
      <div className="p-4 bg-red-500/5 border border-red-500/20 rounded-lg">
        <h3 className="text-sm font-medium text-red-400 uppercase tracking-wide mb-3">Danger Zone</h3>
        <p className="text-sm text-text-secondary mb-4">
          Deleting this flag will remove it permanently. This action cannot be undone.
        </p>
        <button
          onClick={onDelete}
          className="px-4 py-2 bg-red-500/10 text-red-400 border border-red-500/30 rounded-lg hover:bg-red-500/20 transition-colors flex items-center gap-2"
        >
          <Trash2 size={16} />
          Delete Feature Flag
        </button>
      </div>
    </div>
  );
}

// ============================================
// RULE FORM MODAL
// ============================================
interface RuleFormModalProps {
  rule: TargetingRule | null;
  flagType: string;
  tenantId: string;
  appId: string;
  envId: string;
  flagId: string;
  onClose: () => void;
  onSave: () => void;
}

function RuleFormModal({ rule, flagType, tenantId, appId, envId, flagId, onClose, onSave }: RuleFormModalProps) {
  const [name, setName] = useState(rule?.name || '');
  const [priority, setPriority] = useState(rule?.priority ?? 0);
  const [conditions, setConditions] = useState<Condition[]>(rule?.conditions || [{ attribute: '', operator: 'eq', value: '' }]);
  const [value, setValue] = useState(rule?.value !== undefined ? JSON.stringify(rule.value) : 'true');
  const [enabled, setEnabled] = useState(rule?.enabled ?? true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditing = !!rule;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      let parsedValue: unknown;
      try {
        parsedValue = JSON.parse(value);
      } catch {
        parsedValue = value;
      }

      const validConditions = conditions.filter(c => c.attribute.trim() !== '');

      if (validConditions.length === 0) {
        setError('At least one condition is required');
        setLoading(false);
        return;
      }

      const data = {
        name,
        priority,
        conditions: validConditions,
        value: parsedValue,
        enabled,
      };

      if (isEditing) {
        await targetingRulesApi.update(tenantId, appId, envId, flagId, rule.id, data);
      } else {
        await targetingRulesApi.create(tenantId, appId, envId, flagId, data);
      }
      onSave();
      onClose();
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { error?: string } } };
      setError(axiosError.response?.data?.error || 'Failed to save rule');
    } finally {
      setLoading(false);
    }
  };

  const addCondition = () => {
    setConditions([...conditions, { attribute: '', operator: 'eq', value: '' }]);
  };

  const removeCondition = (index: number) => {
    setConditions(conditions.filter((_, i) => i !== index));
  };

  const updateCondition = (index: number, field: keyof Condition, fieldValue: unknown) => {
    setConditions(conditions.map((c, i) => 
      i === index ? { ...c, [field]: fieldValue } : c
    ));
  };

  return (
    <Modal
      title={isEditing ? 'Edit Rule' : 'Create Rule'}
      onClose={onClose}
      maxWidth="xl"
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-text-secondary hover:text-text-primary transition-colors text-sm"
          >
            Cancel
          </button>
          <button
            type="submit"
            form="flag-rule-form"
            disabled={loading}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center gap-2 text-sm"
          >
            {loading && <Loader2 className="animate-spin" size={14} />}
            {isEditing ? 'Save' : 'Create Rule'}
          </button>
        </>
      }
    >
      <form id="flag-rule-form" onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-1.5">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Beta Users"
              required
              className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-1.5">Priority</label>
            <input
              type="number"
              value={priority}
              onChange={(e) => setPriority(parseInt(e.target.value) || 0)}
              min={0}
              className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple"
            />
          </div>
        </div>

        {/* Conditions */}
        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">
            Conditions <span className="font-normal">(ALL must match)</span>
          </label>
          <div className="space-y-2">
            {conditions.map((condition, index) => (
              <div key={index} className="flex items-center gap-2">
                <input
                  type="text"
                  value={condition.attribute}
                  onChange={(e) => updateCondition(index, 'attribute', e.target.value)}
                  placeholder="user.plan"
                  className="flex-1 px-2 py-1.5 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple"
                />
                <select
                  value={condition.operator}
                  onChange={(e) => updateCondition(index, 'operator', e.target.value)}
                  className="px-2 py-1.5 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple"
                >
                  {OPERATORS.map(op => (
                    <option key={op.value} value={op.value}>{op.label}</option>
                  ))}
                </select>
                <input
                  type="text"
                  value={typeof condition.value === 'object' ? JSON.stringify(condition.value) : String(condition.value)}
                  onChange={(e) => {
                    let val: unknown = e.target.value;
                    try { val = JSON.parse(e.target.value); } catch { /* keep string */ }
                    updateCondition(index, 'value', val);
                  }}
                  placeholder="premium"
                  className="flex-1 px-2 py-1.5 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple"
                />
                {conditions.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeCondition(index)}
                    className="p-1.5 hover:bg-red-500/10 rounded transition-colors"
                  >
                    <Trash2 size={14} className="text-red-400" />
                  </button>
                )}
              </div>
            ))}
          </div>
          <button
            type="button"
            onClick={addCondition}
            className="mt-2 text-sm text-accent-purple hover:text-accent-purple/80 transition-colors"
          >
            + Add condition
          </button>
        </div>

        {/* Return Value */}
        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">
            Return Value
          </label>
          <input
            type="text"
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder={flagType === 'boolean' ? 'true or false' : 'Value'}
            required
            className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary font-mono text-sm focus:outline-none focus:border-accent-purple"
          />
        </div>

        {/* Enabled */}
        <div className="flex items-center justify-between p-3 bg-bg-primary rounded-lg">
          <span className="text-text-primary text-sm">Enabled</span>
          <button
            type="button"
            onClick={() => setEnabled(!enabled)}
            className={`relative w-10 h-5 rounded-full transition-colors ${enabled ? 'bg-green-500' : 'bg-bg-tertiary'}`}
          >
            <div className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${enabled ? 'translate-x-5' : 'translate-x-0.5'}`} />
          </button>
        </div>
      </form>
    </Modal>
  );
}

// ============================================
// CREATE FLAG MODAL
// ============================================
interface CreateFlagModalProps {
  tenantId: string;
  appId: string;
  envId: string;
  onClose: () => void;
  onSave: () => void;
}

function CreateFlagModal({ tenantId, appId, envId, onClose, onSave }: CreateFlagModalProps) {
  const [key, setKey] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [type, setType] = useState<FlagType>('boolean');
  const [defaultValue, setDefaultValue] = useState('false');
  const [enabled, setEnabled] = useState(false);
  const [tags, setTags] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      let parsedValue: unknown;
      try {
        parsedValue = JSON.parse(defaultValue);
      } catch {
        parsedValue = defaultValue;
      }

      const parsedTags = tags.split(',').map(t => t.trim()).filter(Boolean);

      await featureFlagsApi.create(tenantId, appId, envId, {
        key,
        name,
        description,
        type,
        default_value: parsedValue,
        enabled,
        tags: parsedTags,
      });
      onSave();
      onClose();
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { error?: string } } };
      setError(axiosError.response?.data?.error || 'Failed to create flag');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (name) {
      setKey(name.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, ''));
    }
  }, [name]);

  useEffect(() => {
    switch (type) {
      case 'boolean': setDefaultValue('false'); break;
      case 'string': setDefaultValue('""'); break;
      case 'number': setDefaultValue('0'); break;
      case 'json': setDefaultValue('{}'); break;
    }
  }, [type]);

  return (
    <Modal
      title="Create Feature Flag"
      onClose={onClose}
      maxWidth="lg"
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-text-secondary hover:text-text-primary transition-colors text-sm"
          >
            Cancel
          </button>
          <button
            type="submit"
            form="create-flag-form"
            disabled={loading}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center gap-2 text-sm"
          >
            {loading && <Loader2 className="animate-spin" size={14} />}
            Create Flag
          </button>
        </>
      }
    >
      <form id="create-flag-form" onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Dark Mode Feature"
            required
            className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Key</label>
          <input
            type="text"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            placeholder="dark_mode_feature"
            required
            pattern="^[a-zA-Z][a-zA-Z0-9_-]*$"
            className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary font-mono text-sm focus:outline-none focus:border-accent-purple"
          />
          <p className="mt-1 text-xs text-text-secondary">Auto-generated from name</p>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Type</label>
          <select
            value={type}
            onChange={(e) => setType(e.target.value as FlagType)}
            className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple"
          >
            <option value="boolean">Boolean</option>
            <option value="string">String</option>
            <option value="number">Number</option>
            <option value="json">JSON</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Enable dark mode for users"
            rows={2}
            className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple resize-none"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Default Value</label>
          {type === 'json' ? (
            <JsonEditor
              value={defaultValue}
              onChange={setDefaultValue}
            />
          ) : (
            <input
              type="text"
              value={defaultValue}
              onChange={(e) => setDefaultValue(e.target.value)}
              required
              className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary font-mono text-sm focus:outline-none focus:border-accent-purple"
            />
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-text-secondary mb-1.5">Tags</label>
          <input
            type="text"
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="beta, feature, ui"
            className="w-full px-3 py-2 bg-bg-primary border border-border rounded-lg text-text-primary text-sm focus:outline-none focus:border-accent-purple"
          />
        </div>

        <div className="flex items-center justify-between p-3 bg-bg-primary rounded-lg">
          <span className="text-text-primary text-sm">Enable on creation</span>
          <button
            type="button"
            onClick={() => setEnabled(!enabled)}
            className={`relative w-10 h-5 rounded-full transition-colors ${enabled ? 'bg-green-500' : 'bg-bg-tertiary'}`}
          >
            <div className={`absolute top-0.5 w-4 h-4 bg-white rounded-full transition-transform ${enabled ? 'translate-x-5' : 'translate-x-0.5'}`} />
          </button>
        </div>
      </form>
    </Modal>
  );
}
