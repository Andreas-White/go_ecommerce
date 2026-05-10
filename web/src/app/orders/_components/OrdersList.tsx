'use client';
import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../../context/AuthContext';
import { api } from '../../../lib/api';
import { Button } from '@/components/ui';
import OrderCard from '@/components/orders/OrderCard';
import '@/components/orders/OrderCard.css';
import '../page.css';
import { useTopProgress } from '@/context/TopProgressContext';
import ListItemSkeleton from '@/components/ui/ListItemSkeleton';

export interface Order {
  id: string;
  total_amount: number;
  status: string;
  payment_status: string;
  created_at: string;
}

interface OrdersListProps {
  initialOrders: Order[];
}

function formatDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}

function OrdersListContent({ initialOrders }: OrdersListProps) {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { start: startProgress, complete: completeProgress } = useTopProgress();
  const [orders, setOrders] = useState<Order[]>(initialOrders);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
      return;
    }
  }, [user, authLoading, router]);

  const handleDeleteOrder = useCallback(async (orderId: string) => {
    setOrders((prev) => prev.filter((o) => o.id !== orderId));
    try {
      startProgress();
      await api.post('/orders/delete', { order_id: orderId }, {}, true);
    } catch {
      const restoredOrders = await api.get<Order[]>('/orders/user');
      setOrders(restoredOrders || []);
    } finally {
      completeProgress();
    }
  }, [startProgress, completeProgress]);

  if (authLoading || loading) {
    return (
      <div className="orders-loading">
        <ListItemSkeleton />
      </div>
    );
  }

  if (error) {
    return (
      <div className="orders-error">
        <div className="error-icon">❌</div>
        <h2>Error Loading Orders</h2>
        <p>{error}</p>
        <Button onClick={() => router.push('/orders')} variant="primary">
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
          <Button variant="secondary" href='/products'>
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
                formatDate={formatDate}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default function OrdersList({ initialOrders }: OrdersListProps) {
  return <OrdersListContent initialOrders={initialOrders} />;
}