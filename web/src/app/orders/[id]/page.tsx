"use client";
import { useState, useEffect, useMemo } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '../../../context/AuthContext';
import { api } from '../../../lib/api';
import Link from 'next/link';
import './page.css';
import Button from '@/components/ui/Button';
import { useTopProgress } from '@/context/TopProgressContext';
import OrderDetailsSkeleton from '@/components/ui/OrderDetailsSkeleton';

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

export default function OrderDetailsPage() {
  const params = useParams();
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const { start: startProgress, complete: completeProgress } = useTopProgress();
  const [orderDetails, setOrderDetails] = useState<OrderWithDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [productMap, setProductMap] = useState<{ [productId: string]: any }>({});

  const orderId = params.id as string;

  // Memoize unique product IDs to avoid unnecessary re-computations
  const uniqueProductIds = useMemo(() => {
    if (!orderDetails) return [];
    return Array.from(new Set(orderDetails.items.map(item => item.product_id)));
  }, [orderDetails]);

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
  }, [uniqueProductIds]); // Only depend on unique product IDs, not orderDetails

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
      return;
    }

    if (orderId && user) {
      fetchOrderDetails();
    }
  }, [orderId, user, authLoading, router]);

  const fetchOrderDetails = async () => {
    startProgress();
    try {
      setLoading(true);
      const details = await api.post<OrderWithDetails>('/orders/details', { order_id: orderId });
      setOrderDetails(details);
    } catch (error) {
      setError('Failed to load order details. Please try again.');
    } finally {
      setLoading(false);
      completeProgress();
    }
  };

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
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (authLoading || loading) {
    return (
      <div className="order-details-loading">
        <OrderDetailsSkeleton />
      </div>
    );
  }

  if (error) {
    return (
      <div className="order-details-error">
        <div className="error-icon">❌</div>
        <h2>Error Loading Order</h2>
        <p>{error}</p>
        <Button onClick={() => router.back()} variant="tertiary">
          ← Back to Orders
        </Button>
      </div>
    );
  }

  if (!orderDetails) {
    return (
      <div className="order-details-error">
        <div className="error-icon">❌</div>
        <h2>Order Not Found</h2>
        <p>The order you're looking for could not be found.</p>
        <Button onClick={() => router.back()} variant="tertiary">
          ← Back to Orders
        </Button>
      </div>
    );
  }

  // Calculate subtotal using product prices from fetched data
  const subtotal = orderDetails.items.reduce((sum, item) => {
    const product = productMap[item.product_id];
    const price = product?.price || 0;
    return sum + (price * item.quantity);
  }, 0);

  return (
    <div className="order-details-container">
      <div className="order-details-header">
        <Button onClick={() => router.back()} variant="tertiary">
          ← Back to Orders
        </Button>
        <h1 className="order-details-title">Order Details</h1>
        <p className="order-details-subtitle">Order #{orderDetails.order.id.slice(0, 8)}</p>
      </div>

      <div className="order-details-content">
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
                <span className={`status-badge ${getStatusBadgeClass(orderDetails.order.status)}`}>
                  {orderDetails.order.status}
                </span>
              </div>
              <div className="info-row">
                <span className="info-label">Payment Status:</span>
                <span className={`status-badge ${getPaymentStatusBadgeClass(orderDetails.order.payment_status)}`}>
                  {orderDetails.order.payment_status}
                </span>
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
                <span className={`status-badge ${getPaymentStatusBadgeClass(orderDetails.payment.status)}`}>
                  {orderDetails.payment.status}
                </span>
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
    </div>
  );
} 