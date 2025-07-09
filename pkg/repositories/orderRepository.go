package repositories

import (
	"context"
	"database/sql"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type IOrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	CreateOrderItems(ctx context.Context, orderItems []models.OrderItem) error
	CreatePayment(ctx context.Context, payment *models.Payment) error
	CreateShipping(ctx context.Context, shipping *models.Shipping) error
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
	GetOrderWithDetails(ctx context.Context, orderID uuid.UUID) (*models.OrderWithDetails, error)
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, paymentStatus string) error
	UpdatePaymentStatus(ctx context.Context, paymentID uuid.UUID, status string, transactionID *string) error
	UpdateProductStock(ctx context.Context, productID uuid.UUID, quantity int) error
}

type OrderRepository struct {
	DB *sql.DB
}

func NewOrderRepository(db *sql.DB) IOrderRepository {
	return &OrderRepository{DB: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO orders (id, user_id, total_amount, status, payment_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = &now

	_, err := r.DB.ExecContext(ctx, query,
		order.ID, order.UserID, order.TotalAmount, order.Status, order.PaymentStatus,
		order.CreatedAt, order.UpdatedAt)

	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreateOrder", order.ID.String())
	}

	return nil
}

func (r *OrderRepository) CreateOrderItems(ctx context.Context, orderItems []models.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4, $5)
	`

	for _, item := range orderItems {
		_, err := r.DB.ExecContext(ctx, query,
			item.ID, item.OrderID, item.ProductID, item.Quantity, item.Price)

		if err != nil {
			return utils.HandleRepositoryErrors(ctx, err, "repository/CreateOrderItems", item.ProductID.String())
		}
	}

	return nil
}

func (r *OrderRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, amount, payment_method, status, transaction_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	payment.CreatedAt = now

	_, err := r.DB.ExecContext(ctx, query,
		payment.ID, payment.OrderID, payment.Amount, payment.PaymentMethod,
		payment.Status, payment.TransactionID, payment.CreatedAt)

	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreatePayment", payment.OrderID.String())
	}

	return nil
}

func (r *OrderRepository) CreateShipping(ctx context.Context, shipping *models.Shipping) error {
	query := `
		INSERT INTO shippings (id, order_id, method, tracking_code, cost, address, city, country, zip_code, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	now := time.Now()
	shipping.CreatedAt = now
	shipping.UpdatedAt = &now

	_, err := r.DB.ExecContext(ctx, query,
		shipping.ID, shipping.OrderID, shipping.Method, shipping.TrackingCode, shipping.Cost,
		shipping.Address, shipping.City, shipping.Country, shipping.ZipCode,
		shipping.CreatedAt, shipping.UpdatedAt)

	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/CreateShipping", shipping.OrderID.String())
	}

	return nil
}

func (r *OrderRepository) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, payment_status, created_at, updated_at
		FROM orders WHERE id = $1
	`

	var order models.Order
	err := r.DB.QueryRowContext(ctx, query, orderID).Scan(
		&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.PaymentStatus,
		&order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetOrderByID", orderID.String())
	}

	return &order, nil
}

func (r *OrderRepository) GetOrderWithDetails(ctx context.Context, orderID uuid.UUID) (*models.OrderWithDetails, error) {
	// Get order
	order, err := r.GetOrderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Get order items
	itemsQuery := `
		SELECT id, order_id, product_id, quantity, price
		FROM order_items WHERE order_id = $1
	`

	rows, err := r.DB.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetOrderWithDetails", orderID.String())
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price)
		if err != nil {
			return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetOrderWithDetails", orderID.String())
		}
		items = append(items, item)
	}

	// Get payment
	paymentQuery := `
		SELECT id, order_id, amount, payment_method, status, transaction_id, created_at
		FROM payments WHERE order_id = $1
	`

	var payment models.Payment
	err = r.DB.QueryRowContext(ctx, paymentQuery, orderID).Scan(
		&payment.ID, &payment.OrderID, &payment.Amount, &payment.PaymentMethod,
		&payment.Status, &payment.TransactionID, &payment.CreatedAt)

	if err != nil && err != sql.ErrNoRows {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetOrderWithDetails", orderID.String())
	}

	// Get shipping
	shippingQuery := `
		SELECT id, order_id, method, tracking_code, cost, address, city, country, zip_code, 
		       shipped_at, delivered_at, created_at, updated_at
		FROM shippings WHERE order_id = $1
	`

	var shipping models.Shipping
	err = r.DB.QueryRowContext(ctx, shippingQuery, orderID).Scan(
		&shipping.ID, &shipping.OrderID, &shipping.Method, &shipping.TrackingCode, &shipping.Cost,
		&shipping.Address, &shipping.City, &shipping.Country, &shipping.ZipCode,
		&shipping.ShippedAt, &shipping.DeliveredAt, &shipping.CreatedAt, &shipping.UpdatedAt)

	if err != nil && err != sql.ErrNoRows {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetOrderWithDetails", orderID.String())
	}

	return &models.OrderWithDetails{
		Order:    *order,
		Items:    items,
		Payment:  payment,
		Shipping: shipping,
	}, nil
}

func (r *OrderRepository) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, payment_status, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetOrdersByUserID", userID.String())
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.TotalAmount, &order.Status, &order.PaymentStatus,
			&order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetOrdersByUserID", userID.String())
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, paymentStatus string) error {
	query := `
		UPDATE orders SET status = $2, payment_status = $3, updated_at = $4
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.DB.ExecContext(ctx, query, orderID, status, paymentStatus, now)

	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateOrderStatus", orderID.String())
	}

	return nil
}

func (r *OrderRepository) UpdatePaymentStatus(ctx context.Context, paymentID uuid.UUID, status string, transactionID *string) error {
	query := `
		UPDATE payments SET status = $2, transaction_id = $3
		WHERE id = $1
	`

	result, err := r.DB.ExecContext(ctx, query, paymentID, status, transactionID)

	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdatePaymentStatus", paymentID.String())
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdatePaymentStatus", paymentID.String())
	}
	if affectedRows == 0 {
		return utils.HandleRepositoryErrors(ctx, sql.ErrNoRows, "repository/UpdatePaymentStatus", paymentID.String())
	}

	return nil
}

func (r *OrderRepository) UpdateProductStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	query := `
		UPDATE products SET stock = stock - $2, updated_at = $3
		WHERE id = $1 AND stock >= $2
	`

	now := time.Now()
	result, err := r.DB.ExecContext(ctx, query, productID, quantity, now)

	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateProductStock", productID.String())
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateProductStock", productID.String())
	}

	if affectedRows == 0 {
		return utils.HandleRepositoryErrors(ctx, sql.ErrNoRows, "repository/UpdateProductStock", productID.String())
	}

	return nil
}
