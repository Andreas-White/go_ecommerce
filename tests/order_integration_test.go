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
	_, producerAuthData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	customerEmail := "customer-checkout@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerAuthData, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
	require.NoError(t, err)

	// 2. Producer creates a product
	product := createTestProduct(t, producerAuthData, models.Product{
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
	addAuthHeaders(addReq, customerAuthData)

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
	addAuthHeaders(checkoutReq, customerAuthData)

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
	addAuthHeaders(confirmReq, customerAuthData)

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
	addAuthHeaders(detailsReq, customerAuthData)

	detailsRR := httptest.NewRecorder()
	testRouter.ServeHTTP(detailsRR, detailsReq)
	require.Equal(t, http.StatusOK, detailsRR.Code, "Failed to get order details")

	// 7. Verify user orders can be retrieved
	userOrdersReq, _ := http.NewRequest("GET", "/orders/user", nil)
	userOrdersReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(userOrdersReq, customerAuthData)

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
	_, customerAuthData, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
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
	addAuthHeaders(checkoutReq, customerAuthData)

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

func TestOrderFulfillmentFlow_Complete(t *testing.T) {
	clearTables(t)

	// 1. Register a producer and a customer
	producerEmail := "producer-fulfillment@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerAuthData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	customerEmail := "customer-fulfillment@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerAuthData, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
	require.NoError(t, err)

	// 2. Producer creates a product
	product := createTestProduct(t, producerAuthData, models.Product{
		Name:  "Test Product for Fulfillment",
		Price: 49.99,
		Stock: 5,
	})

	// 3. Customer completes a purchase
	orderWithDetails := completeTestPurchase(t, customerAuthData, product, 2)

	// 4. Producer views their orders
	producerOrdersReq, _ := http.NewRequest("GET", "/orders/producer", nil)
	producerOrdersReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(producerOrdersReq, producerAuthData)

	producerOrdersRR := httptest.NewRecorder()
	testRouter.ServeHTTP(producerOrdersRR, producerOrdersReq)
	require.Equal(t, http.StatusOK, producerOrdersRR.Code, "Failed to get producer orders")

	var producerOrders []models.OrderWithDetails
	err = json.Unmarshal(producerOrdersRR.Body.Bytes(), &producerOrders)
	require.NoError(t, err)
	assert.Len(t, producerOrders, 1)
	assert.Equal(t, orderWithDetails.Order.ID, producerOrders[0].Order.ID)
	assert.Equal(t, "processing", producerOrders[0].Order.Status)
	assert.Equal(t, "paid", producerOrders[0].Order.PaymentStatus)

	// 5. Producer accepts the order
	acceptRequest := models.OrderFulfillmentRequest{
		OrderID:   orderWithDetails.Order.ID,
		NewStatus: "accepted",
	}

	acceptBody, _ := json.Marshal(acceptRequest)
	acceptReq, _ := http.NewRequest("POST", "/orders/fulfill", bytes.NewBuffer(acceptBody))
	acceptReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(acceptReq, producerAuthData)

	acceptRR := httptest.NewRecorder()
	testRouter.ServeHTTP(acceptRR, acceptReq)
	require.Equal(t, http.StatusOK, acceptRR.Code, "Failed to accept order")

	var acceptResponse models.OrderFulfillmentResponse
	err = json.Unmarshal(acceptRR.Body.Bytes(), &acceptResponse)
	require.NoError(t, err)
	assert.Equal(t, orderWithDetails.Order.ID, acceptResponse.OrderID)
	assert.Equal(t, "accepted", acceptResponse.Status)
	assert.Nil(t, acceptResponse.TrackingCode)
	assert.Nil(t, acceptResponse.ShippedAt)

	// 6. Producer marks order as preparing
	preparingRequest := models.OrderFulfillmentRequest{
		OrderID:   orderWithDetails.Order.ID,
		NewStatus: "preparing",
	}

	preparingBody, _ := json.Marshal(preparingRequest)
	preparingReq, _ := http.NewRequest("POST", "/orders/fulfill", bytes.NewBuffer(preparingBody))
	preparingReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(preparingReq, producerAuthData)

	preparingRR := httptest.NewRecorder()
	testRouter.ServeHTTP(preparingRR, preparingReq)
	require.Equal(t, http.StatusOK, preparingRR.Code, "Failed to mark order as preparing")

	var preparingResponse models.OrderFulfillmentResponse
	err = json.Unmarshal(preparingRR.Body.Bytes(), &preparingResponse)
	require.NoError(t, err)
	assert.Equal(t, orderWithDetails.Order.ID, preparingResponse.OrderID)
	assert.Equal(t, "preparing", preparingResponse.Status)

	// 7. Producer ships the order with tracking code
	trackingCode := "TRK123456789"
	shipRequest := models.OrderFulfillmentRequest{
		OrderID:      orderWithDetails.Order.ID,
		NewStatus:    "shipped",
		TrackingCode: &trackingCode,
	}

	shipBody, _ := json.Marshal(shipRequest)
	shipReq, _ := http.NewRequest("POST", "/orders/fulfill", bytes.NewBuffer(shipBody))
	shipReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(shipReq, producerAuthData)

	shipRR := httptest.NewRecorder()
	testRouter.ServeHTTP(shipRR, shipReq)
	require.Equal(t, http.StatusOK, shipRR.Code, "Failed to ship order")

	var shipResponse models.OrderFulfillmentResponse
	err = json.Unmarshal(shipRR.Body.Bytes(), &shipResponse)
	require.NoError(t, err)
	assert.Equal(t, orderWithDetails.Order.ID, shipResponse.OrderID)
	assert.Equal(t, "shipped", shipResponse.Status)
	assert.Equal(t, trackingCode, *shipResponse.TrackingCode)
	assert.NotNil(t, shipResponse.ShippedAt)

	// 8. Verify customer can see updated order status
	detailsRequest := struct {
		OrderID uuid.UUID `json:"order_id"`
	}{
		OrderID: orderWithDetails.Order.ID,
	}

	detailsBody, _ := json.Marshal(detailsRequest)
	detailsReq, _ := http.NewRequest("POST", "/orders/details", bytes.NewBuffer(detailsBody))
	detailsReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(detailsReq, customerAuthData)

	detailsRR := httptest.NewRecorder()
	testRouter.ServeHTTP(detailsRR, detailsReq)
	require.Equal(t, http.StatusOK, detailsRR.Code, "Failed to get updated order details")

	var updatedOrderDetails models.OrderWithDetails
	err = json.Unmarshal(detailsRR.Body.Bytes(), &updatedOrderDetails)
	require.NoError(t, err)
	assert.Equal(t, "shipped", updatedOrderDetails.Order.Status)
	assert.Equal(t, trackingCode, *updatedOrderDetails.Shipping.TrackingCode)
	assert.NotNil(t, updatedOrderDetails.Shipping.ShippedAt)
}

func TestOrderFulfillmentFlow_NegativeTests(t *testing.T) {
	clearTables(t)

	// 1. Register a customer (non-producer)
	customerEmail := "customer-negative@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerAuthData, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
	require.NoError(t, err)

	// 2. Try to access producer orders endpoint (should fail)
	producerOrdersReq, _ := http.NewRequest("GET", "/orders/producer", nil)
	producerOrdersReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(producerOrdersReq, customerAuthData)

	producerOrdersRR := httptest.NewRecorder()
	testRouter.ServeHTTP(producerOrdersRR, producerOrdersReq)
	assert.Equal(t, http.StatusForbidden, producerOrdersRR.Code, "Non-producer should not be able to access producer orders")

	// 3. Try to fulfill an order (should fail)
	fulfillRequest := models.OrderFulfillmentRequest{
		OrderID:   uuid.New(),
		NewStatus: "accepted",
	}

	fulfillBody, _ := json.Marshal(fulfillRequest)
	fulfillReq, _ := http.NewRequest("POST", "/orders/fulfill", bytes.NewBuffer(fulfillBody))
	fulfillReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(fulfillReq, customerAuthData)

	fulfillRR := httptest.NewRecorder()
	testRouter.ServeHTTP(fulfillRR, fulfillReq)
	assert.Equal(t, http.StatusForbidden, fulfillRR.Code, "Non-producer should not be able to fulfill orders")

	// 4. Register a producer and create a product
	producerEmail := "producer-negative@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerAuthData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	product := createTestProduct(t, producerAuthData, models.Product{
		Name:  "Test Product for Negative Tests",
		Price: 29.99,
		Stock: 3,
	})

	// 5. Try to fulfill an order that doesn't exist (should fail)
	invalidFulfillRequest := models.OrderFulfillmentRequest{
		OrderID:   uuid.New(),
		NewStatus: "accepted",
	}

	invalidFulfillBody, _ := json.Marshal(invalidFulfillRequest)
	invalidFulfillReq, _ := http.NewRequest("POST", "/orders/fulfill", bytes.NewBuffer(invalidFulfillBody))
	invalidFulfillReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(invalidFulfillReq, producerAuthData)

	invalidFulfillRR := httptest.NewRecorder()
	testRouter.ServeHTTP(invalidFulfillRR, invalidFulfillReq)
	assert.Equal(t, http.StatusInternalServerError, invalidFulfillRR.Code, "Should fail when trying to fulfill non-existent order")

	// 6. Try to ship without tracking code (should fail)
	customer2Email := "customer2-negative@example.com"
	customer2Password := "password123"
	customer2Payload := createUserDTO(customer2Email, customer2Password, false)
	registerTestUserAuth(t, customer2Payload)
	_, customer2AuthData, err := loginUserAndGetTokenAuth(t, customer2Email, customer2Password)
	require.NoError(t, err)

	orderWithDetails := completeTestPurchase(t, customer2AuthData, product, 1)

	shipWithoutTrackingRequest := models.OrderFulfillmentRequest{
		OrderID:   orderWithDetails.Order.ID,
		NewStatus: "shipped",
		// No tracking code provided
	}

	shipWithoutTrackingBody, _ := json.Marshal(shipWithoutTrackingRequest)
	shipWithoutTrackingReq, _ := http.NewRequest("POST", "/orders/fulfill", bytes.NewBuffer(shipWithoutTrackingBody))
	shipWithoutTrackingReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(shipWithoutTrackingReq, producerAuthData)

	shipWithoutTrackingRR := httptest.NewRecorder()
	testRouter.ServeHTTP(shipWithoutTrackingRR, shipWithoutTrackingReq)
	assert.Equal(t, http.StatusInternalServerError, shipWithoutTrackingRR.Code, "Should fail when trying to ship without tracking code")
}

