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
	producerPayload := createUserDTO(producerEmail, producerPassword, true)

	registerTestUserAuth(t, producerPayload)

	// 2. Login to get token
	_, authData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)
	require.NotEmpty(t, authData.JWTToken)

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
	addAuthHeaders(createReq, authData)

	createRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createRR, createReq)

	require.Equal(t, http.StatusCreated, createRR.Code, "Failed to create product. Body: "+createRR.Body.String())

	var createdProduct models.Product
	err = json.Unmarshal(createRR.Body.Bytes(), &createdProduct)
	require.NoError(t, err, "Failed to unmarshal product response")
	assert.Equal(t, productPayload.Name, createdProduct.Name)
	assert.Equal(t, productPayload.Description, createdProduct.Description)
	assert.Equal(t, productPayload.Price, createdProduct.Price)
	assert.Equal(t, productPayload.Stock, createdProduct.Stock)
	assert.NotEmpty(t, createdProduct.ID)
}

func TestGetProduct_ByID_And_All(t *testing.T) {
	clearTables(t)

	// 1. Register user and create a product
	producerEmail := "producer2@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, authData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
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
	addAuthHeaders(createReq, authData)
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
	regPayload := createUserDTO("producer3@example.com", "password123", true)
	registerTestUserAuth(t, regPayload)
	_, authData, err := loginUserAndGetTokenAuth(t, "producer3@example.com", "password123")
	require.NoError(t, err)

	productsToCreate := []models.Product{
		{Name: "Apple iPhone", Description: "A premium smartphone", Price: 1200.00, Category: "Phones"},
		{Name: "Banana Laptop", Description: "A yellow laptop", Price: 800.00, Category: "Computers"},
	}

	for _, p := range productsToCreate {
		productBytes, _ := json.Marshal(p)
		createReq, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(productBytes))
		createReq.Header.Set("Content-Type", "application/json")
		addAuthHeaders(createReq, authData)
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
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, authData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
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
	addAuthHeaders(createReq, authData)
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
	addAuthHeaders(updateReq, authData)

	updateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateRR, updateReq)

	require.Equal(t, http.StatusOK, updateRR.Code, "Failed to update product. Body: "+updateRR.Body.String())

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
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, authData, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
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
	addAuthHeaders(createReq, authData)
	createRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createRR, createReq)
	require.Equal(t, http.StatusCreated, createRR.Code)

	// Retrieve the created product from DB to get its ID
	var createdProduct models.Product
	err = testDB.QueryRow("SELECT id FROM products WHERE name = $1", productPayload.Name).Scan(&createdProduct.ID)
	require.NoError(t, err, "Failed to query created product from DB")

	// 2. Delete the product
	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/products/delete?id=%s", createdProduct.ID), nil)
	addAuthHeaders(deleteReq, authData) // Assuming deletes are protected

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

func TestGetProductsByUserID_Ownership(t *testing.T) {
	clearTables(t)

	// 1. Register two producers
	producer1Email := "producer1-ownership@example.com"
	producer1Password := "password123"
	producer1Payload := createUserDTO(producer1Email, producer1Password, true)
	_, producer1AuthData := registerTestUserAuth(t, producer1Payload)

	producer2Email := "producer2-ownership@example.com"
	producer2Password := "password123"
	producer2Payload := createUserDTO(producer2Email, producer2Password, true)
	_, producer2AuthData := registerTestUserAuth(t, producer2Payload)

	// 2. Producer1 creates a product
	product1 := models.Product{
		Name:     "Producer1's Product",
		Price:    50.00,
		Category: "Electronics",
		Stock:    10,
	}
	product1Bytes, _ := json.Marshal(product1)
	create1Req, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(product1Bytes))
	create1Req.Header.Set("Content-Type", "application/json")
	addAuthHeaders(create1Req, producer1AuthData)
	create1RR := httptest.NewRecorder()
	testRouter.ServeHTTP(create1RR, create1Req)
	require.Equal(t, http.StatusCreated, create1RR.Code)

	// 3. Producer2 creates a product
	product2 := models.Product{
		Name:     "Producer2's Product",
		Price:    75.00,
		Category: "Books",
		Stock:    5,
	}
	product2Bytes, _ := json.Marshal(product2)
	create2Req, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(product2Bytes))
	create2Req.Header.Set("Content-Type", "application/json")
	addAuthHeaders(create2Req, producer2AuthData)
	create2RR := httptest.NewRecorder()
	testRouter.ServeHTTP(create2RR, create2Req)
	require.Equal(t, http.StatusCreated, create2RR.Code)

	// 4. Producer1 gets their products (should only see their own)
	getMyProductsReq, _ := http.NewRequest("GET", "/products/my-products", nil)
	addAuthHeaders(getMyProductsReq, producer1AuthData)
	getMyProductsRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getMyProductsRR, getMyProductsReq)
	require.Equal(t, http.StatusOK, getMyProductsRR.Code)

	var producer1Products []models.Product
	err := json.Unmarshal(getMyProductsRR.Body.Bytes(), &producer1Products)
	require.NoError(t, err)
	assert.Len(t, producer1Products, 1, "Producer1 should only see their own product")
	assert.Equal(t, "Producer1's Product", producer1Products[0].Name)

	// 5. Producer2 gets their products (should only see their own)
	getMyProducts2Req, _ := http.NewRequest("GET", "/products/my-products", nil)
	addAuthHeaders(getMyProducts2Req, producer2AuthData)
	getMyProducts2RR := httptest.NewRecorder()
	testRouter.ServeHTTP(getMyProducts2RR, getMyProducts2Req)
	require.Equal(t, http.StatusOK, getMyProducts2RR.Code)

	var producer2Products []models.Product
	err = json.Unmarshal(getMyProducts2RR.Body.Bytes(), &producer2Products)
	require.NoError(t, err)
	assert.Len(t, producer2Products, 1, "Producer2 should only see their own product")
	assert.Equal(t, "Producer2's Product", producer2Products[0].Name)
}

