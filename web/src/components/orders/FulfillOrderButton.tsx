import React, { useState } from 'react';
import { api } from '@/lib/api';

interface FulfillOrderButtonProps {
  orderId: string;
  onFulfilled: () => void;
  status: string;
  nextStatus: string;
}

export default function FulfillOrderButton({
  orderId,
  onFulfilled,
  status,
  nextStatus,
}: FulfillOrderButtonProps) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFulfill = async () => {
    setLoading(true);
    setError(null);
    try {
      // CSRF token will be fetched automatically
      await api.post(
        '/orders/fulfill',
        {
          order_id: orderId,
          new_status: nextStatus,
          tracking_code: '1234567890',
        },
        {},
        true
      );
      onFulfilled();
    } catch (err: any) {
      setError(err.message || 'Failed to fulfill order');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'inline-block' }}>
      <button
        className="btn-primary"
        onClick={handleFulfill}
        disabled={loading || status === 'delivered'}
        style={{ minWidth: 100 }}
      >
        {loading
          ? 'Fulfilling...'
          : status === 'delivered'
          ? 'Delivered'
          : 'Mark as ' + nextStatus}
      </button>
      {error && (
        <div
          className="order-fulfill-error"
          style={{ color: 'red', fontSize: 12 }}
        >
          {error}
        </div>
      )}
    </div>
  );
}
