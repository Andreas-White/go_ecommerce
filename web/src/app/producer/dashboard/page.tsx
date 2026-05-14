'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { useAuth } from '@/context/AuthContext';
import { api } from '@/lib/api';
import CompanyProfile from '@/components/company/CompanyProfile';
import ProducerProductList from '@/components/products/ProducerProductList';
import Alert from '@/components/ui/Alert';
import Spinner from '@/components/ui/Spinner';
import './page.css';
import ProducerOrderList, {
  ProducerOrderLike,
} from '@/components/orders/ProducerOrderList';
import { Button } from '@/components/ui';

const CreateProductForm = dynamic(
  () => import('@/components/products/CreateProductForm'),
  { loading: () => <div>Loading form...</div>, ssr: false }
);

const CreateUpdateCompanyForm = dynamic(
  () => import('@/components/company/CreateUpdateCompanyForm'),
  { loading: () => <div>Loading form...</div>, ssr: false }
);

interface Company {
  id: string;
  name: string;
  address: string;
  city: string;
  country: string;
  zip_code: string;
  review_average: number;
  review_count: number;
  created_at: string;
  updated_at: string;
}

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  stock: number;
  category_id: string;
  image_url: string;
  company: Company;
}

export default function ProducerDashboard() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const [company, setCompany] = useState<Company | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [showCreateCompany, setShowCreateCompany] = useState(false);
  const [showCreateProduct, setShowCreateProduct] = useState(false);
  const [loadingData, setLoadingData] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [orders, setOrders] = useState<ProducerOrderLike[]>([]);

  useEffect(() => {
    if (!loading) {
      if (!user) {
        router.push('/login');
        return;
      }

      if (!user.is_producer) {
        router.push('/');
        return;
      }

      fetchProducerData();
    }
  }, [loading, user, router]);

  const fetchProducerData = async () => {
    try {
      setLoadingData(true);
      setError(null);

      // Fetch company details
      try {
        const companyData = await api.get<Company>('/companies/get-by-user');
        setCompany(companyData);
      } catch (err) {
        // Company not found is expected for new producers
        setCompany(null);
      }

      // Fetch products
      try {
        const productsData = await api.get<Product[]>('/products/my-products');
        setProducts(productsData || []);
      } catch (err) {
        // No products is expected for new producers
        setProducts([]);
      }

      // Fetch orders for producer
      try {
        const ordersData = await api.get<ProducerOrderLike[]>(
          '/orders/producer'
        );
        setOrders(ordersData || []);
      } catch (err) {
        setOrders([]);
      }
    } catch (err) {
      setError('Failed to load producer data');
      setProducts([]);
    } finally {
      setLoadingData(false);
    }
  };

  const handleCompanyCreated = (newCompany: Company) => {
    setCompany(newCompany);
    setShowCreateCompany(false);
  };

  const handleCompanyUpdated = (updatedCompany: Company) => {
    setCompany(updatedCompany);
  };

  const handleCompanyDeleted = () => {
    setCompany(null);
  };

  const handleProductCreated = (newProduct: Product) => {
    setProducts([...products, newProduct]);
    setShowCreateProduct(false);
  };

  const handleProductUpdated = (updatedProduct: Product) => {
    setProducts(
      products.map((p) => (p.id === updatedProduct.id ? updatedProduct : p))
    );
  };

  const handleProductDeleted = (productId: string) => {
    setProducts(products.filter((p) => p.id !== productId));
  };

  const handleOrderFulfilled = (orderId: string, newStatus: string) => {
    setOrders((orders) =>
      orders.map((order) => {
        if (order.id === orderId) {
          return { ...order, status: newStatus };
        }
        if (order.order?.id === orderId) {
          return {
            ...order,
            order: { ...order.order, status: newStatus },
          };
        }
        return order;
      })
    );
  };

  if (loading || loadingData) {
    return (
      <div className="producer-dashboard">
        <div className="loading-container">
          <Spinner />
        </div>
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect to login
  }

  if (!user.is_producer) {
    return null; // Will redirect to home
  }

  return (
    <div className="producer-dashboard">
      <div className="dashboard-header">
        <h1>Producer Dashboard</h1>
        <p>Manage your company profile and products</p>
      </div>

      {error && (
        <Alert type="error" onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <div className="dashboard-content">
        {/* Company Section */}
        <section className="company-section">
          <div className="section-header">
            <h2>Company Profile</h2>
            {!company && (
              <Button
                variant="primary"
                onClick={() => setShowCreateCompany(true)}
              >
                Create Company
              </Button>
            )}
          </div>

          {showCreateCompany ? (
            <CreateUpdateCompanyForm
              onCompanyCreated={handleCompanyCreated}
              onCancel={() => setShowCreateCompany(false)}
              onCompanyDeleted={handleCompanyDeleted}
            />
          ) : company ? (
            <div className="company-content">
              <CompanyProfile
                company={company}
                onCompanyUpdated={handleCompanyUpdated}
                onCompanyDeleted={handleCompanyDeleted}
              />
            </div>
          ) : (
            <div className="no-company">
              <p>You haven't created a company profile yet.</p>
              <Button
                variant="primary"
                onClick={() => setShowCreateCompany(true)}
              >
                Create Your First Company
              </Button>
            </div>
          )}
        </section>

        {/* Products Section */}
        <section className="products-section">
          <div className="section-header">
            <h2>Your Products</h2>
            {company && (
              <Button
                variant="primary"
                onClick={() => setShowCreateProduct(true)}
              >
                Add New Product
              </Button>
            )}
          </div>

          {showCreateProduct ? (
            <CreateProductForm
              onProductCreated={handleProductCreated}
              onCancel={() => setShowCreateProduct(false)}
            />
          ) : (
            <ProducerProductList
              products={products || []}
              onProductUpdated={handleProductUpdated}
              onProductDeleted={handleProductDeleted}
            />
          )}

          {!company && (products || []).length === 0 && (
            <div className="no-products">
              <p>Create a company profile first to add products.</p>
            </div>
          )}

          {company && (products || []).length === 0 && (
            <div className="no-products">
              <p>You haven't added any products yet.</p>
              <Button
                variant="primary"
                onClick={() => setShowCreateProduct(true)}
              >
                Add Your First Product
              </Button>
            </div>
          )}
        </section>

        {/* Orders Section */}
        <section className="orders-section">
          <div className="section-header">
            <h2>Orders for Your Products</h2>
          </div>
          <ProducerOrderList
            orders={orders}
            onOrderFulfilled={handleOrderFulfilled}
          />
        </section>
      </div>
    </div>
  );
}
