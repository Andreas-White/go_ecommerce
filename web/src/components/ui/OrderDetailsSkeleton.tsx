import React from 'react';
import './OrderDetailsSkeleton.css';

export default React.memo(function OrderDetailsSkeleton() {
  return (
    <div className="order-details-skeleton">
      <div className="skeleton-header">
        <div className="skeleton-back skeleton" />
        <div className="skeleton-title skeleton" />
        <div className="skeleton-subtitle skeleton" />
      </div>

      <div className="skeleton-grid">
        <div className="skeleton-section">
          <div className="skeleton-section-title skeleton" />
          <div className="skeleton-info-card">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="skeleton-info-row">
                <div className="skeleton-label skeleton" />
                <div className="skeleton-value skeleton" />
              </div>
            ))}
          </div>
        </div>

        <div className="skeleton-section">
          <div className="skeleton-section-title skeleton" />
          <div className="skeleton-items-list">
            {[1, 2, 3].map((i) => (
              <div key={i} className="skeleton-item-row">
                <div className="skeleton-item-image skeleton" />
                <div className="skeleton-item-info">
                  <div className="skeleton-item-name skeleton" />
                  <div className="skeleton-item-details skeleton" />
                </div>
                <div className="skeleton-item-price skeleton" />
              </div>
            ))}
          </div>
        </div>

        <div className="skeleton-section">
          <div className="skeleton-section-title skeleton" />
          <div className="skeleton-info-card">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="skeleton-info-row">
                <div className="skeleton-label skeleton" />
                <div className="skeleton-value skeleton" />
              </div>
            ))}
          </div>
        </div>

        <div className="skeleton-section">
          <div className="skeleton-section-title skeleton" />
          <div className="skeleton-info-card">
            {[1, 2, 3].map((i) => (
              <div key={i} className="skeleton-info-row">
                <div className="skeleton-label skeleton" />
                <div className="skeleton-value skeleton" />
              </div>
            ))}
          </div>
        </div>

        <div className="skeleton-section">
          <div className="skeleton-section-title skeleton" />
          <div className="skeleton-total-card">
            <div className="skeleton-total-row skeleton" />
            <div className="skeleton-total-row skeleton" />
            <div className="skeleton-total-row skeleton" style={{ width: '60%' }} />
          </div>
        </div>
      </div>
    </div>
  );
});
