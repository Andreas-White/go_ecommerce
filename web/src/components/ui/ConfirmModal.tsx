import React from 'react';
import ReactDOM from 'react-dom';
import { Button } from '@/components/ui';

interface ConfirmModalProps {
  open: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmLabel?: string;
  cancelLabel?: string;
  loading?: boolean;
  confirmClassName?: string;
}

const modalContent = ({
  title,
  message,
  onConfirm,
  onCancel,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  loading = false,
  confirmClassName = '',
}: Omit<ConfirmModalProps, 'open'>) => (
  <div className="modal-overlay" tabIndex={-1}>
    <div className="modal-dialog" role="dialog" aria-modal="true">
      <h3>{title}</h3>
      <p>{message}</p>
      <div style={{ display: 'flex', gap: 12, marginTop: 16 }}>
        <Button variant="secondary" className="btn-cancel" onClick={onCancel} disabled={loading}>
          {cancelLabel}
        </Button>
        <Button variant="secondary" className={confirmClassName} onClick={onConfirm} disabled={loading} style={{ minWidth: 100 }}>
          {loading ? 'Processing...' : confirmLabel}
        </Button>
      </div>
    </div>
  </div>
);

export default function ConfirmModal(props: ConfirmModalProps) {
  if (!props.open) return null;
  return ReactDOM.createPortal(
    <>
      <div className="modal-backdrop" onClick={props.onCancel} />
      {modalContent(props)}
    </>,
    typeof window !== 'undefined' && window.document.body ? window.document.body : document.body
  );
} 