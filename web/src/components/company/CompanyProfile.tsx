'use client';

import './CompanyProfile.css';
import { useState } from 'react';
import dynamic from 'next/dynamic';
import DeleteCompanyButton from './DeleteCompanyButton';
import { Button } from '@/components/ui';

const CreateUpdateCompanyForm = dynamic(
  () => import('./CreateUpdateCompanyForm'),
  { loading: () => <div>Loading form...</div>, ssr: false }
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

interface CompanyProfileProps {
  company: Company;
  onCompanyUpdated?: (company: Company) => void;
  onCompanyDeleted?: () => void;
}

export default function CompanyProfile({ company, onCompanyUpdated, onCompanyDeleted }: CompanyProfileProps) {
  const [editing, setEditing] = useState(false);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  if (editing) {
    return (
      <div className="company-profile">
        <CreateUpdateCompanyForm
          company={company}
          onCompanyUpdated={(updated) => {
            setEditing(false);
            onCompanyUpdated?.(updated);
          }}
          onCancel={() => setEditing(false)}
          onCompanyDeleted={() => {
            setEditing(false);
            onCompanyDeleted?.();
          }}
          saveLabel="Save"
        />
      </div>
    );
  }

  return (
    <div className="company-profile">
      <div className="company-header">
        <h3>{company.name}</h3>
        <div className="company-rating">
          <span className="rating-stars">
            {'★'.repeat(Math.round(company.review_average || 0))}
            {'☆'.repeat(5 - Math.round(company.review_average || 0))}
          </span>
          <span className="rating-text">
            {company.review_average ? company.review_average.toFixed(1) : ''} 
            ({company.review_count || 0} reviews)
          </span>
        </div>
      </div>

      <div className="company-details">
        <div className="detail-group">
          <label>Address:</label>
          <p>{company.address}</p>
        </div>
        <div className="detail-group">
          <label>City:</label>
          <p>{company.city}</p>
        </div>
        <div className="detail-group">
          <label>Country:</label>
          <p>{company.country}</p>
        </div>
        <div className="detail-group">
          <label>ZIP Code:</label>
          <p>{company.zip_code}</p>
        </div>
      </div>
      <div className="company-actions">
        <Button variant="primary" onClick={() => setEditing(true)}>
          Update Company
        </Button>
        <DeleteCompanyButton companyId={company.id} onCompanyDeleted={onCompanyDeleted} />
      </div>
    </div>
  );
} 