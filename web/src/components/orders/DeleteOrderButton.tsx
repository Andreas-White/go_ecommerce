import React, { useState } from 'react';
import { api } from '@/lib/api';
import ConfirmModal from '../ui/ConfirmModal';

interface DeleteOrderButtonProps {
  orderId: string;
  disabled?: boolean;
  onDeleted: (orderId: string) => void;
}

export default function DeleteOrderButton({ orderId, disabled, onDeleted }: DeleteOrderButtonProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showModal, setShowModal] = useState(false);

  const handleDelete = async () => {
    setLoading(true);
    setError(null);
    try {
      await api.getCSRFToken('/users/register');
      await api.post(
        '/orders/delete',
        { order_id: orderId },
        {},
        true
      );
      onDeleted(orderId);
      setShowModal(false);
    } catch (err: any) {
      setError(err.message || 'Failed to delete order');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="delete-order-btn-wrapper">
      <button
        className="btn-secondary delete-order-btn"
        onClick={() => setShowModal(true)}
        disabled={loading || disabled}
        style={{ minWidth: 100 }}
      >
        {loading ? 'Deleting...' : 'Delete Order'}
      </button>
      <ConfirmModal
        open={showModal}
        title="Delete Order"
        message="Are you sure you want to delete this order? This action cannot be undone."
        onConfirm={handleDelete}
        onCancel={() => setShowModal(false)}
        loading={loading}
        confirmLabel="Delete Order"
        confirmClassName="delete-order-btn"
      />
      {error && !showModal && (
        <div className="order-delete-error">
          {error}
        </div>
      )}
    </div>
  );
} 