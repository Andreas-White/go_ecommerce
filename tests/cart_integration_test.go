package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go_ecommerce/pkg/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a product for cart tests
func createTestProduct(t *testing.T, token string, product models.Product) models.Product {
	productBytes, err := json.Marshal(product)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(productBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)
	require.Equal(t, http.StatusCreated, rr.Code, "Failed to create product for cart test")

	var createdProduct models.Product
	err = json.Unmarshal(rr.Body.Bytes(), &createdProduct)
	require.NoError(t, err)
	return createdProduct
}

func TestCartFlow_RegisteredUser(t *testing.T) {
	clearTables(t)

	// 1. Register a producer and a regular user
	producerEmail := "producer-cart@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword)
	producerPayload.IsProducer = true
	registerTestUserAuth(t, producerPayload)
	_, producerToken, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	userEmail := "customer-cart@example.com"
	userPassword := "password123"
	userPayload := createUserDTO(userEmail, userPassword)
	registerTestUserAuth(t, userPayload)
	_, userToken, err := loginUserAndGetTokenAuth(t, userEmail, userPassword)
	require.NoError(t, err)

	// 2. Producer creates two products
	product1 := createTestProduct(t, producerToken, models.Product{
		Name:  "Test Book",
		Price: 25.50,
		Stock: 10,
	})
	product2 := createTestProduct(t, producerToken, models.Product{
		Name:  "Test Pen",
		Price: 1.50,
		Stock: 100,
	})

	// 3. User adds both products to the cart
	cartItemsToAdd := []models.CartItemDTO{
		{ProductID: product1.ID, Quantity: 2},
		{ProductID: product2.ID, Quantity: 5},
	}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.Header.Set("Authorization", "Bearer "+userToken)

	addRR := httptest.NewRecorder()
	testRouter.ServeHTTP(addRR, addReq)
	require.Equal(t, http.StatusOK, addRR.Code, "Failed to add items to cart")

	var createdCart models.Cart
	err = json.Unmarshal(addRR.Body.Bytes(), &createdCart)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdCart.ID)
	cartID := createdCart.ID

	// 3a. Verify items were added by getting the cart
	getBody, _ := json.Marshal(cartID)
	getReq, _ := http.NewRequest("POST", "/cart/get", bytes.NewBuffer(getBody))
	getReq.Header.Set("Content-Type", "application/json")
	getReq.Header.Set("Authorization", "Bearer "+userToken)

	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	require.Equal(t, http.StatusOK, getRR.Code, "Failed to get cart items")

	var addedItems []models.CartItem
	err = json.Unmarshal(getRR.Body.Bytes(), &addedItems)
	require.NoError(t, err)
	require.Len(t, addedItems, 2, "Cart should have 2 items after adding")

	// 4. User removes one product from the cart
	itemToRemove := []models.CartItemDTO{
		{CartID: cartID, ProductID: product2.ID},
	}
	removeBody, _ := json.Marshal(itemToRemove)
	removeReq, _ := http.NewRequest("POST", "/cart/remove", bytes.NewBuffer(removeBody))
	removeReq.Header.Set("Content-Type", "application/json")
	removeReq.Header.Set("Authorization", "Bearer "+userToken)

	removeRR := httptest.NewRecorder()
	testRouter.ServeHTTP(removeRR, removeReq)
	require.Equal(t, http.StatusOK, removeRR.Code, "Failed to remove item from cart")

	// 5. Retrieve cart items and verify
	getBodyAfterRemove, _ := json.Marshal(cartID)
	getReqAfterRemove, _ := http.NewRequest("POST", "/cart/get", bytes.NewBuffer(getBodyAfterRemove))
	getReqAfterRemove.Header.Set("Content-Type", "application/json")
	getReqAfterRemove.Header.Set("Authorization", "Bearer "+userToken)

	getRRAfterRemove := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterRemove, getReqAfterRemove)
	require.Equal(t, http.StatusOK, getRRAfterRemove.Code, "Failed to get cart items")

	var remainingItems []models.CartItem
	err = json.Unmarshal(getRRAfterRemove.Body.Bytes(), &remainingItems)
	require.NoError(t, err)
	require.Len(t, remainingItems, 1, "Cart should have 1 item left")
	assert.Equal(t, product1.ID, remainingItems[0].ProductID)
	assert.Equal(t, 2, remainingItems[0].Quantity)

	// 6. Clear the cart
	clearBody, _ := json.Marshal(cartID)
	clearReq, _ := http.NewRequest("POST", "/cart/clear", bytes.NewBuffer(clearBody))
	clearReq.Header.Set("Content-Type", "application/json")
	clearReq.Header.Set("Authorization", "Bearer "+userToken)

	clearRR := httptest.NewRecorder()
	testRouter.ServeHTTP(clearRR, clearReq)
	require.Equal(t, http.StatusOK, clearRR.Code, "Failed to clear cart")

	// 7. Verify cart is empty
	getReqAfterClear, _ := http.NewRequest("POST", "/cart/get", bytes.NewBuffer(getBody))
	getReqAfterClear.Header.Set("Content-Type", "application/json")
	getReqAfterClear.Header.Set("Authorization", "Bearer "+userToken)

	getRRAfterClear := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterClear, getReqAfterClear)
	require.Equal(t, http.StatusOK, getRRAfterClear.Code)
	var finalItems []models.CartItem
	json.Unmarshal(getRRAfterClear.Body.Bytes(), &finalItems)
	assert.Len(t, finalItems, 0, "Cart should be empty after clearing")
}

