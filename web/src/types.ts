export interface Product {
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

export interface CartItem {
  id?: string;
  product_id: string;
  quantity: number;
  price?: number;
  product_name?: string;
  image_url?: string;
}

export interface Order {
  id: string;
  total_amount: number;
  status: string;
  payment_status: string;
  created_at: string;
}

export interface ShippingInfo {
  address: string;
  city: string;
  country: string;
  zip_code: string;
  method: string;
  cost: number;
}

export interface PaymentInfo {
  payment_method: string;
}

export interface OrderItem {
  product_id: string;
  product_name: string;
  quantity: number;
  price: number;
  subtotal: number;
}

export interface OrderSummary {
  order_id: string;
  total_amount: number;
  shipping_cost: number;
  items: OrderItem[];
  shipping_info: ShippingInfo;
  payment_info: PaymentInfo;
}

export interface CheckoutShippingInfo {
  address: string;
  city: string;
  country: string;
  zip_code: string;
  method: string;
  cost: number;
}

export interface CheckoutPaymentInfo {
  payment_method: string;
}

export interface CheckoutOrderSummary {
  order_id: string;
  total_amount: number;
  shipping_cost: number;
  items: OrderItem[];
  shipping_info: CheckoutShippingInfo;
  payment_info: CheckoutPaymentInfo;
}

export interface CheckoutOrderGroupSummary {
  order_group_id: string;
  total_amount: number;
  orders: CheckoutOrderSummary[];
}

export interface OrderGroupSummary {
  order_group_id: string;
  total_amount: number;
  orders: OrderSummary[];
}

export interface OrderWithDetails {
  order: {
    id: string;
    total_amount: number;
    status: string;
    payment_status: string;
    created_at: string;
  };
  items: OrderItem[];
  payment: {
    payment_method: string;
    amount: number;
    status: string;
    transaction_id?: string;
  };
  shipping: ShippingInfo & {
    status: string;
    tracking_code?: string;
    shipped_at?: string;
    delivered_at?: string;
  };
}

export interface ProductMap {
  [productId: string]: {
    price?: number;
    name?: string;
    image_url?: string;
  };
}

export interface User {
  first_name: string;
  last_name: string;
  email: string;
  is_producer: boolean;
}

export interface Company {
  id: string;
  name: string;
  description: string;
  logo_url?: string;
}

export interface RegisterData {
  first_name: string;
  last_name: string;
  email: string;
  password: string;
  is_producer?: boolean;
  company_id?: string;
}