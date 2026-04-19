import React, { useId } from 'react';
import './Input.css';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  required?: boolean;
}

export default function Input({ label, required, className, id, ...props }: InputProps) {
  const generatedId = useId();
  const inputId = id || generatedId;
  
  return (
    <div className={`ui-input-wrapper ${className || ''}`}>
      {label && (
        <label htmlFor={inputId} className="ui-input-label">
          {label}
          {required && <span className="ui-input-required" title="Required">*</span>}
        </label>
      )}
      <input
        {...props}
        id={inputId}
        className="ui-input"
        required={required}
      />
    </div>
  );
}