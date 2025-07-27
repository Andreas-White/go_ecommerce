import React from 'react';
import DeleteOrderButton from './DeleteOrderButton';
import { Button } from '@/components/ui';

interface Order {
  id: string;
  total_amount: number;
  status: string;
  payment_status: string;
  created_at: string;
}

interface OrderCardProps {
  order: Order;
  onViewDetails: (orderId: string) => void;
  onDelete: (orderId: string) => void;
  getStatusBadgeClass: (status: string) => string;
  getPaymentStatusBadgeClass: (status: string) => string;
  formatDate: (dateString: string) => string;
}

const OrderCard: React.FC<OrderCardProps> = ({
  order,
  onViewDetails,
  onDelete,
  getStatusBadgeClass,
  getPaymentStatusBadgeClass,
  formatDate,
}) => (
  <div className="order-card">
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
      <Button
        variant="secondary"
        style={{ minWidth: 100 }}
        onClick={() => onViewDetails(order.id)}
      >
        View Details
      </Button>
      {(order.status === 'pending' || order.status === 'processing') && (
        <DeleteOrderButton orderId={order.id} onDeleted={onDelete} />
      )}
    </div>
  </div>
);

export default OrderCard; 