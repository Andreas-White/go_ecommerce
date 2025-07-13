import React from 'react';

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  children: React.ReactNode;
  className?: string;
};

export default function Button({ children, className = '', ...props }: ButtonProps) {
  return (
    <button
      className={`button-primary ${className}`}
      {...props}
    >
      {children}
    </button>
  );
} 