import React from 'react';
import './ProductFilterSort.css';

const categories = [
  { value: '', label: 'All Categories' },
  { value: 'Electronics', label: 'Electronics' },
  { value: 'Clothing', label: 'Clothing' },
  { value: 'Books', label: 'Books' },
  { value: 'Home & Garden', label: 'Home & Garden' },
  { value: 'Sports', label: 'Sports' },
  { value: 'Toys', label: 'Toys' },
  { value: 'Health', label: 'Health & Beauty' },
  { value: 'Automotive', label: 'Automotive' },
  { value: 'Food', label: 'Food & Beverages' },
  { value: 'Shoes', label: 'Shoes' },
  { value: 'Other', label: 'Other' },
];

const sortOptions = [
  { value: 'name', label: 'Name' },
  { value: 'price', label: 'Price' },
  { value: 'created_at', label: 'Date Added' },
];

const orderOptions = [
  { value: 'asc', label: 'Ascending' },
  { value: 'desc', label: 'Descending' },
];

export default React.memo(function ProductFilterSort({
  category, sortBy, sortOrder,
  onCategoryChange, onSortByChange, onOrderChange
}: {
  category: string;
  sortBy: string;
  sortOrder: string;
  onCategoryChange: (v: string) => void;
  onSortByChange: (v: string) => void;
  onOrderChange: (v: string) => void;
}) {
  return (
    <div className="product-filter-sort">
      <select 
        value={category} 
        onChange={e => onCategoryChange(e.target.value)} 
        className={`product-filter-select ${category ? 'has-value' : ''}`}
      >
        {categories.map(c => <option key={c.value} value={c.value}>{c.label}</option>)}
      </select>
      <select 
        value={sortBy} 
        onChange={e => onSortByChange(e.target.value)}
        className="product-filter-select has-value"
      >
        {sortOptions.map(s => <option key={s.value} value={s.value}>{s.label}</option>)}
      </select>
      <select 
        value={sortOrder} 
        onChange={e => onOrderChange(e.target.value)}
        className="product-filter-select has-value"
      >
        {orderOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  );
}); 