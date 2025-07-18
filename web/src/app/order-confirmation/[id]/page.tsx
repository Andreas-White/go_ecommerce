"use client";
import { useState, useEffect, useMemo } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuth } from '../../../context/AuthContext';
import { api } from '../../../lib/api';
import Link from 'next/link';
import './page.css';

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
  };
  shipping: {
    address: string;
    city: string;
    country: string;
    zip_code: string;
    method: string;
    cost: number;
  };
}

export default function OrderConfirmationPage() {
  const params = useParams();
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [orderDetails, setOrderDetails] = useState<OrderWithDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [productMap, setProductMap] = useState<{ [productId: string]: any }>({});

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

  const orderId = params.id as string;

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
    try {
      setLoading(true);
      const details = await api.post<OrderWithDetails>('/orders/details', { order_id: orderId });
      setOrderDetails(details);
    } catch (error) {
      console.error('Failed to fetch order details:', error);
      setError('Failed to load order details. Please try again.');
    } finally {
      setLoading(false);
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
      <div className="confirmation-loading">
        <div className="loading-spinner"></div>
        <p>Loading order details...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="confirmation-error">
        <div className="error-icon">❌</div>
        <h2>Error Loading Order</h2>
        <p>{error}</p>
        <Link href="/profile" className="btn-primary">
          Go to Profile
        </Link>
      </div>
    );
  }

  if (!orderDetails) {
    return (
      <div className="confirmation-error">
        <div className="error-icon">❌</div>
        <h2>Order Not Found</h2>
        <p>The order you're looking for could not be found.</p>
        <Link href="/profile" className="btn-primary">
          Go to Profile
        </Link>
      </div>
    );
  }

  return (
    <div className="confirmation-container">
      <div className="confirmation-content">
        <div className="confirmation-header">
          <div className="success-icon">✅</div>
          <h1 className="confirmation-title">Order Confirmed!</h1>
          <p className="confirmation-subtitle">
            Thank you for your order. We've sent a confirmation email with your order details.
          </p>
        </div>

        <div className="order-details">
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
                <span className="info-value status-badge">{orderDetails.order.status}</span>
              </div>
              <div className="info-row">
                <span className="info-label">Payment Status:</span>
                <span className="info-value status-badge">{orderDetails.order.payment_status}</span>
              </div>
            </div>
          </div>

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
            </div>
          </div>

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
                <span className="info-value status-badge">{orderDetails.payment.status}</span>
              </div>
            </div>
          </div>

          <div className="detail-section">
            <h2 className="section-title">Order Total</h2>
            <div className="total-breakdown">
              <div className="total-row">
                <span className="total-label">Subtotal:</span>
                <span className="total-value">
                  ${(orderDetails.order.total_amount - orderDetails.shipping.cost).toFixed(2)}
                </span>
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

        <div className="confirmation-actions">
          <Link href="/products" className="btn-secondary">
            Continue Shopping
          </Link>
          <Link href="/profile" className="btn-primary">
            View My Orders
          </Link>
        </div>
      </div>
    </div>
  );
} 