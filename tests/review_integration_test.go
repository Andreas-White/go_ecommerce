package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_ecommerce/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReviewFlow_Integration(t *testing.T) {
	clearTables(t)

	// 1. Register producer and customer
	producerEmail := "producer-review@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerToken, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	customerEmail := "customer-review@example.com"
	customerPassword := "password123"
	customerPayload := createUserDTO(customerEmail, customerPassword, false)
	registerTestUserAuth(t, customerPayload)
	_, customerToken, err := loginUserAndGetTokenAuth(t, customerEmail, customerPassword)
	require.NoError(t, err)

	// 2. Producer creates a product
	product := createTestProduct(t, producerToken, models.Product{
		Name:  "Reviewable Product",
		Price: 49.99,
		Stock: 10,
	})

	// 3. Customer purchases the product (add to cart, checkout, confirm)
	cartItemsToAdd := []models.CartItemDTO{{ProductID: product.ID, Quantity: 1, Price: product.Price}}
	addBody, _ := json.Marshal(cartItemsToAdd)
	addReq, _ := http.NewRequest("POST", "/cart/add", bytes.NewBuffer(addBody))
	addReq.Header.Set("Content-Type", "application/json")
	addReq.Header.Set("Authorization", "Bearer "+customerToken)
	addRR := httptest.NewRecorder()
	testRouter.ServeHTTP(addRR, addReq)
	require.Equal(t, http.StatusOK, addRR.Code)
	var createdCart models.Cart
	require.NoError(t, json.Unmarshal(addRR.Body.Bytes(), &createdCart))

	checkoutRequest := models.CheckoutRequest{
		CartID: createdCart.ID,
		ShippingInfo: models.ShippingInfo{
			Address: "123 Main St",
			City:    "Testville",
			Country: "Testland",
			ZipCode: "12345",
			Method:  "standard",
			Cost:    4.99,
		},
		PaymentInfo: models.PaymentInfo{PaymentMethod: "credit_card"},
	}
	checkoutBody, _ := json.Marshal(checkoutRequest)
	checkoutReq, _ := http.NewRequest("POST", "/orders/checkout", bytes.NewBuffer(checkoutBody))
	checkoutReq.Header.Set("Content-Type", "application/json")
	checkoutReq.Header.Set("Authorization", "Bearer "+customerToken)
	checkoutRR := httptest.NewRecorder()
	testRouter.ServeHTTP(checkoutRR, checkoutReq)
	require.Equal(t, http.StatusOK, checkoutRR.Code)
	var orderSummary models.OrderSummary
	require.NoError(t, json.Unmarshal(checkoutRR.Body.Bytes(), &orderSummary))

	confirmRequest := struct {
		OrderID uuid.UUID `json:"order_id"`
	}{OrderID: orderSummary.OrderID}
	confirmBody, _ := json.Marshal(confirmRequest)
	confirmReq, _ := http.NewRequest("POST", "/orders/confirm", bytes.NewBuffer(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	confirmReq.Header.Set("Authorization", "Bearer "+customerToken)
	confirmRR := httptest.NewRecorder()
	testRouter.ServeHTTP(confirmRR, confirmReq)
	require.Equal(t, http.StatusCreated, confirmRR.Code)

	// 4. Customer adds a review
	reviewComment := "Great product!"
	reviewPayload := models.ReviewDTO{ProductID: product.ID, Rating: 5, Comment: &reviewComment}
	reviewBytes, _ := json.Marshal(reviewPayload)
	reviewReq, _ := http.NewRequest("POST", "/reviews/add", bytes.NewBuffer(reviewBytes))
	reviewReq.Header.Set("Content-Type", "application/json")
	reviewReq.Header.Set("Authorization", "Bearer "+customerToken)
	reviewRR := httptest.NewRecorder()
	testRouter.ServeHTTP(reviewRR, reviewReq)
	require.Equal(t, http.StatusCreated, reviewRR.Code)

	// 5. Verify review exists
	getReviewsReq, _ := http.NewRequest("GET", fmt.Sprintf("/reviews/get?product_id=%s", product.ID), nil)
	getReviewsRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getReviewsRR, getReviewsReq)
	require.Equal(t, http.StatusOK, getReviewsRR.Code)
	var reviews []models.Review
	require.NoError(t, json.Unmarshal(getReviewsRR.Body.Bytes(), &reviews))
	require.Len(t, reviews, 1)
	assert.Equal(t, 5, reviews[0].Rating)
	assert.Equal(t, reviewComment, *reviews[0].Comment)

	// 6. Update the review
	updatedComment := "Updated review: still great!"
	updatePayload := struct {
		Rating  int     `json:"rating"`
		Comment *string `json:"comment"`
	}{Rating: 4, Comment: &updatedComment}
	updateBytes, _ := json.Marshal(updatePayload)
	updateReq, _ := http.NewRequest("PUT", fmt.Sprintf("/reviews/update?id=%s", reviews[0].ID), bytes.NewBuffer(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+customerToken)
	updateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateRR, updateReq)
	require.Equal(t, http.StatusOK, updateRR.Code)

	// 7. Verify review is updated
	getReviewsRR2 := httptest.NewRecorder()
	testRouter.ServeHTTP(getReviewsRR2, getReviewsReq)
	require.Equal(t, http.StatusOK, getReviewsRR2.Code)
	var reviews2 []models.Review
	require.NoError(t, json.Unmarshal(getReviewsRR2.Body.Bytes(), &reviews2))
	require.Len(t, reviews2, 1)
	assert.Equal(t, 4, reviews2[0].Rating)
	assert.Equal(t, updatedComment, *reviews2[0].Comment)

	// 8. Delete the review
	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/reviews/delete?id=%s", reviews2[0].ID), nil)
	deleteReq.Header.Set("Authorization", "Bearer "+customerToken)
	deleteRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteRR, deleteReq)
	require.Equal(t, http.StatusOK, deleteRR.Code)

	// 9. Verify review is deleted
	getReviewsRR3 := httptest.NewRecorder()
	testRouter.ServeHTTP(getReviewsRR3, getReviewsReq)
	require.Equal(t, http.StatusOK, getReviewsRR3.Code)
	var reviews3 []models.Review
	require.NoError(t, json.Unmarshal(getReviewsRR3.Body.Bytes(), &reviews3))
	assert.Len(t, reviews3, 0)
}

