import React, { useCallback } from 'react';
import { useCart } from '@/context/CartContext';
import ProductCard from './ProductCard';
import ProductCardSkeleton from './ProductCardSkeleton';

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

interface ProductGridProps {
  products: Product[];
  onViewDetails: (productId: string) => void;
  isLoading?: boolean;
}

const SKELETON_COUNT = 8;

export default React.memo(function ProductGrid({
  products,
  onViewDetails,
  isLoading = false,
}: ProductGridProps) {
  const { cartItems } = useCart();

  const handleViewDetails = useCallback((productId: string) => {
    onViewDetails(productId);
  }, [onViewDetails]);

  if (isLoading) {
    return (
      <div className="products-grid">
        {Array.from({ length: SKELETON_COUNT }).map((_, index) => (
          <ProductCardSkeleton key={index} />
        ))}
      </div>
    );
  }

  return (
    <div className="products-grid fade-in">
      {products.map((product) => {
        const cartItem = cartItems.find(item => item.product_id === product.id);
        const cartQuantity = cartItem ? cartItem.quantity : 0;
        return (
          <ProductCard
            key={product.id}
            product={product}
            onViewDetails={handleViewDetails}
            cartQuantity={cartQuantity}
          />
        );
      })}
    </div>
  );
});
