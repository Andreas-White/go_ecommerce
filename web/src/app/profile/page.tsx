"use client";
import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '../../context/AuthContext';
import Link from 'next/link';
import './page.css';
import { Button } from '@/components/ui';

export default function ProfilePage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();

  // Redirect if not authenticated
  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login?message=Please log in to view your profile.');
    }
  }, [user, authLoading, router]);

  // Show loading if checking authentication
  if (authLoading) {
    return (
      <div className="profile-loading-container">
        <div>Checking authentication...</div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect
  }

  return (
    <div className="profile-container">
      <h1 className="profile-title">
        Profile
      </h1>

      <div className="profile-section">
        <h2 className="profile-section-title">Account Information</h2>
        
        <div className="profile-info-card">
          <div className="profile-info-row">
            <strong className="profile-info-label">Name:</strong>
            <span className="profile-info-value">
              {user.first_name} {user.last_name}
            </span>
          </div>
          
          <div className="profile-info-row">
            <strong className="profile-info-label">Email:</strong>
            <span className="profile-info-value">
              {user.email}
            </span>
          </div>
          
          <div className="profile-info-row">
            <strong className="profile-info-label">Account Type:</strong>
            <span className="profile-info-value">
              {user.is_producer ? 'Producer (Seller)' : 'Customer (Buyer)'}
            </span>
          </div>
        </div>
      </div>

      <div className="profile-section">
        <h2 className="profile-section-title">Account Settings</h2>
        
        <div className="profile-settings">
          <Button variant="primary" onClick={() => router.push('/change-password')}>
            Change Password
          </Button>

          {user.is_producer && (
            <Button variant="primary" onClick={() => router.push('/my-products')}>
              Manage My Products
            </Button>
          )}

          <Button variant="secondary" onClick={() => router.push('/orders')}>
            View My Orders
          </Button>
        </div>
      </div>

      <div className="profile-back-link">
        <Button variant="tertiary" onClick={() => router.push('/')}>
          Back to Home
        </Button>
      </div>
    </div>
  );
} 