'use client';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useCart } from '../../context/CartContext';
import { useAuth } from '../../context/AuthContext';
import { api } from '../../lib/api';
import CheckoutStepper from '../../components/checkout/CheckoutStepper';
import ShippingForm from '../../components/checkout/ShippingForm';
import PaymentForm from '../../components/checkout/PaymentForm';
import OrderSummaryDisplay from '../../components/checkout/OrderSummaryDisplay';
import './page.css';
import Alert from '@/components/ui/Alert';
import Spinner from '@/components/ui/Spinner';
import { useTopProgress } from '@/context/TopProgressContext';
import { ShippingInfo, PaymentInfo, OrderGroupSummary } from '@/types';

export default function CheckoutPage() {
  const router = useRouter();
  const { cartItems, loading: cartLoading, clearCart } = useCart();
  const { user, loading: authLoading } = useAuth();
  const { start: startProgress, complete: completeProgress } = useTopProgress();
  const [currentStep, setCurrentStep] = useState(1);
  const [loading, setLoading] = useState(false);
  const [shippingInfo, setShippingInfo] = useState<ShippingInfo>({
    address: '',
    city: '',
    country: '',
    zip_code: '',
    method: 'standard',
    cost: 5.99,
  });
  const [paymentInfo, setPaymentInfo] = useState<PaymentInfo>({
    payment_method: 'credit_card',
  });
  const [orderSummary, setOrderSummary] = useState<OrderGroupSummary | null>(null);
  const [checkoutAlert, setCheckoutAlert] = useState<{
    type: 'success' | 'error' | 'info';
    message: string;
  } | null>(null);

  // Redirect if not authenticated or cart is empty
  useEffect(() => {
    if (!authLoading && !user) {
      router.push('/login');
      return;
    }

    if (!cartLoading && cartItems.length === 0) {
      router.push('/cart');
      return;
    }
  }, [user, authLoading, cartItems, cartLoading, router]);

  const handleShippingSubmit = (shipping: ShippingInfo) => {
    setShippingInfo(shipping);
    setCurrentStep(2);
  };

  const handlePaymentSubmit = async (payment: PaymentInfo) => {
    setPaymentInfo(payment);
    setCurrentStep(3);

    setLoading(true);
    startProgress();
    try {
      const cartItemsWithDetails = await api.post<Array<{ cart_id: string }>>(
        '/cart/get'
      );
      if (!cartItemsWithDetails || cartItemsWithDetails.length === 0) {
        throw new Error('Cart is empty');
      }

      const cartId = cartItemsWithDetails[0].cart_id;

      const summary = await api.post<OrderGroupSummary>(
        '/orders/checkout',
        {
          cart_id: cartId,
          shipping_info: shippingInfo,
          payment_info: payment,
        },
        undefined,
        true
      );
      setOrderSummary(summary);
      setCheckoutAlert({
        type: 'success',
        message: 'Order processed successfully.',
      });
    } catch (error) {
      setCheckoutAlert({
        type: 'error',
        message: 'Failed to process order. Please try again.',
      });
      setCurrentStep(2);
    } finally {
      setLoading(false);
      completeProgress();
    }
  };

  const handleOrderConfirm = async () => {
    if (!orderSummary) return;

    setLoading(true);
    startProgress();
    try {
      await api.post(
        '/orders/confirm',
        {
          order_group_id: orderSummary.order_group_id,
        },
        undefined,
        true
      );

      await clearCart();

      router.push(`/order-confirmation/${orderSummary.order_group_id}`);
    } catch (error) {
      setCheckoutAlert({
        type: 'error',
        message: 'Failed to confirm order. Please try again.',
      });
    } finally {
      setLoading(false);
      completeProgress();
    }
  };

  const handleBackToStep = (step: number) => {
    setCurrentStep(step);
  };

  if (authLoading || cartLoading) {
    return (
      <div className="checkout-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="checkout-container">
      <div className="checkout-content">
        <h1 className="checkout-title">Checkout</h1>

        <CheckoutStepper currentStep={currentStep} />
        {checkoutAlert && (
          <Alert
            type={checkoutAlert.type}
            onClose={() => setCheckoutAlert(null)}
          >
            {checkoutAlert.message}
          </Alert>
        )}

        <div className="checkout-steps">
          {currentStep === 1 && (
            <ShippingForm
              shippingInfo={shippingInfo}
              onSubmit={handleShippingSubmit}
            />
          )}

          {currentStep === 2 && (
            <PaymentForm
              paymentInfo={paymentInfo}
              onSubmit={handlePaymentSubmit}
              onBack={() => handleBackToStep(1)}
            />
          )}

          {currentStep === 3 && (
            <OrderSummaryDisplay
              orderSummary={orderSummary}
              loading={loading}
              onConfirm={handleOrderConfirm}
              onBack={() => handleBackToStep(2)}
            />
          )}
        </div>
      </div>
    </div>
  );
}
