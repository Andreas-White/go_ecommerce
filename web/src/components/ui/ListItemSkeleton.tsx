import React from 'react';
import './ListItemSkeleton.css';

export default React.memo(function ListItemSkeleton() {
  return (
    <div className="list-item-skeleton">
      {[1, 2, 3].map((i) => (
        <div key={i} className="list-item-skeleton-row">
          <div className="list-item-skeleton-image skeleton" />
          <div className="list-item-skeleton-content">
            <div className="list-item-skeleton-title skeleton" />
            <div className="list-item-skeleton-subtitle skeleton" />
          </div>
        </div>
      ))}
    </div>
  );
});
