import React, { useState } from 'react';
import { api } from '@/lib/api';
import ConfirmModal from '../ui/ConfirmModal';

interface CancelOrderButtonProps {
  orderId: string;
  status: string;
  onCanceled: (orderId: string) => void;
}

export default function CancelOrderButton({ orderId, status, onCanceled }: CancelOrderButtonProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showModal, setShowModal] = useState(false);

  if (status !== 'processing' && status !== 'accepted') {
    return null;
  }

  const handleCancel = async () => {
    setLoading(true);
    setError(null);
    try {
      await api.getCSRFToken('/users/register');
      await api.post(
        '/orders/cancel',
        { order_id: orderId },
        {},
        true
      );
      onCanceled(orderId);
      setShowModal(false);
    } catch (err: any) {
      setError(err.message || 'Failed to cancel order');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'inline-block', marginLeft: 8 }}>
      <button
        className="btn-secondary delete-order-btn"
        onClick={() => setShowModal(true)}
        disabled={loading}
        style={{ minWidth: 100 }}
      >
        {loading ? 'Cancelling...' : 'Cancel Order'}
      </button>
      <ConfirmModal
        open={showModal}
        title="Cancel Order"
        message="Are you sure you want to cancel this order? This action cannot be undone."
        onConfirm={handleCancel}
        onCancel={() => setShowModal(false)}
        loading={loading}
        confirmLabel="Cancel Order"
        confirmClassName="delete-order-btn"
      />
      {error && !showModal && (
        <div className="order-cancel-error" style={{ color: 'red', fontSize: 12 }}>
          {error}
        </div>
      )}
    </div>
  );
} 