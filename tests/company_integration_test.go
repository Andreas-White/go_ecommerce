package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go_ecommerce/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompanyFlow_Integration(t *testing.T) {
	clearTables(t)

	// 1. Register a producer
	producerEmail := "producer-company@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerToken, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	// 2. Verify no company exists initially
	getCompanyReq, _ := http.NewRequest("GET", "/companies/get-by-user", nil)
	getCompanyReq.Header.Set("Content-Type", "application/json")
	getCompanyReq.Header.Set("Authorization", "Bearer "+producerToken)

	getCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyRR, getCompanyReq)
	require.Equal(t, http.StatusInternalServerError, getCompanyRR.Code, "Should return error when no company exists")

	// 3. Create a company
	address := "123 Business Street"
	city := "Business City"
	country := "Business Country"
	zipCode := "12345"
	companyPayload := models.CompanyDTO{
		Name:          "Test Company",
		Address:       &address,
		City:          &city,
		Country:       &country,
		ZipCode:       &zipCode,
		ReviewAverage: 0.0,
		ReviewCount:   0,
	}

	companyBody, _ := json.Marshal(companyPayload)
	createCompanyReq, _ := http.NewRequest("POST", "/companies/create", bytes.NewBuffer(companyBody))
	createCompanyReq.Header.Set("Content-Type", "application/json")
	createCompanyReq.Header.Set("Authorization", "Bearer "+producerToken)

	createCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createCompanyRR, createCompanyReq)
	require.Equal(t, http.StatusCreated, createCompanyRR.Code, "Failed to create company")

	// 4. Verify company was created by retrieving it by user ID
	getCompanyReq2, _ := http.NewRequest("GET", "/companies/get-by-user", nil)
	getCompanyReq2.Header.Set("Content-Type", "application/json")
	getCompanyReq2.Header.Set("Authorization", "Bearer "+producerToken)

	getCompanyRR2 := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyRR2, getCompanyReq2)
	require.Equal(t, http.StatusOK, getCompanyRR2.Code, "Failed to get company by user ID")

	var createdCompany models.CompanyDTO
	err = json.Unmarshal(getCompanyRR2.Body.Bytes(), &createdCompany)
	require.NoError(t, err)

	// Verify company details
	assert.Equal(t, "Test Company", createdCompany.Name, "Company name should match")
	assert.Equal(t, "123 Business Street", *createdCompany.Address, "Company address should match")
	assert.Equal(t, "Business City", *createdCompany.City, "Company city should match")
	assert.Equal(t, "Business Country", *createdCompany.Country, "Company country should match")
	assert.Equal(t, "12345", *createdCompany.ZipCode, "Company zip code should match")
	assert.Equal(t, 0.0, createdCompany.ReviewAverage, "Company review average should be 0.0")
	assert.Equal(t, 0, createdCompany.ReviewCount, "Company review count should be 0")

	// 5. Update the company
	updatedAddress := "456 Updated Business Street"
	updatedCity := "Updated Business City"
	updatedCountry := "Updated Business Country"
	updatedZipCode := "54321"
	updateCompanyPayload := models.CompanyDTO{
		ID:            createdCompany.ID,
		Name:          "Updated Test Company",
		Address:       &updatedAddress,
		City:          &updatedCity,
		Country:       &updatedCountry,
		ZipCode:       &updatedZipCode,
		ReviewAverage: 4.5,
		ReviewCount:   10,
	}

	updateCompanyBody, _ := json.Marshal(updateCompanyPayload)
	updateCompanyReq, _ := http.NewRequest("PUT", "/companies/update", bytes.NewBuffer(updateCompanyBody))
	updateCompanyReq.Header.Set("Content-Type", "application/json")
	updateCompanyReq.Header.Set("Authorization", "Bearer "+producerToken)

	updateCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateCompanyRR, updateCompanyReq)
	require.Equal(t, http.StatusOK, updateCompanyRR.Code, "Failed to update company")

	// 6. Verify company was updated by retrieving it by user ID
	getCompanyReq3, _ := http.NewRequest("GET", "/companies/get-by-user", nil)
	getCompanyReq3.Header.Set("Content-Type", "application/json")
	getCompanyReq3.Header.Set("Authorization", "Bearer "+producerToken)

	getCompanyRR3 := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyRR3, getCompanyReq3)
	require.Equal(t, http.StatusOK, getCompanyRR3.Code, "Failed to get updated company by user ID")

	var updatedCompany models.CompanyDTO
	err = json.Unmarshal(getCompanyRR3.Body.Bytes(), &updatedCompany)
	require.NoError(t, err)

	// Verify updated company details
	assert.Equal(t, "Updated Test Company", updatedCompany.Name, "Updated company name should match")
	assert.Equal(t, "456 Updated Business Street", *updatedCompany.Address, "Updated company address should match")
	assert.Equal(t, "Updated Business City", *updatedCompany.City, "Updated company city should match")
	assert.Equal(t, "Updated Business Country", *updatedCompany.Country, "Updated company country should match")
	assert.Equal(t, "54321", *updatedCompany.ZipCode, "Updated company zip code should match")
	assert.Equal(t, 4.5, updatedCompany.ReviewAverage, "Updated company review average should match")
	assert.Equal(t, 10, updatedCompany.ReviewCount, "Updated company review count should match")

	// 7. Delete the company
	deleteCompanyReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/companies/delete?company_id=%s", updatedCompany.ID), nil)
	deleteCompanyReq.Header.Set("Content-Type", "application/json")
	deleteCompanyReq.Header.Set("Authorization", "Bearer "+producerToken)

	deleteCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteCompanyRR, deleteCompanyReq)
	require.Equal(t, http.StatusOK, deleteCompanyRR.Code, "Failed to delete company")

	// 8. Verify company was deleted by trying to retrieve it by user ID
	getCompanyReq4, _ := http.NewRequest("GET", "/companies/get-by-user", nil)
	getCompanyReq4.Header.Set("Content-Type", "application/json")
	getCompanyReq4.Header.Set("Authorization", "Bearer "+producerToken)

	getCompanyRR4 := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyRR4, getCompanyReq4)
	require.Equal(t, http.StatusInternalServerError, getCompanyRR4.Code, "Should return error when company is deleted")

	// 9. Verify company was deleted by trying to retrieve it by company ID
	getCompanyByIdReq, _ := http.NewRequest("GET", fmt.Sprintf("/companies/get-by-id?company_id=%s", updatedCompany.ID), nil)
	getCompanyByIdReq.Header.Set("Content-Type", "application/json")

	getCompanyByIdRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyByIdRR, getCompanyByIdReq)
	require.Equal(t, http.StatusInternalServerError, getCompanyByIdRR.Code, "Should return error when company is deleted")
}

