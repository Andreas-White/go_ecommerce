import React from 'react';
import './ProductCardSkeleton.css';

export default React.memo(function ProductCardSkeleton() {
  return (
    <div className="product-card-skeleton">
      <div className="skeleton-image skeleton" />
      <div className="skeleton-title skeleton" />
      <div className="skeleton-subtitle skeleton" />
      <div className="skeleton-rating skeleton" />
      <div className="skeleton-price skeleton" />
      <div className="skeleton-actions">
        <div className="skeleton-button skeleton" />
      </div>
    </div>
  );
});
