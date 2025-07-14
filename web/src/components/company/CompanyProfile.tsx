'use client';

import './CompanyProfile.css';

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
}

export default function CompanyProfile({ company }: CompanyProfileProps) {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

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
            {company.review_average ? company.review_average.toFixed(1) : 'No'} 
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

        <div className="detail-group">
          <label>Created:</label>
          <p>{formatDate(company.created_at)}</p>
        </div>

        <div className="detail-group">
          <label>Last Updated:</label>
          <p>{formatDate(company.updated_at)}</p>
        </div>
      </div>
    </div>
  );
} 