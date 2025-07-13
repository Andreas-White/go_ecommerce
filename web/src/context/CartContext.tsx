"use client";
import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { api } from '../lib/api';
import { useAuth } from './AuthContext';

interface CartItem {
  id?: string;
  product_id: string;
  quantity: number;
  price?: number;
  product_name?: string;
  image_url?: string;
}

interface CartContextType {
  cartItems: CartItem[];
  cartId: string | null;
  loading: boolean;
  addToCart: (items: CartItem[]) => Promise<void>;
  removeFromCart: (items: CartItem[]) => Promise<void>;
  updateCartItems: (items: CartItem[]) => Promise<void>;
  clearCart: () => Promise<void>;
  getCart: () => Promise<void>;
  refreshCart: () => Promise<void>;
}

const CartContext = createContext<CartContextType | undefined>(undefined);

export const CartProvider = ({ children }: { children: ReactNode }) => {
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [cartId, setCartId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const { user, loading: authLoading } = useAuth();

  // Generate or get session ID for guest carts
  const getSessionId = (): string => {
    if (typeof window !== 'undefined') {
      let sessionId = localStorage.getItem('cart_session_id');
      if (!sessionId) {
        sessionId = 'session_' + Math.random().toString(36).substr(2, 9);
        localStorage.setItem('cart_session_id', sessionId);
      }
      return sessionId;
    }
    return '';
  };

  const getCart = async () => {
    if (authLoading) {
      return;
    }

    setLoading(true);
    try {
      // The backend GetCartItems handler doesn't expect a request body
      // It gets cart info from JWT (authenticated users) or session cookie (guest users)
      const items = await api.post<CartItem[]>('/cart/get');
      setCartItems(items || []);
    } catch (error) {
      setCartItems([]);
    } finally {
      setLoading(false);
    }
  };

  const refreshCart = async () => {
    await getCart();
  };

  // Initialize cart when user changes or on mount
  useEffect(() => {
    if (authLoading) {
      return;
    }

    if (user) {
      // For authenticated users, we'll need to get their cart ID
      // This could be stored in user context or fetched from user profile
      // For now, we'll use a placeholder
      setCartId('user_cart');
    } else {
      // For guest users, use session ID
      const sessionId = getSessionId();
      setCartId(sessionId);
    }
  }, [user, authLoading]);

  // Load cart when user/auth state changes
  useEffect(() => {
    if (!authLoading) {
      getCart();
    }
  }, [user, authLoading]);

  const addToCart = async (items: CartItem[]) => {
    if (authLoading) {
      return;
    }

    setLoading(true);
    try {
      // First get CSRF token
      await api.getCSRFToken('/cart/add');
      
      await api.post('/cart/add', items, {}, true);
      await getCart(); // Refresh cart after adding items
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const removeFromCart = async (items: CartItem[]) => {
    if (authLoading) {
      return;
    }

    setLoading(true);
    try {
      // First get CSRF token
      await api.getCSRFToken('/cart/remove');
      
      await api.post('/cart/remove', items, {}, true);
      await getCart(); // Refresh cart after removing items
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const updateCartItems = async (items: CartItem[]) => {
    if (authLoading) {
      return;
    }

    setLoading(true);
    try {
      // First get CSRF token
      await api.getCSRFToken('/cart/update');
      
      await api.post('/cart/update', items, {}, true);
      await getCart(); // Refresh cart after updating items
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const clearCart = async () => {
    if (authLoading) {
      return;
    }

    setLoading(true);
    try {
      // First get CSRF token
      await api.getCSRFToken('/cart/clear');
      
      await api.post('/cart/clear', {}, {}, true);
      setCartItems([]);
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
    }
  };

  return (
    <CartContext.Provider value={{
      cartItems,
      cartId,
      loading,
      addToCart,
      removeFromCart,
      updateCartItems,
      clearCart,
      getCart,
      refreshCart
    }}>
      {children}
    </CartContext.Provider>
  );
};

export const useCart = () => {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error('useCart must be used within CartProvider');
  return ctx;
}; 