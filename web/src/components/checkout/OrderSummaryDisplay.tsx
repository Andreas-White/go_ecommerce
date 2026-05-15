import React, { useState, useEffect, useMemo } from 'react';
import { api } from '../../lib/api';
import './OrderSummaryDisplay.css';
import { Button } from '@/components/ui';
import { OrderGroupSummary, Product } from '@/types';

interface OrderSummaryDisplayProps {
  orderSummary: OrderGroupSummary | null;
  loading: boolean;
  onConfirm: () => void;
  onBack: () => void;
}

const getShippingMethodLabel = (method: string): string => {
  switch (method) {
    case 'standard': return 'Standard Shipping';
    case 'express': return 'Express Shipping';
    case 'overnight': return 'Overnight Shipping';
    default: return method;
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

export default function OrderSummaryDisplay({ 
  orderSummary, 
  loading, 
  onConfirm, 
  onBack 
}: OrderSummaryDisplayProps) {
  const [productMap, setProductMap] = useState<{ [productId: string]: Product }>({});

  // Memoize unique product IDs to avoid unnecessary re-computations
  const uniqueProductIds = useMemo(() => {
    if (!orderSummary) return [];
    const allItems = orderSummary.orders.flatMap(o => o.items);
    return Array.from(new Set(allItems.map(item => item.product_id)));
  }, [orderSummary]);

  // Fetch products only when the set of unique product IDs changes
  useEffect(() => {
    const fetchProducts = async () => {
      if (uniqueProductIds.length === 0) {
        setProductMap({});
        return;
      }

      const newMap: { [productId: string]: Product } = {};
      await Promise.all(uniqueProductIds.map(async (id) => {
        // Only fetch if we don't already have this product
        if (!productMap[id]) {
          try {
            const product = await api.get<Product>(`/product?id=${id}`);
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
  }, [uniqueProductIds]); // Only depend on unique product IDs, not orderSummary

  if (!orderSummary) {
    return (
      <div className="order-summary">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Processing your order...</p>
        </div>
      </div>
    );
  }

  const allItems = orderSummary.orders.flatMap(o => o.items);
  const totalShippingCost = orderSummary.orders.reduce((sum, o) => sum + o.shipping_cost, 0);

  // Calculate subtotal using product prices from fetched data
  const subtotal = allItems.reduce((sum, item) => {
    const product = productMap[item.product_id];
    const price = product?.price || 0;
    return sum + (price * item.quantity);
  }, 0);
  const total = orderSummary.total_amount || (subtotal + totalShippingCost);

  // Take shipping and payment info from the first order, as they are shared
  const firstOrderInfo = orderSummary.orders[0];

  return (
    <div className="order-summary">
      <h2 className="form-title">Order Summary</h2>
      <p className="form-subtitle">Please review your order details before confirming.</p>
      
      <div className="summary-sections">
        {/* Order Items */}
        <div className="summary-section">
          <h3 className="section-title">Order Items</h3>
          <div className="order-items">
            {allItems.map((item, index) => {
              const product = productMap[item.product_id];
              const price = product?.price || 0;
              const itemSubtotal = price * item.quantity;
              
              return (
                <div key={index} className="order-item">
                  <div className="item-info">
                    <h4 className="item-name">{item.product_name}</h4>
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
        <div className="summary-section">
          <h3 className="section-title">Shipping Information</h3>
          <div className="info-card">
            <div className="info-row">
              <span className="info-label">Address:</span>
              <span className="info-value">{firstOrderInfo.shipping_info.address}</span>
            </div>
            <div className="info-row">
              <span className="info-label">City:</span>
              <span className="info-value">{firstOrderInfo.shipping_info.city}</span>
            </div>
            <div className="info-row">
              <span className="info-label">Country:</span>
              <span className="info-value">{firstOrderInfo.shipping_info.country}</span>
            </div>
            <div className="info-row">
              <span className="info-label">ZIP Code:</span>
              <span className="info-value">{firstOrderInfo.shipping_info.zip_code}</span>
            </div>
            <div className="info-row">
              <span className="info-label">Method:</span>
              <span className="info-value">{getShippingMethodLabel(firstOrderInfo.shipping_info.method)}</span>
            </div>
          </div>
        </div>

        {/* Payment Information */}
        <div className="summary-section">
          <h3 className="section-title">Payment Information</h3>
          <div className="info-card">
            <div className="info-row">
              <span className="info-label">Payment Method:</span>
              <span className="info-value">{getPaymentMethodLabel(firstOrderInfo.payment_info.payment_method)}</span>
            </div>
          </div>
        </div>

        {/* Order Total */}
        <div className="summary-section">
          <h3 className="section-title">Order Total</h3>
          <div className="total-breakdown">
            <div className="total-row">
              <span className="total-label">Subtotal:</span>
              <span className="total-value">${subtotal.toFixed(2)}</span>
            </div>
            <div className="total-row">
              <span className="total-label">Shipping (Multiple Producers):</span>
              <span className="total-value">${totalShippingCost.toFixed(2)}</span>
            </div>
            <div className="total-row total-final">
              <span className="total-label">Total:</span>
              <span className="total-value">${total.toFixed(2)}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="order-note">
        <div className="note-icon">📧</div>
        <div className="note-content">
          <h4>Order Confirmation</h4>
          <p>You will receive an email confirmation with your order details and tracking information once your order is confirmed.</p>
        </div>
      </div>

      <div className="form-actions">
        <Button type="button" onClick={onBack} variant="secondary">
          Back to Payment
        </Button>
        <Button 
          type="button" 
          onClick={onConfirm} 
          variant="primary"
          disabled={loading}
        >
          {loading ? 'Confirming Order...' : 'Confirm Order'}
        </Button>
      </div>
    </div>
  );
} 