func TestReview_NotPurchased_CannotReview(t *testing.T) {
	clearTables(t)

	// Register producer and non-buyer
	producerEmail := "producer-nobuy@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerToken, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	nonBuyerEmail := "nonbuyer@example.com"
	nonBuyerPassword := "password123"
	nonBuyerPayload := createUserDTO(nonBuyerEmail, nonBuyerPassword, false)
	registerTestUserAuth(t, nonBuyerPayload)
	_, nonBuyerToken, err := loginUserAndGetTokenAuth(t, nonBuyerEmail, nonBuyerPassword)
	require.NoError(t, err)

	// Producer creates a product
	product := createTestProduct(t, producerToken, models.Product{
		Name:  "Unpurchased Product",
		Price: 19.99,
		Stock: 5,
	})

	// Non-buyer tries to add a review
	comment := "Should not be able to review"
	reviewPayload := models.ReviewDTO{ProductID: product.ID, Rating: 3, Comment: &comment}
	reviewBytes, _ := json.Marshal(reviewPayload)
	reviewReq, _ := http.NewRequest("POST", "/reviews/add", bytes.NewBuffer(reviewBytes))
	reviewReq.Header.Set("Content-Type", "application/json")
	reviewReq.Header.Set("Authorization", "Bearer "+nonBuyerToken)
	reviewRR := httptest.NewRecorder()
	testRouter.ServeHTTP(reviewRR, reviewReq)
	assert.NotEqual(t, http.StatusCreated, reviewRR.Code)
	assert.Contains(t, reviewRR.Body.String(), "Failed to add review")
}
