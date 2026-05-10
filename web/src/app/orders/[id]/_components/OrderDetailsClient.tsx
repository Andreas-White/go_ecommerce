'use client';
import { useState, useEffect, useMemo, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { api } from '@/lib/api';
import Button from '@/components/ui/Button';
import OrderDetailsSkeleton from '@/components/ui/OrderDetailsSkeleton';
import OrderWithDetailsDisplay, { 
  OrderWithDetails, 
  ProductMap 
} from '@/components/orders/OrderWithDetailsDisplay';
import '../page.css';

interface OrderDetailsClientProps {
  orderDetails: OrderWithDetails;
}

function OrderDetailsClientContent({ orderDetails }: OrderDetailsClientProps) {
  const router = useRouter();
  const { user, loading: authLoading } = useAuth();
  const [productMap, setProductMap] = useState<ProductMap>({});
  const [loadingProducts, setLoadingProducts] = useState(true);

  const uniqueProductIds = useMemo(() => {
    return Array.from(new Set(orderDetails.items.map(item => item.product_id)));
  }, [orderDetails.items]);

  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
      return;
    }
  }, [user, authLoading, router]);

  const fetchProducts = useCallback(async () => {
    if (uniqueProductIds.length === 0) {
      setLoadingProducts(false);
      return;
    }

    try {
      const results = await Promise.all(
        uniqueProductIds.map(async (id) => {
          try {
            const product = await api.get<{ price?: number; name?: string; image_url?: string }>(`/product?id=${id}`);
            return { id, product };
          } catch {
            return { id, product: null };
          }
        })
      );

      const newMap: ProductMap = {};
      results.forEach(({ id, product }) => {
        if (product) {
          newMap[id] = product;
        }
      });
      setProductMap(newMap);
    } finally {
      setLoadingProducts(false);
    }
  }, [uniqueProductIds]);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  if (authLoading || loadingProducts) {
    return (
      <div className="order-details-loading">
        <OrderDetailsSkeleton />
      </div>
    );
  }

  return (
    <div className="order-details-container">
      <div className="order-details-header">
        <Button onClick={() => router.back()} variant="tertiary">
          ← Back to Orders
        </Button>
        <h1 className="order-details-title">Order Details</h1>
        <p className="order-details-subtitle">Order #{orderDetails.order.id.slice(0, 8)}</p>
      </div>
      <OrderWithDetailsDisplay orderDetails={orderDetails} productMap={productMap} />
    </div>
  );
}

export default function OrderDetailsClient({ orderDetails }: OrderDetailsClientProps) {
  return <OrderDetailsClientContent orderDetails={orderDetails} />;
}