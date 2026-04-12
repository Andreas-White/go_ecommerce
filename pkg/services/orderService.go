package services

import (
	"context"
	"errors"
	"fmt"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type IOrderService interface {
	ProcessCheckout(ctx context.Context, userID uuid.UUID, checkoutRequest models.CheckoutRequest) (*models.OrderGroupSummary, error)
	GetOrderGroupSummary(ctx context.Context, groupID uuid.UUID) (*models.OrderGroupSummary, error)
	GetOrderGroupDetails(ctx context.Context, groupID uuid.UUID) ([]models.OrderWithDetails, error)
	GetOrderWithDetails(ctx context.Context, orderID uuid.UUID) (*models.OrderWithDetails, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	GetProducerOrders(ctx context.Context, producerID uuid.UUID) ([]models.OrderWithDetails, error)
	ConfirmOrderGroup(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) ([]models.OrderWithDetails, error)
	ProcessGroupPayment(ctx context.Context, groupID uuid.UUID) error
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

func (s *OrderService) ProcessCheckout(ctx context.Context, userID uuid.UUID, checkoutRequest models.CheckoutRequest) (*models.OrderGroupSummary, error) {
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

	// 3. Group cart items by producer (ProductUserID)
	groupedItems := make(map[uuid.UUID][]models.CartItemProductDetails)
	for _, item := range cartItems {
		groupedItems[item.ProductUserID] = append(groupedItems[item.ProductUserID], item)
	}

	orderGroupID := uuid.New()
	var groupTotalAmount float64
	var orderSummaries []models.OrderSummary

	// 4. Create an order for each producer
	for producerID, items := range groupedItems {
		_ = producerID // if we need it

		var subtotal float64
		var dbOrderItems []models.OrderItem
		var summaryItems []models.OrderItemSummary

		orderID := uuid.New()

		for _, item := range items {
			itemSubtotal := float64(item.Quantity) * item.Price
			subtotal += itemSubtotal

			summaryItems = append(summaryItems, models.OrderItemSummary{
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				Price:       item.Price,
				Subtotal:    itemSubtotal,
			})

			dbOrderItems = append(dbOrderItems, models.OrderItem{
				ID:        uuid.New(),
				OrderID:   orderID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			})
		}

		// Calculate shipping cost per producer
		// For simplicity, we use a flat rate of $5.00 per producer
		producerShippingCost := 5.0
		producerTotalAmount := subtotal + producerShippingCost

		groupTotalAmount += producerTotalAmount

		// Create the order with pending status
		order := &models.Order{
			ID:            orderID,
			OrderGroupID:  &orderGroupID,
			UserID:        userID,
			TotalAmount:   producerTotalAmount,
			Status:        "pending",
			PaymentStatus: "pending",
		}

		err = s.orderRepo.CreateOrder(ctx, order)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
		}

		err = s.orderRepo.CreateOrderItems(ctx, dbOrderItems)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
		}

		// Create payment record for this specific order
		payment := &models.Payment{
			ID:            uuid.New(),
			OrderID:       orderID,
			Amount:        producerTotalAmount,
			PaymentMethod: checkoutRequest.PaymentInfo.PaymentMethod,
			Status:        "pending",
			TransactionID: nil,
		}

		err = s.orderRepo.CreatePayment(ctx, payment)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
		}

		// Create shipping record for this specific order
		orderShippingCost := producerShippingCost
		shipping := &models.Shipping{
			ID:           uuid.New(),
			OrderID:      orderID,
			Method:       &checkoutRequest.ShippingInfo.Method,
			TrackingCode: nil,
			Cost:         &orderShippingCost,
			Address:      checkoutRequest.ShippingInfo.Address,
			City:         checkoutRequest.ShippingInfo.City,
			Country:      checkoutRequest.ShippingInfo.Country,
			ZipCode:      checkoutRequest.ShippingInfo.ZipCode,
		}

		err = s.orderRepo.CreateShipping(ctx, shipping)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/ProcessCheckout")
		}

		orderSummaries = append(orderSummaries, models.OrderSummary{
			OrderID:      orderID,
			TotalAmount:  producerTotalAmount,
			ShippingCost: producerShippingCost,
			Items:        summaryItems,
			ShippingInfo: checkoutRequest.ShippingInfo,
			PaymentInfo:  checkoutRequest.PaymentInfo,
		})
	}

	// 5. Create order group summary for review
	groupSummary := &models.OrderGroupSummary{
		OrderGroupID: orderGroupID,
		TotalAmount:  groupTotalAmount,
		Orders:       orderSummaries,
	}

	return groupSummary, nil
}

