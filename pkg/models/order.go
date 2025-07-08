package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	TotalAmount   float64    `json:"total_amount" db:"total_amount"`
	Status        string     `json:"status" db:"status"`
	PaymentStatus string     `json:"payment_status" db:"payment_status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// CheckoutRequest represents the data needed to process a checkout
type CheckoutRequest struct {
	CartID        uuid.UUID           `json:"cart_id"`
	ShippingInfo  ShippingInfo        `json:"shipping_info"`
	PaymentInfo   PaymentInfo         `json:"payment_info"`
}

// ShippingInfo contains shipping address and method
type ShippingInfo struct {
	Address string  `json:"address"`
	City    string  `json:"city"`
	Country string  `json:"country"`
	ZipCode string  `json:"zip_code"`
	Method  string  `json:"method"`
	Cost    float64 `json:"cost"`
}

// PaymentInfo contains payment method and details
type PaymentInfo struct {
	PaymentMethod string `json:"payment_method"`
	// Note: Actual payment details are handled by payment gateway
	// This is just for method selection
}

// OrderSummary represents the order details for review
type OrderSummary struct {
	OrderID       uuid.UUID                `json:"order_id"`
	TotalAmount   float64                  `json:"total_amount"`
	ShippingCost  float64                  `json:"shipping_cost"`
	Items         []OrderItemSummary       `json:"items"`
	ShippingInfo  ShippingInfo             `json:"shipping_info"`
	PaymentInfo   PaymentInfo              `json:"payment_info"`
}

// OrderItemSummary represents an item in the order summary
type OrderItemSummary struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Price       float64   `json:"price"`
	Subtotal    float64   `json:"subtotal"`
}

// OrderWithDetails represents a complete order with all related data
type OrderWithDetails struct {
	Order       Order        `json:"order"`
	Items       []OrderItem  `json:"items"`
	Payment     Payment      `json:"payment"`
	Shipping    Shipping     `json:"shipping"`
}