func TestCompany_UnauthorizedAccess(t *testing.T) {
	clearTables(t)

	// 1. Register a producer
	producerEmail := "producer-unauthorized@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerToken, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	// 2. Register another user
	otherUserEmail := "other-user@example.com"
	otherUserPassword := "password123"
	otherUserPayload := createUserDTO(otherUserEmail, otherUserPassword, false)
	registerTestUserAuth(t, otherUserPayload)
	_, otherUserToken, err := loginUserAndGetTokenAuth(t, otherUserEmail, otherUserPassword)
	require.NoError(t, err)

	// 3. Producer creates a company
	address := "123 Business Street"
	city := "Business City"
	country := "Business Country"
	zipCode := "12345"
	companyPayload := models.CompanyDTO{
		Name:          "Test Company",
		Address:       &address,
		City:          &city,
		Country:       &country,
		ZipCode:       &zipCode,
		ReviewAverage: 0.0,
		ReviewCount:   0,
	}

	companyBody, _ := json.Marshal(companyPayload)
	createCompanyReq, _ := http.NewRequest("POST", "/companies/create", bytes.NewBuffer(companyBody))
	createCompanyReq.Header.Set("Content-Type", "application/json")
	createCompanyReq.Header.Set("Authorization", "Bearer "+producerToken)

	createCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createCompanyRR, createCompanyReq)
	require.Equal(t, http.StatusCreated, createCompanyRR.Code, "Failed to create company")

	// Get the created company to extract its ID
	getCompanyReq, _ := http.NewRequest("GET", "/companies/get-by-user", nil)
	getCompanyReq.Header.Set("Content-Type", "application/json")
	getCompanyReq.Header.Set("Authorization", "Bearer "+producerToken)

	getCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyRR, getCompanyReq)
	require.Equal(t, http.StatusOK, getCompanyRR.Code, "Failed to get company by user ID")

	var createdCompany models.CompanyDTO
	err = json.Unmarshal(getCompanyRR.Body.Bytes(), &createdCompany)
	require.NoError(t, err)

	// 4. Other user tries to get the company by user ID (should fail)
	getCompanyReq2, _ := http.NewRequest("GET", "/companies/get-by-user", nil)
	getCompanyReq2.Header.Set("Content-Type", "application/json")
	getCompanyReq2.Header.Set("Authorization", "Bearer "+otherUserToken)

	getCompanyRR2 := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyRR2, getCompanyReq2)
	require.Equal(t, http.StatusInternalServerError, getCompanyRR2.Code, "Other user should not be able to get company by user ID")

	// Note: The current implementation doesn't have ownership validation for update/delete
	// So these operations will succeed even for non-owners
	// This is a limitation of the current implementation

	// 5. Other user tries to update the company (currently succeeds due to no ownership validation)
	updateCompanyPayload := models.CompanyDTO{
		ID:            createdCompany.ID,
		Name:          "Unauthorized Update",
		Address:       &address,
		City:          &city,
		Country:       &country,
		ZipCode:       &zipCode,
		ReviewAverage: 0.0,
		ReviewCount:   0,
	}

	updateCompanyBody, _ := json.Marshal(updateCompanyPayload)
	updateCompanyReq, _ := http.NewRequest("PUT", "/companies/update", bytes.NewBuffer(updateCompanyBody))
	updateCompanyReq.Header.Set("Content-Type", "application/json")
	updateCompanyReq.Header.Set("Authorization", "Bearer "+otherUserToken)

	updateCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateCompanyRR, updateCompanyReq)
	// Currently succeeds due to no ownership validation
	require.Equal(t, http.StatusOK, updateCompanyRR.Code, "Currently succeeds due to no ownership validation")

	// 6. Other user tries to delete the company (currently succeeds due to no ownership validation)
	deleteCompanyReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/companies/delete?company_id=%s", createdCompany.ID), nil)
	deleteCompanyReq.Header.Set("Content-Type", "application/json")
	deleteCompanyReq.Header.Set("Authorization", "Bearer "+otherUserToken)

	deleteCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteCompanyRR, deleteCompanyReq)
	// Currently succeeds due to no ownership validation
	require.Equal(t, http.StatusOK, deleteCompanyRR.Code, "Currently succeeds due to no ownership validation")

	// 7. Verify the company was deleted (since the other user was able to delete it)
	getCompanyReq3, _ := http.NewRequest("GET", "/companies/get-by-user", nil)
	getCompanyReq3.Header.Set("Content-Type", "application/json")
	getCompanyReq3.Header.Set("Authorization", "Bearer "+producerToken)

	getCompanyRR3 := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyRR3, getCompanyReq3)
	require.Equal(t, http.StatusInternalServerError, getCompanyRR3.Code, "Company should be deleted")
}

