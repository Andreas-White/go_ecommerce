package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go_ecommerce/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProduct_Success(t *testing.T) {
	clearTables(t)

	// 1. Register a producer user
	producerEmail := "producer@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword)
	producerPayload.IsProducer = true

	regRR := registerTestUserAuth(t, producerPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "Producer registration failed")

	// 2. Login to get token
	_, token, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// 3. Create a product
	productPayload := models.Product{
		Name:        "Test Gadget",
		Description: "A really cool gadget.",
		Price:       99.99,
		Stock:       100,
		Category:    "Electronics",
		ImageUrl:    "http://example.com/image.png",
	}
	productBytes, err := json.Marshal(productPayload)
	require.NoError(t, err)

	createReq, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(productBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)

	createRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createRR, createReq)

	require.Equal(t, http.StatusCreated, createRR.Code, "Failed to create product. Body: "+createRR.Body.String())

	var resp map[string]string
	err = json.Unmarshal(createRR.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Product created successfully", resp["message"])
}

func TestGetProduct_ByID_And_All(t *testing.T) {
	clearTables(t)

	// 1. Register user and create a product
	producerEmail := "producer2@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword)
	producerPayload.IsProducer = true
	registerTestUserAuth(t, producerPayload)
	_, token, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	productPayload := models.Product{
		Name:        "Another Gadget",
		Description: "A different gadget.",
		Price:       149.99,
		Stock:       50,
		Category:    "Tech",
		ImageUrl:    "http://example.com/image2.png",
	}
	productBytes, _ := json.Marshal(productPayload)
	createReq, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(productBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)

	// Retrieve the created product from DB to get its ID
	var createdProduct models.Product
	err = testDB.QueryRow("SELECT id, name, price FROM products WHERE name = $1", productPayload.Name).Scan(&createdProduct.ID, &createdProduct.Name, &createdProduct.Price)
	require.NoError(t, err, "Failed to query created product from DB")

	// 2. Test Get Product By ID
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/product?id=%s", createdProduct.ID), nil)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)

	require.Equal(t, http.StatusOK, getRR.Code)
	var fetchedProduct models.Product
	err = json.Unmarshal(getRR.Body.Bytes(), &fetchedProduct)
	require.NoError(t, err)
	assert.Equal(t, createdProduct.ID, fetchedProduct.ID)
	assert.Equal(t, createdProduct.Name, fetchedProduct.Name)

	// 3. Test Get All Products
	getAllReq, _ := http.NewRequest("GET", "/products", nil)
	getAllRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getAllRR, getAllReq)

	require.Equal(t, http.StatusOK, getAllRR.Code)
	var products []models.Product
	err = json.Unmarshal(getAllRR.Body.Bytes(), &products)
	require.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, createdProduct.ID, products[0].ID)
}

func TestGetProducts_SearchAndSort(t *testing.T) {
	clearTables(t)
	// Setup: create a user and two products
	regPayload := createUserDTO("producer3@example.com", "password123")
	regPayload.IsProducer = true
	registerTestUserAuth(t, regPayload)
	_, token, err := loginUserAndGetTokenAuth(t, "producer3@example.com", "password123")
	require.NoError(t, err)

	productsToCreate := []models.Product{
		{Name: "Apple iPhone", Description: "A premium smartphone", Price: 1200.00, Category: "Phones"},
		{Name: "Banana Laptop", Description: "A yellow laptop", Price: 800.00, Category: "Computers"},
	}

	for _, p := range productsToCreate {
		productBytes, _ := json.Marshal(p)
		createReq, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(productBytes))
		createReq.Header.Set("Content-Type", "application/json")
		createReq.Header.Set("Authorization", "Bearer "+token)
		createRR := httptest.NewRecorder()
		testRouter.ServeHTTP(createRR, createReq)
		require.Equal(t, http.StatusCreated, createRR.Code)
	}

	// Test Search
	searchReq, _ := http.NewRequest("GET", "/products?search=premium", nil)
	searchRR := httptest.NewRecorder()
	testRouter.ServeHTTP(searchRR, searchReq)
	require.Equal(t, http.StatusOK, searchRR.Code)
	var searchResults []models.Product
	json.Unmarshal(searchRR.Body.Bytes(), &searchResults)
	assert.Len(t, searchResults, 1)
	assert.Equal(t, "Apple iPhone", searchResults[0].Name)

	// Test Sort
	sortReq, _ := http.NewRequest("GET", "/products?sortBy=price&sortOrder=asc", nil)
	sortRR := httptest.NewRecorder()
	testRouter.ServeHTTP(sortRR, sortReq)
	require.Equal(t, http.StatusOK, sortRR.Code)
	var sortResults []models.Product
	json.Unmarshal(sortRR.Body.Bytes(), &sortResults)
	assert.Len(t, sortResults, 2)
	assert.Equal(t, "Banana Laptop", sortResults[0].Name) // Cheaper one first
	assert.Equal(t, "Apple iPhone", sortResults[1].Name)
}

