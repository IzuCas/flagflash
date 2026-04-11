import { AlertTriangle, Loader2 } from 'lucide-react';
import { useState } from 'react';
import { Modal } from './Modal';

interface ConfirmDeleteModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => Promise<void>;
  title?: string;
  message?: string;
  itemName?: string;
  confirmText?: string;
  cancelText?: string;
  requireConfirmation?: boolean;
  confirmationWord?: string;
}

export function ConfirmDeleteModal({
  isOpen,
  onClose,
  onConfirm,
  title = 'Confirm Delete',
  message = 'Are you sure you want to delete this item? This action cannot be undone.',
  itemName,
  confirmText = 'Delete',
  cancelText = 'Cancel',
  requireConfirmation = false,
  confirmationWord = 'DELETE',
}: ConfirmDeleteModalProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [confirmInput, setConfirmInput] = useState('');

  if (!isOpen) return null;

  const canDelete = !requireConfirmation || confirmInput === confirmationWord;

  const handleConfirm = async () => {
    if (!canDelete) return;
    
    setLoading(true);
    setError(null);
    
    try {
      await onConfirm();
      setConfirmInput('');
      onClose();
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { error?: string } } };
      setError(axiosError.response?.data?.error || 'Failed to delete. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (loading) return;
    setConfirmInput('');
    setError(null);
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={
        <div className="flex items-center gap-3">
          <div className="p-2 bg-red-500/10 rounded-lg">
            <AlertTriangle className="text-red-500" size={24} />
          </div>
          {title}
        </div>
      }
      footer={
        <>
          <button
            onClick={handleClose}
            disabled={loading}
            className="px-4 py-2 bg-bg-tertiary border border-border rounded-lg text-text-primary hover:bg-border transition-colors disabled:opacity-50"
          >
            {cancelText}
          </button>
          <button
            onClick={handleConfirm}
            disabled={loading || !canDelete}
            className="px-4 py-2 bg-red-600 border border-red-600 rounded-lg text-white hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {loading && <Loader2 className="animate-spin" size={16} />}
            {confirmText}
          </button>
        </>
      }
    >
      <div className="space-y-4">
        {error && (
          <div className="p-3 bg-red-500/10 border border-red-500/50 rounded-lg text-red-400 text-sm">
            {error}
          </div>
        )}

        <p className="text-text-secondary">
          {message}
        </p>

        {itemName && (
          <div className="p-3 bg-bg-primary border border-border rounded-lg">
            <p className="text-sm text-text-secondary">Item to be deleted:</p>
            <p className="font-medium text-text-primary mt-1">{itemName}</p>
          </div>
        )}

        {requireConfirmation && (
          <div className="space-y-2">
            <p className="text-sm text-text-secondary">
              Type <span className="font-mono text-red-400 font-semibold">{confirmationWord}</span> to confirm:
            </p>
            <input
              type="text"
              value={confirmInput}
              onChange={(e) => setConfirmInput(e.target.value)}
              placeholder={confirmationWord}
              className="w-full px-4 py-2 bg-bg-primary border border-border rounded-lg text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-red-500"
              autoFocus
            />
          </div>
        )}
      </div>
    </Modal>
  );
}

export default ConfirmDeleteModal;
