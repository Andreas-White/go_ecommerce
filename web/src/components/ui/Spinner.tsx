import React from 'react';
import styles from './Button.module.css';

interface SpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  color?: string;
}

export default function Spinner({ size = 'md', color = 'currentColor' }: SpinnerProps) {
  const sizeMap = {
    sm: 16,
    md: 20,
    lg: 28
  };

  const strokeWidthMap = {
    sm: 4,
    md: 5,
    lg: 4
  };

  const dimension = sizeMap[size];
  const strokeWidth = strokeWidthMap[size];

  return (
    <span className={styles.spinner} aria-label="Loading" role="status">
      <svg 
        width={dimension} 
        height={dimension} 
        viewBox="0 0 50 50"
        style={{ color }}
      >
        <circle
          cx="25"
          cy="25"
          r="20"
          fill="none"
          stroke="currentColor"
          strokeWidth={strokeWidth}
          strokeDasharray="31.4 31.4"
          strokeLinecap="round"
        >
          <animateTransform
            attributeName="transform"
            type="rotate"
            from="0 25 25"
            to="360 25 25"
            dur="0.8s"
            repeatCount="indefinite"
          />
        </circle>
      </svg>
    </span>
  );
}