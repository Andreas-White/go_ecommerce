import React from 'react';
import Link from 'next/link';
import './OrderList.css';

interface Order {
  id: string;
  total_amount: number;
  status: string;
  payment_status: string;
  created_at: string;
}

interface OrderListProps {
  orders: Order[];
  loading?: boolean;
}

const getStatusBadgeClass = (status: string): string => {
  switch (status.toLowerCase()) {
    case 'pending': return 'status-pending';
    case 'processing': return 'status-processing';
    case 'shipped': return 'status-shipped';
    case 'delivered': return 'status-delivered';
    case 'cancelled': return 'status-cancelled';
    default: return 'status-default';
  }
};

const getPaymentStatusBadgeClass = (status: string): string => {
  switch (status.toLowerCase()) {
    case 'pending': return 'payment-pending';
    case 'paid': return 'payment-paid';
    case 'failed': return 'payment-failed';
    case 'refunded': return 'payment-refunded';
    default: return 'payment-default';
  }
};

const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

export default function OrderList({ orders, loading = false }: OrderListProps) {
  if (loading) {
    return (
      <div className="order-list-loading">
        <div className="loading-spinner"></div>
        <p>Loading orders...</p>
      </div>
    );
  }

  if (orders.length === 0) {
    return (
      <div className="order-list-empty">
        <div className="empty-icon">📦</div>
        <h2 className="empty-title">No Orders Found</h2>
        <p className="empty-text">
          No orders match your current criteria.
        </p>
      </div>
    );
  }

  return (
    <div className="order-list">
      {orders.map((order) => (
        <div key={order.id} className="order-card">
          <div className="order-header">
            <div className="order-info">
              <h3 className="order-id">Order #{order.id.slice(0, 8)}</h3>
              <p className="order-date">{formatDate(order.created_at)}</p>
            </div>
            <div className="order-amount">
              ${order.total_amount.toFixed(2)}
            </div>
          </div>
          
          <div className="order-status">
            <div className="status-section">
              <span className="status-label">Order Status:</span>
              <span className={`status-badge ${getStatusBadgeClass(order.status)}`}>
                {order.status}
              </span>
            </div>
            <div className="status-section">
              <span className="status-label">Payment Status:</span>
              <span className={`status-badge ${getPaymentStatusBadgeClass(order.payment_status)}`}>
                {order.payment_status}
              </span>
            </div>
          </div>

          <div className="order-actions">
            <Link href={`/orders/${order.id}`} className="btn-secondary">
              View Details
            </Link>
          </div>
        </div>
      ))}
    </div>
  );
} 