"use client";
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import Link from 'next/link';
import './page.css';
import { Button } from '@/components/ui';

export default function RegisterPage() {
  const [formData, setFormData] = useState({
    first_name: '',
    last_name: '',
    email: '',
    password: '',
    confirmPassword: '',
    is_producer: false
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const { register, user, loading: authLoading } = useAuth();
  const router = useRouter();

  // Check if user is already logged in
  useEffect(() => {
    if (user && !authLoading) {
      router.push('/');
    }
  }, [user, authLoading, router]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.first_name.trim()) {
      newErrors.first_name = 'First name is required';
    }

    if (!formData.last_name.trim()) {
      newErrors.last_name = 'Last name is required';
    }

    if (!formData.email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Email is invalid';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else if (formData.password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    }

    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) return;

    setLoading(true);
    try {
      await register({
        first_name: formData.first_name,
        last_name: formData.last_name,
        email: formData.email,
        password: formData.password,
        is_producer: formData.is_producer
      });
      // User is automatically logged in after registration
      router.push('/');
    } catch (error) {
      console.error('Registration error:', error);
      setErrors({ general: 'Registration failed. Please try again.' });
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value
    }));
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }));
    }
    if (errors.general) {
      setErrors(prev => ({ ...prev, general: '' }));
    }
  };

  // Show loading state while checking authentication
  if (authLoading) {
    return (
      <div className="register-loading-container">
        <div>Loading...</div>
      </div>
    );
  }

  return (
    <div className="register-container">
      <h1 className="register-title">
        Create Account
      </h1>

      {errors.general && (
        <div className="register-message-error">
          {errors.general}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="register-form-group">
          <label className="register-label">
            First Name *
          </label>
          <input
            type="text"
            name="first_name"
            value={formData.first_name}
            onChange={handleChange}
            className={`register-input${errors.first_name ? ' register-input-error' : ''}`}
          />
          {errors.first_name && (
            <div className="register-error-text">
              {errors.first_name}
            </div>
          )}
        </div>

        <div className="register-form-group">
          <label className="register-label">
            Last Name *
          </label>
          <input
            type="text"
            name="last_name"
            value={formData.last_name}
            onChange={handleChange}
            className={`register-input${errors.last_name ? ' register-input-error' : ''}`}
          />
          {errors.last_name && (
            <div className="register-error-text">
              {errors.last_name}
            </div>
          )}
        </div>

        <div className="register-form-group">
          <label className="register-label">
            Email *
          </label>
          <input
            type="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            className={`register-input${errors.email ? ' register-input-error' : ''}`}
          />
          {errors.email && (
            <div className="register-error-text">
              {errors.email}
            </div>
          )}
        </div>

        <div className="register-form-group">
          <label className="register-label">
            Password *
          </label>
          <input
            type="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            className={`register-input${errors.password ? ' register-input-error' : ''}`}
          />
          {errors.password && (
            <div className="register-error-text">
              {errors.password}
            </div>
          )}
        </div>

        <div className="register-form-group">
          <label className="register-label">
            Confirm Password *
          </label>
          <input
            type="password"
            name="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            className={`register-input${errors.confirmPassword ? ' register-input-error' : ''}`}
          />
          {errors.confirmPassword && (
            <div className="register-error-text">
              {errors.confirmPassword}
            </div>
          )}
        </div>

        <div className="register-checkbox-group">
          <label className="register-checkbox-label">
            <input
              type="checkbox"
              name="is_producer"
              checked={formData.is_producer}
              onChange={handleChange}
              className="register-checkbox"
            />
            I want to sell products (Producer Account)
          </label>
        </div>

        <Button type="submit" variant="primary" disabled={loading} isLoading={loading}>
          Register
        </Button>
      </form>

      <div className="register-link-login">
        Already have an account?{' '}
        <Link href="/login" className="register-link-primary">
          Log in
        </Link>
      </div>
    </div>
  );
} 