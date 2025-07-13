"use client";
import { useState } from 'react';
import { useCart } from '../../context/CartContext';
import { useAuth } from '../../context/AuthContext';
import Link from 'next/link';
import './page.css';

export default function CartPage() {
  const { cartItems, loading, removeFromCart, updateCartItems, clearCart } = useCart();
  const { user } = useAuth();
  const [updating, setUpdating] = useState(false);

  const handleQuantityChange = async (productId: string, newQuantity: number) => {
    if (newQuantity <= 0) {
      // Remove item if quantity is 0 or negative
      await handleRemoveItem(productId);
      return;
    }

    setUpdating(true);
    try {
      await updateCartItems([{ product_id: productId, quantity: newQuantity }]);
    } catch (error) {
      console.error('Failed to update quantity:', error);
    } finally {
      setUpdating(false);
    }
  };

  const handleRemoveItem = async (productId: string) => {
    setUpdating(true);
    try {
      await removeFromCart([{ product_id: productId, quantity: 1 }]);
    } catch (error) {
      console.error('Failed to remove item:', error);
    } finally {
      setUpdating(false);
    }
  };

  const handleClearCart = async () => {
    if (confirm('Are you sure you want to clear your cart?')) {
      setUpdating(true);
      try {
        await clearCart();
      } catch (error) {
        console.error('Failed to clear cart:', error);
      } finally {
        setUpdating(false);
      }
    }
  };

  const calculateTotal = () => {
    return cartItems.reduce((total, item) => total + (item.price || 0) * item.quantity, 0);
  };

  if (loading) {
    return (
      <div className="cart-loading-container">
        <div>Loading cart...</div>
      </div>
    );
  }

  return (
    <div className="cart-container">
      <h1 className="cart-title">Shopping Cart</h1>

      {cartItems.length === 0 ? (
        <div className="cart-empty">
          <div className="cart-empty-icon">🛒</div>
          <h2 className="cart-empty-title">Your cart is empty</h2>
          <p className="cart-empty-text">
            Start shopping to add items to your cart!
          </p>
          <Link href="/products" className="button-primary">
            Browse Products
          </Link>
        </div>
      ) : (
        <div className="cart-content">
          <div className="cart-items-section">
            <div className="cart-items-card">
              <h2 className="cart-items-title">Cart Items</h2>
              
              {cartItems.map((item, index) => (
                <div key={index} className={`cart-item${index < cartItems.length - 1 ? ' cart-item-border' : ''}`}>
                  <div className="cart-item-info">
                    <h3 className="cart-item-name">
                      {item.product_name || `Product ${item.product_id}`}
                    </h3>
                    <p className="cart-item-price">
                      ${(item.price || 0).toFixed(2)} each
                    </p>
                  </div>
                  
                  <div className="cart-item-quantity">
                    <button
                      onClick={() => handleQuantityChange(item.product_id, item.quantity - 1)}
                      disabled={updating}
                      className="cart-quantity-btn"
                    >
                      -
                    </button>
                    <span className="cart-quantity-display">
                      {item.quantity}
                    </span>
                    <button
                      onClick={() => handleQuantityChange(item.product_id, item.quantity + 1)}
                      disabled={updating}
                      className="cart-quantity-btn"
                    >
                      +
                    </button>
                  </div>
                  
                  <div className="cart-item-subtotal">
                    <p className="cart-item-total">
                      ${((item.price || 0) * item.quantity).toFixed(2)}
                    </p>
                  </div>
                  
                  <button
                    onClick={() => handleRemoveItem(item.product_id)}
                    disabled={updating}
                    className="cart-remove-btn"
                  >
                    Remove
                  </button>
                </div>
              ))}
              
              <div className="cart-clear-section">
                <button
                  onClick={handleClearCart}
                  disabled={updating}
                  className="cart-clear-btn"
                >
                  Clear Cart
                </button>
              </div>
            </div>
          </div>
          
          <div className="cart-summary-section">
            <div className="cart-summary-card">
              <h2 className="cart-summary-title">Order Summary</h2>
              
              <div className="cart-summary-details">
                <div className="cart-summary-row">
                  <span className="cart-summary-label">Subtotal:</span>
                  <span className="cart-summary-value">${calculateTotal().toFixed(2)}</span>
                </div>
                <div className="cart-summary-row">
                  <span className="cart-summary-label">Shipping:</span>
                  <span className="cart-summary-value">Calculated at checkout</span>
                </div>
                <hr className="cart-summary-divider" />
                <div className="cart-summary-row cart-summary-total">
                  <span className="cart-summary-label">Total:</span>
                  <span className="cart-summary-value">${calculateTotal().toFixed(2)}</span>
                </div>
              </div>
              
              {user ? (
                <Link href="/checkout" className="cart-checkout-btn">
                  Proceed to Checkout
                </Link>
              ) : (
                <div className="cart-login-prompt">
                  <p className="cart-login-text">
                    Please log in to checkout
                  </p>
                  <Link href="/login" className="cart-checkout-btn">
                    Login to Checkout
                  </Link>
                </div>
              )}
              
              <div className="cart-continue-shopping">
                <Link href="/products" className="cart-continue-link">
                  Continue Shopping
                </Link>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
} 