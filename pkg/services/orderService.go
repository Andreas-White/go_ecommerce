package services

import (
	"context"
	"errors"
	"fmt"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"time"

	"github.com/google/uuid"
)

type IOrderService interface {
	ProcessCheckout(ctx context.Context, userID uuid.UUID, checkoutRequest models.CheckoutRequest) (*models.OrderSummary, error)
	GetOrderSummary(ctx context.Context, orderID uuid.UUID) (*models.OrderSummary, error)
	GetOrderWithDetails(ctx context.Context, orderID uuid.UUID) (*models.OrderWithDetails, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	ConfirmOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*models.OrderWithDetails, error)
	ProcessPayment(ctx context.Context, orderID uuid.UUID) error
}

type OrderService struct {
	orderRepo   repositories.IOrderRepository
	cartRepo    repositories.ICartRepository
	productRepo repositories.IProductRepository
}

func NewOrderService(orderRepo repositories.IOrderRepository, cartRepo repositories.ICartRepository, productRepo repositories.IProductRepository) IOrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *OrderService) ProcessCheckout(ctx context.Context, userID uuid.UUID, checkoutRequest models.CheckoutRequest) (*models.OrderSummary, error) {
	// 1. Get cart items
	cartItems, err := s.cartRepo.GetAllCartItemsByCartID(ctx, checkoutRequest.CartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}

	if len(cartItems) == 0 {
		return nil, errors.New("cart is empty")
	}

	// 2. Validate stock availability
	for _, item := range cartItems {
		if item.Quantity > int(item.ProductStock) {
			return nil, fmt.Errorf("insufficient stock for product %s: requested %d, available %d", 
				item.ProductName, item.Quantity, item.ProductStock)
		}
	}

	// 3. Calculate totals
	var subtotal float64
	var orderItems []models.OrderItemSummary
	
	for _, item := range cartItems {
		itemSubtotal := float64(item.Quantity) * item.Price
		subtotal += itemSubtotal
		
		orderItems = append(orderItems, models.OrderItemSummary{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    itemSubtotal,
		})
	}

	totalAmount := subtotal + checkoutRequest.ShippingInfo.Cost

	// 4. Create the order with pending status
	orderID := uuid.New()
	order := &models.Order{
		ID:            orderID,
		UserID:        userID,
		TotalAmount:   totalAmount,
		Status:        "pending",
		PaymentStatus: "pending",
	}

	err = s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 5. Create order items
	var dbOrderItems []models.OrderItem
	for _, item := range orderItems {
		orderItem := models.OrderItem{
			ID:        uuid.New(),
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		dbOrderItems = append(dbOrderItems, orderItem)
	}

	err = s.orderRepo.CreateOrderItems(ctx, dbOrderItems)
	if err != nil {
		return nil, fmt.Errorf("failed to create order items: %w", err)
	}

	// 6. Create payment record with pending status
	payment := &models.Payment{
		ID:            uuid.New(),
		OrderID:       order.ID,
		Amount:        totalAmount,
		PaymentMethod: checkoutRequest.PaymentInfo.PaymentMethod,
		Status:        "pending",
		TransactionID: nil,
	}

	err = s.orderRepo.CreatePayment(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// 7. Create shipping record
	shipping := &models.Shipping{
		ID:           uuid.New(),
		OrderID:      order.ID,
		Method:       &checkoutRequest.ShippingInfo.Method,
		TrackingCode: nil,
		Cost:         &checkoutRequest.ShippingInfo.Cost,
		Address:      checkoutRequest.ShippingInfo.Address,
		City:         checkoutRequest.ShippingInfo.City,
		Country:      checkoutRequest.ShippingInfo.Country,
		ZipCode:      checkoutRequest.ShippingInfo.ZipCode,
	}

	err = s.orderRepo.CreateShipping(ctx, shipping)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipping: %w", err)
	}

	// 8. Create order summary for review
	orderSummary := &models.OrderSummary{
		OrderID:      orderID,
		TotalAmount:  totalAmount,
		ShippingCost: checkoutRequest.ShippingInfo.Cost,
		Items:        orderItems,
		ShippingInfo: checkoutRequest.ShippingInfo,
		PaymentInfo:  checkoutRequest.PaymentInfo,
	}

	return orderSummary, nil
}

func (s *OrderService) GetOrderSummary(ctx context.Context, orderID uuid.UUID) (*models.OrderSummary, error) {
	orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order details: %w", err)
	}

	// Convert to order summary
	var orderItems []models.OrderItemSummary
	for _, item := range orderWithDetails.Items {
		orderItems = append(orderItems, models.OrderItemSummary{
			ProductID:   item.ProductID,
			ProductName: "", // Would need to fetch from product repo
			Quantity:    item.Quantity,
			Price:       item.Price,
			Subtotal:    float64(item.Quantity) * item.Price,
		})
	}

	shippingInfo := models.ShippingInfo{
		Address: orderWithDetails.Shipping.Address,
		City:    orderWithDetails.Shipping.City,
		Country: orderWithDetails.Shipping.Country,
		ZipCode: orderWithDetails.Shipping.ZipCode,
		Method:  *orderWithDetails.Shipping.Method,
		Cost:    *orderWithDetails.Shipping.Cost,
	}

	paymentInfo := models.PaymentInfo{
		PaymentMethod: orderWithDetails.Payment.PaymentMethod,
	}

	return &models.OrderSummary{
		OrderID:      orderWithDetails.Order.ID,
		TotalAmount:  orderWithDetails.Order.TotalAmount,
		ShippingCost: *orderWithDetails.Shipping.Cost,
		Items:        orderItems,
		ShippingInfo: shippingInfo,
		PaymentInfo:  paymentInfo,
	}, nil
}

