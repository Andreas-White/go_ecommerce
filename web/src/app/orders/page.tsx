"use client";
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import { api } from '../../lib/api';
import Link from 'next/link';
import './page.css';
import DeleteOrderButton from '../../components/orders/DeleteOrderButton';

interface Order {
  id: string;
  total_amount: number;
  status: string;
  payment_status: string;
  created_at: string;
}

export default function OrdersPage() {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  // Remove deleteLoading and deleteError state

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
      return;
    }

    if (user) {
      fetchOrders();
    }
  }, [user, authLoading, router]);

  const fetchOrders = async () => {
    try {
      setLoading(true);
      const userOrders = await api.get<Order[]>('/orders/user');
      setOrders(userOrders || []);
    } catch (error) {
      console.error('Failed to fetch orders:', error);
      setError('Failed to load order history. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteOrder = async (orderId: string) => {
    setOrders((prev) => prev.filter((o) => o.id !== orderId));
  };

  const getStatusBadgeClass = (status: string): string => {
    switch (status.toLowerCase()) {
      case 'pending': return 'status-pending';
      case 'processing': return 'status-processing';
      case 'shipped': return 'status-shipped';
      case 'delivered': return 'status-delivered';
      case 'canceled': return 'status-cancelled';
      case 'canceled_by_user': return 'status-cancelled';
      default: return 'status-default';
    }
  };

  const getPaymentStatusBadgeClass = (status: string): string => {
    switch (status.toLowerCase()) {
      case 'pending': return 'payment-pending';
      case 'paid': return 'payment-paid';
      case 'unpaid': return 'payment-failed';
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

  if (authLoading || loading) {
    return (
      <div className="orders-loading">
        <div className="loading-spinner"></div>
        <p>Loading your orders...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="orders-error">
        <div className="error-icon">❌</div>
        <h2>Error Loading Orders</h2>
        <p>{error}</p>
        <button onClick={fetchOrders} className="btn-primary">
          Try Again
        </button>
      </div>
    );
  }

  return (
    <div className="orders-container">
      <div className="orders-header">
        <h1 className="orders-title">Order History</h1>
        <p className="orders-subtitle">View all your past orders and their current status.</p>
      </div>

      {orders.length === 0 ? (
        <div className="orders-empty">
          <div className="empty-icon">📦</div>
          <h2 className="empty-title">No Orders Yet</h2>
          <p className="empty-text">
            You haven't placed any orders yet. Start shopping to see your order history here!
          </p>
          <Link href="/products" className="btn-primary">
            Browse Products
          </Link>
        </div>
      ) : (
        <div className="orders-content">
          <div className="orders-list">
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
                  <button
                    className="btn-secondary"
                    style={{ minWidth: 100 }}
                    onClick={() => router.push(`/orders/${order.id}`)}
                  >
                    View Details
                  </button>
                  {/* Delete button for pending/processing orders */}
                  {(order.status === 'pending' || order.status === 'processing') && (
                    <DeleteOrderButton
                      orderId={order.id}
                      onDeleted={handleDeleteOrder}
                    />
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
} 