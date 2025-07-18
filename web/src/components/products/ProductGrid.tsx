import ProductCard from './ProductCard';
import type { Product } from './ProductCard';
import './ProductGrid.css';

// Not used 
export default function ProductGrid({ products }: { products: Product[] }) {
  return (
    <div className="product-grid">
      {products.map(product => (
        <ProductCard key={product.id} product={product} />
      ))}
    </div>
  );
} 