func (s *OrderService) GetOrderGroupSummary(ctx context.Context, groupID uuid.UUID) (*models.OrderGroupSummary, error) {
	orders, err := s.orderRepo.GetOrdersByGroupID(ctx, groupID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetOrderGroupSummary")
	}

	var groupTotal float64
	var orderSummaries []models.OrderSummary

	for _, order := range orders {
		orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, order.ID)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/GetOrderGroupSummary")
		}

		var orderItems []models.OrderItemSummary
		for _, item := range orderWithDetails.Items {
			// Product Name is empty because GetOrderWithDetails doesn't fetch products table, just order_items
			orderItems = append(orderItems, models.OrderItemSummary{
				ProductID:   item.ProductID,
				ProductName: "", // Would need to fetch from product repo to get name
				Quantity:    item.Quantity,
				Price:       item.Price,
				Subtotal:    float64(item.Quantity) * item.Price,
			})
		}

		var shippingCost float64
		if orderWithDetails.Shipping.Cost != nil {
			shippingCost = *orderWithDetails.Shipping.Cost
		}

		shippingInfo := models.ShippingInfo{
			Address: orderWithDetails.Shipping.Address,
			City:    orderWithDetails.Shipping.City,
			Country: orderWithDetails.Shipping.Country,
			ZipCode: orderWithDetails.Shipping.ZipCode,
		}
		if orderWithDetails.Shipping.Method != nil {
			shippingInfo.Method = *orderWithDetails.Shipping.Method
		}
		shippingInfo.Cost = shippingCost

		paymentInfo := models.PaymentInfo{
			PaymentMethod: orderWithDetails.Payment.PaymentMethod,
		}

		orderSummaries = append(orderSummaries, models.OrderSummary{
			OrderID:      order.ID,
			TotalAmount:  order.TotalAmount,
			ShippingCost: shippingCost,
			Items:        orderItems,
			ShippingInfo: shippingInfo,
			PaymentInfo:  paymentInfo,
		})

		groupTotal += order.TotalAmount
	}

	return &models.OrderGroupSummary{
		OrderGroupID: groupID,
		TotalAmount:  groupTotal,
		Orders:       orderSummaries,
	}, nil
}

func (s *OrderService) GetOrderWithDetails(ctx context.Context, orderID uuid.UUID) (*models.OrderWithDetails, error) {
	return s.orderRepo.GetOrderWithDetails(ctx, orderID)
}

func (s *OrderService) GetOrderGroupDetails(ctx context.Context, groupID uuid.UUID) ([]models.OrderWithDetails, error) {
	orders, err := s.orderRepo.GetOrdersByGroupID(ctx, groupID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetOrderGroupDetails")
	}

	var results []models.OrderWithDetails
	for _, order := range orders {
		details, err := s.orderRepo.GetOrderWithDetails(ctx, order.ID)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/GetOrderGroupDetails")
		}
		results = append(results, *details)
	}
	return results, nil
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return s.orderRepo.GetOrdersByUserID(ctx, userID)
}

func (s *OrderService) GetProducerOrders(ctx context.Context, producerID uuid.UUID) ([]models.OrderWithDetails, error) {
	return s.orderRepo.GetOrdersByProducerID(ctx, producerID)
}

