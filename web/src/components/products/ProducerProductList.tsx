'use client';

import { useState } from 'react';
import Image from 'next/image';
import dynamic from 'next/dynamic';
import DeleteProductButton from './DeleteProductButton';
import './ProducerProductList.css';
import { Button } from '@/components/ui';

const EditProductButton = dynamic(
  () => import('./EditProductButton'),
  { 
    loading: () => <div className="edit-loading">Loading form...</div>,
    ssr: false 
  }
);

interface Company {
  id: string;
  name: string;
  address: string;
  city: string;
  country: string;
  zip_code: string;
  review_average: number;
  review_count: number;
  created_at: string;
  updated_at: string;
}

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  category_id: string;
  image_url: string;
  company: Company;
}

interface ProducerProductListProps {
  products: Product[];
  onProductUpdated: (product: Product) => void;
  onProductDeleted: (productId: string) => void;
}

export default function ProducerProductList({
  products,
  onProductUpdated,
  onProductDeleted,
}: ProducerProductListProps) {
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);

  const handleEditStart = (product: Product) => {
    setEditingProduct(product);
  };

  const handleEditComplete = (updatedProduct: Product) => {
    onProductUpdated(updatedProduct);
    setEditingProduct(null);
  };

  const handleEditCancel = () => {
    setEditingProduct(null);
  };

  if (products.length === 0) {
    return (
      <div className="no-products">
        <p>No products found. Create your first product to get started!</p>
      </div>
    );
  }

  return (
    <div className="producer-product-list">
      {editingProduct && (
        <EditProductButton
          product={editingProduct}
          onProductUpdated={handleEditComplete}
          onCancel={handleEditCancel}
        />
      )}
      <div className="products-grid">
        {products.map((product) => (
          <div key={product.id} className="product-card">
            <div className="product-image">
              {product.image_url ? (
                <Image
                  src={product.image_url}
                  alt={product.name}
                  fill
                  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                />
              ) : (
                <div className="no-image">No Image</div>
              )}
            </div>
            <div className="product-info">
              <h4>{product.name}</h4>
              <p className="product-description">
                {product.description || 'No description available'}
              </p>
              <div className="product-details">
                <div className="detail-item">
                  <span className="label">Price:</span>
                  <span className="value">${product.price.toFixed(2)}</span>
                </div>
                <div className="detail-item">
                  <span className="label">Stock:</span>
                  <span
                    className={`value ${
                      product.stock === 0 ? 'out-of-stock' : ''
                    }`}
                  >
                    {product.stock} units
                  </span>
                </div>
                <div className="detail-item">
                  <span className="label">Category:</span>
                  <span className="value">{product.category_id}</span>
                </div>
              </div>
            </div>
            <div className="product-actions">
              <Button
                variant="secondary"
                onClick={() => handleEditStart(product)}
              >
                Edit
              </Button>
              <DeleteProductButton
                productId={product.id}
                productName={product.name}
                onProductDeleted={onProductDeleted}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
