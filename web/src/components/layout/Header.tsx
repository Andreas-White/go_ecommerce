'use client';
import Link from 'next/link';
import Image from 'next/image';
import { usePathname, useRouter } from 'next/navigation';
import { useState, useEffect, useRef } from 'react';
import { useAuth } from '../../context/AuthContext';
import { useCart } from '../../context/CartContext';
import SearchBar from '../common/SearchBar';
import './Header.css';
import { Button } from '@/components/ui';

export default function Header() {
  const { user, logout, loading } = useAuth();
  const { cartItems } = useCart();
  const pathname = usePathname();
  const router = useRouter();
  const [headerSearchTerm, setHeaderSearchTerm] = useState('');
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [cartPulsing, setCartPulsing] = useState(false);
  const prevCartCount = useRef(0);

  const cartCount = cartItems.reduce((total, item) => total + item.quantity, 0);

  const showSearchBar =
    pathname === '/' ||
    pathname === '/products' ||
    pathname.startsWith('/product/');

  const handleSearchSubmit = (term: string) => {
    if (term.trim()) {
      router.push(`/products?search=${encodeURIComponent(term.trim())}`);
    } else {
      router.push('/products');
    }
  };

  const handleLogout = async () => {
    await logout();
    if (typeof window !== 'undefined') {
      localStorage.removeItem('cart_session_id');
    }
    setMobileMenuOpen(false);
  };

  const isActive = (path: string) => pathname === path;

  useEffect(() => {
    if (cartCount > prevCartCount.current) {
      setCartPulsing(true);
      const timer = setTimeout(() => setCartPulsing(false), 300);
      return () => clearTimeout(timer);
    }
    prevCartCount.current = cartCount;
  }, [cartCount]);

  const closeMobileMenu = () => setMobileMenuOpen(false);

  return (
    <header className="header">
      <nav className="header-nav">
        <div className="header-left">
          <button
            className={`hamburger-btn ${mobileMenuOpen ? 'open' : ''}`}
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            aria-label="Toggle menu"
          >
            <span className="bar"></span>
            <span className="bar"></span>
            <span className="bar"></span>
          </button>
          <Link href="/" onClick={closeMobileMenu}>
            <div className="header-logo">
              <Image
                src="/Logo2.png"
                alt="SnapCart Logo"
                width={240}
                height={80}
                className="logo-dark"
                style={{ objectFit: 'contain' }}
                priority
              />
              <Image
                src="/Logo.png"
                alt="SnapCart Logo"
                width={240}
                height={80}
                className="logo-light"
                style={{ objectFit: 'contain' }}
              />
            </div>
          </Link>
        </div>

        {showSearchBar && (
          <div className="header-center">
            <SearchBar
              value={headerSearchTerm}
              onChange={setHeaderSearchTerm}
              onSubmit={handleSearchSubmit}
            />
          </div>
        )}

        <div className="header-right">
          <Link
            href="/cart"
            className={`header-cart ${isActive('/cart') ? 'active' : ''}`}
          >
            <span className="header-cart-icon">🛒</span>
            {cartCount > 0 && (
              <span
                className={`header-cart-badge ${cartPulsing ? 'pulsing' : ''}`}
              >
                {cartCount}
              </span>
            )}
          </Link>
          {loading ? (
            <span className="header-loading">Loading...</span>
          ) : user ? (
            <>
              <Link
                href="/profile"
                className={isActive('/profile') ? 'active' : ''}
              >
                {user.first_name}
              </Link>
              {user.is_producer && (
                <Link
                  href="/producer/dashboard"
                  className={`producer-link ${
                    isActive('/producer/dashboard') ? 'active' : ''
                  }`}
                >
                  Dashboard
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
              <Link
                href="/login"
                className={isActive('/login') ? 'active' : ''}
              >
                Login
              </Link>
              <Link
                href="/register"
                className={isActive('/register') ? 'active' : ''}
              >
                Register
              </Link>
            </>
          )}
        </div>
      </nav>

      {/* Mobile Menu */}
      <div
        className={`mobile-menu-overlay ${mobileMenuOpen ? 'open' : ''}`}
        onClick={closeMobileMenu}
      />
      <div className={`mobile-menu ${mobileMenuOpen ? 'open' : ''}`}>
        {showSearchBar && (
          <SearchBar
            value={headerSearchTerm}
            onChange={setHeaderSearchTerm}
            onSubmit={(term) => {
              handleSearchSubmit(term);
              closeMobileMenu();
            }}
          />
        )}
        {loading ? (
          <span className="header-loading">Loading...</span>
        ) : user ? (
          <>
            <Link
              href="/cart"
              className={isActive('/cart') ? 'active' : ''}
              onClick={closeMobileMenu}
            >
              🛒 Cart {cartCount > 0 && `(${cartCount})`}
            </Link>
            <Link
              href="/profile"
              className={isActive('/profile') ? 'active' : ''}
              onClick={closeMobileMenu}
            >
              {user.first_name}
            </Link>
            {user.is_producer && (
              <Link
                href="/producer/dashboard"
                className={`producer-link ${
                  isActive('/producer/dashboard') ? 'active' : ''
                }`}
                onClick={closeMobileMenu}
              >
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
            <Link
              href="/cart"
              className={isActive('/cart') ? 'active' : ''}
              onClick={closeMobileMenu}
            >
              🛒 Cart {cartCount > 0 && `(${cartCount})`}
            </Link>
            <Link
              href="/login"
              className={isActive('/login') ? 'active' : ''}
              onClick={closeMobileMenu}
            >
              Login
            </Link>
            <Link
              href="/register"
              className={isActive('/register') ? 'active' : ''}
              onClick={closeMobileMenu}
            >
              Register
            </Link>
          </>
        )}
      </div>
    </header>
  );
}
