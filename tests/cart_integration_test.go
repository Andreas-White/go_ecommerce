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
func createTestProduct(t *testing.T, authData *TestAuthData, product models.Product) models.Product {
	productBytes, err := json.Marshal(product)
	require.NoError(t, err)

	req, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(productBytes))
	req.Header.Set("Content-Type", "application/json")
	addAuthHeaders(req, authData)

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
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerAuthData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	userEmail := "customer-cart@example.com"
	userPassword := "password123"
	userPayload := createUserDTO(userEmail, userPassword, false)
	registerTestUserAuth(t, userPayload)
	_, userAuthData, err := loginUserAndGetTokenAuth(t, userEmail, userPassword)
	require.NoError(t, err)

	// 2. Producer creates two products
	product1 := createTestProduct(t, producerAuthData, models.Product{
		Name:  "Test Book",
		Price: 25.50,
		Stock: 10,
	})
	product2 := createTestProduct(t, producerAuthData, models.Product{
		Name:  "Test Pen",
		Price: 1.50,
		Stock: 100,
	})

	// 3. User adds both products to the cart
	cartItemsToAdd := []models.CartItemDTO{
		{ProductID: product1.ID, Quantity: 2, Price: product1.Price},
		{ProductID: product2.ID, Quantity: 5, Price: product2.Price},
	}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(addReq, userAuthData)

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
	addAuthHeaders(getReq, userAuthData)

	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	require.Equal(t, http.StatusOK, getRR.Code, "Failed to get cart items")

	var addedItems []models.CartItemProductDetails
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
	addAuthHeaders(removeReq, userAuthData)

	removeRR := httptest.NewRecorder()
	testRouter.ServeHTTP(removeRR, removeReq)
	require.Equal(t, http.StatusOK, removeRR.Code, "Failed to remove item from cart")

	// 5. Retrieve cart items and verify
	getBodyAfterRemove, _ := json.Marshal(cartID)
	getReqAfterRemove, _ := http.NewRequest("POST", "/cart/get", bytes.NewBuffer(getBodyAfterRemove))
	getReqAfterRemove.Header.Set("Content-Type", "application/json")
	addAuthHeaders(getReqAfterRemove, userAuthData)

	getRRAfterRemove := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterRemove, getReqAfterRemove)
	require.Equal(t, http.StatusOK, getRRAfterRemove.Code, "Failed to get cart items")

	var remainingItems []models.CartItemProductDetails
	err = json.Unmarshal(getRRAfterRemove.Body.Bytes(), &remainingItems)
	require.NoError(t, err)
	require.Len(t, remainingItems, 1, "Cart should have 1 item left")
	assert.Equal(t, product1.ID, remainingItems[0].ProductID)
	assert.Equal(t, 2, remainingItems[0].Quantity)

	// 6. Clear the cart
	clearBody, _ := json.Marshal(cartID)
	clearReq, _ := http.NewRequest("POST", "/cart/clear", bytes.NewBuffer(clearBody))
	clearReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(clearReq, userAuthData)

	clearRR := httptest.NewRecorder()
	testRouter.ServeHTTP(clearRR, clearReq)
	require.Equal(t, http.StatusOK, clearRR.Code, "Failed to clear cart")

	// 7. Verify cart is empty
	getReqAfterClear, _ := http.NewRequest("POST", "/cart/get", bytes.NewBuffer(getBody))
	getReqAfterClear.Header.Set("Content-Type", "application/json")
	addAuthHeaders(getReqAfterClear, userAuthData)

	getRRAfterClear := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterClear, getReqAfterClear)
	require.Equal(t, http.StatusOK, getRRAfterClear.Code)
	var finalItems []models.CartItemProductDetails
	json.Unmarshal(getRRAfterClear.Body.Bytes(), &finalItems)
	assert.Len(t, finalItems, 0, "Cart should be empty after clearing")
}

