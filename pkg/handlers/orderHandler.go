package handlers

import (
	"encoding/json"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/services"
	"go_ecommerce/pkg/utils"
	"net/http"

	"github.com/google/uuid"
)

type OrderHandler struct {
	orderService services.IOrderService
}

func NewOrderHandler(orderService services.IOrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// ProcessCheckout handles the initial checkout request
// This creates an order summary for review
func (h *OrderHandler) ProcessCheckout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (must be authenticated)
	user := middleware.GetUserFromContext(r, w)
	if user == nil {
		utils.HandleAPIErrors(nil, w, "handler/ProcessCheckout", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	var checkoutRequest models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&checkoutRequest); err != nil {
		utils.HandleAPIErrors(err, w, "handler/ProcessCheckout", http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate required fields
	if checkoutRequest.CartID == uuid.Nil {
		utils.HandleAPIErrors(nil, w, "handler/ProcessCheckout", http.StatusBadRequest, "Cart ID is required")
		return
	}

	if checkoutRequest.ShippingInfo.Address == "" || checkoutRequest.ShippingInfo.City == "" || 
	   checkoutRequest.ShippingInfo.Country == "" || checkoutRequest.ShippingInfo.ZipCode == "" {
		utils.HandleAPIErrors(nil, w, "handler/ProcessCheckout", http.StatusBadRequest, "Complete shipping information is required")
		return
	}

	if checkoutRequest.PaymentInfo.PaymentMethod == "" {
		utils.HandleAPIErrors(nil, w, "handler/ProcessCheckout", http.StatusBadRequest, "Payment method is required")
		return
	}

	// Process checkout and get order summary
	orderSummary, err := h.orderService.ProcessCheckout(ctx, user.ID, checkoutRequest)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/ProcessCheckout", http.StatusInternalServerError, "Failed to process checkout")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, orderSummary)
}

// ConfirmOrder handles the order confirmation after review
// This creates the actual order, payment, and shipping records
func (h *OrderHandler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (must be authenticated)
	user := middleware.GetUserFromContext(r, w)
	if user == nil {
		utils.HandleAPIErrors(nil, w, "handler/ConfirmOrder", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	var request struct {
		OrderID uuid.UUID `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.HandleAPIErrors(err, w, "handler/ConfirmOrder", http.StatusBadRequest, "Invalid request payload")
		return
	}

	if request.OrderID == uuid.Nil {
		utils.HandleAPIErrors(nil, w, "handler/ConfirmOrder", http.StatusBadRequest, "Order ID is required")
		return
	}

	// Confirm the order and process payment
	orderWithDetails, err := h.orderService.ConfirmOrder(ctx, request.OrderID, user.ID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/ConfirmOrder", http.StatusInternalServerError, "Failed to confirm order")
		return
	}

	// Return the complete order details
	utils.RespondWithJSON(w, http.StatusCreated, orderWithDetails)
}

// GetOrderSummary retrieves the order summary for review
func (h *OrderHandler) GetOrderSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (must be authenticated)
	user := middleware.GetUserFromContext(r, w)
	if user == nil {
		utils.HandleAPIErrors(nil, w, "handler/GetOrderSummary", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	var request struct {
		OrderID uuid.UUID `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetOrderSummary", http.StatusBadRequest, "Invalid request payload")
		return
	}

	if request.OrderID == uuid.Nil {
		utils.HandleAPIErrors(nil, w, "handler/GetOrderSummary", http.StatusBadRequest, "Order ID is required")
		return
	}

	orderSummary, err := h.orderService.GetOrderSummary(ctx, request.OrderID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetOrderSummary", http.StatusInternalServerError, "Failed to get order summary")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, orderSummary)
}

// GetOrderDetails retrieves complete order details including items, payment, and shipping
func (h *OrderHandler) GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (must be authenticated)
	user := middleware.GetUserFromContext(r, w)
	if user == nil {
		utils.HandleAPIErrors(nil, w, "handler/GetOrderDetails", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	var request struct {
		OrderID uuid.UUID `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetOrderDetails", http.StatusBadRequest, "Invalid request payload")
		return
	}

	if request.OrderID == uuid.Nil {
		utils.HandleAPIErrors(nil, w, "handler/GetOrderDetails", http.StatusBadRequest, "Order ID is required")
		return
	}

	orderWithDetails, err := h.orderService.GetOrderWithDetails(ctx, request.OrderID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetOrderDetails", http.StatusInternalServerError, "Failed to get order details")
		return
	}

	// Verify the order belongs to the authenticated user
	if orderWithDetails.Order.UserID != user.ID {
		utils.HandleAPIErrors(nil, w, "handler/GetOrderDetails", http.StatusForbidden, "Access denied")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, orderWithDetails)
}

// GetUserOrders retrieves all orders for the authenticated user
func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (must be authenticated)
	user := middleware.GetUserFromContext(r, w)
	if user == nil {
		utils.HandleAPIErrors(nil, w, "handler/GetUserOrders", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	orders, err := h.orderService.GetUserOrders(ctx, user.ID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetUserOrders", http.StatusInternalServerError, "Failed to get user orders")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, orders)
} 