import React from 'react';
import { Button } from '@/components/ui';

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  category_id: string;
  image_url?: string;
  company?: {
    name: string;
  };
}

interface CartItem {
  id?: string;
  product_id: string;
  quantity: number;
  price?: number;
}

interface ProductCardProps {
  product: Product;
  cartItems: CartItem[];
  onViewDetails: (productId: string) => void;
  onAddToCart: (product: Product) => void;
}

const ProductCard: React.FC<ProductCardProps> = ({
  product,
  cartItems,
  onViewDetails,
  onAddToCart,
}) => {
  const cartItem = cartItems.find((item) => item.product_id === product.id);
  const quantityInCart = cartItem ? cartItem.quantity : 0;
  const isStockLimitReached = quantityInCart >= product.stock;

  return (
    <div className="product-card">
      {product.image_url ? (
        <img
          src={product.image_url}
          alt={product.name}
          className="product-card-image"
        />
      ) : (
        <div className="product-image-placeholder">
          <span className="product-image-text">No Image</span>
        </div>
      )}
      <h3 className="product-name">{product.name}</h3>
      {product.company && (
        <p className="product-company">by {product.company.name}</p>
      )}
      <p className="product-description">{product.description}</p>
      <div className="product-price-stock">
        <span className="product-price">${product.price?.toFixed(2)}</span>
        <span
          className={`product-stock${
            product.stock > 0
              ? ' product-stock-available'
              : ' product-stock-unavailable'
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
          onClick={() => onAddToCart(product)}
          disabled={product.stock <= 0 || isStockLimitReached}
          className={`product-add-btn${
            product.stock > 0
              ? ' product-add-btn-available'
              : ' product-add-btn-unavailable'
          }`}
        >
          Add to Cart
        </Button>
      </div>
    </div>
  );
};

export default ProductCard;
