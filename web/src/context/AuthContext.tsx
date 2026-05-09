"use client";
import React, { createContext, useContext, useState, useEffect, useCallback, useMemo, ReactNode } from 'react';
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
  changePassword: (data: { current_password: string; new_password: string }) => Promise<void>;
  checkAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const checkAuth = useCallback(async () => {
    try {
      const userData = await api.get<User>('/users/get-by-id');
      setUser(userData);
    } catch (error) {
      setUser(null);
    }
  }, []);

  useEffect(() => {
    checkAuth().finally(() => {
      setLoading(false);
    });
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    setLoading(true);
    try {
      await api.post<{ message: string; csrf_token: string }>(
        '/users/login', 
        { email, password },
        {},
        true
      );
      await checkAuth();
    } catch (error) {
      if (error instanceof Error && error.message.includes('CSRF')) {
        api.clearCSRFToken();
        try {
          await api.post<{ message: string; csrf_token: string }>(
            '/users/login', 
            { email, password },
            {},
            true
          );
          await checkAuth();
        } catch (retryError) {
          throw retryError;
        }
      } else {
        throw error;
      }
    } finally {
      setLoading(false);
    }
  }, [checkAuth]);

  const logout = useCallback(async () => {
    try {
      await api.post('/auth/logout', {}, {}, true);
    } catch (error) {
    } finally {
      setUser(null);
      api.clearCSRFToken();
    }
  }, []);

  const register = useCallback(async (data: Record<string, any>) => {
    setLoading(true);
    try {
      await api.post<{ message: string; csrf_token: string }>(
        '/users/register', 
        data,
        {},
        true
      );
      await checkAuth();
    } catch (error) {
      if (error instanceof Error && error.message.includes('CSRF')) {
        api.clearCSRFToken();
        try {
          await api.post<{ message: string; csrf_token: string }>(
            '/users/register', 
            data,
            {},
            true
          );
          await checkAuth();
        } catch (retryError) {
          throw retryError;
        }
      } else {
        throw error;
      }
    } finally {
      setLoading(false);
    }
  }, [checkAuth]);

  const changePassword = useCallback(async (data: { current_password: string; new_password: string }) => {
    try {
      await api.post<{ message: string }>(
        '/auth/change-password',
        data,
        {},
        true
      );
    } catch (error) {
      throw error;
    }
  }, []);

  const value = useMemo(() => ({
    user,
    loading,
    login,
    logout,
    register,
    changePassword,
    checkAuth
  }), [user, loading, login, logout, register, changePassword, checkAuth]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}; 