"use client";
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import Link from 'next/link';
import './page.css';
import { Button } from '@/components/ui';

export default function ChangePasswordPage() {
  const [formData, setFormData] = useState({
    current_password: '',
    new_password: '',
    confirm_password: ''
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const { changePassword, user, loading: authLoading } = useAuth();
  const router = useRouter();

  // Check if user is logged in
  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login?message=Please log in to change your password.');
    }
  }, [user, authLoading, router]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.current_password) {
      newErrors.current_password = 'Current password is required';
    }

    if (!formData.new_password) {
      newErrors.new_password = 'New password is required';
    } else if (formData.new_password.length < 8) {
      newErrors.new_password = 'New password must be at least 8 characters';
    }

    if (formData.new_password !== formData.confirm_password) {
      newErrors.confirm_password = 'Passwords do not match';
    }

    if (formData.current_password === formData.new_password) {
      newErrors.new_password = 'New password must be different from current password';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) return;

    setLoading(true);
    try {
      await changePassword({
        current_password: formData.current_password,
        new_password: formData.new_password
      });
      setSuccess(true);
      setFormData({
        current_password: '',
        new_password: '',
        confirm_password: ''
      });
      setErrors({});
    } catch (error) {
      console.error('Change password error:', error);
      setErrors({ general: 'Failed to change password. Please check your current password and try again.' });
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }));
    }
    if (errors.general) {
      setErrors(prev => ({ ...prev, general: '' }));
    }
    if (success) {
      setSuccess(false);
    }
  };

  // Show loading if checking authentication
  if (authLoading) {
    return (
      <div className="change-password-loading-container">
        <div>Checking authentication...</div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect
  }

  return (
    <div className="change-password-container">
      <h1 className="change-password-title">
        Change Password
      </h1>

      {success && (
        <div className="change-password-message-success">
          Password changed successfully!
        </div>
      )}

      {errors.general && (
        <div className="change-password-message-error">
          {errors.general}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div className="change-password-form-group">
          <label className="change-password-label">
            Current Password *
          </label>
          <input
            type="password"
            name="current_password"
            value={formData.current_password}
            onChange={handleChange}
            className={`change-password-input${errors.current_password ? ' change-password-input-error' : ''}`}
          />
          {errors.current_password && (
            <div className="change-password-error-text">
              {errors.current_password}
            </div>
          )}
        </div>

        <div className="change-password-form-group">
          <label className="change-password-label">
            New Password *
          </label>
          <input
            type="password"
            name="new_password"
            value={formData.new_password}
            onChange={handleChange}
            className={`change-password-input${errors.new_password ? ' change-password-input-error' : ''}`}
          />
          {errors.new_password && (
            <div className="change-password-error-text">
              {errors.new_password}
            </div>
          )}
        </div>

        <div className="change-password-form-group">
          <label className="change-password-label">
            Confirm New Password *
          </label>
          <input
            type="password"
            name="confirm_password"
            value={formData.confirm_password}
            onChange={handleChange}
            className={`change-password-input${errors.confirm_password ? ' change-password-input-error' : ''}`}
          />
          {errors.confirm_password && (
            <div className="change-password-error-text">
              {errors.confirm_password}
            </div>
          )}
        </div>
        <div className="change-password-button-container">
          <Button type="submit" variant="primary" disabled={loading} isLoading={loading}>
            Change Password
          </Button>

        </div>
      </form>

      <div className="change-password-back-link">
        <Button variant="tertiary" onClick={() => router.push('/profile')}>
          Back to Profile
        </Button>
      </div>
    </div>
  );
} 