// Helper function to complete a test purchase
func completeTestPurchase(t *testing.T, customerAuthData *TestAuthData, product models.Product, quantity int) models.OrderWithDetails {
	// Add product to cart
	cartItemsToAdd := []models.CartItemDTO{
		{ProductID: product.ID, Quantity: quantity, Price: product.Price},
	}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(addReq, customerAuthData)

	addRR := httptest.NewRecorder()
	testRouter.ServeHTTP(addRR, addReq)
	require.Equal(t, http.StatusOK, addRR.Code, "Failed to add items to cart")

	var createdCart models.Cart
	err := json.Unmarshal(addRR.Body.Bytes(), &createdCart)
	require.NoError(t, err)

	// Process checkout
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
	addAuthHeaders(checkoutReq, customerAuthData)

	checkoutRR := httptest.NewRecorder()
	testRouter.ServeHTTP(checkoutRR, checkoutReq)
	require.Equal(t, http.StatusOK, checkoutRR.Code, "Failed to process checkout")

	var orderSummary models.OrderSummary
	err = json.Unmarshal(checkoutRR.Body.Bytes(), &orderSummary)
	require.NoError(t, err)

	// Confirm order
	confirmRequest := struct {
		OrderID uuid.UUID `json:"order_id"`
	}{
		OrderID: orderSummary.OrderID,
	}

	confirmBody, _ := json.Marshal(confirmRequest)
	confirmReq, _ := http.NewRequest("POST", "/orders/confirm", bytes.NewBuffer(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(confirmReq, customerAuthData)

	confirmRR := httptest.NewRecorder()
	testRouter.ServeHTTP(confirmRR, confirmReq)
	require.Equal(t, http.StatusCreated, confirmRR.Code, "Failed to confirm order")

	var orderWithDetails models.OrderWithDetails
	err = json.Unmarshal(confirmRR.Body.Bytes(), &orderWithDetails)
	require.NoError(t, err)

	return orderWithDetails
}

func TestSalesReportFlow_Complete(t *testing.T) {
	clearTables(t)

	// 1. Register a producer
	producerEmail := "producer-sales@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerAuthData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	// 2. Producer creates multiple products in different categories
	product1 := createTestProduct(t, producerAuthData, models.Product{
		Name:     "Electronics Product",
		Price:    99.99,
		Stock:    10,
		Category: "Electronics",
	})

	product2 := createTestProduct(t, producerAuthData, models.Product{
		Name:     "Clothing Product",
		Price:    49.99,
		Stock:    20,
		Category: "Clothing",
	})

	product3 := createTestProduct(t, producerAuthData, models.Product{
		Name:     "Another Electronics Product",
		Price:    199.99,
		Stock:    5,
		Category: "Electronics",
	})

	// 3. Register multiple customers and create purchases
	customers := []string{"customer1@example.com", "customer2@example.com", "customer3@example.com"}
	var allOrders []models.OrderWithDetails

	for i, customerEmail := range customers {
		customerPassword := "password123"
		customerPayload := createUserDTO(customerEmail, customerPassword, false)
		registerTestUserAuth(t, customerPayload)
		_, customerAuthData, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
		require.NoError(t, err)

		// Customer buys different combinations of products
		var products []models.Product
		var quantities []int

		switch i {
		case 0: // Customer 1: buys electronics
			products = []models.Product{product1, product3}
			quantities = []int{2, 1}
		case 1: // Customer 2: buys clothing and electronics
			products = []models.Product{product2, product1}
			quantities = []int{3, 1}
		case 2: // Customer 3: buys only clothing
			products = []models.Product{product2}
			quantities = []int{2}
		}

		for j, product := range products {
			order := completeTestPurchase(t, customerAuthData, product, quantities[j])
			allOrders = append(allOrders, order)
		}
	}

	// 4. Get overall sales report
	salesReportRequest := models.SalesReportRequest{}

	salesReportBody, _ := json.Marshal(salesReportRequest)
	salesReportReq, _ := http.NewRequest("POST", "/orders/sales-report", bytes.NewBuffer(salesReportBody))
	salesReportReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(salesReportReq, producerAuthData)

	salesReportRR := httptest.NewRecorder()
	testRouter.ServeHTTP(salesReportRR, salesReportReq)
	require.Equal(t, http.StatusOK, salesReportRR.Code, "Failed to get sales report")

	var salesReport models.SalesReportResponse
	err = json.Unmarshal(salesReportRR.Body.Bytes(), &salesReport)
	require.NoError(t, err)

	// 5. Verify sales report data
	assert.Equal(t, len(allOrders), salesReport.TotalOrders)
	assert.True(t, salesReport.TotalRevenue > 0, "Total revenue should be positive")
	assert.True(t, salesReport.TotalItemsSold > 0, "Total items sold should be positive")
	assert.True(t, salesReport.AverageOrderValue > 0, "Average order value should be positive")
	assert.Len(t, salesReport.TopSellingProducts, 3, "Should have 3 products in top selling")
	assert.Len(t, salesReport.SalesByCategory, 2, "Should have 2 categories")

	// 6. Verify top selling products are ordered by units sold
	if len(salesReport.TopSellingProducts) >= 2 {
		assert.GreaterOrEqual(t, salesReport.TopSellingProducts[0].UnitsSold, salesReport.TopSellingProducts[1].UnitsSold)
	}

	// 7. Get sales report filtered by category
	categoryFilterRequest := models.SalesReportRequest{
		Category: &product1.Category, // Electronics
	}

	categoryFilterBody, _ := json.Marshal(categoryFilterRequest)
	categoryFilterReq, _ := http.NewRequest("POST", "/orders/sales-report", bytes.NewBuffer(categoryFilterBody))
	categoryFilterReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(categoryFilterReq, producerAuthData)

	categoryFilterRR := httptest.NewRecorder()
	testRouter.ServeHTTP(categoryFilterRR, categoryFilterReq)
	require.Equal(t, http.StatusOK, categoryFilterRR.Code, "Failed to get category-filtered sales report")

	var categoryFilteredReport models.SalesReportResponse
	err = json.Unmarshal(categoryFilterRR.Body.Bytes(), &categoryFilteredReport)
	require.NoError(t, err)

	// 8. Verify category filter works
	assert.Len(t, categoryFilteredReport.SalesByCategory, 1, "Should have only Electronics category")
	assert.Equal(t, "Electronics", categoryFilteredReport.SalesByCategory[0].Category)
	assert.True(t, categoryFilteredReport.TotalRevenue > 0, "Category revenue should be positive")
}

func TestSalesReportFlow_NegativeTests(t *testing.T) {
	clearTables(t)

	// 1. Register a customer (non-producer)
	customerEmail := "customer-sales-negative@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerAuthData, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
	require.NoError(t, err)

	// 2. Try to access sales report (should fail)
	salesReportRequest := models.SalesReportRequest{}

	salesReportBody, _ := json.Marshal(salesReportRequest)
	salesReportReq, _ := http.NewRequest("POST", "/orders/sales-report", bytes.NewBuffer(salesReportBody))
	salesReportReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(salesReportReq, customerAuthData)

	salesReportRR := httptest.NewRecorder()
	testRouter.ServeHTTP(salesReportRR, salesReportReq)
	assert.Equal(t, http.StatusForbidden, salesReportRR.Code, "Non-producer should not be able to access sales reports")

	// 3. Register a producer with no sales
	producerEmail := "producer-no-sales@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerAuthData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	// 4. Get sales report for producer with no sales
	noSalesRequest := models.SalesReportRequest{}

	noSalesBody, _ := json.Marshal(noSalesRequest)
	noSalesReq, _ := http.NewRequest("POST", "/orders/sales-report", bytes.NewBuffer(noSalesBody))
	noSalesReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(noSalesReq, producerAuthData)

	noSalesRR := httptest.NewRecorder()
	testRouter.ServeHTTP(noSalesRR, noSalesReq)
	require.Equal(t, http.StatusOK, noSalesRR.Code, "Should get empty sales report")

	var noSalesReport models.SalesReportResponse
	err = json.Unmarshal(noSalesRR.Body.Bytes(), &noSalesReport)
	require.NoError(t, err)

	// 5. Verify empty sales report
	assert.Equal(t, 0, noSalesReport.TotalOrders)
	assert.Equal(t, 0.0, noSalesReport.TotalRevenue)
	assert.Equal(t, 0, noSalesReport.TotalItemsSold)
	assert.Equal(t, 0.0, noSalesReport.AverageOrderValue)
	assert.Len(t, noSalesReport.TopSellingProducts, 0)
	assert.Len(t, noSalesReport.SalesByCategory, 0)
}
