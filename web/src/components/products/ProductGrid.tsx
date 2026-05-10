import React, { useMemo } from 'react';
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
  product_id: string;
  quantity: number;
}

interface ProductGridProps {
  products: Product[];
  onViewDetails: (productId: string) => void;
  isLoading?: boolean;
  cartItems?: CartItem[];
}

const SKELETON_COUNT = 8;

export default React.memo(function ProductGrid({
  products,
  onViewDetails,
  isLoading = false,
  cartItems = [],
}: ProductGridProps) {
  const productCartMap = useMemo(() => {
    const map = new Map<string, number>();
    cartItems.forEach(item => map.set(item.product_id, item.quantity));
    return map;
  }, [cartItems]);

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
        const cartQuantity = productCartMap.get(product.id) || 0;
        return (
          <ProductCard
            key={product.id}
            product={product}
            onViewDetails={onViewDetails}
            cartQuantity={cartQuantity}
          />
        );
      })}
    </div>
  );
});
