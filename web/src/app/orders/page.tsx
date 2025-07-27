"use client";
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import { api } from '../../lib/api';
import { Button } from '@/components/ui';
import OrderCard from '@/components/orders/OrderCard';
import '@/components/orders/OrderCard.css';
import './page.css';

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
        <Button onClick={fetchOrders} variant="primary">
          Try Again
        </Button>
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
          <Button variant="secondary" onClick={() => router.push('/products')}>
            Browse Products
          </Button>
        </div>
      ) : (
        <div className="orders-content">
          <div className="orders-list">
            {orders.map((order) => (
              <OrderCard
                key={order.id}
                order={order}
                onViewDetails={(orderId) => router.push(`/orders/${orderId}`)}
                onDelete={handleDeleteOrder}
                getStatusBadgeClass={getStatusBadgeClass}
                getPaymentStatusBadgeClass={getPaymentStatusBadgeClass}
                formatDate={formatDate}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
} 