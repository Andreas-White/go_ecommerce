'use client';
import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';
import { useCart } from '../../../context/CartContext';
import { useAuth } from '../../../context/AuthContext';
import { api } from '../../../lib/api';
import Link from 'next/link';
import './page.css';
import AddReviewForm from '../../../components/products/AddReviewForm';
import { Button } from '@/components/ui';

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
    address?: string;
    city?: string;
    country?: string;
  };
}

interface Review {
  id: string;
  product_id: string;
  user_id: string;
  rating: number;
  comment: string;
  created_at: string;
  user?: {
    first_name: string;
    last_name: string;
  };
}

export default function ProductDetailsPage() {
  const params = useParams();
  const productId = params.id as string;

  const [product, setProduct] = useState<Product | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [quantity, setQuantity] = useState(1);
  const [addingToCart, setAddingToCart] = useState(false);

  const { addToCart } = useCart();
  const { user } = useAuth();

  useEffect(() => {
    if (productId) {
      fetchProductDetails();
      fetchProductReviews();
    }
  }, [productId]);

  const fetchProductDetails = async () => {
    try {
      const productData = await api.get<Product>(`/product?id=${productId}`);
      setProduct(productData);
    } catch (error) {
      console.error('Failed to fetch product:', error);
      setError('Failed to load product details.');
    } finally {
      setLoading(false);
    }
  };

  const fetchProductReviews = async () => {
    try {
      const reviewsData = await api.get<Review[]>(
        `/reviews/get?product_id=${productId}`
      );
      setReviews(reviewsData || []);
    } catch (error) {
      console.error('Failed to fetch reviews:', error);
      // Don't set error for reviews as it's not critical
    } finally {
      setLoading(false);
    }
  };

  const handleAddToCart = async () => {
    if (!product) return;

    setAddingToCart(true);
    try {
      await addToCart([
        {
          product_id: product.id,
          price: product.price,
          quantity: quantity,
        },
      ]);
      alert(`${quantity} ${quantity === 1 ? 'item' : 'items'} added to cart!`);
    } catch (error) {
      console.error('Failed to add to cart:', error);
      alert('Failed to add item to cart. Please try again.');
    } finally {
      setAddingToCart(false);
    }
  };

  const calculateAverageRating = () => {
    if (reviews.length === 0) return 0;
    const total = reviews.reduce((sum, review) => sum + review.rating, 0);
    return total / reviews.length;
  };

  const renderStars = (rating: number) => {
    return '★'.repeat(Math.floor(rating)) + '☆'.repeat(5 - Math.floor(rating));
  };

  if (loading) {
    return (
      <div className="product-details-loading">
        <div>Loading product details...</div>
      </div>
    );
  }

  if (error || !product) {
    return (
      <div className="product-details-error">
        <h2>Product Not Found</h2>
        <p>{error || 'The product you are looking for does not exist.'}</p>
        <Link href="/products" className="product-details-back-link">
          Back to Products
        </Link>
      </div>
    );
  }

  return (
    <div className="product-details-container">
      <div className="product-details-breadcrumb">
        <Link href="/products">Products</Link>
        <span> / </span>
        <span>{product.name}</span>
      </div>

      <div className="product-details-content">
        <div className="product-details-main">
          <div className="product-details-image-section">
            {product.image_url ? (
              <img
                src={product.image_url}
                alt={product.name}
                className="product-details-image"
              />
            ) : (
              <div className="product-details-image-placeholder">
                <span>No Image Available</span>
              </div>
            )}
          </div>

          <div className="product-details-info">
            <h1 className="product-details-title">{product.name}</h1>

            {product.company && (
              <p className="product-details-company">
                by {product.company.name}
              </p>
            )}

            <div className="product-details-rating">
              <span className="product-details-stars">
                {renderStars(parseFloat(calculateAverageRating().toFixed(1)))}
              </span>
              <span className="product-details-rating-text">
                {calculateAverageRating().toFixed(1)} ({reviews.length} reviews)
              </span>
            </div>

            <div className="product-details-price">
              ${product.price.toFixed(2)}
            </div>

            <div className="product-details-stock">
              <span
                className={`product-details-stock-status${
                  product.stock > 0
                    ? ' product-details-stock-available'
                    : ' product-details-stock-unavailable'
                }`}
              >
                {product.stock > 0
                  ? `${product.stock} in stock`
                  : 'Out of stock'}
              </span>
            </div>

            <div className="product-details-description">
              <h3>Description</h3>
              <p>{product.description}</p>
            </div>

            {product.stock > 0 && (
              <div className="product-details-purchase">
                <div className="product-details-quantity">
                  <label htmlFor="quantity">Quantity:</label>
                  <select
                    id="quantity"
                    value={quantity}
                    onChange={(e) => setQuantity(parseInt(e.target.value))}
                    className="product-details-quantity-select"
                  >
                    {[...Array(Math.min(10, product.stock))].map((_, i) => (
                      <option key={i + 1} value={i + 1}>
                        {i + 1}
                      </option>
                    ))}
                  </select>
                </div>

                <Button
                  onClick={handleAddToCart}
                  disabled={addingToCart}
                  variant='primary'
                >
                  {addingToCart ? 'Adding...' : 'Add to Cart'}
                </Button>
              </div>
            )}

            {product.company && (
              <div className="product-details-seller">
                <h3>Seller Information</h3>
                <p>
                  <strong>{product.company.name}</strong>
                </p>
                {product.company.address && <p>{product.company.address}</p>}
                {product.company.city && product.company.country && (
                  <p>
                    {product.company.city}, {product.company.country}
                  </p>
                )}
              </div>
            )}
          </div>
        </div>

        <div className="product-details-reviews">
          <h2>Customer Reviews</h2>

          {reviews.length === 0 ? (
            <p className="product-details-no-reviews">
              No reviews yet. Be the first to review this product!
            </p>
          ) : (
            <div className="product-details-reviews-list">
              {reviews.map((review) => (
                <div key={review.id} className="product-details-review">
                  <div className="product-details-review-header">
                    <span className="product-details-review-stars">
                      {renderStars(review.rating)}
                    </span>
                    <span className="product-details-review-author">
                      {review.user
                        ? `${review.user.first_name} ${review.user.last_name}`
                        : 'Anonymous'}
                    </span>
                    <span className="product-details-review-date">
                      {new Date(review.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  {review.comment && (
                    <p className="product-details-review-comment">
                      {review.comment}
                    </p>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Add Review Form for authenticated users */}
          {user && (
            <AddReviewForm
              productId={product.id}
              onReviewSubmitted={fetchProductReviews}
            />
          )}
        </div>
      </div>
    </div>
  );
}
