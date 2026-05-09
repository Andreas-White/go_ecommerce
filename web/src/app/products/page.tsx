'use client';
import { useState, useEffect, useCallback, Suspense } from 'react';
import { useCart } from '../../context/CartContext';
import { api } from '../../lib/api';
import ProductFilterSort from '../../components/products/ProductFilterSort';
import { useDebounce } from '../../hooks/useDebounce';
import './page.css';
import { useRouter, useSearchParams } from 'next/navigation';
import ProductGrid from '../../components/products/ProductGrid';
import '../../components/products/ProductGrid.css';
import Alert from '@/components/ui/Alert';
import { useTopProgress } from '@/context/TopProgressContext';
import ProductCardSkeleton from '@/components/products/ProductCardSkeleton';

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

function ProductsPageContent() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const debouncedSearchTerm = useDebounce(searchTerm, 400);
  const [category, setCategory] = useState('');
  const [sortBy, setSortBy] = useState('name');
  const [sortOrder, setSortOrder] = useState('asc');
  const { cartItems, addToCart, updateCartItems } = useCart();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [cartAlert, setCartAlert] = useState<{
    type: 'success' | 'error' | 'info';
    message: string;
  } | null>(null);
  const { start: startProgress, complete: completeProgress } = useTopProgress();

  useEffect(() => {
    const initialSearch = searchParams.get('search') || '';
    setSearchTerm(initialSearch);
  }, [searchParams]);

  const fetchProducts = useCallback(
    async (immediateSearchTerm?: string) => {
      setLoading(true);
      startProgress();
      try {
        let productsData: Product[] = [];
        const searchToUse = immediateSearchTerm ?? debouncedSearchTerm;

        if (category) {
          const params = new URLSearchParams();
          params.append('category', category);
          if (sortBy) params.append('sortBy', sortBy);
          if (sortOrder) params.append('sortOrder', sortOrder);

          productsData =
            (await api.get<Product[]>(
              `/products/category?${params.toString()}`
            )) || [];
        } else {
          const params = new URLSearchParams();
          if (searchToUse) params.append('search', searchToUse);
          if (sortBy) params.append('sortBy', sortBy);
          if (sortOrder) params.append('sortOrder', sortOrder);

          productsData =
            (await api.get<Product[]>(`/products?${params.toString()}`)) ||
            [];
        }
        const inStockProducts = productsData.filter(
          (product) => product.stock > 0
        );
        setProducts(inStockProducts);
        setError(null);
      } catch (error) {
        setError('Failed to load products. Please try again.');
      } finally {
        setLoading(false);
        completeProgress();
      }
    },
    [category, debouncedSearchTerm, sortBy, sortOrder, startProgress, completeProgress]
  );

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  const handleAddToCart = useCallback(
    async (product: Product) => {
      const existingCartItem = cartItems.find(
        (item) => item.product_id === product.id
      );

      try {
        if (existingCartItem) {
          if (existingCartItem.quantity < product.stock) {
            await updateCartItems([
              {
                product_id: product.id,
                price: product.price,
                quantity: existingCartItem.quantity + 1,
              },
            ]);
            setCartAlert({
              type: 'info',
              message: `Increased ${product.name} quantity!`,
            });
          } else {
            setCartAlert({
              type: 'error',
              message: `Cannot add more of ${product.name}. Stock limit reached.`,
            });
          }
        } else {
          await addToCart([
            {
              product_id: product.id,
              price: product.price,
              quantity: 1,
            },
          ]);
          setCartAlert({
            type: 'info',
            message: `${product.name} added to cart!`,
          });
        }
      } catch (error) {
        setCartAlert({
          type: 'error',
          message: 'Failed to update cart. Please try again.',
        });
      }
    },
    [cartItems, addToCart, updateCartItems]
  );

  const handleViewDetails = useCallback(
    (productId: string) => {
      router.push(`/product/${productId}`);
    },
    [router]
  );

  return (
    <div className="products-container">
      <ProductFilterSort
        category={category}
        sortBy={sortBy}
        sortOrder={sortOrder}
        onCategoryChange={setCategory}
        onSortByChange={setSortBy}
        onOrderChange={setSortOrder}
      />

      {cartAlert && (
        <Alert type={cartAlert.type} onClose={() => setCartAlert(null)}>
          {cartAlert.message}
        </Alert>
      )}

      {error && <div className="products-error">{error}</div>}

      {!loading && products.length === 0 && (
        <div className="products-empty">
          <div className="products-empty-icon">📦</div>
          <h2 className="products-empty-title">No products found</h2>
          <p className="products-empty-text">
            {searchTerm || category
              ? 'Try adjusting your search terms or category filter.'
              : 'No products are available at the moment.'}
          </p>
        </div>
      )}

      <ProductGrid
        products={products}
        onViewDetails={handleViewDetails}
        isLoading={loading}
      />
    </div>
  );
}

export default function ProductsPage() {
  return (
    <Suspense fallback={
      <div className="products-container">
        <div className="products-grid">
          {Array.from({ length: 8 }).map((_, i) => (
            <ProductCardSkeleton key={i} />
          ))}
        </div>
      </div>
    }>
      <ProductsPageContent />
    </Suspense>
  );
}