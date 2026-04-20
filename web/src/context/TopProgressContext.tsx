'use client';
import React, { createContext, useContext, useState, useCallback, useEffect, useRef, ReactNode } from 'react';

interface TopProgressContextType {
  start: () => void;
  complete: () => void;
}

const TopProgressContext = createContext<TopProgressContextType | undefined>(undefined);

export const useTopProgress = () => {
  const context = useContext(TopProgressContext);
  if (!context) {
    throw new Error('useTopProgress must be used within TopProgressProvider');
  }
  return context;
};

export function TopProgressProvider({ children }: { children: ReactNode }) {
  const [progress, setProgress] = useState(0);
  const [visible, setVisible] = useState(false);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const completingRef = useRef(false);

  const start = useCallback(() => {
    completingRef.current = false;
    setProgress(0);
    setVisible(true);
    const timeout = setTimeout(() => {
      if (!completingRef.current) {
        setProgress(80);
      }
    }, 150);
    timerRef.current = timeout;
  }, []);

  const complete = useCallback(() => {
    completingRef.current = true;
    setProgress(100);
    setTimeout(() => {
      setVisible(false);
      setProgress(0);
    }, 300);
  }, []);

  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  return (
    <TopProgressContext.Provider value={{ start, complete }}>
      {children}
      {visible && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            height: '4px',
            width: `${progress}%`,
            backgroundColor: 'var(--primary-color)',
            zIndex: 9999,
            transition: progress < 100 ? 'width 1.5s ease-out' : 'width 0.3s ease-out, opacity 0.3s ease-out',
            opacity: progress < 100 ? 1 : 0,
          }}
        />
      )}
    </TopProgressContext.Provider>
  );
}
