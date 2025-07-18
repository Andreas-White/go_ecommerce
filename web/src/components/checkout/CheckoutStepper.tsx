import React from 'react';
import './CheckoutStepper.css';

interface CheckoutStepperProps {
  currentStep: number;
}

const steps = [
  { number: 1, title: 'Shipping', description: 'Delivery Information' },
  { number: 2, title: 'Payment', description: 'Payment Method' },
  { number: 3, title: 'Review', description: 'Order Summary' }
];

export default function CheckoutStepper({ currentStep }: CheckoutStepperProps) {
  return (
    <div className="checkout-stepper">
      {steps.map((step, index) => (
        <div key={step.number} className="stepper-step">
          <div className={`step-number ${currentStep >= step.number ? 'active' : ''}`}>
            {currentStep > step.number ? (
              <span className="step-check">✓</span>
            ) : (
              step.number
            )}
          </div>
          <div className="step-content">
            <h3 className={`step-title ${currentStep >= step.number ? 'active' : ''}`}>
              {step.title}
            </h3>
            <p className={`step-description ${currentStep >= step.number ? 'active' : ''}`}>
              {step.description}
            </p>
          </div>
          {index < steps.length - 1 && (
            <div className={`step-connector ${currentStep > step.number ? 'active' : ''}`} />
          )}
        </div>
      ))}
    </div>
  );
} 