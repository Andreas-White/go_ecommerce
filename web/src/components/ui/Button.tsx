'use client';

import React from 'react';
import Link from 'next/link';
import styles from './Button.module.css';
import Spinner from './Spinner';

type Variant = 'primary' | 'secondary' | 'tertiary' | 'destructive';
type Size = 'sm' | 'md' | 'lg';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  size?: Size;
  isLoading?: boolean;
  href?: string;
  className?: string;
  children: React.ReactNode;
}

export default function Button({
  variant = 'primary',
  size = 'md',
  isLoading = false,
  disabled = false,
  href,
  className = '',
  children,
  ...props
}: ButtonProps) {
  const classes = [
    styles.button,
    styles[variant],
    styles[size],
    isLoading ? styles.loading : '',
    className
  ].filter(Boolean).join(' ');

  if (href) {
    return (
      <Link
        href={href}
        className={classes}
        role="button"
        aria-disabled={disabled || isLoading}
        tabIndex={disabled || isLoading ? -1 : 0}
        onClick={(e) => {
          if (disabled || isLoading) {
            e.preventDefault();
            return;
          }
          if (props.onClick) {
            props.onClick(e as any);
          }
        }}
      >
        {isLoading ? <Spinner /> : children}
      </Link>
    );
  }

  return (
    <button
      type={props.type || 'button'}
      className={classes}
      disabled={disabled || isLoading}
      aria-disabled={disabled || isLoading}
      {...props}
    >
      {isLoading ? <Spinner /> : children}
    </button>
  );
} 