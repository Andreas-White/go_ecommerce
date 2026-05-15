import React from 'react';
import Image from 'next/image';
import './CartItem.css';
import { Button } from '@/components/ui';

interface CartItemType {
  id?: string;
  product_id: string;
  quantity: number;
  price?: number;
  product_name?: string;
  image_url?: string;
}

interface ProductType {
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

interface CartItemProps {
  item: CartItemType;
  product: ProductType;
  updating: boolean;
  onQuantityChange: (productId: string, newQuantity: number) => void;
  onRemove: (productId: string) => void;
}

export default function CartItem({
  item,
  product,
  updating,
  onQuantityChange,
  onRemove,
}: CartItemProps) {
  const price = product?.price || 0;
  return (
    <div className={`cart-item`}>
      <div className="cart-item-info">
        {product?.image_url && (
          <div className="cart-item-image-wrapper">
            <Image
              src={product.image_url}
              alt={product?.name || `Product ${item.product_id}`}
              fill
              sizes="120px"
              className="cart-item-image"
            />
          </div>
        )}
        <div className="cart-item-details">
          <h3 className="cart-item-name">
            {product?.name || `Product ${item.product_id}`}
          </h3>
          <p className="cart-item-price">${price.toFixed(2)} each</p>
        </div>
      </div>
      <div className="cart-item-quantity">
        <Button
          variant="secondary"
          onClick={() => onQuantityChange(item.product_id, item.quantity - 1)}
          disabled={updating}
          className="cart-quantity-btn"
        >
          -
        </Button>
        <span className="cart-quantity-display">{item.quantity}</span>
        <Button
          variant="secondary"
          onClick={() => onQuantityChange(item.product_id, item.quantity + 1)}
          disabled={updating}
          className="cart-quantity-btn"
        >
          +
        </Button>
      </div>
      <div className="cart-item-subtotal">
        <p className="cart-item-total">${(price * item.quantity).toFixed(2)}</p>
      </div>
      <Button
        variant="destructive"
        onClick={() => onRemove(item.product_id)}
        disabled={updating}
        className="cart-remove-btn"
      >
        Remove
      </Button>
    </div>
  );
}
