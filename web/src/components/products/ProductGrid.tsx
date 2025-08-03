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

interface ProductGridProps {
  products: Product[];
  onViewDetails: (productId: string) => void;
  onAddToCart: (product: Product) => void;
}

const ProductGrid: React.FC<ProductGridProps> = ({ products, onViewDetails, onAddToCart }) => (
  <div className="products-grid">
    {products.map((product) => (
      <ProductCard
        key={product.id}
        product={product}
        onViewDetails={onViewDetails}
        onAddToCart={onAddToCart}
      />
    ))}
  </div>
);

export default ProductGrid; 