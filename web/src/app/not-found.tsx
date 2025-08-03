import './not-found.css';
import { Button } from '@/components/ui';

export default function NotFound() {
  return (
    <div className="not-found-container">
      <h1 className="not-found-title">404 - Page Not Found</h1>
      <p className="not-found-message">Sorry, the page you are looking for does not exist.</p>
      <div className="not-found-links">
        <Button variant="tertiary" href='/'>Go to Home</Button>
        <Button variant="tertiary" href='/products'>Browse Products</Button>
      </div>
    </div>
  );
} 