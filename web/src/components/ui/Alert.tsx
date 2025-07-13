import React from 'react';
import './Alert.css';

type AlertProps = {
  type?: 'success' | 'error' | 'info';
  children: React.ReactNode;
};

export default function Alert({ type = 'info', children }: AlertProps) {
  return (
    <div className={`alert alert-${type}`}>
      {children}
    </div>
  );
} 