import React from 'react';
import './Input.css';

type InputProps = React.InputHTMLAttributes<HTMLInputElement>;

export default function Input(props: InputProps) {
  return (
    <input
      {...props}
      className={`ui-input ${props.className || ''}`}
    />
  );
} 