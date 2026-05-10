import { serverFetch } from '@/lib/api.server';
import OrdersList, { Order } from './_components/OrdersList';
import ListItemSkeleton from '@/components/ui/ListItemSkeleton';

async function getOrders(): Promise<Order[]> {
  try {
    const orders = await serverFetch<Order[]>('/orders/user');
    return orders || [];
  } catch {
    return [];
  }
}

export default async function OrdersPage() {
  const orders = await getOrders();

  if (orders.length === 0) {
    return (
      <div className="orders-container">
        <div className="orders-loading">
          <ListItemSkeleton />
        </div>
      </div>
    );
  }

  return <OrdersList initialOrders={orders} />;
}

export const metadata = {
  title: 'Orders | SnapCart',
  description: 'View your order history',
};