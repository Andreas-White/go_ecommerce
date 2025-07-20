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
  changePassword: (data: { current_password: string; new_password: string }) => Promise<void>;
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
      // Login - CSRF token will be fetched automatically
      await api.post<{ message: string; csrf_token: string }>(
        '/users/login', 
        { email, password },
        {},
        true // require CSRF
      );
      
      // User is now authenticated via cookies, fetch user data
      await checkAuth();
    } catch (error) {
      // If it's a CSRF error, clear the token and retry once
      if (error instanceof Error && error.message.includes('CSRF')) {
        api.clearCSRFToken();
        try {
          await api.post<{ message: string; csrf_token: string }>(
            '/users/login', 
            { email, password },
            {},
            true // require CSRF
          );
          await checkAuth();
        } catch (retryError) {
          throw retryError; // Re-throw the retry error
        }
      } else {
        throw error; // Re-throw non-CSRF errors
      }
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    try {
      // Call logout endpoint to clear cookies - CSRF token will be fetched automatically
      await api.post('/auth/logout', {}, {}, true);
    } catch (error) {
      // Even if logout fails, clear local state
    } finally {
      setUser(null);
      // Clear CSRF token using the API client's function
      api.clearCSRFToken();
    }
  };

  const register = async (data: Record<string, any>) => {
    setLoading(true);
    try {
      // Register - CSRF token will be fetched automatically
      await api.post<{ message: string; csrf_token: string }>(
        '/users/register', 
        data,
        {},
        true // require CSRF
      );
      
      // User is automatically logged in after registration
      await checkAuth();
    } catch (error) {
      // If it's a CSRF error, clear the token and retry once
      if (error instanceof Error && error.message.includes('CSRF')) {
        api.clearCSRFToken();
        try {
          await api.post<{ message: string; csrf_token: string }>(
            '/users/register', 
            data,
            {},
            true // require CSRF
          );
          await checkAuth();
        } catch (retryError) {
          throw retryError; // Re-throw the retry error
        }
      } else {
        throw error; // Re-throw non-CSRF errors
      }
    } finally {
      setLoading(false);
    }
  };

  const changePassword = async (data: { current_password: string; new_password: string }) => {
    try {
      // Change password - CSRF token will be fetched automatically
      await api.post<{ message: string }>(
        '/auth/change-password',
        data,
        {},
        true // require CSRF
      );
    } catch (error) {
      throw error; // Re-throw to let the component handle it
    }
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, register, changePassword, checkAuth }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}; 