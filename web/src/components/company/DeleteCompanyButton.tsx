'use client';

import { useState } from 'react';
import { api } from '@/lib/api';
import Alert from '@/components/ui/Alert';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import './DeleteCompanyButton.css';
import { Button } from '@/components/ui';

interface DeleteCompanyButtonProps {
  companyId: string;
  onCompanyDeleted?: () => void;
}

export default function DeleteCompanyButton({ companyId, onCompanyDeleted }: DeleteCompanyButtonProps) {
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDelete = async () => {
    setLoading(true);
    setError(null);

    try {
      // Delete company - CSRF token will be fetched automatically
      await api.delete(`/companies/delete?company_id=${companyId}`, undefined, {}, true);
      
      onCompanyDeleted?.();
      setShowConfirmation(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  if (showConfirmation) {
    return (
      <div className="delete-company-confirmation-overlay">
        <div className="delete-confirmation-modal">
          {error && <Alert type="error">{error}</Alert>}
          <div className="confirmation-content">
            <h4>Delete Company</h4>
            <p>Are you sure you want to delete this company? This action cannot be undone.</p>
            <div className="confirmation-actions">
              <Button
                variant="destructive"
                onClick={handleDelete}
                disabled={loading}
              >
                {loading ? (
                  <>
                    <LoadingSpinner />
                    Deleting...
                  </>
                ) : (
                  'Delete Company'
                )}
              </Button>
              <Button
                variant="secondary"
                onClick={() => setShowConfirmation(false)}
                disabled={loading}
              >
                Cancel
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <Button
      variant="destructive"
      onClick={() => setShowConfirmation(true)}
    >
      Delete Company
    </Button>
  );
} 