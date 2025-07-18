import React, { useState } from 'react';
import { api } from '../../lib/api';
import { useAuth } from '../../context/AuthContext';
import './AddReviewForm.css';

interface AddReviewFormProps {
  productId: string;
  onReviewSubmitted: () => void;
}

export default function AddReviewForm({ productId, onReviewSubmitted }: AddReviewFormProps) {
  const { user } = useAuth();
  const [rating, setRating] = useState<number>(0);
  const [comment, setComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(false);
    if (!rating) {
      setError('Please select a rating.');
      return;
    }
    setSubmitting(true);
    try {
      // Fetch CSRF token
      const payload = {
        product_id: productId,
        rating,
        comment,
      }
      await api.post('/reviews/add', payload, undefined, true);
      setSuccess(true);
      setRating(0);
      setComment('');
      onReviewSubmitted();
    } catch (err: any) {
      setError(err?.message || 'Failed to submit review.');
    } finally {
      setSubmitting(false);
    }
  };

  if (!user) return null;

  return (
    <form className="add-review-form" onSubmit={handleSubmit} style={{ marginTop: 24 }}>
      <h3>Write a Review</h3>
      {error && <div className="add-review-error">{error}</div>}
      {success && <div className="add-review-success">Review submitted!</div>}
      <div className="add-review-rating">
        <label>Rating:</label>
        <select value={rating} onChange={e => setRating(Number(e.target.value))} required>
          <option value={0}>Select...</option>
          {[1,2,3,4,5].map(star => (
            <option key={star} value={star}>{star} Star{star > 1 ? 's' : ''}</option>
          ))}
        </select>
      </div>
      <div className="add-review-comment">
        <label>Comment:</label>
        <textarea
          value={comment}
          onChange={e => setComment(e.target.value)}
          rows={3}
          placeholder="Share your thoughts about this product..."
        />
      </div>
      <button type="submit" disabled={submitting} className="add-review-submit-btn">
        {submitting ? 'Submitting...' : 'Submit Review'}
      </button>
    </form>
  );
} 