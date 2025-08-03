'use client';
import React, { useState, useEffect } from 'react';
import './Alert.css';

type AlertProps = {
  type?: 'success' | 'error' | 'info';
  children: React.ReactNode;
  onClose: () => void;
  duration?: number;
};

export default function Alert({
  type = 'info',
  children,
  onClose,
  duration = 5000,
}: AlertProps) {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    // Animate in
    const enterTimeout = setTimeout(() => setVisible(true), 50);

    // Set timer to animate out
    const exitTimeout = setTimeout(() => {
      setVisible(false);
    }, duration);

    // Set timer to fully remove from DOM after animation
    const removeTimeout = setTimeout(onClose, duration + 300); // 300ms for transition

    return () => {
      clearTimeout(enterTimeout);
      clearTimeout(exitTimeout);
      clearTimeout(removeTimeout);
    };
  }, [duration, onClose]);

  const handleClose = () => {
    setVisible(false);
    setTimeout(onClose, 300); // Wait for animation to finish
  };

  return (
    <div className={`alert alert-${type} ${visible ? 'visible' : ''}`}>
      {children}
      <button
        onClick={handleClose}
        className="alert-close-button"
        aria-label="Close"
      >
        &times;
      </button>
    </div>
  );
}