func (s *OrderService) GetOrderWithDetails(ctx context.Context, orderID uuid.UUID) (*models.OrderWithDetails, error) {
	return s.orderRepo.GetOrderWithDetails(ctx, orderID)
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return s.orderRepo.GetOrdersByUserID(ctx, userID)
}

// ConfirmOrder updates the order status and processes payment
// This is called after the user confirms the order summary
func (s *OrderService) ConfirmOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*models.OrderWithDetails, error) {
	// 1. Get the existing order with details
	orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order details: %w", err)
	}

	// 2. Verify the order belongs to the user
	if orderWithDetails.Order.UserID != userID {
		return nil, errors.New("order does not belong to user")
	}

	// 3. Verify the order is still pending
	if orderWithDetails.Order.Status != "pending" {
		return nil, fmt.Errorf("order is not in pending status: %s", orderWithDetails.Order.Status)
	}

	// 4. Update product stock
	for _, item := range orderWithDetails.Items {
		err = s.orderRepo.UpdateProductStock(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	// 5. Process payment
	err = s.ProcessPayment(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}

	// 6. Clear the cart after successful order confirmation
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err == nil {
		s.cartRepo.ClearCart(ctx, &cart.ID)
	}

	// 7. Get the updated order details
	updatedOrderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated order details: %w", err)
	}

	return updatedOrderWithDetails, nil
}

// ProcessPayment simulates payment processing
func (s *OrderService) ProcessPayment(ctx context.Context, orderID uuid.UUID) error {
	// In a real implementation, this would integrate with a payment gateway
	// For now, we'll simulate a successful payment

	// Get the payment record for this order
	orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return err
	}
	paymentID := orderWithDetails.Payment.ID

	// Generate a mock transaction ID
	transactionID := fmt.Sprintf("TXN_%s_%d", orderID.String()[:8], time.Now().Unix())

	// Update payment status using payment ID
	err = s.orderRepo.UpdatePaymentStatus(ctx, paymentID, "paid", &transactionID)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Update order status
	err = s.orderRepo.UpdateOrderStatus(ctx, orderID, "processing", "paid")
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
} 