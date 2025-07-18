// Centralized API client for Go backend with cookie-based authentication and CSRF protection
// This file is not a React component and should not trigger Fast Refresh
// @refresh reset
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

// CSRF token management
let csrfToken: string | null = null;

function getCSRFToken(): string | null {
  if (typeof window !== 'undefined') {
    // Try to get from cookie first
    const cookies = document.cookie.split(';');
    const csrfCookie = cookies.find(cookie => cookie.trim().startsWith('csrf_token='));
    if (csrfCookie) {
      return csrfCookie.split('=')[1];
    }
  }
  return csrfToken;
}

function setCSRFToken(token: string) {
  csrfToken = token;
}

async function request<T>(
  method: string,
  path: string,
  body?: any,
  customHeaders?: Record<string, string>,
  requireCSRF: boolean = false
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...customHeaders,
  };

  // Add CSRF token for state-changing operations
  if (requireCSRF) {
    const token = getCSRFToken();
    if (token) {
      headers['X-CSRF-Token'] = token;
    }
  }

  const res = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
    credentials: 'include', // Include cookies for authentication
  });

  if (res.status === 401) {
    // Clear any stored tokens but don't redirect automatically
    // Let the calling code handle the redirect
    if (typeof window !== 'undefined') {
      localStorage.removeItem('csrf_token');
      csrfToken = null;
    }
    throw new Error('Unauthorized');
  }

  if (!res.ok) {
    const error = await res.text();
    throw new Error(error || 'API error');
  }

  if (res.status === 204) return undefined as T;
  
  const data = await res.json();
  
  // Extract CSRF token from response if present
  if (data && data.csrf_token) {
    setCSRFToken(data.csrf_token);
    if (typeof window !== 'undefined') {
      localStorage.setItem('csrf_token', data.csrf_token);
    }
  }
  
  return data;
}

// Get CSRF token from server
async function getCSRFTokenFromServer(endpoint: string): Promise<string> {
  try {
    await fetch(`${API_BASE_URL}${endpoint}`, {
      method: 'GET',
      credentials: 'include',
    });
    
    // The CSRF token should now be in the cookie
    const token = getCSRFToken();
    if (token) {
      return token;
    }
    throw new Error('Failed to get CSRF token');
  } catch (error) {
    throw new Error('Failed to get CSRF token from server');
  }
}

export const api = {
  get: <T>(path: string, headers?: Record<string, string>) => request<T>('GET', path, undefined, headers),
  post: <T>(path: string, body?: any, headers?: Record<string, string>, requireCSRF: boolean = false) => 
    request<T>('POST', path, body, headers, requireCSRF),
  put: <T>(path: string, body?: any, headers?: Record<string, string>, requireCSRF: boolean = false) => 
    request<T>('PUT', path, body, headers, requireCSRF),
  delete: <T>(path: string, body?: any, headers?: Record<string, string>, requireCSRF: boolean = false) => 
    request<T>('DELETE', path, body, headers, requireCSRF),
  getCSRFToken: getCSRFTokenFromServer,
}; 