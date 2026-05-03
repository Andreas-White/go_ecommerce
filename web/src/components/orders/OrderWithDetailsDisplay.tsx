import React, { useState, useEffect, useMemo } from 'react';
import { api } from '../../lib/api';
import { StatusBadge } from '@/components/ui';
import './OrderWithDetailsDisplay.css';

interface OrderWithDetails {
  order: {
    id: string;
    total_amount: number;
    status: string;
    payment_status: string;
    created_at: string;
  };
  items: Array<{
    product_id: string;
    product_name: string;
    quantity: number;
    price: number;
    subtotal: number;
  }>;
  payment: {
    payment_method: string;
    amount: number;
    status: string;
    transaction_id?: string;
  };
  shipping: {
    address: string;
    city: string;
    country: string;
    zip_code: string;
    method: string;
    cost: number;
    status: string;
    tracking_code?: string;
    shipped_at?: string;
    delivered_at?: string;
  };
}

interface OrderWithDetailsDisplayProps {
  orderDetails: OrderWithDetails;
}

const getPaymentMethodLabel = (method: string): string => {
  switch (method) {
    case 'credit_card': return 'Credit Card';
    case 'paypal': return 'PayPal';
    case 'bank_transfer': return 'Bank Transfer';
    default: return method;
  }
};

const getShippingMethodLabel = (method: string): string => {
  switch (method) {
    case 'standard': return 'Standard Shipping';
    case 'express': return 'Express Shipping';
    case 'overnight': return 'Overnight Shipping';
    default: return method;
  }
};

const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

export default function OrderWithDetailsDisplay({ orderDetails }: OrderWithDetailsDisplayProps) {
  const [productMap, setProductMap] = useState<{ [productId: string]: any }>({});

  // Memoize unique product IDs to avoid unnecessary re-computations
  const uniqueProductIds = useMemo(() => {
    return Array.from(new Set(orderDetails.items.map(item => item.product_id)));
  }, [orderDetails.items]);

  // Fetch products only when the set of unique product IDs changes
  useEffect(() => {
    const fetchProducts = async () => {
      if (uniqueProductIds.length === 0) {
        setProductMap({});
        return;
      }

      const newMap: { [productId: string]: any } = {};
      await Promise.all(uniqueProductIds.map(async (id) => {
        // Only fetch if we don't already have this product
        if (!productMap[id]) {
          try {
            const product = await api.get(`/product?id=${id}`);
            newMap[id] = product;
          } catch (e) {
            // ignore error, leave undefined
          }
        } else {
          // Keep existing product data
          newMap[id] = productMap[id];
        }
      }));
      setProductMap(newMap);
    };
    
    fetchProducts();
  }, [uniqueProductIds]); // Only depend on unique product IDs

  // Calculate subtotal using product prices from fetched data
  const subtotal = orderDetails.items.reduce((sum, item) => {
    const product = productMap[item.product_id];
    const price = product?.price || 0;
    return sum + (price * item.quantity);
  }, 0);

  return (
    <div className="order-details-display">
      <div className="details-grid">
        {/* Order Information */}
        <div className="detail-section">
          <h2 className="section-title">Order Information</h2>
          <div className="info-card">
            <div className="info-row">
              <span className="info-label">Order ID:</span>
              <span className="info-value">{orderDetails.order.id}</span>
            </div>
            <div className="info-row">
              <span className="info-label">Order Date:</span>
              <span className="info-value">{formatDate(orderDetails.order.created_at)}</span>
            </div>
            <div className="info-row">
              <span className="info-label">Status:</span>
              <StatusBadge type="order" status={orderDetails.order.status} />
            </div>
            <div className="info-row">
              <span className="info-label">Payment Status:</span>
              <StatusBadge type="payment" status={orderDetails.order.payment_status} />
            </div>
          </div>
        </div>

        {/* Order Items */}
        <div className="detail-section">
          <h2 className="section-title">Order Items</h2>
          <div className="order-items">
            {orderDetails.items.map((item, index) => {
              const product = productMap[item.product_id];
              const price = product?.price || 0;
              const itemSubtotal = price * item.quantity;
              
              return (
                <div key={index} className="order-item">
                  <div className="item-info">
                    <h3 className="item-name">{item.product_name}</h3>
                    <p className="item-details">
                      Quantity: {item.quantity} × ${price.toFixed(2)}
                    </p>
                  </div>
                  <div className="item-subtotal">
                    ${itemSubtotal.toFixed(2)}
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Shipping Information */}
        <div className="detail-section">
          <h2 className="section-title">Shipping Information</h2>
          <div className="info-card">
            <div className="info-row">
              <span className="info-label">Address:</span>
              <span className="info-value">{orderDetails.shipping.address}</span>
            </div>
            <div className="info-row">
              <span className="info-label">City:</span>
              <span className="info-value">{orderDetails.shipping.city}</span>
            </div>
            <div className="info-row">
              <span className="info-label">Country:</span>
              <span className="info-value">{orderDetails.shipping.country}</span>
            </div>
            <div className="info-row">
              <span className="info-label">ZIP Code:</span>
              <span className="info-value">{orderDetails.shipping.zip_code}</span>
            </div>
            <div className="info-row">
              <span className="info-label">Method:</span>
              <span className="info-value">{getShippingMethodLabel(orderDetails.shipping.method)}</span>
            </div>
            {orderDetails.shipping.tracking_code && (
              <div className="info-row">
                <span className="info-label">Tracking Code:</span>
                <span className="info-value tracking-code">{orderDetails.shipping.tracking_code}</span>
              </div>
            )}
            {orderDetails.shipping.shipped_at && (
              <div className="info-row">
                <span className="info-label">Shipped:</span>
                <span className="info-value">{formatDate(orderDetails.shipping.shipped_at)}</span>
              </div>
            )}
            {orderDetails.shipping.delivered_at && (
              <div className="info-row">
                <span className="info-label">Delivered:</span>
                <span className="info-value">{formatDate(orderDetails.shipping.delivered_at)}</span>
              </div>
            )}
          </div>
        </div>

        {/* Payment Information */}
        <div className="detail-section">
          <h2 className="section-title">Payment Information</h2>
          <div className="info-card">
            <div className="info-row">
              <span className="info-label">Payment Method:</span>
              <span className="info-value">{getPaymentMethodLabel(orderDetails.payment.payment_method)}</span>
            </div>
            <div className="info-row">
              <span className="info-label">Amount Paid:</span>
              <span className="info-value">${orderDetails.payment.amount.toFixed(2)}</span>
            </div>
            <div className="info-row">
              <span className="info-label">Payment Status:</span>
              <StatusBadge type="payment" status={orderDetails.payment.status} />
            </div>
            {orderDetails.payment.transaction_id && (
              <div className="info-row">
                <span className="info-label">Transaction ID:</span>
                <span className="info-value transaction-id">{orderDetails.payment.transaction_id}</span>
              </div>
            )}
          </div>
        </div>

        {/* Order Total */}
        <div className="detail-section">
          <h2 className="section-title">Order Total</h2>
          <div className="total-breakdown">
            <div className="total-row">
              <span className="total-label">Subtotal:</span>
              <span className="total-value">${subtotal.toFixed(2)}</span>
            </div>
            <div className="total-row">
              <span className="total-label">Shipping:</span>
              <span className="total-value">${orderDetails.shipping.cost.toFixed(2)}</span>
            </div>
            <div className="total-row total-final">
              <span className="total-label">Total:</span>
              <span className="total-value">${orderDetails.order.total_amount.toFixed(2)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
} 