import React from 'react';
import Input from '../ui/Input';
import './SearchBar.css';
import { Button } from '@/components/ui';

export default function SearchBar({ value, onChange, onSubmit }: {
  value: string;
  onChange: (v: string) => void;
  onSubmit?: () => void;
}) {
  return (
    <form
      onSubmit={e => {
        e.preventDefault();
        onSubmit && onSubmit();
      }}
      className="search-bar-form"
    >
      <Input
        type="text"
        value={value}
        onChange={e => onChange(e.target.value)}
        placeholder="Search products..."
        className="search-bar-input"
      />
      <Button type="submit" variant="primary" className="search-bar-button">
        🔍
      </Button>
    </form>
  );
} 