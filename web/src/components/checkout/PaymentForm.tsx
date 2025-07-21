import React, { useState } from 'react';
import './PaymentForm.css';
import { Button } from '@/components/ui';

interface PaymentInfo {
  payment_method: string;
}

interface PaymentFormProps {
  paymentInfo: PaymentInfo;
  onSubmit: (payment: PaymentInfo) => void;
  onBack: () => void;
}

const paymentMethods = [
  { 
    value: 'credit_card', 
    label: 'Credit Card', 
    description: 'Visa, Mastercard, American Express',
    icon: '💳'
  },
  { 
    value: 'paypal', 
    label: 'PayPal', 
    description: 'Pay with your PayPal account',
    icon: '🔗'
  },
  { 
    value: 'bank_transfer', 
    label: 'Bank Transfer', 
    description: 'Direct bank transfer',
    icon: '🏦'
  }
];

export default function PaymentForm({ paymentInfo, onSubmit, onBack }: PaymentFormProps) {
  const [formData, setFormData] = useState<PaymentInfo>(paymentInfo);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});

  const handleInputChange = (field: keyof PaymentInfo, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  const validateForm = (): boolean => {
    const newErrors: { [key: string]: string } = {};

    if (!formData.payment_method) {
      newErrors.payment_method = 'Please select a payment method';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (validateForm()) {
      onSubmit(formData);
    }
  };

  return (
    <div className="payment-form">
      <h2 className="form-title">Payment Information</h2>
      <p className="form-subtitle">Choose your preferred payment method.</p>
      
      <form onSubmit={handleSubmit}>
        <div className="form-section">
          <h3 className="section-title">Payment Method</h3>
          
          <div className="payment-methods">
            {paymentMethods.map((method) => (
              <div
                key={method.value}
                className={`payment-method ${formData.payment_method === method.value ? 'selected' : ''}`}
                onClick={() => handleInputChange('payment_method', method.value)}
              >
                <div className="method-info">
                  <div className="method-header">
                    <input
                      type="radio"
                      name="payment_method"
                      value={method.value}
                      checked={formData.payment_method === method.value}
                      onChange={() => handleInputChange('payment_method', method.value)}
                      className="method-radio"
                    />
                    <span className="method-icon">{method.icon}</span>
                    <div className="method-details">
                      <span className="method-label">{method.label}</span>
                      <span className="method-description">{method.description}</span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
          
          {errors.payment_method && (
            <span className="error-message">{errors.payment_method}</span>
          )}
        </div>

        <div className="payment-note">
          <div className="note-icon">🔒</div>
          <div className="note-content">
            <h4>Secure Payment</h4>
            <p>Your payment information is encrypted and secure. We never store your full payment details.</p>
          </div>
        </div>

        <div className="form-actions">
          <Button type="button" onClick={onBack} variant="secondary">
            Back to Shipping
          </Button>
          <Button type="submit" variant="primary">
            Review Order
          </Button>
        </div>
      </form>
    </div>
  );
} 