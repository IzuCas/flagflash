import { useState, useEffect, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { 
  Users, 
  Plus, 
  Search, 
  Trash2, 
  Edit2,
  ChevronLeft,
  Loader2,
  AlertCircle,
  Filter,
} from 'lucide-react';
import { segmentsApi } from '../../services/flagflash-api';
import { useAuth } from '../../contexts/AuthContext';
import { usePermissions } from '../../hooks/usePermissions';
import { ConfirmDeleteModal, Modal } from '../../components';
import type { Segment, SegmentRule, CreateSegmentRequest } from '../../types/flagflash';

const OPERATORS = [
  { value: 'equals', label: 'equals' },
  { value: 'not_equals', label: 'not equals' },
  { value: 'contains', label: 'contains' },
  { value: 'not_contains', label: 'not contains' },
  { value: 'starts_with', label: 'starts with' },
  { value: 'ends_with', label: 'ends with' },
  { value: 'in', label: 'in list' },
  { value: 'not_in', label: 'not in list' },
  { value: 'greater_than', label: 'greater than' },
  { value: 'less_than', label: 'less than' },
  { value: 'regex', label: 'matches regex' },
];

export default function SegmentsPage() {
  const { tenantId: urlTenantId } = useParams<{ tenantId: string }>();
  const { selectedTenant } = useAuth();
  const { isAtLeast } = usePermissions();
  const isAdmin = isAtLeast('admin');
  const activeTenantId = urlTenantId || selectedTenant?.id || '';
  
  const [segments, setSegments] = useState<Segment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [segmentToDelete, setSegmentToDelete] = useState<Segment | null>(null);
  const [editingSegment, setEditingSegment] = useState<Segment | null>(null);

  const fetchData = useCallback(async () => {
    if (!activeTenantId) {
      setLoading(false);
      return;
    }
    
    try {
      setLoading(true);
      const response = await segmentsApi.list(activeTenantId);
      setSegments(response.segments || []);
      setError(null);
    } catch (err) {
      setError('Failed to load segments');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [activeTenantId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filteredSegments = segments.filter(segment => 
    segment.name.toLowerCase().includes(search.toLowerCase()) ||
    segment.description?.toLowerCase().includes(search.toLowerCase())
  );

  const handleDelete = async () => {
    if (!segmentToDelete || !activeTenantId) return;
    
    await segmentsApi.delete(activeTenantId, segmentToDelete.id);
    setSegments(prev => prev.filter(s => s.id !== segmentToDelete.id));
    setSegmentToDelete(null);
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
              <Users className="text-accent-purple" />
              User Segments
            </h1>
            <p className="text-text-secondary mt-1">
              Define reusable user segments for targeting
            </p>
          </div>
          {isAdmin && (
            <button
              onClick={() => setShowCreateModal(true)}
              disabled={!activeTenantId}
              className="flex items-center gap-2 px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Plus size={20} />
              Create Segment
            </button>
          )}
        </div>

        {/* Search */}
        <div className="relative mb-6">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-text-secondary" size={20} />
          <input
            type="text"
            placeholder="Search segments..."
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
          /* Segments List */
          <div className="space-y-3">
            {filteredSegments.length === 0 ? (
              <div className="text-center py-12 text-text-secondary">
                <Users size={48} className="mx-auto mb-4 opacity-50" />
                <p>No segments found</p>
                <p className="text-sm mt-1">Create a segment to define reusable targeting rules</p>
              </div>
            ) : (
              filteredSegments.map(segment => (
                <div
                  key={segment.id}
                  className="bg-bg-secondary border border-border rounded-xl p-5 hover:border-accent-purple/50 transition-colors"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="font-semibold text-text-primary">{segment.name}</h3>
                        <span className="px-2 py-0.5 bg-accent-purple/20 text-accent-purple rounded text-xs font-medium">
                          {segment.rules?.length || 0} rules
                        </span>
                      </div>
                      
                      {segment.description && (
                        <p className="text-sm text-text-secondary mb-3">{segment.description}</p>
                      )}

                      {segment.rules && segment.rules.length > 0 && (
                        <div className="flex flex-wrap gap-2">
                          {segment.rules.slice(0, 3).map((rule, idx) => (
                            <span
                              key={idx}
                              className="inline-flex items-center gap-1 px-2 py-1 bg-bg-tertiary rounded text-xs text-text-secondary"
                            >
                              <Filter size={12} />
                              {rule.attribute} {rule.operator} {rule.value}
                            </span>
                          ))}
                          {segment.rules.length > 3 && (
                            <span className="px-2 py-1 text-xs text-text-secondary">
                              +{segment.rules.length - 3} more
                            </span>
                          )}
                        </div>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => setEditingSegment(segment)}
                        className="p-2 hover:bg-bg-tertiary rounded-lg transition-colors text-text-secondary hover:text-accent-purple"
                        title="Edit"
                      >
                        <Edit2 size={18} />
                      </button>
                      {isAdmin && (
                        <button
                          onClick={() => setSegmentToDelete(segment)}
                          className="p-2 hover:bg-red-500/10 rounded-lg transition-colors text-text-secondary hover:text-red-400"
                          title="Delete"
                        >
                          <Trash2 size={18} />
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {/* Create Modal */}
        {showCreateModal && (
          <SegmentModal
            tenantId={activeTenantId}
            onClose={() => setShowCreateModal(false)}
            onSave={(segment) => {
              setSegments(prev => [...prev, segment]);
              setShowCreateModal(false);
            }}
          />
        )}

        {/* Edit Modal */}
        {editingSegment && (
          <SegmentModal
            tenantId={activeTenantId}
            segment={editingSegment}
            onClose={() => setEditingSegment(null)}
            onSave={(segment) => {
              setSegments(prev => prev.map(s => s.id === segment.id ? segment : s));
              setEditingSegment(null);
            }}
          />
        )}

        {/* Delete Modal */}
        <ConfirmDeleteModal
          isOpen={!!segmentToDelete}
          onClose={() => setSegmentToDelete(null)}
          onConfirm={handleDelete}
          title="Delete Segment"
          message={`Are you sure you want to delete "${segmentToDelete?.name}"? This may affect targeting rules that use this segment.`}
        />
      </div>
    </div>
  );
}

interface SegmentModalProps {
  tenantId: string;
  segment?: Segment;
  onClose: () => void;
  onSave: (segment: Segment) => void;
}

function SegmentModal({ tenantId, segment, onClose, onSave }: SegmentModalProps) {
  const [name, setName] = useState(segment?.name || '');
  const [description, setDescription] = useState(segment?.description || '');
  const [rules, setRules] = useState<SegmentRule[]>(segment?.rules || []);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) return;

    try {
      setSaving(true);
      setError(null);
      
      const data: CreateSegmentRequest = {
        name,
        description: description || undefined,
        rules,
      };
      
      const result = segment
        ? await segmentsApi.update(tenantId, segment.id, data)
        : await segmentsApi.create(tenantId, data);
      
      onSave(result);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const addRule = () => {
    setRules(prev => [...prev, { attribute: '', operator: 'equals', value: '' }]);
  };

  const updateRule = (index: number, field: keyof SegmentRule, value: string) => {
    setRules(prev => prev.map((rule, i) => 
      i === index ? { ...rule, [field]: value } : rule
    ));
  };

  const removeRule = (index: number) => {
    setRules(prev => prev.filter((_, i) => i !== index));
  };

  return (
    <Modal isOpen onClose={onClose} title={segment ? 'Edit Segment' : 'Create Segment'}>
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
            placeholder="Beta Users"
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-text-primary mb-1">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Describe what this segment represents..."
            rows={2}
            className="w-full px-3 py-2 bg-bg-secondary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple resize-none"
          />
        </div>

        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-sm font-medium text-text-primary">Rules</label>
            <button
              type="button"
              onClick={addRule}
              className="text-sm text-accent-purple hover:text-accent-purple/80 flex items-center gap-1"
            >
              <Plus size={14} />
              Add Rule
            </button>
          </div>
          
          {rules.length === 0 ? (
            <p className="text-sm text-text-secondary italic py-4 text-center border border-dashed border-border rounded-lg">
              No rules defined. Add rules to define segment criteria.
            </p>
          ) : (
            <div className="space-y-3">
              {rules.map((rule, index) => (
                <div key={index} className="flex items-center gap-2 p-3 bg-bg-tertiary rounded-lg">
                  <input
                    type="text"
                    value={rule.attribute}
                    onChange={(e) => updateRule(index, 'attribute', e.target.value)}
                    placeholder="attribute"
                    className="flex-1 px-2 py-1 bg-bg-secondary border border-border rounded text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
                  />
                  <select
                    value={rule.operator}
                    onChange={(e) => updateRule(index, 'operator', e.target.value)}
                    className="px-2 py-1 bg-bg-secondary border border-border rounded text-sm text-text-primary focus:outline-none focus:border-accent-purple"
                  >
                    {OPERATORS.map(op => (
                      <option key={op.value} value={op.value}>{op.label}</option>
                    ))}
                  </select>
                  <input
                    type="text"
                    value={rule.value}
                    onChange={(e) => updateRule(index, 'value', e.target.value)}
                    placeholder="value"
                    className="flex-1 px-2 py-1 bg-bg-secondary border border-border rounded text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-accent-purple"
                  />
                  <button
                    type="button"
                    onClick={() => removeRule(index)}
                    className="p-1 hover:bg-red-500/10 rounded text-text-secondary hover:text-red-400"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              ))}
            </div>
          )}
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
            disabled={saving || !name}
            className="px-4 py-2 bg-accent-purple text-white rounded-lg hover:bg-accent-purple/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {saving && <Loader2 size={16} className="animate-spin" />}
            {segment ? 'Save Changes' : 'Create Segment'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
