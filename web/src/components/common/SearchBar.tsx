import React from 'react';
import Input from '../ui/Input';
import './SearchBar.css';

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
      <button type="submit" className="search-bar-button">
        🔍
      </button>
    </form>
  );
} 