// ConfirmOrderGroup updates the order statuses and processes payment
// This is called after the user confirms the order summary
func (s *OrderService) ConfirmOrderGroup(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) ([]models.OrderWithDetails, error) {
	// 1. Get all orders for this group
	orders, err := s.orderRepo.GetOrdersByGroupID(ctx, groupID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrderGroup")
	}

	if len(orders) == 0 {
		return nil, utils.HandleServiceErrors(ctx, errors.New("no orders found for this group"), "service/ConfirmOrderGroup")
	}

	var confirmedOrders []models.OrderWithDetails

	// Check for idempotency: if already processing, return safely
	alreadyProcessed := true
	for _, order := range orders {
		if order.Status != "processing" && order.Status != "completed" {
			alreadyProcessed = false
			break
		}
	}

	if alreadyProcessed {
		for _, order := range orders {
			updatedOrderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, order.ID)
			if err != nil {
				return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrderGroup")
			}
			confirmedOrders = append(confirmedOrders, *updatedOrderWithDetails)
		}
		return confirmedOrders, nil
	}

	// 2. Process each order
	for _, order := range orders {
		orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, order.ID)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrderGroup")
		}

		// Verify the order belongs to the user
		if orderWithDetails.Order.UserID != userID {
			return nil, utils.HandleServiceErrors(ctx, errors.New("order does not belong to user"), "service/ConfirmOrderGroup")
		}

		// Verify the order is still pending OR processing (in case of retry a partial failure)
		if orderWithDetails.Order.Status != "pending" && orderWithDetails.Order.Status != "processing" {
			return nil, utils.HandleServiceErrors(ctx, fmt.Errorf("order is not in actionable status: %s", orderWithDetails.Order.Status), "service/ConfirmOrderGroup")
		}

		// Only deduct product stock if still pending
		if orderWithDetails.Order.Status == "pending" {
			for _, item := range orderWithDetails.Items {
				err = s.orderRepo.UpdateProductStock(ctx, item.ProductID, item.Quantity)
				if err != nil {
					return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrderGroup")
				}
			}
		}
	}


	// 3. Process grouped payment
	err = s.ProcessGroupPayment(ctx, groupID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrderGroup")
	}

	// 4. Clear the cart after successful order confirmation
	cart, err := s.cartRepo.GetCartByUserID(ctx, userID)
	if err == nil {
		s.cartRepo.ClearCart(ctx, &cart.ID)
	}

	// 5. Get the updated order details
	for _, order := range orders {
		updatedOrderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, order.ID)
		if err != nil {
			return nil, utils.HandleServiceErrors(ctx, err, "service/ConfirmOrderGroup")
		}
		confirmedOrders = append(confirmedOrders, *updatedOrderWithDetails)
	}

	return confirmedOrders, nil
}

// ProcessGroupPayment simulates payment processing for the whole group
func (s *OrderService) ProcessGroupPayment(ctx context.Context, groupID uuid.UUID) error {
	// In a real implementation, this would integrate with a payment gateway
	// For now, we'll simulate a successful payment

	orders, err := s.orderRepo.GetOrdersByGroupID(ctx, groupID)
	if err != nil {
		return utils.HandleServiceErrors(ctx, err, "service/ProcessGroupPayment")
	}

	for _, order := range orders {
		orderWithDetails, err := s.orderRepo.GetOrderWithDetails(ctx, order.ID)
		if err != nil {
			return utils.HandleServiceErrors(ctx, err, "service/ProcessGroupPayment")
		}
		
		paymentID := orderWithDetails.Payment.ID

		// Safely skip if already paid (in case this is a retry from a previous half-failure)
		if orderWithDetails.Payment.Status == "paid" {
			continue
		}

		// Generate a mocked transaction ID unique per payment avoiding UNIQ constraints
		transactionID := fmt.Sprintf("TXN_%s_%d", paymentID.String()[:8], time.Now().UnixNano())

		// Update payment status using payment ID
		err = s.orderRepo.UpdatePaymentStatus(ctx, paymentID, "paid", &transactionID)
		if err != nil {
			return utils.HandleServiceErrors(ctx, err, "service/ProcessGroupPayment")
		}

		// Update order status
		err = s.orderRepo.UpdateOrderStatus(ctx, order.ID, "processing", "paid")
		if err != nil {
			return utils.HandleServiceErrors(ctx, err, "service/ProcessGroupPayment")
		}
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
