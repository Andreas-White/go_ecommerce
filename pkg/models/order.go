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
	CartID       uuid.UUID    `json:"cart_id"`
	ShippingInfo ShippingInfo `json:"shipping_info"`
	PaymentInfo  PaymentInfo  `json:"payment_info"`
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
	OrderID      uuid.UUID          `json:"order_id"`
	TotalAmount  float64            `json:"total_amount"`
	ShippingCost float64            `json:"shipping_cost"`
	Items        []OrderItemSummary `json:"items"`
	ShippingInfo ShippingInfo       `json:"shipping_info"`
	PaymentInfo  PaymentInfo        `json:"payment_info"`
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
	Order    Order       `json:"order"`
	Items    []OrderItem `json:"items"`
	Payment  Payment     `json:"payment"`
	Shipping Shipping    `json:"shipping"`
}

// OrderFulfillmentRequest represents the request to accept and fulfill an order
type OrderFulfillmentRequest struct {
	OrderID      uuid.UUID `json:"order_id"`
	NewStatus    string    `json:"new_status"` // "accepted", "preparing", "shipped"
	TrackingCode *string   `json:"tracking_code,omitempty"`
}

// OrderFulfillmentResponse represents the response after order fulfillment
type OrderFulfillmentResponse struct {
	OrderID      uuid.UUID  `json:"order_id"`
	Status       string     `json:"status"`
	TrackingCode *string    `json:"tracking_code,omitempty"`
	ShippedAt    *time.Time `json:"shipped_at,omitempty"`
	Message      string     `json:"message"`
}

// SalesReportRequest represents the request for sales reports with optional filters
type SalesReportRequest struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Category  *string    `json:"category,omitempty"`
}

// SalesReportResponse represents the sales report data for producers
type SalesReportResponse struct {
	ProducerID         uuid.UUID           `json:"producer_id"`
	TotalRevenue       float64             `json:"total_revenue"`
	TotalOrders        int                 `json:"total_orders"`
	TotalItemsSold     int                 `json:"total_items_sold"`
	AverageOrderValue  float64             `json:"average_order_value"`
	TopSellingProducts []TopSellingProduct `json:"top_selling_products"`
	SalesByCategory    []CategorySales     `json:"sales_by_category"`
	Period             SalesPeriod         `json:"period"`
}

// TopSellingProduct represents a product with its sales metrics
type TopSellingProduct struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Category    string    `json:"category"`
	UnitsSold   int       `json:"units_sold"`
	Revenue     float64   `json:"revenue"`
	Price       float64   `json:"price"`
}

// CategorySales represents sales metrics for a product category
type CategorySales struct {
	Category   string  `json:"category"`
	UnitsSold  int     `json:"units_sold"`
	Revenue    float64 `json:"revenue"`
	OrderCount int     `json:"order_count"`
}

// SalesPeriod represents the time period for the sales report
type SalesPeriod struct {
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}
