import React from 'react';
import './CartSkeleton.css';

export default React.memo(function CartSkeleton() {
  return (
    <div className="cart-skeleton">
      <div className="skeleton-title skeleton" />

      <div className="skeleton-items-section">
        <div className="skeleton-items-card">
          <div className="skeleton-card-title skeleton" />
          {[1, 2, 3].map((i) => (
            <div key={i} className="skeleton-cart-item">
              <div className="skeleton-item-image skeleton" />
              <div className="skeleton-item-details">
                <div className="skeleton-item-name skeleton" />
                <div className="skeleton-item-price skeleton" />
              </div>
              <div className="skeleton-item-quantity skeleton" />
              <div className="skeleton-item-remove skeleton" />
            </div>
          ))}
        </div>
      </div>

      <div className="skeleton-summary-section">
        <div className="skeleton-summary-card">
          <div className="skeleton-summary-title skeleton" />
          <div className="skeleton-summary-row skeleton" />
          <div className="skeleton-summary-row skeleton" />
          <div className="skeleton-summary-divider skeleton" />
          <div className="skeleton-summary-row skeleton" style={{ width: '60%' }} />
          <div className="skeleton-summary-button skeleton" />
        </div>
      </div>
    </div>
  );
});