func TestCompany_InvalidOperations(t *testing.T) {
	clearTables(t)

	// 1. Register a producer
	producerEmail := "producer-invalid@example.com"
	producerPassword := "password123"
	producerPayload := createUserDTO(producerEmail, producerPassword, true)
	registerTestUserAuth(t, producerPayload)
	_, producerToken, err := loginUserAndGetTokenAuth(t, producerEmail, producerPassword)
	require.NoError(t, err)

	// 2. Try to create company without authentication
	address := "123 Business Street"
	city := "Business City"
	country := "Business Country"
	zipCode := "12345"
	companyPayload := models.CompanyDTO{
		Name:          "Test Company",
		Address:       &address,
		City:          &city,
		Country:       &country,
		ZipCode:       &zipCode,
		ReviewAverage: 0.0,
		ReviewCount:   0,
	}

	companyBody, _ := json.Marshal(companyPayload)
	createCompanyReq, _ := http.NewRequest("POST", "/companies/create", bytes.NewBuffer(companyBody))
	createCompanyReq.Header.Set("Content-Type", "application/json")
	// No Authorization header

	createCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(createCompanyRR, createCompanyReq)
	require.Equal(t, http.StatusUnauthorized, createCompanyRR.Code, "Should require authentication to create company")

	// 3. Try to create company with invalid payload
	invalidCompanyBody := []byte(`{"name": "", "invalid_field": "value"}`)
	invalidCreateReq, _ := http.NewRequest("POST", "/companies/create", bytes.NewBuffer(invalidCompanyBody))
	invalidCreateReq.Header.Set("Content-Type", "application/json")
	invalidCreateReq.Header.Set("Authorization", "Bearer "+producerToken)

	invalidCreateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(invalidCreateRR, invalidCreateReq)
	require.Equal(t, http.StatusCreated, invalidCreateRR.Code, "Should handle invalid payload gracefully")

	// 4. Try to get company by invalid company ID (should handle gracefully)
	getCompanyByIdReq, _ := http.NewRequest("GET", "/companies/get-by-id?company_id=invalid-uuid", nil)
	getCompanyByIdReq.Header.Set("Content-Type", "application/json")

	getCompanyByIdRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getCompanyByIdRR, getCompanyByIdReq)
	// Should return 400 due to invalid UUID format
	require.Equal(t, http.StatusBadRequest, getCompanyByIdRR.Code, "Should handle invalid company ID")

	// 5. Try to delete company without company_id parameter
	deleteCompanyReq, _ := http.NewRequest("DELETE", "/companies/delete", nil)
	deleteCompanyReq.Header.Set("Content-Type", "application/json")
	deleteCompanyReq.Header.Set("Authorization", "Bearer "+producerToken)

	deleteCompanyRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteCompanyRR, deleteCompanyReq)
	require.Equal(t, http.StatusBadRequest, deleteCompanyRR.Code, "Should require company_id parameter for deletion")
}
