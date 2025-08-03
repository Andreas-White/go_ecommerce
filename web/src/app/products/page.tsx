'use client';
import { useState, useEffect } from 'react';
import { useCart } from '../../context/CartContext';
import { api } from '../../lib/api';
import ProductFilterSort from '../../components/products/ProductFilterSort';
import SearchBar from '../../components/common/SearchBar';
import './page.css';
import { useRouter } from 'next/navigation';
import ProductGrid from '../../components/products/ProductGrid';
import '../../components/products/ProductGrid.css';
import Alert from '@/components/ui/Alert';

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
  const router = useRouter();
  const [cartAlert, setCartAlert] = useState<{
    type: 'success' | 'error';
    message: string;
  } | null>(null);
  useEffect(() => {
    fetchProducts();
  }, [searchTerm, category, sortBy, sortOrder]);

  const fetchProducts = async () => {
    setLoading(true);
    try {
      let productsData: Product[] = [];

      if (category) {
        // Use category-specific endpoint
        productsData =
          (await api.get<Product[]>(
            `/products/category?category=${category}`
          )) || [];
      } else {
        // Use general products endpoint with search and sort
        const params = new URLSearchParams();
        if (searchTerm) params.append('search', searchTerm);
        if (sortBy) params.append('sortBy', sortBy);
        if (sortOrder) params.append('sortOrder', sortOrder);

        productsData =
          (await api.get<Product[]>(`/products?${params.toString()}`)) || [];
      }

      setProducts(productsData);
      setError(null);
    } catch (error) {
      setError('Failed to load products. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleAddToCart = async (product: Product) => {
    try {
      await addToCart([
        {
          product_id: product.id,
          price: product.price,
          quantity: 1,
        },
      ]);
      setCartAlert({
        type: 'success',
        message: `${product.name} added to cart!`,
      });
    } catch (error) {
      setCartAlert({
        type: 'error',
        message: 'Failed to add item to cart. Please try again.',
      });
    }
    setTimeout(() => setCartAlert(null), 2000);
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

      {cartAlert && (
        <Alert type={cartAlert.type} onClose={() => setCartAlert(null)}>
          {cartAlert.message}
        </Alert>
      )}

      {error && <div className="products-error">{error}</div>}

      {products.length === 0 ? (
        <div className="products-empty">
          <div className="products-empty-icon">📦</div>
          <h2 className="products-empty-title">No products found</h2>
          <p className="products-empty-text">
            {searchTerm || category
              ? 'Try adjusting your search terms or category filter.'
              : 'No products are available at the moment.'}
          </p>
        </div>
      ) : (
        <ProductGrid
          products={products}
          onViewDetails={(productId) => router.push(`/product/${productId}`)}
          onAddToCart={handleAddToCart}
        />
      )}
    </div>
  );
}