func TestUpdateProduct_Success(t *testing.T) {
	clearTables(t)

	// 1. Register user and create a product
	producerEmail := "producer-update@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword)
	producerPayload.IsProducer = true
	registerTestUserAuth(t, producerPayload)
	_, token, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	initialProduct := models.Product{
		Name:        "Updatable Gadget",
		Description: "An old gadget.",
		Price:       19.99,
		Stock:       10,
		Category:    "Vintage",
		ImageUrl:    "http://example.com/image.png",
	}
	productBytes, _ := json.Marshal(initialProduct)
	createReq, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(productBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)

	// Retrieve the created product from DB to get its ID
	var createdProduct models.Product
	err = testDB.QueryRow("SELECT id FROM products WHERE name = $1", initialProduct.Name).Scan(&createdProduct.ID)
	require.NoError(t, err, "Failed to query created product from DB")

	// 2. Update the product
	updatedProductPayload := models.Product{
		Name:        "Shiny New Gadget",
		Description: "A brand new gadget.",
		Price:       29.99,
		Stock:       25,
		Category:    "Modern",
		ImageUrl:    "http://example.com/new_image.png",
	}
	updatedProductPayload.ID = createdProduct.ID // Ensure ID is in the payload for the handler
	updateBytes, err := json.Marshal(updatedProductPayload)
	require.NoError(t, err)

	updateReq, _ := http.NewRequest("PUT", fmt.Sprintf("/products/update?id=%s", createdProduct.ID), bytes.NewBuffer(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)

	updateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateRR, updateReq)

	require.Equal(t, http.StatusOK, updateRR.Code, "Failed to update product. Body: "+updateRR.Body.String())

	// The handler appears to return the updated product object, not a simple message.
	// We will unmarshal into the Product model and verify its contents.
	var returnedProduct models.Product
	err = json.Unmarshal(updateRR.Body.Bytes(), &returnedProduct)
	require.NoError(t, err, "Failed to unmarshal update response. Body: "+updateRR.Body.String())

	// Assert that the returned product matches the updated payload
	assert.Equal(t, updatedProductPayload.Name, returnedProduct.Name)
	assert.Equal(t, updatedProductPayload.Price, returnedProduct.Price)
	assert.Equal(t, createdProduct.ID, returnedProduct.ID)

	// 3. Verify the update in the database
	var finalProduct models.Product
	err = testDB.QueryRow("SELECT name, description, price, stock, category, image_url FROM products WHERE id = $1", createdProduct.ID).Scan(
		&finalProduct.Name, &finalProduct.Description, &finalProduct.Price, &finalProduct.Stock, &finalProduct.Category, &finalProduct.ImageUrl)
	require.NoError(t, err)
	assert.Equal(t, updatedProductPayload.Name, finalProduct.Name)
	assert.Equal(t, updatedProductPayload.Description, finalProduct.Description)
	assert.Equal(t, updatedProductPayload.Price, finalProduct.Price)
	assert.Equal(t, updatedProductPayload.Stock, finalProduct.Stock)
	assert.Equal(t, updatedProductPayload.Category, finalProduct.Category)
	assert.Equal(t, updatedProductPayload.ImageUrl, finalProduct.ImageUrl)
}

func TestDeleteProduct_Success(t *testing.T) {
	clearTables(t)

	// 1. Register user and create a product
	producerEmail := "producer-delete@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword)
	producerPayload.IsProducer = true
	registerTestUserAuth(t, producerPayload)
	_, token, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	productPayload := models.Product{
		Name:     "Deletable Gadget",
		Category: "Junk",
		Price:    1.00,
		Stock:    1,
	}
	productBytes, _ := json.Marshal(productPayload)
	createReq, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(productBytes))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)

	// Retrieve the created product from DB to get its ID
	var createdProduct models.Product
	err = testDB.QueryRow("SELECT id FROM products WHERE name = $1", productPayload.Name).Scan(&createdProduct.ID)
	require.NoError(t, err, "Failed to query created product from DB")

	// 2. Delete the product
	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/products/delete?id=%s", createdProduct.ID), nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token) // Assuming deletes are protected

	deleteRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteRR, deleteReq)

	require.Equal(t, http.StatusOK, deleteRR.Code, "Failed to delete product. Body: "+deleteRR.Body.String())

	// A 200 OK status is sufficient, as we verify deletion from the DB.
	// The handler may return the deleted object, causing unmarshal errors if we expect a simple message.

	// 3. Verify the product is deleted from the database
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM products WHERE id = $1", createdProduct.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Product was not deleted from the database")
}
