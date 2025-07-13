"use client";
import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { api } from '../lib/api';

interface User {
  first_name: string;
  last_name: string;
  email: string;
  is_producer: boolean;
  // Add other fields as needed
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  register: (data: Record<string, any>) => Promise<void>;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const checkAuth = async () => {
    try {
      // Try to fetch user data - if successful, user is authenticated
      const userData = await api.get<User>('/users/get-by-id');
      setUser(userData);
    } catch (error) {
      // User is not authenticated
      setUser(null);
    }
  };

  useEffect(() => {
    checkAuth().finally(() => {
      setLoading(false);
    });
  }, []);

  const login = async (email: string, password: string) => {
    setLoading(true);
    try {
      // First get CSRF token
      await api.getCSRFToken('/users/login');
      
      // Then login
      const res = await api.post<{ message: string; csrf_token: string }>(
        '/users/login', 
        { email, password },
        {},
        true // require CSRF
      );
      
      // User is now authenticated via cookies, fetch user data
      await checkAuth();
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    try {
      // First get CSRF token
      await api.getCSRFToken('/auth/logout');
      
      // Call logout endpoint to clear cookies
      await api.post('/auth/logout', {}, {}, true);
    } catch (error) {
      // Even if logout fails, clear local state
    } finally {
      setUser(null);
      // Clear any stored CSRF tokens
      if (typeof window !== 'undefined') {
        localStorage.removeItem('csrf_token');
      }
    }
  };

  const register = async (data: Record<string, any>) => {
    setLoading(true);
    try {
      // First get CSRF token
      await api.getCSRFToken('/users/register');
      
      // Then register
      const res = await api.post<{ message: string; csrf_token: string }>(
        '/users/register', 
        data,
        {},
        true // require CSRF
      );
      
      // User is automatically logged in after registration
      await checkAuth();
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, register, checkAuth }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}; 