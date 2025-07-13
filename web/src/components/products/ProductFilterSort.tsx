import React from 'react';
import './ProductFilterSort.css';

const categories = [
  { value: '', label: 'All Categories' },
  { value: 'electronics', label: 'Electronics' },
  { value: 'fashion', label: 'Fashion' },
  { value: 'books', label: 'Books' },
  // Add more categories as needed
];

const sortOptions = [
  { value: '', label: 'Default' },
  { value: 'price', label: 'Price' },
  { value: 'name', label: 'Name' },
];

const orderOptions = [
  { value: 'asc', label: 'Ascending' },
  { value: 'desc', label: 'Descending' },
];

export default function ProductFilterSort({
  category, sortBy, sortOrder,
  onCategoryChange, onSortByChange, onSortOrderChange
}: {
  category: string;
  sortBy: string;
  sortOrder: string;
  onCategoryChange: (v: string) => void;
  onSortByChange: (v: string) => void;
  onSortOrderChange: (v: string) => void;
}) {
  return (
    <div className="product-filter-sort">
      <select value={category} onChange={e => onCategoryChange(e.target.value)} className="product-filter-select">
        {categories.map(c => <option key={c.value} value={c.value}>{c.label}</option>)}
      </select>
      <select value={sortBy} onChange={e => onSortByChange(e.target.value)} className="product-filter-select">
        {sortOptions.map(s => <option key={s.value} value={s.value}>{s.label}</option>)}
      </select>
      <select value={sortOrder} onChange={e => onSortOrderChange(e.target.value)} className="product-filter-select">
        {orderOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  );
} 