package services

import (
	"context"
	"errors"
	"fmt"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/utils"
	"log"
	"time"

	"github.com/google/uuid"
)

type IOrderService interface {
	ProcessCheckout(ctx context.Context, userID uuid.UUID, checkoutRequest models.CheckoutRequest) (*models.OrderSummary, error)
	GetOrderSummary(ctx context.Context, orderID uuid.UUID) (*models.OrderSummary, error)
	GetOrderWithDetails(ctx context.Context, orderID uuid.UUID) (*models.OrderWithDetails, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	GetProducerOrders(ctx context.Context, producerID uuid.UUID) ([]models.OrderWithDetails, error)
	ConfirmOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*models.OrderWithDetails, error)
	ProcessPayment(ctx context.Context, orderID uuid.UUID) error
	FulfillOrder(ctx context.Context, producerID uuid.UUID, fulfillmentRequest models.OrderFulfillmentRequest) (*models.OrderFulfillmentResponse, error)
	GetSalesReport(ctx context.Context, producerID uuid.UUID, request models.SalesReportRequest) (*models.SalesReportResponse, error)
	CancelOrder(ctx context.Context, orderID uuid.UUID) error
	SoftDeleteOrder(ctx context.Context, orderID uuid.UUID) error
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
		return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
	}

	if len(cartItems) == 0 {
		return nil, utils.HandleServiceErrors(ctx, errors.New("cart is empty"), "service/ProcessCheckout")
	}

	// 2. Validate stock availability
	for _, item := range cartItems {
		if item.Quantity > int(item.ProductStock) {
			return nil, utils.HandleServiceErrors(ctx, fmt.Errorf("insufficient stock for product %s: requested %d, available %d",
				item.ProductName, item.Quantity, item.ProductStock), "service/ProcessCheckout")
		}
	}

	// 3. Calculate totals
	var subtotal float64
	var orderItems []models.OrderItemSummary

	log.Println("cartItems: ", cartItems)
	log.Println("checkoutRequest: ", checkoutRequest)

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

	log.Println("subtotal: ", subtotal)
	log.Println("checkoutRequest.ShippingInfo.Cost: ", checkoutRequest.ShippingInfo.Cost)

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
		return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
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
		return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
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
		return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
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
		return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
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
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetOrderSummary")
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

func (s *OrderService) GetProducerOrders(ctx context.Context, producerID uuid.UUID) ([]models.OrderWithDetails, error) {
	return s.orderRepo.GetOrdersByProducerID(ctx, producerID)
}

// ConfirmOrder updates the order status and processes payment
// This is called after the user confirms the order summary
func (s *OrderService) ConfirmOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*models.OrderWithDetails, error) {
	// 1. Get the existing order with details
	orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrder")
	}

	// 2. Verify the order belongs to the user
	if orderWithDetails.Order.UserID != userID {
		return nil, utils.HandleServiceErrors(ctx, errors.New("order does not belong to user"), "service/ConfirmOrder")
	}

	// 3. Verify the order is still pending
	if orderWithDetails.Order.Status != "pending" {
		return nil, utils.HandleServiceErrors(ctx, fmt.Errorf("order is not in pending status: %s", orderWithDetails.Order.Status), "service/ConfirmOrder")
	}

	// 4. Update product stock
	for _, item := range orderWithDetails.Items {
		err = s.orderRepo.UpdateProductStock(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrder")
		}
	}

	// 5. Process payment
	err = s.ProcessPayment(ctx, orderID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrder")
	}

	// 6. Clear the cart after successful order confirmation
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err == nil {
		s.cartRepo.ClearCart(ctx, &cart.ID)
	}

	// 7. Get the updated order details
	updatedOrderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrder")
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
		return utils.HandleServiceErrors(ctx, err, "service/ProcessPayment")
	}
	paymentID := orderWithDetails.Payment.ID

	// Generate a mock transaction ID
	transactionID := fmt.Sprintf("TXN_%s_%d", orderID.String()[:8], time.Now().Unix())

	// Update payment status using payment ID
	err = s.orderRepo.UpdatePaymentStatus(ctx, paymentID, "paid", &transactionID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/ProcessPayment")
	}

	// Update order status
	err = s.orderRepo.UpdateOrderStatus(ctx, orderID, "processing", "paid")
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/ProcessPayment")
	}

	return nil
}

