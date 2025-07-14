'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import Alert from '@/components/ui/Alert';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import './DeleteProductButton.css';

interface DeleteProductButtonProps {
  productId: string;
  productName: string;
  onProductDeleted: (productId: string) => void;
}

export default function DeleteProductButton({ productId, productName, onProductDeleted }: DeleteProductButtonProps) {
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDelete = async () => {
    setLoading(true);
    setError(null);

    try {
      // Get CSRF token first
      await api.getCSRFToken('/users/register');

      // Delete product
      await api.delete(`/products/delete?id=${productId}`, undefined, {}, true);
      
      onProductDeleted(productId);
      setShowConfirmation(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  if (showConfirmation) {
    return (
      <div className="delete-product-confirmation">
        {error && <Alert type="error">{error}</Alert>}
        
        <div className="confirmation-content">
          <h4>Delete Product</h4>
          <p>Are you sure you want to delete "{productName}"? This action cannot be undone.</p>
          
          <div className="confirmation-actions">
            <button
              className="btn-danger"
              onClick={handleDelete}
              disabled={loading}
            >
              {loading ? (
                <>
                  <LoadingSpinner />
                  Deleting...
                </>
              ) : (
                'Delete Product'
              )}
            </button>
            
            <button
              className="btn-secondary"
              onClick={() => setShowConfirmation(false)}
              disabled={loading}
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <button
      className="btn-delete"
      onClick={() => setShowConfirmation(true)}
    >
      Delete
    </button>
  );
} 