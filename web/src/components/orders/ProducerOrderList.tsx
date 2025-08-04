import React from 'react';
import FulfillOrderButton from './FulfillOrderButton';
import './ProducerOrderList.css';
import CancelOrderButton from './CancelOrderButton';
import { Button } from '../ui';

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
  onOrderFulfilled: (orderId: string, newStatus: string) => void;
}

const orderStatusDisplay: Record<string, string> = {
  processing: 'Processing',
  pending: 'Pending',
  accepted: 'Accepted',
  shipped: 'Shipped',
  delivered: 'Delivered',
  canceled: 'Canceled',
  canceled_by_user: 'Canceled by User',
};

const orderPaymentStatusDisplay: Record<string, string> = {
  unpaid: 'Unpaid',
  paid: 'Paid',
  refunded: 'Refunded',
};

export default function ProducerOrderList({
  orders,
  onOrderFulfilled,
}: ProducerOrderListProps) {
  if (!orders.length) {
    return (
      <div className="order-list-empty">No orders for your products yet.</div>
    );
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

  function getOrderPaymentStatus(order: ProducerOrderLike): string {
    return order.payment_status || order.order?.payment_status || '';
  }

  function getOrderCreatedAt(order: ProducerOrderLike): string {
    return order.created_at || order.order?.created_at || '';
  }

  function getOrderNextStatus(order: ProducerOrderLike): string {
    const status = getOrderStatus(order);
    switch (status) {
      case 'processing':
        return 'accepted';
      case 'accepted':
        return 'shipped';
    }
    return status;
  }

  function isActionable(order: ProducerOrderLike): boolean {
    const status = getOrderStatus(order);
    return status !== 'shipped' && !isCanceled(order);
  }

  function isCanceled(order: ProducerOrderLike): boolean {
    const status = getOrderStatus(order);
    return status === 'canceled' || status === 'canceled_by_user';
  }

  function isShipped(order: ProducerOrderLike): boolean {
    const status = getOrderStatus(order);
    return status === 'shipped';
  }

  return (
    <div className="order-list">
      <table className="order-list-table">
        <thead>
          <tr>
            <th>Status</th>
            <th>Payment Status</th>
            <th>Total</th>
            <th>Date</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((order) => {
            const orderId = getOrderId(order);
            const status = getOrderStatus(order);
            const nextStatus = getOrderNextStatus(order);
            const paymentStatus = getOrderPaymentStatus(order);
            const createdAt = getOrderCreatedAt(order);
            return (
              <tr key={orderId}>
                <td>
                  {isShipped(order) && (
                    <span className="order-status-shipped">
                      {orderStatusDisplay[status] || status}
                    </span>
                  )}
                  {isCanceled(order) && (
                    <span className="order-status-canceled">
                      {orderStatusDisplay[status] || status}
                    </span>
                  )}
                  {!isShipped(order) && !isCanceled(order) && (
                    <span className="order-status-default">
                      {orderStatusDisplay[status] || status}
                    </span>
                  )}
                </td>
                <td>
                  {orderPaymentStatusDisplay[paymentStatus] || paymentStatus}
                </td>
                <td>{getOrderTotal(order)}</td>
                <td>
                  {createdAt ? new Date(createdAt).toLocaleString() : 'N/A'}
                </td>
                <td>
                  {isActionable(order) && (
                    <>
                      {
                        <FulfillOrderButton
                          orderId={orderId}
                          onFulfilled={() =>
                            onOrderFulfilled(orderId, nextStatus)
                          }
                          status={status}
                          nextStatus={nextStatus}
                        />
                      }
                      {
                        <CancelOrderButton
                          orderId={orderId}
                          status={status}
                          onCanceled={() => window.location.reload()}
                        />
                      }
                    </>
                  )}
                  {
                    <Button href={`/orders/${orderId}`} variant="secondary">
                      View Details
                    </Button>
                  }
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