func TestCartFlow_GuestUser(t *testing.T) {
	clearTables(t)

	// 1. Register a producer to create products
	producerEmail := "producer-guest-cart@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword)
	producerPayload.IsProducer = true
	registerTestUserAuth(t, producerPayload)
	_, producerToken, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	// 2. Producer creates two products
	product1 := createTestProduct(t, producerToken, models.Product{
		Name:  "Guest Book",
		Price: 35.00,
		Stock: 5,
	})
	product2 := createTestProduct(t, producerToken, models.Product{
		Name:  "Guest Marker",
		Price: 2.50,
		Stock: 50,
	})

	// 3. Guest adds both products to the cart
	cartItemsToAdd := []models.CartItemDTO{
		{ProductID: product1.ID, Quantity: 1},
		{ProductID: product2.ID, Quantity: 10},
	}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")

	addRR := httptest.NewRecorder()
	testRouter.ServeHTTP(addRR, addReq)
	require.Equal(t, http.StatusOK, addRR.Code, "Failed to add items to guest cart")

	var createdCart models.Cart
	err = json.Unmarshal(addRR.Body.Bytes(), &createdCart)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdCart.ID)
	cartID := createdCart.ID

	// Capture the session cookie
	sessionCookie := addRR.Result().Cookies()[0]
	require.NotNil(t, sessionCookie, "Session cookie not set for guest user")

	// 3a. Verify items were added by getting the cart
	getBody, _ := json.Marshal(cartID)
	getReq, _ := http.NewRequest("POST", "/cart/get", bytes.NewBuffer(getBody))
	getReq.Header.Set("Content-Type", "application/json")
	getReq.AddCookie(sessionCookie)

	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	require.Equal(t, http.StatusOK, getRR.Code, "Failed to get guest cart items")

	var addedItems []models.CartItem
	err = json.Unmarshal(getRR.Body.Bytes(), &addedItems)
	require.NoError(t, err)
	require.Len(t, addedItems, 2, "Guest cart should have 2 items after adding")

	// 4. Guest removes one product from the cart
	itemToRemove := []models.CartItemDTO{
		{CartID: cartID, ProductID: product1.ID},
	}
	removeBody, _ := json.Marshal(itemToRemove)
	removeReq, _ := http.NewRequest("POST", "/cart/remove", bytes.NewBuffer(removeBody))
	removeReq.Header.Set("Content-Type", "application/json")
	removeReq.AddCookie(sessionCookie)

	removeRR := httptest.NewRecorder()
	testRouter.ServeHTTP(removeRR, removeReq)
	require.Equal(t, http.StatusOK, removeRR.Code, "Failed to remove item from guest cart")

	// 5. Retrieve cart items and verify
	getBodyAfterRemove, _ := json.Marshal(cartID)
	getReqAfterRemove, _ := http.NewRequest("POST", "/cart/get", bytes.NewBuffer(getBodyAfterRemove))
	getReqAfterRemove.Header.Set("Content-Type", "application/json")
	getReqAfterRemove.AddCookie(sessionCookie)

	getRRAfterRemove := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterRemove, getReqAfterRemove)
	require.Equal(t, http.StatusOK, getRRAfterRemove.Code, "Failed to get guest cart items")

	var remainingItems []models.CartItem
	err = json.Unmarshal(getRRAfterRemove.Body.Bytes(), &remainingItems)
	require.NoError(t, err)
	require.Len(t, remainingItems, 1, "Guest cart should have 1 item left")
	assert.Equal(t, product2.ID, remainingItems[0].ProductID)

	// 6. Clear the cart
	clearBody, _ := json.Marshal(cartID)
	clearReq, _ := http.NewRequest("POST", "/cart/clear", bytes.NewBuffer(clearBody))
	clearReq.Header.Set("Content-Type", "application/json")
	clearReq.AddCookie(sessionCookie)

	clearRR := httptest.NewRecorder()
	testRouter.ServeHTTP(clearRR, clearReq)
	require.Equal(t, http.StatusOK, clearRR.Code, "Failed to clear guest cart")

	// 7. Verify cart is empty
	getReqAfterClear, _ := http.NewRequest("POST", "/cart/get", bytes.NewBuffer(getBody))
	getReqAfterClear.Header.Set("Content-Type", "application/json")
	getReqAfterClear.AddCookie(sessionCookie)

	getRRAfterClear := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterClear, getReqAfterClear)
	require.Equal(t, http.StatusOK, getRRAfterClear.Code)
	var finalItems []models.CartItem
	json.Unmarshal(getRRAfterClear.Body.Bytes(), &finalItems)
	assert.Len(t, finalItems, 0, "Guest cart should be empty after clearing")
}
