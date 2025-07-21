'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import Alert from '@/components/ui/Alert';
import LoadingSpinner from '@/components/ui/LoadingSpinner';
import './CreateUpdateCompanyForm.css';
import DeleteCompanyButton from './DeleteCompanyButton';
import { Button } from '@/components/ui';

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

interface CreateUpdateCompanyFormProps {
  company?: Company;
  onCompanyCreated?: (company: Company) => void;
  onCompanyUpdated?: (company: Company) => void;
  onCompanyDeleted?: () => void;
  onCancel?: () => void;
  saveLabel?: string;
}

export default function CreateUpdateCompanyForm({ 
  company, 
  onCompanyCreated, 
  onCompanyUpdated, 
  onCompanyDeleted,
  onCancel,
  saveLabel 
}: CreateUpdateCompanyFormProps) {
  const [formData, setFormData] = useState({
    id: '',
    name: '',
    address: '',
    city: '',
    country: '',
    zip_code: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const isEditing = !!company;

  useEffect(() => {
    if (company) {
      setFormData({
        id: company.id || '',
        name: company.name || '',
        address: company.address || '',
        city: company.city || '',
        country: company.country || '',
        zip_code: company.zip_code || ''
      });
    }
  }, [company]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      if (isEditing) {
        // Update existing company - CSRF token will be fetched automatically
        const result = await api.put<Company>('/companies/update', formData, {}, true);
        setSuccess('Company updated successfully!');
        if (onCompanyUpdated) {
          onCompanyUpdated(result);
        }
      } else {
        // Create new company - CSRF token will be fetched automatically
        const result = await api.post<Company>('/companies/create', 
        {
          name: formData.name,
          address: formData.address,
          city: formData.city,
          country: formData.country,
          zip_code: formData.zip_code
        }, {}, true);
        setSuccess('Company created successfully!');
        if (onCompanyCreated) {
          onCompanyCreated(result);
        }

        // Reset form if creating new company
        setFormData({
          id: '',
          name: '',
          address: '',
          city: '',
          country: '',
          zip_code: ''
        });
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="company-form-container">
      <h3>{isEditing ? 'Update Company Profile' : 'Create Company Profile'}</h3>
      
      {error && <Alert type="error">{error}</Alert>}
      {success && <Alert type="success">{success}</Alert>}

      <form onSubmit={handleSubmit} className="company-form">
        <div className="form-group">
          <label htmlFor="name">Company Name *</label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleInputChange}
            required
            placeholder="Enter company name"
          />
        </div>

        <div className="form-group">
          <label htmlFor="address">Address</label>
          <input
            type="text"
            id="address"
            name="address"
            value={formData.address}
            onChange={handleInputChange}
            placeholder="Enter company address"
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="city">City</label>
            <input
              type="text"
              id="city"
              name="city"
              value={formData.city}
              onChange={handleInputChange}
              placeholder="Enter city"
            />
          </div>

          <div className="form-group">
            <label htmlFor="country">Country</label>
            <input
              type="text"
              id="country"
              name="country"
              value={formData.country}
              onChange={handleInputChange}
              placeholder="Enter country"
            />
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="zip_code">ZIP Code</label>
          <input
            type="text"
            id="zip_code"
            name="zip_code"
            value={formData.zip_code}
            onChange={handleInputChange}
            placeholder="Enter ZIP code"
          />
        </div>

        <div className="form-actions">
          <Button
            type="submit"
            variant="primary"
            disabled={loading}
          >
            {loading ? (
              <>
                <LoadingSpinner />
                {isEditing ? (saveLabel || 'Updating...') : 'Creating...'}
              </>
            ) : (
              isEditing ? (saveLabel || 'Update Company') : 'Create Company'
            )}
          </Button>
          {isEditing && (
            <DeleteCompanyButton companyId={company.id} onCompanyDeleted={onCompanyDeleted} />
          )}
          
          {onCancel && (
            <Button
              type="button"
              variant="secondary"
              onClick={onCancel}
              disabled={loading}
            >
              Cancel
            </Button>
          )}
        </div>
      </form>
    </div>
  );
} 