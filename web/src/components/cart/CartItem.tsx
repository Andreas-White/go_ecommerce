import React from 'react';
import './CartItem.css';
import { Button } from '@/components/ui';

export default function CartItem({
  item,
  product,
  updating,
  onQuantityChange,
  onRemove,
}: {
  item: any;
  product: any;
  updating: boolean;
  onQuantityChange: (productId: string, newQuantity: number) => void;
  onRemove: (productId: string) => void;
}) {
  const price = item?.price || 0;
  return (
    <div className={`cart-item`}>
      <div className="cart-item-info">
        {product?.image_url && (
          <img
            src={product.image_url}
            alt={item.product_name || `Product ${item.product_id}`}
            className="cart-item-image"
          />
        )}
        <h3 className="cart-item-name">
          {item.product_name || `Product ${item.product_id}`}
        </h3>
        <p className="cart-item-price">${price.toFixed(2)} each</p>
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