func TestCartFlow_GuestUser(t *testing.T) {
	clearTables(t)

	// 1. Register a producer to create products
	producerEmail := "producer-guest-cart@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	_, producerAuthData := registerTestUserAuth(t, producerPayload)

	// 2. Producer creates two products
	product1 := createTestProduct(t, producerAuthData, models.Product{
		Name:        "Guest Book",
		Price:       35.00,
		Stock:       5,
		Category:    "Books",
		ImageUrl:    "https://example.com/guest-book.jpg",
		Description: "A guest book for visitors to sign.",
	})
	product2 := createTestProduct(t, producerAuthData, models.Product{
		Name:        "Guest Marker",
		Price:       2.50,
		Stock:       50,
		Category:    "Stationery",
		ImageUrl:    "https://example.com/guest-marker.jpg",
		Description: "A guest marker for visitors to sign.",
	})

	// 3. Guest adds both products to the cart (no authentication required)
	// First, get CSRF token for guest user
	csrfReq, _ := http.NewRequest("GET", "/csrf", nil)
	csrfRR := httptest.NewRecorder()
	testRouter.ServeHTTP(csrfRR, csrfReq)
	require.Equal(t, http.StatusNoContent, csrfRR.Code, "Failed to get CSRF token")

	// Get CSRF cookie and token
	csrfCookie := csrfRR.Result().Cookies()[0]
	require.NotNil(t, csrfCookie, "CSRF cookie not set")

	cartItemsToAdd := []models.CartItemDTO{
		{ProductID: product1.ID, Quantity: 1, Price: product1.Price},
		{ProductID: product2.ID, Quantity: 10, Price: product2.Price},
	}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.AddCookie(csrfCookie)
	addReq.Header.Set("X-CSRF-Token", csrfCookie.Value)

	addRR := httptest.NewRecorder()
	testRouter.ServeHTTP(addRR, addReq)
	require.Equal(t, http.StatusOK, addRR.Code, "Failed to add items to guest cart")

	var createdCart models.Cart
	err := json.Unmarshal(addRR.Body.Bytes(), &createdCart)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdCart.ID)
	cartID := createdCart.ID

	// Capture the session cookie
	sessionCookie := addRR.Result().Cookies()[0]
	require.NotNil(t, sessionCookie, "Session cookie not set for guest user")

	// 4. Verify items were added by getting the cart
	getReq, _ := http.NewRequest("GET", "/cart/get", nil)
	getReq.AddCookie(sessionCookie)

	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	require.Equal(t, http.StatusOK, getRR.Code, "Failed to get guest cart items")

	var addedItems []models.CartItemProductDetails
	err = json.Unmarshal(getRR.Body.Bytes(), &addedItems)
	require.NoError(t, err)
	require.Len(t, addedItems, 2, "Guest cart should have 2 items after adding")

	// 5. Update cart items
	cartItemsToUpdate := []models.CartItemDTO{
		{CartID: cartID, ProductID: product1.ID, Quantity: 2, Price: product1.Price},
		{CartID: cartID, ProductID: product2.ID, Quantity: 20, Price: product2.Price},
	}
	updateBody, _ := json.Marshal(cartItemsToUpdate)
	updateReq, _ := http.NewRequest("POST", "/cart/update", bytes.NewBuffer(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(sessionCookie)
	updateReq.AddCookie(csrfCookie)
	updateReq.Header.Set("X-CSRF-Token", csrfCookie.Value)

	updateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateRR, updateReq)
	require.Equal(t, http.StatusOK, updateRR.Code, "Failed to update guest cart items")

	// 6. Verify items were updated by getting the cart
	getReqAfterUpdate, _ := http.NewRequest("GET", "/cart/get", nil)
	getReqAfterUpdate.AddCookie(sessionCookie)

	getRRAfterUpdate := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterUpdate, getReqAfterUpdate)
	require.Equal(t, http.StatusOK, getRRAfterUpdate.Code, "Failed to get guest cart items after update")

	var updatedItems []models.CartItemProductDetails
	err = json.Unmarshal(getRRAfterUpdate.Body.Bytes(), &updatedItems)
	require.NoError(t, err)
	require.Len(t, updatedItems, 2, "Guest cart should have 2 items after updating")
	assert.Equal(t, product1.ID, updatedItems[0].ProductID)
	assert.Equal(t, cartID, updatedItems[0].CartID)
	assert.Equal(t, 2, updatedItems[0].Quantity)
	assert.Equal(t, product2.ID, updatedItems[1].ProductID)
	assert.Equal(t, cartID, updatedItems[1].CartID)
	assert.Equal(t, 20, updatedItems[1].Quantity)

	// 7. Guest removes one product from the cart
	itemToRemove := []models.CartItemDTO{
		{CartID: cartID, ProductID: product1.ID},
	}
	removeBody, _ := json.Marshal(itemToRemove)
	removeReq, _ := http.NewRequest("POST", "/cart/remove", bytes.NewBuffer(removeBody))
	removeReq.Header.Set("Content-Type", "application/json")
	removeReq.AddCookie(sessionCookie)
	removeReq.AddCookie(csrfCookie)
	removeReq.Header.Set("X-CSRF-Token", csrfCookie.Value)

	removeRR := httptest.NewRecorder()
	testRouter.ServeHTTP(removeRR, removeReq)
	require.Equal(t, http.StatusOK, removeRR.Code, "Failed to remove item from guest cart")

	// 8. Retrieve cart items and verify
	getReqAfterRemove, _ := http.NewRequest("GET", "/cart/get", nil)
	getReqAfterRemove.AddCookie(sessionCookie)

	getRRAfterRemove := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterRemove, getReqAfterRemove)
	require.Equal(t, http.StatusOK, getRRAfterRemove.Code, "Failed to get guest cart items")

	var remainingItems []models.CartItemProductDetails
	err = json.Unmarshal(getRRAfterRemove.Body.Bytes(), &remainingItems)
	require.NoError(t, err)
	require.Len(t, remainingItems, 1, "Guest cart should have 1 item left")
	assert.Equal(t, product2.ID, remainingItems[0].ProductID)

	// 9. Clear the cart
	clearReq, _ := http.NewRequest("POST", "/cart/clear", nil)
	clearReq.Header.Set("Content-Type", "application/json")
	clearReq.AddCookie(sessionCookie)
	clearReq.AddCookie(csrfCookie)
	clearReq.Header.Set("X-CSRF-Token", csrfCookie.Value)

	clearRR := httptest.NewRecorder()
	testRouter.ServeHTTP(clearRR, clearReq)
	require.Equal(t, http.StatusOK, clearRR.Code, "Failed to clear guest cart")

	// 10. Verify cart is empty
	getReqAfterClear, _ := http.NewRequest("GET", "/cart/get", nil)
	getReqAfterClear.AddCookie(sessionCookie)

	getRRAfterClear := httptest.NewRecorder()
	testRouter.ServeHTTP(getRRAfterClear, getReqAfterClear)
	require.Equal(t, http.StatusOK, getRRAfterClear.Code)
	var finalItems []models.CartItemProductDetails
	json.Unmarshal(getRRAfterClear.Body.Bytes(), &finalItems)
	assert.Len(t, finalItems, 0, "Guest cart should be empty after clearing")
}
