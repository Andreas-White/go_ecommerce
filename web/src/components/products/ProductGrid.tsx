import React from 'react';
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

interface CartItem {
  id?: string;
  product_id: string;
  quantity: number;
  price?: number;
}

interface ProductGridProps {
  products: Product[];
  cartItems: CartItem[];
  onViewDetails: (productId: string) => void;
  onAddToCart: (product: Product) => void;
  isLoading?: boolean;
}

const SKELETON_COUNT = 8;

export default React.memo(function ProductGrid({
  products,
  cartItems,
  onViewDetails,
  onAddToCart,
  isLoading = false,
}: ProductGridProps) {
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
      {products.map((product) => (
        <ProductCard
          key={product.id}
          product={product}
          onViewDetails={onViewDetails}
          onAddToCart={onAddToCart}
          cartItems={cartItems}
        />
      ))}
    </div>
  );
});
