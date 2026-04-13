package tests

import (
	"bytes"
	"encoding/json"
	"go_ecommerce/pkg/models"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

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

	var orderGroupSummary models.OrderGroupSummary
	err = json.Unmarshal(checkoutRR.Body.Bytes(), &orderGroupSummary)
	require.NoError(t, err)
	require.Len(t, orderGroupSummary.Orders, 1, "Should have exactly 1 order in the group")

	orderSummary := orderGroupSummary.Orders[0]
	assert.NotEqual(t, uuid.Nil, orderSummary.OrderID)
	assert.True(t, floatEquals(orderSummary.TotalAmount, 64.98, 0.001), "Expected 64.98, got %v", orderSummary.TotalAmount)
	assert.Equal(t, 5.0, orderSummary.ShippingCost)

	// 5. Customer confirms the order group
	confirmRequest := struct {
		OrderGroupID uuid.UUID `json:"order_group_id"`
	}{
		OrderGroupID: orderGroupSummary.OrderGroupID,
	}

	confirmBody, _ := json.Marshal(confirmRequest)
	confirmReq, _ := http.NewRequest("POST", "/orders/confirm", bytes.NewBuffer(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(confirmReq, customerAuthData)

	confirmRR := httptest.NewRecorder()
	testRouter.ServeHTTP(confirmRR, confirmReq)
	require.Equal(t, http.StatusCreated, confirmRR.Code, "Failed to confirm order")

	var confirmedOrders []models.OrderWithDetails
	err = json.Unmarshal(confirmRR.Body.Bytes(), &confirmedOrders)
	require.NoError(t, err)
	require.Len(t, confirmedOrders, 1, "Should have exactly 1 confirmed order")

	orderWithDetails := confirmedOrders[0]
	assert.Equal(t, orderSummary.OrderID, orderWithDetails.Order.ID)
	assert.Equal(t, "processing", orderWithDetails.Order.Status)
	assert.Equal(t, "paid", orderWithDetails.Order.PaymentStatus)
	assert.True(t, floatEquals(orderWithDetails.Order.TotalAmount, 64.98, 0.001), "Expected 64.98, got %v", orderWithDetails.Order.TotalAmount)
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

func TestOrderCancelAndDeleteEndpoints(t *testing.T) {
	clearTables(t)

	// Register producer and customer
	producerEmail := "producer-cancel@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerAuthData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	customerEmail := "customer-cancel@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerAuthData, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
	require.NoError(t, err)

	// Producer creates a product
	product := createTestProduct(t, producerAuthData, models.Product{
		Name:  "Cancelable Product",
		Price: 19.99,
		Stock: 5,
	})

	// Customer completes a purchase
	orderWithDetails := completeTestPurchase(t, customerAuthData, product, 1)

	// --- Test /orders/cancel (producer cancels order) ---
	cancelRequest := struct {
		OrderID uuid.UUID `json:"order_id"`
	}{OrderID: orderWithDetails.Order.ID}
	cancelBody, _ := json.Marshal(cancelRequest)
	cancelReq, _ := http.NewRequest("POST", "/orders/cancel", bytes.NewBuffer(cancelBody))
	cancelReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(cancelReq, producerAuthData)

	cancelRR := httptest.NewRecorder()
	testRouter.ServeHTTP(cancelRR, cancelReq)
	assert.Equal(t, http.StatusOK, cancelRR.Code, "Producer should be able to cancel order")
	assert.Contains(t, cancelRR.Body.String(), "Order cancelled successfully")

	// Try to cancel again (should fail)
	cancelBody2, _ := json.Marshal(cancelRequest)
	cancelReq2, _ := http.NewRequest("POST", "/orders/cancel", bytes.NewBuffer(cancelBody2))
	cancelReq2.Header.Set("Content-Type", "application/json")
	addAuthHeaders(cancelReq2, producerAuthData)
	cancelRR2 := httptest.NewRecorder()
	testRouter.ServeHTTP(cancelRR2, cancelReq2)
	assert.Equal(t, http.StatusInternalServerError, cancelRR2.Code, "Should not be able to cancel already canceled order")

	// --- Test /orders/delete (customer soft-deletes order) ---
	// First, create a new order for deletion
	orderWithDetails2 := completeTestPurchase(t, customerAuthData, product, 1)
	deleteRequest := struct {
		OrderID uuid.UUID `json:"order_id"`
	}{OrderID: orderWithDetails2.Order.ID}
	deleteBody, _ := json.Marshal(deleteRequest)
	deleteReq, _ := http.NewRequest("POST", "/orders/delete", bytes.NewBuffer(deleteBody))
	deleteReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(deleteReq, customerAuthData)

	deleteRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteRR, deleteReq)
	assert.Equal(t, http.StatusOK, deleteRR.Code, "Customer should be able to soft-delete order")
	assert.Contains(t, deleteRR.Body.String(), "Order deleted successfully")

	// Try to delete again (should fail)
	deleteBody2, _ := json.Marshal(deleteRequest)
	deleteReq2, _ := http.NewRequest("POST", "/orders/delete", bytes.NewBuffer(deleteBody2))
	deleteReq2.Header.Set("Content-Type", "application/json")
	addAuthHeaders(deleteReq2, customerAuthData)
	deleteRR2 := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteRR2, deleteReq2)
	assert.Equal(t, http.StatusInternalServerError, deleteRR2.Code, "Should not be able to delete already deleted order")
}

func TestMultiProducerOrderWithOneDecline(t *testing.T) {
	clearTables(t)

	// 1. Register 3 producers with products
	producer1Email := "producer1-multi@example.com"
	producer1Password := "password123"
	producer1Payload := createUserDTO(producer1Email, producer1Password, true)
	registerTestUserAuth(t, producer1Payload)
	_, producer1AuthData, _ := loginUserAndGetTokenAuth(t, producer1Email, producer1Password)
	product1 := createTestProduct(t, producer1AuthData, models.Product{
		Name:  "Product from Producer 1",
		Price: 29.99,
		Stock: 10,
	})

	producer2Email := "producer2-multi@example.com"
	producer2Password := "password123"
	producer2Payload := createUserDTO(producer2Email, producer2Password, true)
	registerTestUserAuth(t, producer2Payload)
	_, producer2AuthData, _ := loginUserAndGetTokenAuth(t, producer2Email, producer2Password)
	product2 := createTestProduct(t, producer2AuthData, models.Product{
		Name:  "Product from Producer 2",
		Price: 19.99,
		Stock: 10,
	})

	producer3Email := "producer3-multi@example.com"
	producer3Password := "password123"
	producer3Payload := createUserDTO(producer3Email, producer3Password, true)
	registerTestUserAuth(t, producer3Payload)
	_, producer3AuthData, _ := loginUserAndGetTokenAuth(t, producer3Email, producer3Password)
	product3 := createTestProduct(t, producer3AuthData, models.Product{
		Name:  "Product from Producer 3",
		Price: 39.99,
		Stock: 10,
	})

	// 2. Register customer
	customerEmail := "customer-multi@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerAuthData, _ := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)

	// 3. Customer adds all 3 products to cart
	cartItemsToAdd := []models.CartItemDTO{
		{ProductID: product1.ID, Quantity: 1, Price: product1.Price},
		{ProductID: product2.ID, Quantity: 2, Price: product2.Price},
		{ProductID: product3.ID, Quantity: 1, Price: product3.Price},
	}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(addReq, customerAuthData)
	addRR := httptest.NewRecorder()
	testRouter.ServeHTTP(addRR, addReq)
	require.Equal(t, http.StatusOK, addRR.Code)

	var createdCart models.Cart
	require.NoError(t, json.Unmarshal(addRR.Body.Bytes(), &createdCart))

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
		PaymentInfo: models.PaymentInfo{PaymentMethod: "credit_card"},
	}
	checkoutBody, _ := json.Marshal(checkoutRequest)
	checkoutReq, _ := http.NewRequest("POST", "/orders/checkout", bytes.NewBuffer(checkoutBody))
	checkoutReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(checkoutReq, customerAuthData)
	checkoutRR := httptest.NewRecorder()
	testRouter.ServeHTTP(checkoutRR, checkoutReq)
	require.Equal(t, http.StatusOK, checkoutRR.Code)

	var orderGroupSummary models.OrderGroupSummary
	require.NoError(t, json.Unmarshal(checkoutRR.Body.Bytes(), &orderGroupSummary))
	require.Len(t, orderGroupSummary.Orders, 3, "Should have 3 orders from 3 producers")

	// 5. Customer confirms the order group
	confirmRequest := struct {
		OrderGroupID uuid.UUID `json:"order_group_id"`
	}{OrderGroupID: orderGroupSummary.OrderGroupID}
	confirmBody, _ := json.Marshal(confirmRequest)
	confirmReq, _ := http.NewRequest("POST", "/orders/confirm", bytes.NewBuffer(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(confirmReq, customerAuthData)
	confirmRR := httptest.NewRecorder()
	testRouter.ServeHTTP(confirmRR, confirmReq)
	require.Equal(t, http.StatusCreated, confirmRR.Code)

	var confirmedOrders []models.OrderWithDetails
	require.NoError(t, json.Unmarshal(confirmRR.Body.Bytes(), &confirmedOrders))
	require.Len(t, confirmedOrders, 3)
	for _, order := range confirmedOrders {
		assert.Equal(t, "processing", order.Order.Status)
		assert.Equal(t, "paid", order.Order.PaymentStatus)
	}

	// 6. Producer 2 cancels their order (the middle-priced one)
	// Find producer2's order by total amount (19.99*2 + 5.00 = 44.98)
	producer2OrderID := uuid.Nil
	for _, order := range confirmedOrders {
		if floatEquals(order.Order.TotalAmount, 44.98, 0.01) {
			producer2OrderID = order.Order.ID
			break
		}
	}
	require.NotEqual(t, uuid.Nil, producer2OrderID, "Could not find producer2's order")

	cancelBody, _ := json.Marshal(struct {
		OrderID uuid.UUID `json:"order_id"`
	}{OrderID: producer2OrderID})
	cancelReq, _ := http.NewRequest("POST", "/orders/cancel", bytes.NewBuffer(cancelBody))
	cancelReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(cancelReq, producer2AuthData)
	cancelRR := httptest.NewRecorder()
	testRouter.ServeHTTP(cancelRR, cancelReq)
	require.Equal(t, http.StatusOK, cancelRR.Code)

	// 7. Verify Producer 2's order is canceled and refunded
	detailsBody, _ := json.Marshal(struct {
		OrderID uuid.UUID `json:"order_id"`
	}{OrderID: producer2OrderID})
	detailsReq, _ := http.NewRequest("POST", "/orders/details", bytes.NewBuffer(detailsBody))
	detailsReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(detailsReq, customerAuthData)
	detailsRR := httptest.NewRecorder()
	testRouter.ServeHTTP(detailsRR, detailsReq)
	require.Equal(t, http.StatusOK, detailsRR.Code)

	var canceledOrder models.OrderWithDetails
	require.NoError(t, json.Unmarshal(detailsRR.Body.Bytes(), &canceledOrder))
	assert.Equal(t, "canceled", canceledOrder.Order.Status)
	assert.Equal(t, "refunded", canceledOrder.Order.PaymentStatus)
	assert.Equal(t, "refunded", canceledOrder.Payment.Status)

	// 8. Verify Producer 1's and Producer 3's orders are still processing
	// Find producer1 and producer3 orders by their total amounts
	var producer1OrderID, producer3OrderID uuid.UUID
	for _, order := range confirmedOrders {
		if floatEquals(order.Order.TotalAmount, 34.99, 0.01) {
			producer1OrderID = order.Order.ID
		} else if floatEquals(order.Order.TotalAmount, 44.99, 0.01) {
			producer3OrderID = order.Order.ID
		}
	}
	require.NotEqual(t, uuid.Nil, producer1OrderID, "Could not find producer1's order")
	require.NotEqual(t, uuid.Nil, producer3OrderID, "Could not find producer3's order")

	detailsBody1, _ := json.Marshal(struct {
		OrderID uuid.UUID `json:"order_id"`
	}{OrderID: producer1OrderID})
	detailsReq1, _ := http.NewRequest("POST", "/orders/details", bytes.NewBuffer(detailsBody1))
	detailsReq1.Header.Set("Content-Type", "application/json")
	addAuthHeaders(detailsReq1, customerAuthData)
	detailsRR1 := httptest.NewRecorder()
	testRouter.ServeHTTP(detailsRR1, detailsReq1)
	require.Equal(t, http.StatusOK, detailsRR1.Code)

	var producer1Order models.OrderWithDetails
	require.NoError(t, json.Unmarshal(detailsRR1.Body.Bytes(), &producer1Order))
	assert.Equal(t, "processing", producer1Order.Order.Status)
	assert.Equal(t, "paid", producer1Order.Order.PaymentStatus)

	detailsBody3, _ := json.Marshal(struct {
		OrderID uuid.UUID `json:"order_id"`
	}{OrderID: producer3OrderID})
	detailsReq3, _ := http.NewRequest("POST", "/orders/details", bytes.NewBuffer(detailsBody3))
	detailsReq3.Header.Set("Content-Type", "application/json")
	addAuthHeaders(detailsReq3, customerAuthData)
	detailsRR3 := httptest.NewRecorder()
	testRouter.ServeHTTP(detailsRR3, detailsReq3)
	require.Equal(t, http.StatusOK, detailsRR3.Code)

	var producer3Order models.OrderWithDetails
	require.NoError(t, json.Unmarshal(detailsRR3.Body.Bytes(), &producer3Order))
	assert.Equal(t, "processing", producer3Order.Order.Status)
	assert.Equal(t, "paid", producer3Order.Order.PaymentStatus)

	// 9. Verify customer can still see all 3 orders
	userOrdersReq, _ := http.NewRequest("GET", "/orders/user", nil)
	addAuthHeaders(userOrdersReq, customerAuthData)
	userOrdersRR := httptest.NewRecorder()
	testRouter.ServeHTTP(userOrdersRR, userOrdersReq)
	require.Equal(t, http.StatusOK, userOrdersRR.Code)

	var userOrders []models.Order
	require.NoError(t, json.Unmarshal(userOrdersRR.Body.Bytes(), &userOrders))
	assert.Len(t, userOrders, 3)
}

func TestMultiProducerOrderAllDeclined(t *testing.T) {
	clearTables(t)

	// 1. Register 3 producers with products
	producer1Email := "producer1-decline@example.com"
	producer1Password := "password123"
	producer1Payload := createUserDTO(producer1Email, producer1Password, true)
	registerTestUserAuth(t, producer1Payload)
	_, producer1AuthData, _ := loginUserAndGetTokenAuth(t, producer1Email, producer1Password)
	product1 := createTestProduct(t, producer1AuthData, models.Product{
		Name:  "Declinable Product 1",
		Price: 25.00,
		Stock: 5,
	})

	producer2Email := "producer2-decline@example.com"
	producer2Password := "password123"
	producer2Payload := createUserDTO(producer2Email, producer2Password, true)
	registerTestUserAuth(t, producer2Payload)
	_, producer2AuthData, _ := loginUserAndGetTokenAuth(t, producer2Email, producer2Password)
	product2 := createTestProduct(t, producer2AuthData, models.Product{
		Name:  "Declinable Product 2",
		Price: 35.00,
		Stock: 5,
	})

	producer3Email := "producer3-decline@example.com"
	producer3Password := "password123"
	producer3Payload := createUserDTO(producer3Email, producer3Password, true)
	registerTestUserAuth(t, producer3Payload)
	_, producer3AuthData, _ := loginUserAndGetTokenAuth(t, producer3Email, producer3Password)
	product3 := createTestProduct(t, producer3AuthData, models.Product{
		Name:  "Declinable Product 3",
		Price: 45.00,
		Stock: 5,
	})

	// 2. Register customer
	customerEmail := "customer-decline@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerAuthData, _ := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)

	// 3. Customer adds all 3 products to cart
	cartItemsToAdd := []models.CartItemDTO{
		{ProductID: product1.ID, Quantity: 1, Price: product1.Price},
		{ProductID: product2.ID, Quantity: 1, Price: product2.Price},
		{ProductID: product3.ID, Quantity: 1, Price: product3.Price},
	}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(addReq, customerAuthData)
	addRR := httptest.NewRecorder()
	testRouter.ServeHTTP(addRR, addReq)
	require.Equal(t, http.StatusOK, addRR.Code)

	var createdCart models.Cart
	require.NoError(t, json.Unmarshal(addRR.Body.Bytes(), &createdCart))

	// 4. Customer initiates checkout
	checkoutRequest := models.CheckoutRequest{
		CartID: createdCart.ID,
		ShippingInfo: models.ShippingInfo{
			Address: "456 Decline St",
			City:    "Decline City",
			Country: "Declineland",
			ZipCode: "99999",
			Method:  "express",
			Cost:    15.00,
		},
		PaymentInfo: models.PaymentInfo{PaymentMethod: "credit_card"},
	}
	checkoutBody, _ := json.Marshal(checkoutRequest)
	checkoutReq, _ := http.NewRequest("POST", "/orders/checkout", bytes.NewBuffer(checkoutBody))
	checkoutReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(checkoutReq, customerAuthData)
	checkoutRR := httptest.NewRecorder()
	testRouter.ServeHTTP(checkoutRR, checkoutReq)
	require.Equal(t, http.StatusOK, checkoutRR.Code)

	var orderGroupSummary models.OrderGroupSummary
	require.NoError(t, json.Unmarshal(checkoutRR.Body.Bytes(), &orderGroupSummary))
	require.Len(t, orderGroupSummary.Orders, 3)

	// 5. Customer confirms the order group
	confirmRequest := struct {
		OrderGroupID uuid.UUID `json:"order_group_id"`
	}{OrderGroupID: orderGroupSummary.OrderGroupID}
	confirmBody, _ := json.Marshal(confirmRequest)
	confirmReq, _ := http.NewRequest("POST", "/orders/confirm", bytes.NewBuffer(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(confirmReq, customerAuthData)
	confirmRR := httptest.NewRecorder()
	testRouter.ServeHTTP(confirmRR, confirmReq)
	require.Equal(t, http.StatusCreated, confirmRR.Code)

	var confirmedOrders []models.OrderWithDetails
	require.NoError(t, json.Unmarshal(confirmRR.Body.Bytes(), &confirmedOrders))
	require.Len(t, confirmedOrders, 3)

	// 6. All 3 producers cancel their orders
	for i, producerAuthData := range []*TestAuthData{producer1AuthData, producer2AuthData, producer3AuthData} {
		cancelBody, _ := json.Marshal(struct {
			OrderID uuid.UUID `json:"order_id"`
		}{OrderID: confirmedOrders[i].Order.ID})
		cancelReq, _ := http.NewRequest("POST", "/orders/cancel", bytes.NewBuffer(cancelBody))
		cancelReq.Header.Set("Content-Type", "application/json")
		addAuthHeaders(cancelReq, producerAuthData)
		cancelRR := httptest.NewRecorder()
		testRouter.ServeHTTP(cancelRR, cancelReq)
		require.Equal(t, http.StatusOK, cancelRR.Code)
	}

	// 7. Verify all orders are canceled and refunded
	userOrdersReq, _ := http.NewRequest("GET", "/orders/user", nil)
	addAuthHeaders(userOrdersReq, customerAuthData)
	userOrdersRR := httptest.NewRecorder()
	testRouter.ServeHTTP(userOrdersRR, userOrdersReq)
	require.Equal(t, http.StatusOK, userOrdersRR.Code)

	var userOrders []models.Order
	require.NoError(t, json.Unmarshal(userOrdersRR.Body.Bytes(), &userOrders))
	require.Len(t, userOrders, 3)

	for _, order := range userOrders {
		assert.Equal(t, "canceled", order.Status)
		assert.Equal(t, "refunded", order.PaymentStatus)
	}

	// 8. Verify the total refunded amount is correct
	var totalRefunded float64
	for _, order := range userOrders {
		totalRefunded += order.TotalAmount
	}
	expectedTotal := 25.00 + 5.0 + 35.00 + 5.0 + 45.00 + 5.0 // products + shipping per order
	assert.True(t, floatEquals(totalRefunded, expectedTotal, 0.01), "Expected total refunded %v, got %v", expectedTotal, totalRefunded)
}
