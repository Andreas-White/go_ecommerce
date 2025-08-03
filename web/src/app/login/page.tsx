'use client';
import { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import Link from 'next/link';
import './page.css';
import { Alert, Button } from '@/components/ui';

function LoginForm() {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [loginAlert, setLoginAlert] = useState<{
    type: 'success' | 'error' | 'info';
    message: string;
  } | null>(null);
  const { login, user, loading: authLoading } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();

  // Check if user is already logged in
  useEffect(() => {
    if (user && !authLoading) {
      router.push('/');
    }
  }, [user, authLoading, router]);

  // Check for success message from registration
  useEffect(() => {
    const messageParam = searchParams.get('message');
    if (messageParam) {
      setLoginAlert({ type: 'info', message: messageParam });
    }
  }, [searchParams]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) return;

    setLoading(true);
    try {
      await login(formData.email, formData.password);
      router.push('/');
    } catch (error) {
      setErrors({ general: 'Invalid email or password. Please try again.' });
      setLoginAlert({
        type: 'error',
        message: 'Invalid email or password. Please try again.',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: '' }));
    }
    if (errors.general) {
      setErrors((prev) => ({ ...prev, general: '' }));
    }
  };

  // Show loading state while checking authentication
  if (authLoading) {
    return (
      <div className="login-loading-container">
        <div>Loading...</div>
      </div>
    );
  }

  return (
    <div className="login-container">
      <h1 className="login-title">Log In</h1>

      {errors.general && (
        <div className="login-message-error">{errors.general}</div>
      )}

      {loginAlert && (
        <Alert type={loginAlert.type} onClose={() => setLoginAlert(null)}>
          {loginAlert.message}
        </Alert>
      )}

      <form onSubmit={handleSubmit}>
        <div className="login-form-group">
          <label className="login-label">Email *</label>
          <input
            type="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            className={`login-input${errors.email ? ' login-input-error' : ''}`}
          />
          {errors.email && (
            <div className="login-error-text">{errors.email}</div>
          )}
        </div>

        <div className="login-form-group">
          <label className="login-label">Password *</label>
          <input
            type="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            className={`login-input${
              errors.password ? ' login-input-error' : ''
            }`}
          />
          {errors.password && (
            <div className="login-error-text">{errors.password}</div>
          )}
        </div>

        <Button
          type="submit"
          variant="primary"
          disabled={loading}
          isLoading={loading}
        >
          Login
        </Button>
      </form>

      <div className="login-link-register">
        Don't have an account?{' '}
        <Link href="/register" className="login-link-primary">
          Sign up
        </Link>
      </div>

      <div className="login-link-forgot">
        <Button variant="tertiary" href="/change-password">
          Forgot your password?
        </Button>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="login-loading-container">
          <div>Loading...</div>
        </div>
      }
    >
      <LoginForm />
    </Suspense>
  );
}
