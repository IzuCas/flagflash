import { X } from 'lucide-react';
import { ReactNode } from 'react';

const MAX_WIDTH_CLASSES = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  '2xl': 'max-w-2xl',
} as const;

interface ModalProps {
  /** When provided and false, renders nothing (useful for controlled modals). */
  isOpen?: boolean;
  /** If provided, clicking the backdrop or the X button calls this. */
  onClose?: () => void;
  title: ReactNode;
  children: ReactNode;
  /** Rendered in a fixed footer bar below the scrollable body. */
  footer?: ReactNode;
  maxWidth?: keyof typeof MAX_WIDTH_CLASSES;
}

export function Modal({
  isOpen = true,
  onClose,
  title,
  children,
  footer,
  maxWidth = 'md',
}: ModalProps) {
  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className={`bg-bg-secondary border border-border rounded-xl w-full ${MAX_WIDTH_CLASSES[maxWidth]} max-h-[90vh] flex flex-col shadow-2xl shadow-black/50`}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="px-6 py-5 border-b border-border flex items-center justify-between gap-4 shrink-0">
          <div className="text-xl font-semibold text-text-primary">{title}</div>
          {onClose && (
            <button
              type="button"
              onClick={onClose}
              className="p-1.5 hover:bg-bg-tertiary rounded-lg transition-colors text-text-secondary hover:text-text-primary shrink-0"
            >
              <X size={18} />
            </button>
          )}
        </div>

        {/* Scrollable body */}
        <div className="p-6 overflow-y-auto flex-1">
          {children}
        </div>

        {/* Fixed footer */}
        {footer && (
          <div className="px-6 py-4 border-t border-border flex justify-end gap-3 shrink-0">
            {footer}
          </div>
        )}
      </div>
    </div>
  );
}