func TestUpdateDeleteProduct_OwnershipCheck(t *testing.T) {
	clearTables(t)

	// 1. Register two producers
	producer1Email := "producer1-ownership-check@example.com"
	producer1Password := "password123"
	producer1Payload := createUserDTO(producer1Email, producer1Password, true)
	_, producer1AuthData := registerTestUserAuth(t, producer1Payload)

	producer2Email := "producer2-ownership-check@example.com"
	producer2Password := "password123"
	producer2Payload := createUserDTO(producer2Email, producer2Password, true)
	_, producer2AuthData := registerTestUserAuth(t, producer2Payload)

	// 2. Producer1 creates a product
	product1 := models.Product{
		Name:     "Producer1's Protected Product",
		Price:    100.00,
		Category: "Electronics",
		Stock:    15,
	}
	product1Bytes, _ := json.Marshal(product1)
	create1Req, _ := http.NewRequest("POST", "/products/create", bytes.NewBuffer(product1Bytes))
	create1Req.Header.Set("Content-Type", "application/json")
	addAuthHeaders(create1Req, producer1AuthData)
	create1RR := httptest.NewRecorder()
	testRouter.ServeHTTP(create1RR, create1Req)
	require.Equal(t, http.StatusCreated, create1RR.Code)

	var createdProduct models.Product
	err := json.Unmarshal(create1RR.Body.Bytes(), &createdProduct)
	require.NoError(t, err)

	// 3. Producer2 tries to update Producer1's product (should fail)
	updatePayload := models.Product{
		Name:     "Hacked Product",
		Price:    999.99,
		Category: "Hacked",
		Stock:    1,
	}
	updatePayload.ID = createdProduct.ID
	updateBytes, _ := json.Marshal(updatePayload)
	updateReq, _ := http.NewRequest("PUT", fmt.Sprintf("/products/update?id=%s", createdProduct.ID), bytes.NewBuffer(updateBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(updateReq, producer2AuthData)
	updateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateRR, updateReq)
	assert.Equal(t, http.StatusForbidden, updateRR.Code, "Producer2 should not be able to update Producer1's product")

	// 4. Producer2 tries to delete Producer1's product (should fail)
	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/products/delete?id=%s", createdProduct.ID), nil)
	addAuthHeaders(deleteReq, producer2AuthData)
	deleteRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteRR, deleteReq)
	assert.Equal(t, http.StatusForbidden, deleteRR.Code, "Producer2 should not be able to delete Producer1's product")

	// 5. Verify the product still exists and is unchanged
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/product?id=%s", createdProduct.ID), nil)
	addAuthHeaders(getReq, producer1AuthData)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	require.Equal(t, http.StatusOK, getRR.Code)

	var retrievedProduct models.Product
	err = json.Unmarshal(getRR.Body.Bytes(), &retrievedProduct)
	require.NoError(t, err)
	assert.Equal(t, "Producer1's Protected Product", retrievedProduct.Name, "Product should remain unchanged")
	assert.Equal(t, 100.00, retrievedProduct.Price, "Product price should remain unchanged")

	// 6. Producer1 should be able to update their own product
	validUpdatePayload := models.Product{
		Name:     "Producer1's Updated Product",
		Price:    150.00,
		Category: "Updated Electronics",
		Stock:    20,
	}
	validUpdatePayload.ID = createdProduct.ID
	validUpdateBytes, _ := json.Marshal(validUpdatePayload)
	validUpdateReq, _ := http.NewRequest("PUT", fmt.Sprintf("/products/update?id=%s", createdProduct.ID), bytes.NewBuffer(validUpdateBytes))
	validUpdateReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(validUpdateReq, producer1AuthData)
	validUpdateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(validUpdateRR, validUpdateReq)
	assert.Equal(t, http.StatusOK, validUpdateRR.Code, "Producer1 should be able to update their own product")
}
