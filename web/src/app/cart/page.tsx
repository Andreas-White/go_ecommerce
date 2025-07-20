'use client';
import { useState, useEffect, useMemo, useCallback } from 'react';
import { useCart } from '../../context/CartContext';
import { useAuth } from '../../context/AuthContext';
import Link from 'next/link';
import './page.css';
import { api } from '../../lib/api';
import CartItem from '../../components/cart/CartItem';

export default function CartPage() {
  const { cartItems, loading, removeFromCart, updateCartItems, clearCart } =
    useCart();
  const { user } = useAuth();
  const [updating, setUpdating] = useState(false);
  const [productMap, setProductMap] = useState<{ [productId: string]: any }>(
    {}
  );

  // Memoize unique product IDs to avoid unnecessary re-computations
  const uniqueProductIds = useMemo(() => {
    return Array.from(new Set(cartItems.map((item) => item.product_id)));
  }, [cartItems]);

  // Fetch products only when the set of unique product IDs changes
  useEffect(() => {
    const fetchProducts = async () => {
      if (uniqueProductIds.length === 0) {
        setProductMap({});
        return;
      }

      const newMap: { [productId: string]: any } = {};
      await Promise.all(
        uniqueProductIds.map(async (id) => {
          // Only fetch if we don't already have this product
          if (!productMap[id]) {
            try {
              const product = await api.get(`/product?id=${id}`);
              newMap[id] = product;
            } catch (e) {
              // ignore error, leave undefined
            }
          } else {
            // Keep existing product data
            newMap[id] = productMap[id];
          }
        })
      );
      setProductMap(newMap);
    };

    fetchProducts();
  }, [uniqueProductIds]); // Only depend on unique product IDs, not cartItems

  const handleQuantityChange = useCallback(
    async (productId: string, newQuantity: number) => {
      if (newQuantity <= 0) {
        // Remove item if quantity is 0 or negative
        await handleRemoveItem(productId);
        return;
      }

      setUpdating(true);
      try {
        await updateCartItems([
          { product_id: productId, quantity: newQuantity },
        ]);
      } catch (error) {
        console.error('Failed to update quantity:', error);
      } finally {
        setUpdating(false);
      }
    },
    [updateCartItems]
  );

  const handleRemoveItem = useCallback(
    async (productId: string) => {
      setUpdating(true);
      try {
        await removeFromCart([{ product_id: productId, quantity: 1 }]);
      } catch (error) {
        console.error('Failed to remove item:', error);
      } finally {
        setUpdating(false);
      }
    },
    [removeFromCart]
  );

  const handleClearCart = useCallback(async () => {
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
  }, [clearCart]);

  const calculateTotal = useCallback(() => {
    return cartItems.reduce((total, item) => {
      const price = item?.price || 0;
      return total + price * item.quantity;
    }, 0);
  }, [cartItems]);

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

              {cartItems.map((item, index) => {
                const product = productMap[item.product_id];
                return (
                  <CartItem
                    key={item.product_id}
                    item={item}
                    product={product}
                    updating={updating}
                    onQuantityChange={handleQuantityChange}
                    onRemove={handleRemoveItem}
                  />
                );
              })}

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
                  <span className="cart-summary-value">
                    ${calculateTotal().toFixed(2)}
                  </span>
                </div>
                <div className="cart-summary-row">
                  <span className="cart-summary-label">Shipping:</span>
                  <span className="cart-summary-value">
                    Calculated at checkout
                  </span>
                </div>
                <hr className="cart-summary-divider" />
                <div className="cart-summary-row cart-summary-total">
                  <span className="cart-summary-label">Total:</span>
                  <span className="cart-summary-value">
                    ${calculateTotal().toFixed(2)}
                  </span>
                </div>
              </div>

              {user ? (
                <Link href="/checkout" className="cart-checkout-btn">
                  Proceed to Checkout
                </Link>
              ) : (
                <div className="cart-login-prompt">
                  <p className="cart-login-text">Please log in to checkout</p>
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
