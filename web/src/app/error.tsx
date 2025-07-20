'use client';
import Link from 'next/link';
import './error.css';

export default function Error({ error, reset }: { error: Error & { digest?: string }, reset: () => void }) {
  return (
    <div className="error-container">
      <h1 className="error-title">500 - Server Error</h1>
      <p className="error-message">Oops! Something went wrong on our end.</p>
      <p className="error-message">Please try again, or come back later.</p>
      <div className="error-actions">
        <button className="error-btn" onClick={() => reset()}>Retry</button>
        <Link href="/" className="error-link">Go to Home</Link>
        <Link href="/products" className="error-link">Browse Products</Link>
      </div>
    </div>
  );
} 