import Link from 'next/link';
import Button from '../ui/Button';
import './ProductCard.css';

export type Product = {
  id: string;
  name: string;
  price: number;
  image_url?: string;
  description?: string;
};

// Not used 
export default function ProductCard({ product }: { product: Product }) {
  return (
    <div className="product-card">
      <Link href={`/product/${product.id}`} className="product-card-link">
        <div className="product-card-image-container">
          {product.image_url ? (
            <img src={product.image_url} alt={product.name} className="product-card-image" />
          ) : (
            <div className="product-card-placeholder">
              No Image
            </div>
          )}
        </div>
        <h3 className="product-card-title">{product.name}</h3>
        <div className="product-card-price">
          ${product.price.toFixed(2)}
        </div>
      </Link>
      <Button className="product-card-button">Add to to Cart</Button>
    </div>
  );
} 