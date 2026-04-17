import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  Target, 
  Plus, 
  Edit2, 
  Trash2, 
  ChevronLeft,
  Loader2,
  AlertCircle,
  GripVertical,
  ToggleLeft,
  ToggleRight,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';
import { targetingRulesApi, featureFlagsApi } from '../../services/flagflash-api';
import type { TargetingRule, FeatureFlag, Condition, Operator } from '../../types/flagflash';
import { ConfirmDeleteModal, Modal } from '../../components';
import { usePermissions } from '../../hooks/usePermissions';

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

export default function TargetingRulesPage() {
  const { tenantId, appId, envId, flagId } = useParams<{ 
    tenantId: string; 
    appId: string; 
    envId: string;
    flagId: string;
  }>();
  const { canCreateTargetingRule, canUpdateTargetingRule, canDeleteTargetingRule } = usePermissions();
  const [rules, setRules] = useState<TargetingRule[]>([]);
  const [flag, setFlag] = useState<FeatureFlag | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingRule, setEditingRule] = useState<TargetingRule | null>(null);
  const [deletingRule, setDeletingRule] = useState<TargetingRule | null>(null);
  const [expandedRules, setExpandedRules] = useState<Set<string>>(new Set());

  const fetchData = useCallback(async () => {
    if (!tenantId || !appId || !envId || !flagId) return;
    
    try {
      setLoading(true);
      const [rulesResponse, flagData] = await Promise.all([
        targetingRulesApi.list(tenantId, appId, envId, flagId),
        featureFlagsApi.get(tenantId, appId, envId, flagId)
      ]);
      setRules(rulesResponse.rules || []);
      setFlag(flagData);
      setError(null);
    } catch (err) {
      setError('Failed to load targeting rules');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [tenantId, appId, envId, flagId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleDelete = async () => {
    if (!deletingRule || !tenantId || !appId || !envId || !flagId) return;
    
    await targetingRulesApi.delete(tenantId, appId, envId, flagId, deletingRule.id);
    setRules(prev => prev.filter(r => r.id !== deletingRule.id));
  };

  const handleToggle = async (rule: TargetingRule) => {
    if (!tenantId || !appId || !envId || !flagId) return;
    
    try {
      const updated = await targetingRulesApi.update(tenantId, appId, envId, flagId, rule.id, {
        enabled: !rule.enabled
      });
      setRules(prev => prev.map(r => r.id === rule.id ? updated : r));
    } catch (err) {
      setError('Failed to toggle rule');
    }
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

  return (
    <div className="min-h-screen bg-bg-primary p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          <Link
            to={`/tenants/${tenantId}/applications/${appId}/environments/${envId}/flags`}
            className="p-2 hover:bg-bg-secondary rounded-lg transition-colors"
          >
            <ChevronLeft size={20} className="text-text-secondary" />
          </Link>
          <div className="flex-1">
            <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
              <Target className="text-accent-purple" />
              Targeting Rules
            </h1>
            {flag && (
              <p className="text-text-secondary mt-1">
                {flag.name} <span className="font-mono text-xs bg-bg-tertiary px-2 py-0.5 rounded ml-2">{flag.key}</span>
              </p>
            )}
          </div>
          {canCreateTargetingRule && (
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors"
            >
              <Plus size={20} />
              Add Rule
            </button>
          )}
        </div>

        {/* Info Box */}
        <div className="mb-6 p-4 bg-bg-secondary border border-border rounded-lg">
          <p className="text-sm text-text-secondary">
            Rules are evaluated in priority order (lower number = higher priority). 
            The first matching rule determines the flag value. If no rules match, the default value is used.
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
          /* Rules List */
          <div className="space-y-3">
            {rules.length === 0 ? (
              <div className="text-center py-12 bg-bg-secondary border border-border rounded-xl">
                <Target size={48} className="mx-auto text-text-secondary mb-4" />
                <p className="text-text-secondary">No targeting rules yet.</p>
                <p className="text-text-secondary text-sm mt-1">Add a rule to customize flag behavior based on user attributes.</p>
              </div>
            ) : (
              rules
                .sort((a, b) => a.priority - b.priority)
                .map((rule) => (
                  <div
                    key={rule.id}
                    className={`bg-bg-secondary border rounded-xl transition-colors ${
                      rule.enabled ? 'border-border hover:border-accent-purple/50' : 'border-border/50 opacity-60'
                    }`}
                  >
                    {/* Rule Header */}
                    <div className="p-4 flex items-center gap-4">
                      <div className="text-text-secondary cursor-move">
                        <GripVertical size={18} />
                      </div>
                      
                      <div className="w-8 h-8 flex items-center justify-center bg-bg-tertiary rounded-lg text-text-secondary font-mono text-sm">
                        {rule.priority}
                      </div>

                      <div className="flex-1">
                        <div className="flex items-center gap-3">
                          <h3 className="font-semibold text-text-primary">{rule.name}</h3>
                          <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                            rule.enabled 
                              ? 'bg-green-500/20 text-green-400'
                              : 'bg-gray-500/20 text-gray-400'
                          }`}>
                            {rule.enabled ? 'Active' : 'Inactive'}
                          </span>
                        </div>
                        <div className="text-sm text-text-secondary mt-1">
                          {rule.conditions.length} condition{rule.conditions.length !== 1 ? 's' : ''} → 
                          <span className="font-mono text-accent-purple ml-1">
                            {JSON.stringify(rule.value)}
                          </span>
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <button
                          onClick={() => toggleExpand(rule.id)}
                          className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
                          title="Expand"
                        >
                          {expandedRules.has(rule.id) ? (
                            <ChevronUp size={18} className="text-text-secondary" />
                          ) : (
                            <ChevronDown size={18} className="text-text-secondary" />
                          )}
                        </button>

                        {canUpdateTargetingRule && (
                          <button
                            onClick={() => handleToggle(rule)}
                            className={`p-2 rounded-lg transition-colors ${
                              rule.enabled 
                                ? 'hover:bg-green-500/10 text-green-400' 
                                : 'hover:bg-bg-tertiary text-text-secondary'
                            }`}
                            title={rule.enabled ? 'Disable' : 'Enable'}
                          >
                            {rule.enabled ? <ToggleRight size={18} /> : <ToggleLeft size={18} />}
                          </button>
                        )}

                        {canUpdateTargetingRule && (
                          <button
                            onClick={() => setEditingRule(rule)}
                            className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors"
                            title="Edit"
                          >
                            <Edit2 size={18} className="text-text-secondary" />
                          </button>
                        )}

                        {canDeleteTargetingRule && (
                          <button
                            onClick={() => setDeletingRule(rule)}
                            className="p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                            title="Delete"
                          >
                            <Trash2 size={18} className="text-red-400" />
                          </button>
                        )}
                      </div>
                    </div>

                    {/* Expanded Conditions */}
                    {expandedRules.has(rule.id) && (
                      <div className="px-4 pb-4 pt-0">
                        <div className="border-t border-border pt-4">
                          <h4 className="text-sm font-medium text-text-secondary mb-2">Conditions (ALL must match):</h4>
                          <div className="space-y-2">
                            {rule.conditions.map((condition, idx) => (
                              <div key={idx} className="flex items-center gap-2 text-sm">
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
        )}

        {/* Create/Edit Modal */}
        {(showCreateModal || editingRule) && tenantId && appId && envId && flagId && (
          <RuleModal
            rule={editingRule}
            flagType={flag?.type || 'boolean'}
            tenantId={tenantId}
            appId={appId}
            envId={envId}
            flagId={flagId}
            onClose={() => {
              setShowCreateModal(false);
              setEditingRule(null);
            }}
            onSave={fetchData}
          />
        )}

        {/* Delete Modal */}
        <ConfirmDeleteModal
          isOpen={!!deletingRule}
          onClose={() => setDeletingRule(null)}
          onConfirm={handleDelete}
          title="Delete Targeting Rule"
          message={`Are you sure you want to delete the rule "${deletingRule?.name}"? This action cannot be undone.`}
          itemName={deletingRule?.name}
          confirmText="Delete"
        />
      </div>
    </div>
  );
}

interface RuleModalProps {
  rule: TargetingRule | null;
  flagType: string;
  tenantId: string;
  appId: string;
  envId: string;
  flagId: string;
  onClose: () => void;
  onSave: () => void;
}

function RuleModal({ rule, flagType, tenantId, appId, envId, flagId, onClose, onSave }: RuleModalProps) {
  const [name, setName] = useState(rule?.name || '');
  const [priority, setPriority] = useState(rule?.priority ?? 0);
  const [conditions, setConditions] = useState<Condition[]>(rule?.conditions || [{ attribute: '', operator: 'eq', value: '' }]);
  const [boolValue, setBoolValue] = useState<boolean>(() => {
    if (rule?.value !== undefined) {
      return rule.value === true || rule.value === 'true';
    }
    return true;
  });
  const [stringValue, setStringValue] = useState<string>(() => {
    if (rule?.value !== undefined && typeof rule.value === 'string') {
      return rule.value;
    }
    return '';
  });
  const [numberValue, setNumberValue] = useState<number>(() => {
    if (rule?.value !== undefined && typeof rule.value === 'number') {
      return rule.value;
    }
    return 0;
  });
  const [jsonValue, setJsonValue] = useState<string>(() => {
    if (rule?.value !== undefined) {
      return typeof rule.value === 'object' ? JSON.stringify(rule.value, null, 2) : JSON.stringify(rule.value);
    }
    return '{}';
  });
  const [enabled, setEnabled] = useState(rule?.enabled ?? true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const isEditing = !!rule;

  const getValueByType = (): unknown => {
    switch (flagType) {
      case 'boolean':
        return boolValue;
      case 'string':
        return stringValue;
      case 'number':
        return numberValue;
      case 'json':
        try {
          return JSON.parse(jsonValue);
        } catch {
          return jsonValue;
        }
      default:
        return boolValue;
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const parsedValue = getValueByType();

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
      title={isEditing ? 'Edit Targeting Rule' : 'Create Targeting Rule'}
      onClose={onClose}
      maxWidth="2xl"
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
            form="rule-form"
            disabled={loading}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 flex items-center gap-2"
          >
            {loading && <Loader2 className="animate-spin" size={16} />}
            {isEditing ? 'Save Changes' : 'Create Rule'}
          </button>
        </>
      }
    >
      <form id="rule-form" onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Beta Users"
              required
              className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Priority
            </label>
            <input
              type="number"
              value={priority}
              onChange={(e) => setPriority(parseInt(e.target.value) || 0)}
              min={0}
              className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple"
            />
            <p className="mt-1 text-xs text-text-secondary">Lower = higher priority</p>
          </div>
        </div>

        {/* Conditions */}
        <div>
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Conditions <span className="text-text-secondary font-normal">(ALL must match)</span>
          </label>
          <div className="space-y-3">
            {conditions.map((condition, index) => (
              <div key={index} className="p-3 bg-bg-primary border border-border rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs text-text-secondary">Condition {index + 1}</span>
                  <button
                    type="button"
                    onClick={() => removeCondition(index)}
                    disabled={conditions.length <= 1}
                    className="p-1.5 hover:bg-red-500/10 rounded transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                    title={conditions.length <= 1 ? 'At least one condition required' : 'Remove condition'}
                  >
                    <Trash2 size={14} className="text-red-400" />
                  </button>
                </div>
                <div className="grid grid-cols-3 gap-2">
                  <input
                    type="text"
                    value={condition.attribute}
                    onChange={(e) => updateCondition(index, 'attribute', e.target.value)}
                    placeholder="attribute"
                    className="px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple text-sm"
                  />
                  <select
                    value={condition.operator}
                    onChange={(e) => updateCondition(index, 'operator', e.target.value)}
                    className="px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary focus:outline-none focus:border-accent-purple text-sm"
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
                      try {
                        val = JSON.parse(e.target.value);
                      } catch {
                        // Keep as string
                      }
                      updateCondition(index, 'value', val);
                    }}
                    placeholder="value"
                    className="px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple text-sm"
                  />
                </div>
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
          <label className="block text-sm font-medium text-text-secondary mb-2">
            Return Value <span className="text-xs text-text-secondary font-normal ml-2">Type: {flagType}</span>
          </label>
          
          {flagType === 'boolean' && (
            <div className="flex items-center gap-4">
              <button
                type="button"
                onClick={() => setBoolValue(true)}
                className={`flex-1 px-4 py-3 rounded-lg border-2 transition-colors ${
                  boolValue 
                    ? 'border-green-500 bg-green-500/10 text-green-400' 
                    : 'border-border bg-bg-primary text-text-secondary hover:border-border/80'
                }`}
              >
                <span className="font-mono">true</span>
              </button>
              <button
                type="button"
                onClick={() => setBoolValue(false)}
                className={`flex-1 px-4 py-3 rounded-lg border-2 transition-colors ${
                  !boolValue 
                    ? 'border-red-500 bg-red-500/10 text-red-400' 
                    : 'border-border bg-bg-primary text-text-secondary hover:border-border/80'
                }`}
              >
                <span className="font-mono">false</span>
              </button>
            </div>
          )}

          {flagType === 'string' && (
            <input
              type="text"
              value={stringValue}
              onChange={(e) => setStringValue(e.target.value)}
              placeholder="Enter string value"
              required
              className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
            />
          )}

          {flagType === 'number' && (
            <input
              type="number"
              value={numberValue}
              onChange={(e) => setNumberValue(parseFloat(e.target.value) || 0)}
              placeholder="Enter number value"
              required
              className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
            />
          )}

          {flagType === 'json' && (
            <div className="space-y-2">
              <div className="relative">
                <textarea
                  value={jsonValue}
                  onChange={(e) => setJsonValue(e.target.value)}
                  placeholder='{"key": "value"}'
                  required
                  rows={8}
                  className="w-full px-4 py-3 bg-bg-primary border border-border rounded-lg text-text-primary font-mono text-sm placeholder:text-text-secondary focus:outline-none focus:border-accent-purple resize-y"
                  style={{ tabSize: 2 }}
                />
                <button
                  type="button"
                  onClick={() => {
                    try {
                      const parsed = JSON.parse(jsonValue);
                      setJsonValue(JSON.stringify(parsed, null, 2));
                    } catch {
                      // Invalid JSON, do nothing
                    }
                  }}
                  className="absolute top-2 right-2 px-2 py-1 text-xs bg-bg-secondary border border-border rounded hover:bg-bg-tertiary transition-colors text-text-secondary"
                >
                  Format
                </button>
              </div>
              {(() => {
                try {
                  JSON.parse(jsonValue);
                  return <p className="text-xs text-green-400">✓ Valid JSON</p>;
                } catch {
                  return <p className="text-xs text-red-400">✗ Invalid JSON</p>;
                }
              })()}
            </div>
          )}

          {!['boolean', 'string', 'number', 'json'].includes(flagType) && (
            <div className="flex items-center gap-4">
              <button
                type="button"
                onClick={() => setBoolValue(true)}
                className={`flex-1 px-4 py-3 rounded-lg border-2 transition-colors ${
                  boolValue 
                    ? 'border-green-500 bg-green-500/10 text-green-400' 
                    : 'border-border bg-bg-primary text-text-secondary hover:border-border/80'
                }`}
              >
                <span className="font-mono">true</span>
              </button>
              <button
                type="button"
                onClick={() => setBoolValue(false)}
                className={`flex-1 px-4 py-3 rounded-lg border-2 transition-colors ${
                  !boolValue 
                    ? 'border-red-500 bg-red-500/10 text-red-400' 
                    : 'border-border bg-bg-primary text-text-secondary hover:border-border/80'
                }`}
              >
                <span className="font-mono">false</span>
              </button>
            </div>
          )}
        </div>

        {/* Enabled Toggle */}
        <div className="flex items-center justify-between p-3 bg-bg-primary rounded-lg">
          <span className="text-text-primary">Enabled</span>
          <button
            type="button"
            onClick={() => setEnabled(!enabled)}
            className={`relative w-12 h-6 rounded-full transition-colors ${
              enabled ? 'bg-green-500' : 'bg-bg-tertiary'
            }`}
          >
            <div
              className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform ${
                enabled ? 'translate-x-7' : 'translate-x-1'
              }`}
            />
          </button>
        </div>
      </form>
    </Modal>
  );
}
