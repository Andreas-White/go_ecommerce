import React from 'react';
import FulfillOrderButton from './FulfillOrderButton';
import './OrderList.css';

// Accept both flat and nested order structures
export type ProducerOrderLike = {
  id?: string;
  user_id?: string;
  total_amount?: number;
  status?: string;
  payment_status?: string;
  created_at?: string;
  updated_at?: string;
  order?: {
    id: string;
    total_amount: number;
    status: string;
    payment_status: string;
    created_at: string;
    updated_at: string;
  };
  payment?: {
    amount: number;
  };
};

interface ProducerOrderListProps {
  orders: ProducerOrderLike[];
  onOrderFulfilled: (orderId: string) => void;
}

const statusDisplay: Record<string, string> = {
  pending: 'Pending',
  paid: 'Paid',
  shipped: 'Shipped',
  fulfilled: 'Fulfilled',
  cancelled: 'Cancelled',
};

export default function ProducerOrderList({ orders, onOrderFulfilled }: ProducerOrderListProps) {
  if (!orders.length) {
    return <div className="order-list-empty">No orders for your products yet.</div>;
  }

  function getOrderTotal(order: ProducerOrderLike): string {
    if (typeof order.total_amount === 'number') {
      return `$${order.total_amount.toFixed(2)}`;
    }
    if (order.order && typeof order.order.total_amount === 'number') {
      return `$${order.order.total_amount.toFixed(2)}`;
    }
    if (order.payment && typeof order.payment.amount === 'number') {
      return `$${order.payment.amount.toFixed(2)}`;
    }
    return 'N/A';
  }

  function getOrderId(order: ProducerOrderLike): string {
    return order.id || order.order?.id || '';
  }

  function getOrderStatus(order: ProducerOrderLike): string {
    return order.status || order.order?.status || '';
  }

  function getOrderCreatedAt(order: ProducerOrderLike): string {
    return order.created_at || order.order?.created_at || '';
  }

  function isActionable(order: ProducerOrderLike): boolean {
    const status = getOrderStatus(order);
    return status !== 'shipped' && status !== 'fulfilled' && status !== 'cancelled';
  }

  return (
    <div className="order-list">
      <table className="order-list-table">
        <thead>
          <tr>
            <th>Order ID</th>
            <th>Status</th>
            <th>Total</th>
            <th>Date</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {orders.map(order => {
            const orderId = getOrderId(order);
            const status = getOrderStatus(order);
            const createdAt = getOrderCreatedAt(order);
            return (
              <tr key={orderId}>
                <td>{orderId}</td>
                <td>{statusDisplay[status] || status}</td>
                <td>{getOrderTotal(order)}</td>
                <td>{createdAt ? new Date(createdAt).toLocaleString() : 'N/A'}</td>
                <td>
                  {isActionable(order) ? (
                    <FulfillOrderButton orderId={orderId} onFulfilled={() => onOrderFulfilled(orderId)} />
                  ) : (
                    <span className="order-status-fulfilled">{statusDisplay[status] || status}</span>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
} 