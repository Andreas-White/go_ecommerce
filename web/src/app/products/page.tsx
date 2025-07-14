"use client";
import { useState, useEffect } from 'react';
import { useCart } from '../../context/CartContext';
import { api } from '../../lib/api';
import Link from 'next/link';
import ProductFilterSort from '../../components/products/ProductFilterSort';
import SearchBar from '../../components/common/SearchBar';
import './page.css';

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  category_id: string;
  image_url?: string;
  company?: {
    name: string;
  };
}

export default function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [category, setCategory] = useState('');
  const [sortBy, setSortBy] = useState('name');
  const [sortOrder, setSortOrder] = useState('asc');
  const { addToCart } = useCart();

  useEffect(() => {
    fetchProducts();
  }, [searchTerm, category, sortBy, sortOrder]);

  const fetchProducts = async () => {
    setLoading(true);
    try {
      let productsData: Product[] = [];

      if (category) {
        // Use category-specific endpoint
        productsData = await api.get<Product[]>(`/products/category?category=${category}`) || [];
      } else {
        // Use general products endpoint with search and sort
        const params = new URLSearchParams();
        if (searchTerm) params.append('search', searchTerm);
        if (sortBy) params.append('sortBy', sortBy);
        if (sortOrder) params.append('sortOrder', sortOrder);

        productsData = await api.get<Product[]>(`/products?${params.toString()}`) || [];
      }

      setProducts(productsData);
      setError(null);
    } catch (error) {
      console.error('Failed to fetch products:', error);
      setError('Failed to load products. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleAddToCart = async (product: Product) => {
    try {
      await addToCart([{
        product_id: product.id,
        quantity: 1
      }]);
      // Show success message (you could add a toast notification here)
      alert(`${product.name} added to cart!`);
    } catch (error) {
      console.error('Failed to add to cart:', error);
      alert('Failed to add item to cart. Please try again.');
    }
  };

  const handleSearchSubmit = () => {
    // Trigger search - this will be handled by the useEffect
    fetchProducts();
  };

  if (loading) {
    return (
      <div className="products-loading-container">
        <div>Loading products...</div>
      </div>
    );
  }

  return (
    <div className="products-container">
      <h1 className="products-title">Products</h1>

      {/* Search Bar */}
      <div className="products-search-section">
        <SearchBar
          value={searchTerm}
          onChange={setSearchTerm}
          onSubmit={handleSearchSubmit}
        />
      </div>

      {/* Filter and Sort Controls */}
      <ProductFilterSort
        category={category}
        sortBy={sortBy}
        sortOrder={sortOrder}
        onCategoryChange={setCategory}
        onSortByChange={setSortBy}
        onSortOrderChange={setSortOrder}
      />

      {error && (
        <div className="products-error">
          {error}
        </div>
      )}

      {products.length === 0 ? (
        <div className="products-empty">
          <div className="products-empty-icon">📦</div>
          <h2 className="products-empty-title">No products found</h2>
          <p className="products-empty-text">
            {searchTerm || category ? 'Try adjusting your search terms or category filter.' : 'No products are available at the moment.'}
          </p>
        </div>
      ) : (
        <div className="products-grid">
          {products.map((product) => (
            <div key={product.id} className="product-card">
              {product.image_url && (
                <div className="product-image-placeholder">
                  <span className="product-image-text">Image</span>
                </div>
              )}
              
              <h3 className="product-name">
                {product.name}
              </h3>
              
              {product.company && (
                <p className="product-company">
                  by {product.company.name}
                </p>
              )}
              
              <p className="product-description">
                {product.description}
              </p>
              
              <div className="product-price-stock">
                <span className="product-price">
                  ${product.price.toFixed(2)}
                </span>
                <span className={`product-stock${product.stock > 0 ? ' product-stock-available' : ' product-stock-unavailable'}`}>
                  {product.stock > 0 ? `${product.stock} in stock` : 'Out of stock'}
                </span>
              </div>
              
              <div className="product-actions">
                <Link 
                  href={`/product/${product.id}`}
                  className="product-view-btn"
                >
                  View Details
                </Link>
                <button
                  onClick={() => handleAddToCart(product)}
                  disabled={product.stock <= 0}
                  className={`product-add-btn${product.stock > 0 ? ' product-add-btn-available' : ' product-add-btn-unavailable'}`}
                >
                  Add to Cart
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
} 