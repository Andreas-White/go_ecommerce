import { serverFetch } from '@/lib/api.server';
import ProductsList, { Product } from './_components/ProductsList';
import ProductCardSkeleton from '@/components/products/ProductCardSkeleton';

async function getProducts(): Promise<Product[]> {
  try {
    const products = await serverFetch<Product[]>('/products');
    return products?.filter((p: Product) => p.stock > 0) || [];
  } catch {
    return [];
  }
}

export default async function ProductsPage() {
  const products = await getProducts();

  if (products.length === 0) {
    return (
      <div className="products-container">
        <div className="products-grid">
          {Array.from({ length: 8 }).map((_, i) => (
            <ProductCardSkeleton key={i} />
          ))}
        </div>
      </div>
    );
  }

  return <ProductsList initialProducts={products} />;
}

export const metadata = {
  title: 'Products | SnapCart',
  description: 'Browse our product catalog',
};