import React, { useState, useCallback } from 'react';
import { Button } from '@/components/ui';
import { useCart } from '@/context/CartContext';
import './ProductCard.css';

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  original_price?: number;
  stock: number;
  rating?: number;
  review_count?: number;
  category_id: string;
  image_url?: string;
  company?: {
    name: string;
  };
}

interface ProductCardProps {
  product: Product;
  cartQuantity: number;
  onViewDetails: (productId: string) => void;
}

export default React.memo(function ProductCard({
  product,
  cartQuantity,
  onViewDetails,
}: ProductCardProps) {
  const { addToCart } = useCart();
  const [isAddingToCart, setIsAddingToCart] = useState(false);
  const isStockLimitReached = cartQuantity >= product.stock;

  const isOnSale = product.original_price && product.original_price > product.price;
  const discountPercentage = isOnSale
    ? Math.round(((product.original_price! - product.price) / product.original_price!) * 100)
    : 0;

  const handleAddToCart = useCallback(async () => {
    setIsAddingToCart(true);
    try {
      await addToCart([{ product_id: product.id, quantity: 1, price: product.price }]);
    } finally {
      setIsAddingToCart(false);
    }
  }, [addToCart, product.id, product.price]);

  return (
    <div className="product-card">
      {isOnSale && (
        <span className="sale-badge">-{discountPercentage}%</span>
      )}
      <div className="product-card-image-wrapper" onClick={() => onViewDetails(product.id)}>
        {product.image_url ? (
          <img
            src={product.image_url}
            alt={product.name}
            className="product-card-image"
            loading="lazy"
          />
        ) : (
          <div className="product-image-placeholder">
            <span className="product-image-text">No Image</span>
          </div>
        )}
        <span className="quick-view-tooltip">Tap for details</span>
      </div>
      <h3 className="product-name" onClick={() => onViewDetails(product.id)}>{product.name}</h3>
      {product.company && (
        <p className="product-company">by {product.company.name}</p>
      )}
      <div className="product-rating">
        <span className={`product-stars ${product.rating === undefined ? 'product-stars-hidden' : ''}`}>
          {'★'.repeat(Math.floor(product.rating ?? 0))}
          {'☆'.repeat(5 - Math.floor(product.rating ?? 0))}
        </span>
        {product.review_count !== undefined && product.rating !== undefined && (
          <span className="review-count">({product.review_count})</span>
        )}
      </div>
      <div className="product-price-stock">
        <span className="product-price">
          {isOnSale ? (
            <span className="sale-price">
              <span className="original-price">${product.original_price?.toFixed(2)}</span>
              ${product.price?.toFixed(2)}
            </span>
          ) : (
            `$${product.price?.toFixed(2)}`
          )}
        </span>
        <span
          className={`product-stock ${
            product.stock > 0
              ? 'product-stock-available'
              : 'product-stock-unavailable'
          }`}
        >
          {product.stock > 0 ? `${product.stock} in stock` : 'Out of stock'}
        </span>
      </div>
      <div className="product-actions">
        <Button variant="secondary" onClick={() => onViewDetails(product.id)}>
          View Details
        </Button>
        <Button
          variant="primary"
          onClick={handleAddToCart}
          disabled={product.stock <= 0 || isStockLimitReached}
          isLoading={isAddingToCart}
          className={`product-add-btn ${
            product.stock > 0
              ? 'product-add-btn-available'
              : 'product-add-btn-unavailable'
          }`}
        >
          Add to Cart
        </Button>
      </div>
    </div>
  );
});
