import React, { useState } from 'react';
import './ShippingForm.css';
import { Button } from '@/components/ui';
import { ShippingInfo } from '@/types';

interface ShippingFormProps {
  shippingInfo: ShippingInfo;
  onSubmit: (shipping: ShippingInfo) => void;
}

const shippingMethods = [
  {
    value: 'standard',
    label: 'Standard Shipping',
    cost: 5.99,
    days: '3-5 business days',
  },
  {
    value: 'express',
    label: 'Express Shipping',
    cost: 12.99,
    days: '1-2 business days',
  },
  {
    value: 'overnight',
    label: 'Overnight Shipping',
    cost: 24.99,
    days: 'Next business day',
  },
];

export default function ShippingForm({
  shippingInfo,
  onSubmit,
}: ShippingFormProps) {
  const [formData, setFormData] = useState<ShippingInfo>(shippingInfo);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});
  const [shakeErrorFields, setShakeErrorFields] = useState(false);

  const handleInputChange = (
    field: keyof ShippingInfo,
    value: string | number
  ) => {
    validateForm();
    if (field === 'method') {
      const selected = shippingMethods.find((m) => m.value === value);
      setFormData((prev) => ({
        ...prev,
        method: value as string,
        cost: selected ? selected.cost : prev.cost,
      }));
    } else {
      setFormData((prev) => ({ ...prev, [field]: value }));
    }

    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: '' }));
    }
  };

  const validateForm = (): boolean => {
    const newErrors: { [key: string]: string } = {};

    if (!formData.address.trim()) {
      newErrors.address = 'Address is required';
    }
    if (!formData.city.trim()) {
      newErrors.city = 'City is required';
    }
    if (!formData.country.trim()) {
      newErrors.country = 'Country is required';
    }
    if (!formData.zip_code.trim()) {
      newErrors.zip_code = 'ZIP code is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (validateForm()) {
      onSubmit(formData);
    } else {
      setShakeErrorFields(true);
      setTimeout(() => setShakeErrorFields(false), 400);
    }
  };

  const selectedMethod = shippingMethods.find(
    (method) => method.value === formData.method
  );

  return (
    <div className="shipping-form">
      <h2 className="form-title">Shipping Information</h2>
      <p className="form-subtitle">
        Please provide your delivery address and shipping method.
      </p>

      <form onSubmit={handleSubmit}>
        <div className="form-section">
          <h3 className="section-title">Delivery Address</h3>

          <div className="form-group">
            <label htmlFor="address" className="form-label">
              Street Address *
            </label>
            <input
              type="text"
              id="address"
              value={formData.address}
              onChange={(e) => handleInputChange('address', e.target.value)}
              className={`form-input ${errors.address ? 'error' : ''} ${shakeErrorFields && errors.address ? 'shake' : ''}`}
              placeholder="Enter your street address"
            />
            {errors.address && (
              <span className="error-message">{errors.address}</span>
            )}
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="city" className="form-label">
                City *
              </label>
              <input
                type="text"
                id="city"
                value={formData.city}
                onChange={(e) => handleInputChange('city', e.target.value)}
                className={`form-input ${errors.city ? 'error' : ''} ${shakeErrorFields && errors.city ? 'shake' : ''}`}
                placeholder="Enter city"
              />
              {errors.city && (
                <span className="error-message">{errors.city}</span>
              )}
            </div>

            <div className="form-group">
              <label htmlFor="zip_code" className="form-label">
                ZIP Code *
              </label>
              <input
                type="text"
                id="zip_code"
                value={formData.zip_code}
                onChange={(e) => handleInputChange('zip_code', e.target.value)}
                className={`form-input ${errors.zip_code ? 'error' : ''} ${shakeErrorFields && errors.zip_code ? 'shake' : ''}`}
                placeholder="Enter ZIP code"
              />
              {errors.zip_code && (
                <span className="error-message">{errors.zip_code}</span>
              )}
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="country" className="form-label">
              Country *
            </label>
            <input
              type="text"
              id="country"
              value={formData.country}
              onChange={(e) => handleInputChange('country', e.target.value)}
              className={`form-input ${errors.country ? 'error' : ''} ${shakeErrorFields && errors.country ? 'shake' : ''}`}
              placeholder="Enter country"
            />
            {errors.country && (
              <span className="error-message">{errors.country}</span>
            )}
          </div>
        </div>

        <div className="form-section">
          <h3 className="section-title">Shipping Method</h3>

          <div className="shipping-methods">
            {shippingMethods.map((method) => (
              <div
                key={method.value}
                className={`shipping-method ${
                  formData.method === method.value ? 'selected' : ''
                }`}
                onClick={() => handleInputChange('method', method.value)}
              >
                <div className="method-info">
                  <div className="method-header">
                    <input
                      type="radio"
                      name="shipping_method"
                      value={method.value}
                      checked={formData.method === method.value}
                      onChange={() => handleInputChange('method', method.value)}
                      className="method-radio"
                    />
                    <span className="method-label">{method.label}</span>
                    <span className="method-cost">
                      ${method.cost.toFixed(2)}
                    </span>
                  </div>
                  <p className="method-days">{method.days}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="form-actions">
          <Button type="submit" variant="primary">
            Continue to Payment
          </Button>
        </div>
      </form>
    </div>
  );
}
