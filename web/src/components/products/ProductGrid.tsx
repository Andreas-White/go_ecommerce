import React from 'react';
import ProductCard from './ProductCard';

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
}

export default React.memo(function ProductGrid({
  products,
  cartItems,
  onViewDetails,
  onAddToCart,
}: ProductGridProps) {
  return (
    <div className="products-grid">
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
