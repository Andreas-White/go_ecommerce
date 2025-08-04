'use client';
import { useState, useEffect, useMemo, useCallback } from 'react';
import { useCart } from '../../context/CartContext';
import { useAuth } from '../../context/AuthContext';
import Link from 'next/link';
import './page.css';
import { api } from '../../lib/api';
import CartItem from '../../components/cart/CartItem';
import { Alert, Button } from '@/components/ui';
import ConfirmModal from '../../components/ui/ConfirmModal';

export default function CartPage() {
  const { cartItems, loading, removeFromCart, updateCartItems, clearCart } =
    useCart();
  const { user } = useAuth();
  const [updating, setUpdating] = useState(false);
  const [productMap, setProductMap] = useState<{ [productId: string]: any }>(
    {}
  );
  const [cartAlert, setCartAlert] = useState<{
    type: 'success' | 'error' | 'info';
    message: string;
  } | null>(null);
  const [isClearCartModalOpen, setIsClearCartModalOpen] = useState(false);

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
        setCartAlert({ type: 'info', message: 'Quantity updated' });
      } catch (error) {
        setCartAlert({ type: 'error', message: 'Failed to update quantity' });
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
        setCartAlert({ type: 'info', message: 'Item removed from cart' });
      } catch (error) {
        setCartAlert({
          type: 'error',
          message: 'Failed to remove item from cart',
        });
      } finally {
        setUpdating(false);
      }
    },
    [removeFromCart]
  );

  const handleClearCart = useCallback(() => {
    setIsClearCartModalOpen(true);
  }, []);

  const confirmClearCart = useCallback(async () => {
    setUpdating(true);
    try {
      await clearCart();
      setCartAlert({ type: 'info', message: 'Cart cleared' });
    } catch (error) {
      setCartAlert({ type: 'error', message: 'Failed to clear cart' });
    } finally {
      setUpdating(false);
      setIsClearCartModalOpen(false);
    }
  }, [clearCart]);

  const calculateTotal = useCallback(() => {
    console.log('cartItems: ', cartItems);
    return cartItems.reduce((total, item) => {
      console.log('item: ', item);
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

      {cartAlert && (
        <Alert type={cartAlert.type} onClose={() => setCartAlert(null)}>
          {cartAlert.message}
        </Alert>
      )}

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
                <Button
                  onClick={handleClearCart}
                  disabled={updating}
                  variant="secondary"
                >
                  Clear Cart
                </Button>
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
                <Button variant="secondary" href="/checkout">
                  Proceed to Checkout
                </Button>
              ) : (
                <div className="cart-login-prompt">
                  <p className="cart-login-text">Please log in to checkout</p>
                  <Button variant="secondary" href="/login">
                    Login to Checkout
                  </Button>
                </div>
              )}

              <div className="cart-continue-shopping">
                <Button variant="tertiary" href="/products">
                  Continue Shopping
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
      <ConfirmModal
        open={isClearCartModalOpen}
        onCancel={() => setIsClearCartModalOpen(false)}
        onConfirm={confirmClearCart}
        title="Clear Cart"
        message="Are you sure you want to clear your cart?"
        loading={updating}
        confirmLabel="Clear"
      />
    </div>
  );
}
