"use client";
import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from '../../context/AuthContext';
import { useCart } from '../../context/CartContext';
import './Header.css';
import { Button } from '@/components/ui';

export default function Header() {
  const { user, logout, loading } = useAuth();
  const { cartItems } = useCart();
  
  // Calculate total items in cart
  const cartCount = cartItems.reduce((total, item) => total + item.quantity, 0);

  const handleLogout = async () => {
    await logout();
    // Clear cart session for guest users
    if (typeof window !== 'undefined') {
      localStorage.removeItem('cart_session_id');
    }
  };

  return (
    <header className="header">
      <nav className="header-nav">
        <div className="header-left">
          <Link href="/">
            <div className="header-logo">
              <Image
                src="/Logo.png"
                alt="SnapCart Logo"
                width={240}
                height={80}
                className="logo-dark"
                style={{ objectFit: 'contain' }}
                priority
              />
              <Image
                src="/Logo2.png"
                alt="SnapCart Logo"
                width={240}
                height={80}
                className="logo-light"
                style={{ objectFit: 'contain'}}
              />
            </div>
          </Link>
          <Link href="/products">Products</Link>
          <Link href="/categories">Categories</Link>
        </div>
        <div className="header-right">
          <Link href="/cart" className="header-cart">
            <span className="header-cart-icon">🛒</span>
            {cartCount > 0 && (
              <span className="header-cart-badge">{cartCount}</span>
            )}
          </Link>
          {loading ? (
            <span className="header-loading">Loading...</span>
          ) : user ? (
            <>
              <Link href="/profile">{user.first_name}</Link>
              {user.is_producer && (
                <Link href="/producer/dashboard" className="producer-link">
                  Producer Dashboard
                </Link>
              )}
              <Button 
                variant="primary" 
                className="header-logout-btn" 
                onClick={handleLogout}
              >
                Logout
              </Button>
            </>
          ) : (
            <>
              <Link href="/login">Login</Link>
              <Link href="/register">Register</Link>
            </>
          )}
        </div>
      </nav>
    </header>
  );
} 