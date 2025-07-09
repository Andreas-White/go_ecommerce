package tests

import (
	"bytes"
	"encoding/json"
	"go_ecommerce/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckoutFlow_Complete(t *testing.T) {
	clearTables(t)

	// 1. Register a producer and a customer
	producerEmail := "producer-checkout@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerToken, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	customerEmail := "customer-checkout@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerToken, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
	require.NoError(t, err)

	// 2. Producer creates a product
	product := createTestProduct(t, producerToken, models.Product{
		Name:  "Test Product for Checkout",
		Price: 29.99,
		Stock: 10,
	})

	// 3. Customer adds product to cart
	cartItemsToAdd := []models.CartItemDTO{
		{ProductID: product.ID, Quantity: 2, Price: product.Price},
	}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.Header.Set("Authorization", "Bearer "+customerToken)

	addRR := httptest.NewRecorder()
	testRouter.ServeHTTP(addRR, addReq)
	require.Equal(t, http.StatusOK, addRR.Code, "Failed to add items to cart")

	var createdCart models.Cart
	err = json.Unmarshal(addRR.Body.Bytes(), &createdCart)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdCart.ID)

	// 4. Customer initiates checkout
	checkoutRequest := models.CheckoutRequest{
		CartID: createdCart.ID,
		ShippingInfo: models.ShippingInfo{
			Address: "123 Main St",
			City:    "New York",
			Country: "USA",
			ZipCode: "10001",
			Method:  "standard",
			Cost:    5.99,
		},
		PaymentInfo: models.PaymentInfo{
			PaymentMethod: "credit_card",
		},
	}

	checkoutBody, _ := json.Marshal(checkoutRequest)
	checkoutReq, _ := http.NewRequest("POST", "/orders/checkout", bytes.NewBuffer(checkoutBody))
	checkoutReq.Header.Set("Content-Type", "application/json")
	checkoutReq.Header.Set("Authorization", "Bearer "+customerToken)

	checkoutRR := httptest.NewRecorder()
	testRouter.ServeHTTP(checkoutRR, checkoutReq)
	require.Equal(t, http.StatusOK, checkoutRR.Code, "Failed to process checkout")

	var orderSummary models.OrderSummary
	err = json.Unmarshal(checkoutRR.Body.Bytes(), &orderSummary)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, orderSummary.OrderID)
	assert.Equal(t, 65.97, orderSummary.TotalAmount) // (29.99 * 2) + 5.99 shipping
	assert.Equal(t, 5.99, orderSummary.ShippingCost)
	assert.Len(t, orderSummary.Items, 1)
	assert.Equal(t, product.ID, orderSummary.Items[0].ProductID)
	assert.Equal(t, 2, orderSummary.Items[0].Quantity)

	// 5. Customer confirms the order
	confirmRequest := struct {
		OrderID uuid.UUID `json:"order_id"`
	}{
		OrderID: orderSummary.OrderID,
	}

	confirmBody, _ := json.Marshal(confirmRequest)
	confirmReq, _ := http.NewRequest("POST", "/orders/confirm", bytes.NewBuffer(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	confirmReq.Header.Set("Authorization", "Bearer "+customerToken)

	confirmRR := httptest.NewRecorder()
	testRouter.ServeHTTP(confirmRR, confirmReq)
	require.Equal(t, http.StatusCreated, confirmRR.Code, "Failed to confirm order")

	var orderWithDetails models.OrderWithDetails
	err = json.Unmarshal(confirmRR.Body.Bytes(), &orderWithDetails)
	require.NoError(t, err)
	assert.Equal(t, orderSummary.OrderID, orderWithDetails.Order.ID)
	assert.Equal(t, "processing", orderWithDetails.Order.Status)
	assert.Equal(t, "paid", orderWithDetails.Order.PaymentStatus)
	assert.Equal(t, 65.97, orderWithDetails.Order.TotalAmount)
	assert.Len(t, orderWithDetails.Items, 1)
	assert.Equal(t, "paid", orderWithDetails.Payment.Status)
	assert.NotNil(t, orderWithDetails.Payment.TransactionID)

	// 6. Verify the order details can be retrieved
	detailsRequest := struct {
		OrderID uuid.UUID `json:"order_id"`
	}{
		OrderID: orderWithDetails.Order.ID,
	}

	detailsBody, _ := json.Marshal(detailsRequest)
	detailsReq, _ := http.NewRequest("POST", "/orders/details", bytes.NewBuffer(detailsBody))
	detailsReq.Header.Set("Content-Type", "application/json")
	detailsReq.Header.Set("Authorization", "Bearer "+customerToken)

	detailsRR := httptest.NewRecorder()
	testRouter.ServeHTTP(detailsRR, detailsReq)
	require.Equal(t, http.StatusOK, detailsRR.Code, "Failed to get order details")

	// 7. Verify user orders can be retrieved
	userOrdersReq, _ := http.NewRequest("GET", "/orders/user", nil)
	userOrdersReq.Header.Set("Content-Type", "application/json")
	userOrdersReq.Header.Set("Authorization", "Bearer "+customerToken)

	userOrdersRR := httptest.NewRecorder()
	testRouter.ServeHTTP(userOrdersRR, userOrdersReq)
	require.Equal(t, http.StatusOK, userOrdersRR.Code, "Failed to get user orders")

	var userOrders []models.Order
	err = json.Unmarshal(userOrdersRR.Body.Bytes(), &userOrders)
	require.NoError(t, err)
	assert.Len(t, userOrders, 1)
	assert.Equal(t, orderWithDetails.Order.ID, userOrders[0].ID)
}

func TestCheckoutFlow_EmptyCart(t *testing.T) {
	clearTables(t)

	// Register a customer
	customerEmail := "customer-empty-cart@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerToken, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
	require.NoError(t, err)

	// Try to checkout with empty cart
	checkoutRequest := models.CheckoutRequest{
		CartID: uuid.New(), // Random cart ID
		ShippingInfo: models.ShippingInfo{
			Address: "123 Main St",
			City:    "New York",
			Country: "USA",
			ZipCode: "10001",
			Method:  "standard",
			Cost:    5.99,
		},
		PaymentInfo: models.PaymentInfo{
			PaymentMethod: "credit_card",
		},
	}

	checkoutBody, _ := json.Marshal(checkoutRequest)
	checkoutReq, _ := http.NewRequest("POST", "/orders/checkout", bytes.NewBuffer(checkoutBody))
	checkoutReq.Header.Set("Content-Type", "application/json")
	checkoutReq.Header.Set("Authorization", "Bearer "+customerToken)

	checkoutRR := httptest.NewRecorder()
	testRouter.ServeHTTP(checkoutRR, checkoutReq)
	assert.Equal(t, http.StatusInternalServerError, checkoutRR.Code, "Should fail with empty cart")
}

func TestCheckoutFlow_Unauthenticated(t *testing.T) {
	clearTables(t)

	// Try to checkout without authentication
	checkoutRequest := models.CheckoutRequest{
		CartID: uuid.New(),
		ShippingInfo: models.ShippingInfo{
			Address: "123 Main St",
			City:    "New York",
			Country: "USA",
			ZipCode: "10001",
			Method:  "standard",
			Cost:    5.99,
		},
		PaymentInfo: models.PaymentInfo{
			PaymentMethod: "credit_card",
		},
	}

	checkoutBody, _ := json.Marshal(checkoutRequest)
	checkoutReq, _ := http.NewRequest("POST", "/orders/checkout", bytes.NewBuffer(checkoutBody))
	checkoutReq.Header.Set("Content-Type", "application/json")

	checkoutRR := httptest.NewRecorder()
	testRouter.ServeHTTP(checkoutRR, checkoutReq)
	assert.Equal(t, http.StatusUnauthorized, checkoutRR.Code, "Should fail without authentication")
}