// FulfillOrder handles order fulfillment by producers
func (s *OrderService) FulfillOrder(ctx context.Context, producerID uuid.UUID, fulfillmentRequest models.OrderFulfillmentRequest) (*models.OrderFulfillmentResponse, error) {
	// 1. Get the order with details
	orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, fulfillmentRequest.OrderID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/FulfillOrder")
	}

	// 2. Verify the order contains products from this producer
	producerOwnsOrder := false
	for _, item := range orderWithDetails.Items {
		product, err := s.productRepo.GetProductByID(ctx, item.ProductID.String())
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/FulfillOrder")
		}
		if product.UserID == producerID {
			producerOwnsOrder = true
			break
		}
	}

	if !producerOwnsOrder {
		return nil, utils.HandleServiceErrors(ctx, errors.New("order does not contain products from this producer"), "service/FulfillOrder")
	}

	// 3. Verify the order is in a valid state for fulfillment
	if orderWithDetails.Order.PaymentStatus != "paid" {
		return nil, utils.HandleServiceErrors(ctx, errors.New("order payment is not completed"), "service/FulfillOrder")
	}

	// 4. Validate the new status
	validStatuses := map[string]bool{
		"accepted": true,
		"shipped":  true,
		"canceled": true,
	}
	if !validStatuses[fulfillmentRequest.NewStatus] {
		return nil, utils.HandleServiceErrors(ctx, fmt.Errorf("invalid status: %s", fulfillmentRequest.NewStatus), "service/FulfillOrder")
	}

	// 5. Update order status
	err = s.orderRepo.UpdateOrderStatus(ctx, fulfillmentRequest.OrderID, fulfillmentRequest.NewStatus, orderWithDetails.Order.PaymentStatus)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/FulfillOrder")
	}

	// 6. If status is "shipped", update shipping tracking
	var shippedAt *time.Time
	if fulfillmentRequest.NewStatus == "shipped" {
		if fulfillmentRequest.TrackingCode == nil || *fulfillmentRequest.TrackingCode == "" {
			return nil, utils.HandleServiceErrors(ctx, errors.New("tracking code is required when marking order as shipped"), "service/FulfillOrder")
		}

		now := time.Now()
		shippedAt = &now

		err = s.orderRepo.UpdateShippingTracking(ctx, fulfillmentRequest.OrderID, *fulfillmentRequest.TrackingCode, shippedAt)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/FulfillOrder")
		}
	}

	// 7. Create response
	response := &models.OrderFulfillmentResponse{
		OrderID:      fulfillmentRequest.OrderID,
		Status:       fulfillmentRequest.NewStatus,
		TrackingCode: fulfillmentRequest.TrackingCode,
		ShippedAt:    shippedAt,
		Message:      fmt.Sprintf("Order %s successfully updated to %s", fulfillmentRequest.OrderID.String()[:8], fulfillmentRequest.NewStatus),
	}

	return response, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	// 1. Get the order with details
	orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/CancelOrder")
	}

	// 2. Verify the order is in a valid state for cancellation
	if orderWithDetails.Order.Status != "processing" && orderWithDetails.Order.Status != "accepted" {
		return utils.HandleServiceErrors(ctx, errors.New("order is not in processing or accepted status"), "service/CancelOrder")
	}

	// 3. Update product stock
	for _, item := range orderWithDetails.Items {
		err = s.orderRepo.UpdateProductStock(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return utils.HandleServiceErrors(ctx, err, "service/CancelOrder")
		}
	}

	// 4. Refund the payment
	err = s.orderRepo.UpdatePaymentStatus(ctx, orderWithDetails.Payment.ID, "refunded", nil)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/CancelOrder")
	}

	// 5. Update order status
	err = s.orderRepo.UpdateOrderStatus(ctx, orderID, "canceled", "refunded")
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/CancelOrder")
	}

	return nil
}

func (s *OrderService) SoftDeleteOrder(ctx context.Context, orderID uuid.UUID) error {
	// 1. Get the order with details
	orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/SoftDeleteOrder")
	}

	// 2. Verify the order is in a valid state for soft deletion
	if orderWithDetails.Order.Status != "pending" && orderWithDetails.Order.Status != "processing" {
		return utils.HandleServiceErrors(ctx, errors.New("order is not in pending or processing status"), "service/SoftDeleteOrder")
	}

	// 3. Update product stock
	for _, item := range orderWithDetails.Items {
		err = s.orderRepo.UpdateProductStock(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return utils.HandleServiceErrors(ctx, err, "service/SoftDeleteOrder")
		}
	}

	// 4. Refund the payment
	if orderWithDetails.Payment.Status == "paid" {
		err = s.orderRepo.UpdatePaymentStatus(ctx, orderWithDetails.Payment.ID, "refunded", nil)
		if err != nil {
			return utils.HandleServiceErrors(ctx, err, "service/SoftDeleteOrder")
		}
	}

	// 5. Soft delete the order
	return s.orderRepo.SoftDeleteOrder(ctx, orderID)
}

// GetSalesReport retrieves sales analytics for a producer
func (s *OrderService) GetSalesReport(ctx context.Context, producerID uuid.UUID, request models.SalesReportRequest) (*models.SalesReportResponse, error) {
	// Validate date range if both dates are provided
	if request.StartDate != nil && request.EndDate != nil {
		if request.StartDate.After(*request.EndDate) {
			return nil, utils.HandleServiceErrors(ctx, errors.New("start date cannot be after end date"), "service/GetSalesReport")
		}
	}

	// Get sales report from repository
	report, err := s.orderRepo.GetSalesReport(ctx, producerID, request.StartDate, request.EndDate, request.Category)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetSalesReport")
	}

	return report, nil
}
