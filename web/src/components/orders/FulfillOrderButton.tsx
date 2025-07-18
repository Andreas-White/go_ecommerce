import React, { useState } from 'react';
import { api } from '@/lib/api';

interface FulfillOrderButtonProps {
  orderId: string;
  onFulfilled: () => void;
}

export default function FulfillOrderButton({ orderId, onFulfilled }: FulfillOrderButtonProps) {
  const [loading, setLoading] = useState(false);
  const [fulfilled, setFulfilled] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFulfill = async () => {
    setLoading(true);
    setError(null);
    try {
      // Get CSRF token and ensure it's in the cookie
      await api.getCSRFToken('/users/register');
      const csrfToken = document.cookie
        .split(';')
        .map(c => c.trim())
        .find(c => c.startsWith('csrf_token='))?.split('=')[1] || '';
      const res = await fetch('/orders/fulfill', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': csrfToken,
        },
        credentials: 'include',
        body: JSON.stringify({ order_id: orderId, new_status: 'accepted', tracking_code: '1234567890' }),
      });
      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.message || 'Failed to fulfill order');
      }
      setFulfilled(true);
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
        disabled={loading || fulfilled}
        style={{ minWidth: 100 }}
      >
        {loading ? 'Fulfilling...' : fulfilled ? 'Fulfilled' : 'Mark as Fulfilled'}
      </button>
      {error && <div className="order-fulfill-error" style={{ color: 'red', fontSize: 12 }}>{error}</div>}
    </div>
  );
} 