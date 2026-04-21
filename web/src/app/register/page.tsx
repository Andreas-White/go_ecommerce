'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import Link from 'next/link';
import './page.css';
import { Button } from '@/components/ui';
import validateEmail from '@/lib/validation';
import { Alert } from '@/components/ui';

export default function RegisterPage() {
  const [formData, setFormData] = useState({
    first_name: '',
    last_name: '',
    email: '',
    password: '',
    confirmPassword: '',
    is_producer: false,
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [shakeErrorFields, setShakeErrorFields] = useState(false);
  const { register, user, loading: authLoading } = useAuth();
  const [registerAlert, setRegisterAlert] = useState<{
    type: 'success' | 'error' | 'info';
    message: string;
  } | null>(null);
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

    const emailError = validateEmail(formData.email);
    if (emailError) {
      newErrors.email = emailError;
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

    if (!validateForm()) {
      setShakeErrorFields(true);
      setTimeout(() => setShakeErrorFields(false), 400);
      return;
    }

    setLoading(true);
    try {
      await register({
        first_name: formData.first_name,
        last_name: formData.last_name,
        email: formData.email,
        password: formData.password,
        is_producer: formData.is_producer,
      });
      // User is automatically logged in after registration
      setRegisterAlert({
        type: 'success',
        message: 'Registration successful!',
      });
      router.push('/');
    } catch (error) {
      setRegisterAlert({
        type: 'error',
        message: 'Registration failed. Please try again.',
      });
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    validateForm();
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: '' }));
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
      <h1 className="register-title">Create Account</h1>

      {registerAlert && (
        <Alert type={registerAlert.type} onClose={() => setRegisterAlert(null)}>
          {registerAlert.message}
        </Alert>
      )}

      <form onSubmit={handleSubmit}>
        <div className="register-form-group">
          <label className="register-label">First Name *</label>
          <input
            type="text"
            name="first_name"
            value={formData.first_name}
            onChange={handleChange}
            className={`register-input${
              errors.first_name ? ' register-input-error' : ''
            }${shakeErrorFields && errors.first_name ? ' shake' : ''}`}
          />
          {errors.first_name && (
            <div className="register-error-text">{errors.first_name}</div>
          )}
        </div>

        <div className="register-form-group">
          <label className="register-label">Last Name *</label>
          <input
            type="text"
            name="last_name"
            value={formData.last_name}
            onChange={handleChange}
            className={`register-input${
              errors.last_name ? ' register-input-error' : ''
            }${shakeErrorFields && errors.last_name ? ' shake' : ''}`}
          />
          {errors.last_name && (
            <div className="register-error-text">{errors.last_name}</div>
          )}
        </div>

        <div className="register-form-group">
          <label className="register-label">Email *</label>
          <input
            type="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            className={`register-input${
              errors.email ? ' register-input-error' : ''
            }${shakeErrorFields && errors.email ? ' shake' : ''}`}
          />
          {errors.email && (
            <div className="register-error-text">{errors.email}</div>
          )}
        </div>

        <div className="register-form-group">
          <label className="register-label">Password *</label>
          <input
            type="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            className={`register-input${
              errors.password ? ' register-input-error' : ''
            }${shakeErrorFields && errors.password ? ' shake' : ''}`}
          />
          {errors.password && (
            <div className="register-error-text">{errors.password}</div>
          )}
        </div>

        <div className="register-form-group">
          <label className="register-label">Confirm Password *</label>
          <input
            type="password"
            name="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            className={`register-input${
              errors.confirmPassword ? ' register-input-error' : ''
            }${shakeErrorFields && errors.confirmPassword ? ' shake' : ''}`}
          />
          {errors.confirmPassword && (
            <div className="register-error-text">{errors.confirmPassword}</div>
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

        <Button
          type="submit"
          variant="primary"
          disabled={loading}
          isLoading={loading}
        >
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
