import React from 'react';
import styles from './StatusBadge.module.css';

interface StatusBadgeProps {
  type: 'order' | 'payment';
  status: string;
  className?: string;
}

const getStatusClass = (type: 'order' | 'payment', status: string): string => {
  const normalizedStatus = status.toLowerCase();

  if (type === 'order') {
    switch (normalizedStatus) {
      case 'pending':
        return styles.statusPending;
      case 'processing':
        return styles.statusProcessing;
      case 'shipped':
        return styles.statusShipped;
      case 'delivered':
        return styles.statusDelivered;
      case 'cancelled':
      case 'canceled':
      case 'canceled_by_user':
        return styles.statusCancelled;
      default:
        return styles.statusDefault;
    }
  } else {
    switch (normalizedStatus) {
      case 'pending':
      case 'unpaid':
        return styles.paymentPending;
      case 'paid':
        return styles.paymentPaid;
      case 'failed':
        return styles.paymentFailed;
      case 'refunded':
        return styles.paymentRefunded;
      default:
        return styles.paymentDefault;
    }
  }
};

export default function StatusBadge({ type, status, className = '' }: StatusBadgeProps) {
  return (
    <span className={`${styles.statusBadge} ${getStatusClass(type, status)} ${className}`}>
      {status}
    </span>
  );
}