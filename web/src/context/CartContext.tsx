"use client";
import React, { createContext, useContext, useState, useEffect, useCallback, useMemo, ReactNode } from 'react';
import { api } from '../lib/api';
import { useAuth } from './AuthContext';
import { CartItem } from '../types';

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

  const getCart = useCallback(async () => {
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
  }, [authLoading]);

  const refreshCart = useCallback(async () => {
    await getCart();
  }, [getCart]);

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

  const addToCart = useCallback(async (items: CartItem[]) => {
    if (authLoading) {
      return;
    }

    setCartItems(prev => {
      const newItems = [...prev];
      items.forEach(newItem => {
        const existingIndex = newItems.findIndex(item => item.product_id === newItem.product_id);
        if (existingIndex >= 0) {
          newItems[existingIndex].quantity += newItem.quantity;
        } else {
          newItems.push(newItem);
        }
      });
      return newItems;
    });

    try {
      await api.post('/cart/add', items, {}, true);
    } catch (error) {
      setCartItems(prev => [...prev]);
      throw error;
    }
  }, [authLoading]);

  const removeFromCart = useCallback(async (items: CartItem[]) => {
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
  }, [authLoading, cartItems]);

  const updateCartItems = useCallback(async (items: CartItem[]) => {
    if (authLoading) {
      return;
    }

    const prevItems = [...cartItems];
    setCartItems(prev => {
      const newItems = [...prev];
      items.forEach(updateItem => {
        const existingIndex = newItems.findIndex(item => item.product_id === updateItem.product_id);
        if (existingIndex >= 0) {
          newItems[existingIndex].quantity = updateItem.quantity;
        }
      });
      return newItems;
    });

    try {
      await api.post('/cart/update', items, {}, true);
    } catch (error) {
      setCartItems(prevItems);
      throw error;
    }
  }, [authLoading, cartItems]);

  const clearCart = useCallback(async () => {
    if (authLoading) {
      return;
    }

    const prevItems = [...cartItems];
    setCartItems([]);

    try {
      await api.post('/cart/clear', {}, {}, true);
    } catch (error) {
      setCartItems(prevItems);
      throw error;
    }
  }, [authLoading, cartItems]);

  const value = useMemo(() => ({
    cartItems,
    cartId,
    loading,
    addToCart,
    removeFromCart,
    updateCartItems,
    clearCart,
    getCart,
    refreshCart
  }), [cartItems, cartId, loading, addToCart, removeFromCart, updateCartItems, clearCart, getCart, refreshCart]);

  return (
    <CartContext.Provider value={value}>
      {children}
    </CartContext.Provider>
  );
};

export const useCart = () => {
  const ctx = useContext(CartContext);
  if (!ctx) throw new Error('useCart must be used within CartProvider');
  return ctx;
}; 