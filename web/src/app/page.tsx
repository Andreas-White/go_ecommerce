"use client";
import { useAuth } from '../context/AuthContext';
import Link from 'next/link';
import './page.css';

export default function HomePage() {
  const { user } = useAuth();

  return (
    <div className="home-container">
      <div className="home-hero">
        <h1 className="home-title">
          Welcome to SnapCart
        </h1>
        <p className="home-subtitle">
          Your one-stop shop for quality products from trusted producers
        </p>
        
        <div className="home-cta-buttons">
          <Link href="/products" className="home-cta-btn home-cta-btn-primary">
            Browse Products
          </Link>
          {!user && (
            <Link href="/register" className="home-cta-btn home-cta-btn-secondary">
              Create Account
            </Link>
          )}
        </div>
      </div>

      <div className="home-features">
        <h2 className="home-features-title">
          Why Choose Us?
        </h2>
        
        <div className="home-features-grid">
          <div className="home-feature-card">
            <div className="home-feature-icon">🛒</div>
            <h3 className="home-feature-title">Easy Shopping</h3>
            <p className="home-feature-description">
              Browse and purchase products with our intuitive shopping experience
            </p>
          </div>
          
          <div className="home-feature-card">
            <div className="home-feature-icon">🏭</div>
            <h3 className="home-feature-title">Direct from Producers</h3>
            <p className="home-feature-description">
              Connect directly with product producers for better quality and prices
            </p>
          </div>
          
          <div className="home-feature-card">
            <div className="home-feature-icon">🔒</div>
            <h3 className="home-feature-title">Secure Transactions</h3>
            <p className="home-feature-description">
              Your data and transactions are protected with industry-standard security
            </p>
          </div>
          
          <div className="home-feature-card">
            <div className="home-feature-icon">📦</div>
            <h3 className="home-feature-title">Fast Delivery</h3>
            <p className="home-feature-description">
              Quick and reliable shipping to get your products to you faster
            </p>
          </div>
        </div>
      </div>

      {user && (
        <div className="home-user-section">
          <h2 className="home-user-title">
            Welcome back, {user.first_name}!
          </h2>
          <div className="home-user-actions">
            <Link href="/products" className="home-user-btn">
              Continue Shopping
            </Link>
            <Link href="/profile" className="home-user-btn">
              View Profile
            </Link>
            {user.is_producer && (
              <Link href="/my-products" className="home-user-btn">
                Manage Products
              </Link>
            )}
          </div>
        </div>
      )}

      <div className="home-footer">
        <p className="home-footer-text">
          Ready to start shopping? Explore our product catalog today!
        </p>
        <Link href="/products" className="home-footer-btn">
          Shop Now
        </Link>
      </div>
    </div>
  );
}
