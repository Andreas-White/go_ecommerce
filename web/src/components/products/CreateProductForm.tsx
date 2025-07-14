'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import Alert from '@/components/ui/Alert';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import './CreateProductForm.css';

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

interface CreateProductFormProps {
  onProductCreated: (product: Product) => void;
  onCancel: () => void;
}

const CATEGORIES = [
  'Electronics',
  'Clothing',
  'Books',
  'Home & Garden',
  'Sports & Outdoors',
  'Toys & Games',
  'Health & Beauty',
  'Automotive',
  'Food & Beverages',
  'Other'
];

export default function CreateProductForm({ onProductCreated, onCancel }: CreateProductFormProps) {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    price: 0,
    stock: 0,
    category_id: '',
    image_url: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: name === 'price' || name === 'stock' ? parseFloat(value) || 0 : value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // Get CSRF token first
      await api.getCSRFToken('/users/register');

      // Create product
      const newProduct = await api.post<Product>('/products/create', formData, {}, true);
      onProductCreated(newProduct);
      
      // Reset form
      setFormData({
        name: '',
        description: '',
        price: 0,
        stock: 0,
        category_id: '',
        image_url: ''
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="create-product-form">
      <h3>Create New Product</h3>
      
      {error && <Alert type="error">{error}</Alert>}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="name">Product Name *</label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleInputChange}
            required
            placeholder="Enter product name"
          />
        </div>

        <div className="form-group">
          <label htmlFor="description">Description</label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleInputChange}
            rows={3}
            placeholder="Enter product description"
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="price">Price *</label>
            <input
              type="number"
              id="price"
              name="price"
              value={formData.price}
              onChange={handleInputChange}
              min="0"
              step="0.01"
              required
              placeholder="0.00"
            />
          </div>

          <div className="form-group">
            <label htmlFor="stock">Stock *</label>
            <input
              type="number"
              id="stock"
              name="stock"
              value={formData.stock}
              onChange={handleInputChange}
              min="0"
              required
              placeholder="0"
            />
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="category_id">Category *</label>
          <select
            id="category_id"
            name="category_id"
            value={formData.category_id}
            onChange={handleInputChange}
            required
          >
            <option value="">Select a category</option>
            {CATEGORIES.map(category => (
              <option key={category} value={category}>
                {category}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label htmlFor="image_url">Image URL</label>
          <input
            type="url"
            id="image_url"
            name="image_url"
            value={formData.image_url}
            onChange={handleInputChange}
            placeholder="https://example.com/image.jpg"
          />
        </div>

        <div className="form-actions">
          <button
            type="submit"
            className="btn-primary"
            disabled={loading}
          >
            {loading ? (
              <>
                <LoadingSpinner />
                Creating...
              </>
            ) : (
              'Create Product'
            )}
          </button>
          
          <button
            type="button"
            className="btn-secondary"
            onClick={onCancel}
            disabled={loading}
          >
            Cancel
          </button>
        </div>
      </form>
    </div>
  );
} 