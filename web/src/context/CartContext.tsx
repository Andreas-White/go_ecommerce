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
      setCartId('user_cart');
    } else {
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

    // Optimistic update
    const newItems = [...cartItems];
    items.forEach(newItem => {
      const existingIndex = newItems.findIndex(item => item.product_id === newItem.product_id);
      if (existingIndex >= 0) {
        newItems[existingIndex].quantity += newItem.quantity;
      } else {
        newItems.push(newItem);
      }
    });
    setCartItems(newItems);

    try {
      // CSRF token will be fetched automatically
      await api.post('/cart/add', items, {}, true);
      // Don't refresh cart - we already updated optimistically
    } catch (error) {
      // Revert optimistic update on error
      setCartItems(cartItems);
      throw error;
    }
  };

  const removeFromCart = async (items: CartItem[]) => {
    if (authLoading) {
      return;
    }

    // Optimistic update
    const newItems = cartItems.filter(item => 
      !items.some(removeItem => removeItem.product_id === item.product_id)
    );
    setCartItems(newItems);

    try {
      // CSRF token will be fetched automatically
      await api.post('/cart/remove', items, {}, true);
      // Don't refresh cart - we already updated optimistically
    } catch (error) {
      // Revert optimistic update on error
      setCartItems(cartItems);
      throw error;
    }
  };

  const updateCartItems = async (items: CartItem[]) => {
    if (authLoading) {
      return;
    }

    // Optimistic update
    const newItems = [...cartItems];
    items.forEach(updateItem => {
      const existingIndex = newItems.findIndex(item => item.product_id === updateItem.product_id);
      if (existingIndex >= 0) {
        newItems[existingIndex].quantity = updateItem.quantity;
      }
    });
    setCartItems(newItems);

    try {
      // CSRF token will be fetched automatically
      await api.post('/cart/update', items, {}, true);
      // Don't refresh cart - we already updated optimistically
    } catch (error) {
      // Revert optimistic update on error
      setCartItems(cartItems);
      throw error;
    }
  };

  const clearCart = async () => {
    if (authLoading) {
      return;
    }

    // Optimistic update
    setCartItems([]);

    try {
      // CSRF token will be fetched automatically
      await api.post('/cart/clear', {}, {}, true);
      // Don't refresh cart - we already updated optimistically
    } catch (error) {
      // Revert optimistic update on error
      setCartItems(cartItems);
      throw error;
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