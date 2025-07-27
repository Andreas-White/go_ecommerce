// Centralized API client for Go backend with cookie-based authentication and CSRF protection
// This file is not a React component and should not trigger Fast Refresh
// @refresh reset
const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

// CSRF token management
let csrfToken: string | null = null;
let isFetchingCSRFToken = false;

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

function clearCSRFToken() {
  csrfToken = null;
  if (typeof window !== 'undefined') {
    localStorage.removeItem('csrf_token');
  }
}

// Determine which endpoint to use for CSRF token fetching based on the request path
function getCSRFTokenEndpoint(path: string): string {
  if (path.startsWith('/users/login')) {
    return '/users/login';
  } else if (path.startsWith('/users/register')) {
    return '/users/register';
  } else if (path.startsWith('/auth/')) {
    return '/auth/change-password';
  } else {
    // Default endpoint for most operations
    return '/users/register';
  }
}

// Get CSRF token from server
async function getCSRFTokenFromServer(endpoint: string): Promise<string> {
  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      method: 'GET',
      credentials: 'include',
    });
    
    if (!response.ok) {
      throw new Error(`Failed to get CSRF token: ${response.status}`);
    }
    
    // The CSRF token should now be in the cookie
    const token = getCSRFToken();
    if (token) {
      return token;
    }
    throw new Error('CSRF token not found in cookies after fetch');
  } catch (error) {
    throw new Error('Failed to get CSRF token from server');
  }
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
    let token = getCSRFToken();
    
    // If no token available and we're not already fetching one, fetch it automatically
    if (!token && !isFetchingCSRFToken) {
      isFetchingCSRFToken = true;
      try {
        const csrfEndpoint = getCSRFTokenEndpoint(path);
        token = await getCSRFTokenFromServer(csrfEndpoint);
      } catch (error) {
        // Don't throw here, let the request proceed and handle the error response
      } finally {
        isFetchingCSRFToken = false;
      }
    }
    
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
    clearCSRFToken();
    throw new Error('Unauthorized');
  }

  if (res.status === 403) {
    // Clear CSRF token on 403 errors as it might be invalid
    clearCSRFToken();
    throw new Error('Forbidden - CSRF token may be invalid');
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

export const api = {
  get: <T>(path: string, headers?: Record<string, string>) => request<T>('GET', path, undefined, headers),
  post: <T>(path: string, body?: any, headers?: Record<string, string>, requireCSRF: boolean = false) => 
    request<T>('POST', path, body, headers, requireCSRF),
  put: <T>(path: string, body?: any, headers?: Record<string, string>, requireCSRF: boolean = false) => 
    request<T>('PUT', path, body, headers, requireCSRF),
  delete: <T>(path: string, body?: any, headers?: Record<string, string>, requireCSRF: boolean = false) => 
    request<T>('DELETE', path, body, headers, requireCSRF),
  getCSRFToken: getCSRFTokenFromServer,
  clearCSRFToken, // Export for manual clearing if needed
}; 