'use client';
import './error.css';
import { Button } from '@/components/ui';

export default function Error({ error, reset }: { error: Error & { digest?: string }, reset: () => void }) {
  return (
    <div className="error-container">
      <h1 className="error-title">500 - Server Error</h1>
      <p className="error-message">Oops! Something went wrong on our end.</p>
      <p className="error-message">Please try again, or come back later.</p>
      <div className="error-actions">
        <Button variant="secondary" onClick={() => reset()}>
          Retry
        </Button>
        <Button variant="tertiary" href='/'>Go to Home</Button>
        <Button variant="tertiary" href='/products'>Browse Products</Button>
      </div>
    </div>
  );
} 