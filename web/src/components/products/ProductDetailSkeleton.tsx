import React from 'react';
import './ProductDetailSkeleton.css';

export default React.memo(function ProductDetailSkeleton() {
  return (
    <div className="product-detail-skeleton">
      <div className="skeleton-breadcrumb">
        <div className="skeleton-breadcrumb-item skeleton" />
        <div className="skeleton-breadcrumb-separator skeleton" />
        <div className="skeleton-breadcrumb-item skeleton" />
      </div>

      <div className="skeleton-content">
        <div className="skeleton-image-section">
          <div className="skeleton-image skeleton" />
        </div>

        <div className="skeleton-info-section">
          <div className="skeleton-title skeleton" />
          <div className="skeleton-company skeleton" />
          <div className="skeleton-rating skeleton" />
          <div className="skeleton-price skeleton" />
          <div className="skeleton-stock skeleton" />
          <div className="skeleton-description-title skeleton" />
          <div className="skeleton-description skeleton" />
          <div className="skeleton-description skeleton" style={{ width: '80%' }} />
          <div className="skeleton-purchase">
            <div className="skeleton-quantity skeleton" />
            <div className="skeleton-button skeleton" />
          </div>
        </div>
      </div>
    </div>
  );
});
