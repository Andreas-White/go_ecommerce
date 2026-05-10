import { serverFetch } from '@/lib/api.server';
import OrderDetailsClient from './_components/OrderDetailsClient';
import { OrderWithDetails } from '@/components/orders/OrderWithDetailsDisplay';
import OrderDetailsSkeleton from '@/components/ui/OrderDetailsSkeleton';

interface PageProps {
  params: Promise<{ id: string }>;
}

async function getOrderDetails(orderId: string): Promise<OrderWithDetails | null> {
  try {
    const details = await serverFetch<OrderWithDetails>('/orders/details', {
      method: 'POST',
      body: JSON.stringify({ order_id: orderId }),
    });
    return details;
  } catch {
    return null;
  }
}

export default async function OrderDetailsPage({ params }: PageProps) {
  const { id: orderId } = await params;
  const orderDetails = await getOrderDetails(orderId);

  if (!orderDetails) {
    return (
      <div className="order-details-container">
        <div className="order-details-loading">
          <OrderDetailsSkeleton />
        </div>
      </div>
    );
  }

  return <OrderDetailsClient orderDetails={orderDetails} />;
}

export async function generateMetadata({ params }: PageProps) {
  const { id } = await params;
  return {
    title: `Order #${id.slice(0, 8)} | SnapCart`,
    description: 'View order details',
  };
}