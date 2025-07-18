import React from 'react';
import './ShippingTracking.css';

interface ShippingInfo {
  status: string;
  tracking_code?: string;
  shipped_at?: string;
  delivered_at?: string;
  method: string;
  address: string;
  city: string;
  country: string;
  zip_code: string;
}

interface ShippingTrackingProps {
  shipping: ShippingInfo;
}

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

const getStatusIcon = (status: string): string => {
  switch (status.toLowerCase()) {
    case 'pending': return '⏳';
    case 'processing': return '⚙️';
    case 'shipped': return '📦';
    case 'delivered': return '✅';
    case 'cancelled': return '❌';
    default: return '📋';
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

export default function ShippingTracking({ shipping }: ShippingTrackingProps) {
  const getTrackingSteps = () => {
    const steps = [
      { id: 'pending', label: 'Order Placed', completed: true },
      { id: 'processing', label: 'Processing', completed: ['processing', 'shipped', 'delivered'].includes(shipping.status.toLowerCase()) },
      { id: 'shipped', label: 'Shipped', completed: ['shipped', 'delivered'].includes(shipping.status.toLowerCase()) },
      { id: 'delivered', label: 'Delivered', completed: shipping.status.toLowerCase() === 'delivered' }
    ];
    return steps;
  };

  const trackingSteps = getTrackingSteps();

  return (
    <div className="shipping-tracking">
      <div className="tracking-header">
        <h2 className="tracking-title">Shipping & Tracking</h2>
        <div className="tracking-status">
          <span className={`status-badge ${getStatusBadgeClass(shipping.status)}`}>
            {getStatusIcon(shipping.status)} {shipping.status}
          </span>
        </div>
      </div>

      {/* Tracking Progress */}
      <div className="tracking-progress">
        <div className="progress-steps">
          {trackingSteps.map((step, index) => (
            <div key={step.id} className={`progress-step ${step.completed ? 'completed' : ''}`}>
              <div className="step-icon">
                {step.completed ? '✓' : (index + 1)}
              </div>
              <div className="step-content">
                <div className="step-label">{step.label}</div>
                {step.id === 'shipped' && shipping.shipped_at && (
                  <div className="step-date">{formatDate(shipping.shipped_at)}</div>
                )}
                {step.id === 'delivered' && shipping.delivered_at && (
                  <div className="step-date">{formatDate(shipping.delivered_at)}</div>
                )}
              </div>
              {index < trackingSteps.length - 1 && (
                <div className={`step-connector ${step.completed ? 'completed' : ''}`}></div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Shipping Information */}
      <div className="shipping-info">
        <div className="info-section">
          <h3 className="info-title">Shipping Method</h3>
          <p className="info-value">{getShippingMethodLabel(shipping.method)}</p>
        </div>

        <div className="info-section">
          <h3 className="info-title">Delivery Address</h3>
          <div className="address-info">
            <p>{shipping.address}</p>
            <p>{shipping.city}, {shipping.country} {shipping.zip_code}</p>
          </div>
        </div>

        {shipping.tracking_code && (
          <div className="info-section">
            <h3 className="info-title">Tracking Code</h3>
            <div className="tracking-code-container">
              <code className="tracking-code">{shipping.tracking_code}</code>
              <button 
                className="copy-button"
                onClick={() => navigator.clipboard.writeText(shipping.tracking_code!)}
                title="Copy tracking code"
              >
                📋
              </button>
            </div>
          </div>
        )}

        {/* Shipping Timeline */}
        <div className="info-section">
          <h3 className="info-title">Shipping Timeline</h3>
          <div className="timeline">
            {shipping.shipped_at && (
              <div className="timeline-item">
                <div className="timeline-icon shipped">📦</div>
                <div className="timeline-content">
                  <div className="timeline-title">Package Shipped</div>
                  <div className="timeline-date">{formatDate(shipping.shipped_at)}</div>
                </div>
              </div>
            )}
            {shipping.delivered_at && (
              <div className="timeline-item">
                <div className="timeline-icon delivered">✅</div>
                <div className="timeline-content">
                  <div className="timeline-title">Package Delivered</div>
                  <div className="timeline-date">{formatDate(shipping.delivered_at)}</div>
                </div>
              </div>
            )}
            {!shipping.shipped_at && !shipping.delivered_at && (
              <div className="timeline-item">
                <div className="timeline-icon pending">⏳</div>
                <div className="timeline-content">
                  <div className="timeline-title">Order Processing</div>
                  <div className="timeline-date">Your order is being prepared for shipment</div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
} 