package tests

import (
	"context"
	"encoding/json"
	"go_ecommerce/pkg/models"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearTables(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE TABLE users, products RESTART IDENTITY CASCADE")
	require.NoError(t, err, "Failed to truncate tables")
}

func TestUserRegistration_Success(t *testing.T) {
	clearTables(t)

	registrationPayload := createUserDTO("testuser@example.com", "password123", false)

	regRR, authData := registerTestUserAuth(t, registrationPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "User registration failed")

	var responseBody map[string]string
	err := json.Unmarshal(regRR.Body.Bytes(), &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "User created successfully", responseBody["message"])

	// Verify that JWT and CSRF tokens were set
	assert.NotEmpty(t, authData.JWTToken, "JWT token should be set in cookie")
	assert.NotEmpty(t, authData.CSRFToken, "CSRF token should be set in cookie")

	var userID, email string
	err = testDB.QueryRow("SELECT id, email FROM users WHERE email = $1", registrationPayload.Email).Scan(&userID, &email)
	require.NoError(t, err, "User should exist in database after registration")
	assert.Equal(t, registrationPayload.Email, email)

	var cartCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM carts WHERE user_id = $1", userID).Scan(&cartCount)
	require.NoError(t, err, "Failed to query cart count for new user")
	assert.Equal(t, 1, cartCount, "A cart should be created for the new user")
}

func TestUserLogin_Success(t *testing.T) {
	clearTables(t)

	registrationPayload := createUserDTO("loginuser@example.com", "securepassword123", false)

	regRR, _ := registerTestUserAuth(t, registrationPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "User registration failed")

	loginRR, authData, err := loginUserAndGetTokenAuth(t, registrationPayload.Email, registrationPayload.Password)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, loginRR.Code, "Expected status code 200 OK for login")
	var loginResponseBody map[string]string
	err = json.Unmarshal(loginRR.Body.Bytes(), &loginResponseBody)
	require.NoError(t, err)
	assert.NotEmpty(t, authData.JWTToken, "JWT token should not be empty on successful login")
	assert.NotEmpty(t, authData.CSRFToken, "CSRF token should not be empty on successful login")
}

func TestUserLogin_Failure(t *testing.T) {
	clearTables(t)

	rr, _, err := loginUserAndGetTokenAuth(t, "nonexistent@example.com", "anypassword")
	require.ErrorContains(t, err, "login request failed with status 401")
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected 401 for non-existent user")

	registrationPayload := createUserDTO("wrongpass@example.com", "correctpassword", false)
	regRR, _ := registerTestUserAuth(t, registrationPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration for wrong password test failed")
	rrWrong, _, err := loginUserAndGetTokenAuth(t, registrationPayload.Email, "incorrectpassword")
	require.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rrWrong.Code, "Expected 401 for wrong password")
}

func TestGetUserByID_Success(t *testing.T) {
	clearTables(t)
	var registeredUser models.User

	regPayload := createUserDTO("getuser@example.com", "password123", false)
	regRR, _ := registerTestUserAuth(t, regPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	err := testDB.QueryRowContext(context.Background(), "SELECT id, first_name, last_name, email, phone, is_producer, address, city, country, zip_code, created_at FROM users WHERE email = $1", "getuser@example.com").
		Scan(&registeredUser.ID, &registeredUser.FirstName, &registeredUser.LastName, &registeredUser.Email, &registeredUser.Phone, &registeredUser.IsProducer, &registeredUser.Address, &registeredUser.City, &registeredUser.Country, &registeredUser.ZipCode, &registeredUser.CreatedAt)
	require.NoError(t, err, "Failed to retrieve registered user from DB")

	loginRR, authData, err := loginUserAndGetTokenAuth(t, registeredUser.Email, "password123")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	require.NotEmpty(t, authData.JWTToken, "JWT token not found")

	getRR := getUserById(authData)

	assert.Equal(t, http.StatusOK, getRR.Code, "Expected 200 OK for GetUserByID")
	var fetchedUser models.UserDTO
	err = json.Unmarshal(getRR.Body.Bytes(), &fetchedUser)
	require.NoError(t, err)
	assert.Equal(t, registeredUser.FirstName, fetchedUser.FirstName)
	assert.Equal(t, registeredUser.Email, fetchedUser.Email)
}

func TestGetUserByID_Unauthorized(t *testing.T) {
	clearTables(t)
	getRRUnauth := getUserById(nil)
	assert.Equal(t, http.StatusUnauthorized, getRRUnauth.Code, "Expected 401 Unauthorized without token")

	// Test with invalid auth data
	invalidAuthData := &TestAuthData{JWTToken: "invalidtoken123"}
	getRRInvalidToken := getUserById(invalidAuthData)
	assert.Equal(t, http.StatusUnauthorized, getRRInvalidToken.Code, "Expected 401 Unauthorized with invalid token")
}

func TestUpdateUser_Success(t *testing.T) {
	clearTables(t)
	registrationPayload := createUserDTO("updateuser@example.com", "password123", false)
	regRR, _ := registerTestUserAuth(t, registrationPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	loginRR, authData, err := loginUserAndGetTokenAuth(t, registrationPayload.Email, registrationPayload.Password)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	require.NotEmpty(t, authData.JWTToken, "JWT token not found")

	updatePayload := models.UserDTO{
		FirstName: "UpdatedFirst", City: "NewCity",
	}

	updateRR := updateUser(authData, updatePayload)

	assert.Equal(t, http.StatusOK, updateRR.Code, "Expected 200 OK for UpdateUser")
	var updateRespBody map[string]string
	_ = json.Unmarshal(updateRR.Body.Bytes(), &updateRespBody)
	assert.Equal(t, "User updated successfully", updateRespBody["message"])

	var dbFirstName, dbCity string
	err = testDB.QueryRow("SELECT first_name, city FROM users WHERE email = $1", registrationPayload.Email).Scan(&dbFirstName, &dbCity)
	require.NoError(t, err, "Failed to query updated user from DB")
	assert.Equal(t, updatePayload.FirstName, dbFirstName)
	assert.Equal(t, updatePayload.City, dbCity)
}

func TestUpdateUser_Unauthorized(t *testing.T) {
	clearTables(t)
	updatePayload := models.UserDTO{FirstName: "AttemptUpdate"}

	updateRRUnauth := updateUser(nil, updatePayload)
	assert.Equal(t, http.StatusUnauthorized, updateRRUnauth.Code, "Expected 401 Unauthorized without token")

	// Test with invalid auth data
	invalidAuthData := &TestAuthData{JWTToken: "invalidtoken123"}
	updateRRInvalidToken := updateUser(invalidAuthData, updatePayload)
	assert.Equal(t, http.StatusUnauthorized, updateRRInvalidToken.Code, "Expected 401 Unauthorized with invalid token")
}

func TestDeleteUser_Success(t *testing.T) {
	clearTables(t)
	// Register a user for deletion test
	registrationPayload := createUserDTO("deleteuser@example.com", "password123", false)
	regRR, _ := registerTestUserAuth(t, registrationPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	// Login to get a token
	loginRR, authData, err := loginUserAndGetTokenAuth(t, registrationPayload.Email, registrationPayload.Password)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	require.NotEmpty(t, authData.JWTToken, "JWT token not found")

	// Query user to get user Id
	var userId string
	err = testDB.QueryRow("SELECT id FROM users WHERE email = $1", registrationPayload.Email).Scan(&userId)
	require.NoError(t, err, "Failed to query user from DB after delete")

	// Delete the user
	deleteRR := deleteUser(authData)

	assert.Equal(t, http.StatusOK, deleteRR.Code, "Expected 200 OK for DeleteUser")
	var deleteRespBody map[string]string
	_ = json.Unmarshal(deleteRR.Body.Bytes(), &deleteRespBody)
	assert.Equal(t, "User deleted successfully", deleteRespBody["message"])

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", userId).Scan(&count)
	require.NoError(t, err, "Failed to query user count from DB after delete")
	assert.Equal(t, 0, count, "User should not exist in DB after deletion")

	err = testDB.QueryRow("SELECT COUNT(*) FROM auths WHERE user_id = $1", userId).Scan(&count)
	require.NoError(t, err, "Failed to query auth count from DB after delete")
	assert.Equal(t, 0, count, "Auths should not exist in DB after deletion")

	err = testDB.QueryRow("SELECT COUNT(*) FROM carts WHERE user_id = $1", userId).Scan(&count)
	require.NoError(t, err, "Failed to query cart count from DB after user delete")
	assert.Equal(t, 0, count, "Cart should not exist in DB after user deletion")
}

func TestDeleteUser_Unauthorized(t *testing.T) {
	clearTables(t)
	deleteRRUnauth := deleteUser(nil)
	assert.Equal(t, http.StatusUnauthorized, deleteRRUnauth.Code, "Expected 401 Unauthorized without token")

	// Test with invalid auth data
	invalidAuthData := &TestAuthData{JWTToken: "invalidtoken123"}
	deleteRRInvalidToken := deleteUser(invalidAuthData)
	assert.Equal(t, http.StatusUnauthorized, deleteRRInvalidToken.Code, "Expected 401 Unauthorized with invalid token")
}

func TestGetUserByName(t *testing.T) {
	clearTables(t)
	registrationPayload := createUserDTO("byname@example.com", "password123", false)

	regRR, _ := registerTestUserAuth(t, registrationPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	loginRR, authData, err := loginUserAndGetTokenAuth(t, registrationPayload.Email, registrationPayload.Password)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	require.NotEmpty(t, authData.JWTToken, "JWT token not found")

	getRR := getUserByName(authData)

	assert.Equal(t, http.StatusOK, getRR.Code, "Expected 200 OK for GetUserByName")
	var user models.User
	err = json.Unmarshal(getRR.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, "Login", user.FirstName)
	assert.Equal(t, "User", user.LastName)
}

func TestGetUserByEmail(t *testing.T) {
	clearTables(t)
	registrationPayload := createUserDTO("byemail@example.com", "password123", false)

	regRR, _ := registerTestUserAuth(t, registrationPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	loginRR, authData, err := loginUserAndGetTokenAuth(t, registrationPayload.Email, registrationPayload.Password)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	require.NoError(t, err)
	require.NotEmpty(t, authData.JWTToken, "JWT token not found")

	getRR := getUserByEmail(authData)

	assert.Equal(t, http.StatusOK, getRR.Code, "Expected 200 OK for GetUserByEmail")
	var user models.User
	err = json.Unmarshal(getRR.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, registrationPayload.Email, user.Email)
	assert.Equal(t, "Login", user.FirstName)
}

func TestDuplicateRegistration(t *testing.T) {
	clearTables(t)
	regPayload := createUserDTO("duplicate@example.com", "password123", false)
	regRR, _ := registerTestUserAuth(t, regPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "First registration failed")

	// Attempt second registration with same email
	secondRegRR, _ := registerTestUserAuth(t, regPayload)
	assert.Equal(t, http.StatusInternalServerError, secondRegRR.Code, "Expected 500 for duplicate registration")
	var errorResp map[string]string
	err := json.Unmarshal(secondRegRR.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "Failed to create user")
}

func TestInvalidPasswordRegistration(t *testing.T) {
	clearTables(t)
	regPayload := createUserDTO("shortpass@example.com", "123", false)
	regRR, _ := registerTestUserAuth(t, regPayload)
	assert.Equal(t, http.StatusBadRequest, regRR.Code, "Expected 400 for invalid password")
	var errorResp map[string]string
	err := json.Unmarshal(regRR.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "Password must be at least 8 characters long")
}

func TestInvalidEmailRegistration(t *testing.T) {
	clearTables(t)
	regPayload := createUserDTO("invalid-email-format", "password123", false)
	regRR, _ := registerTestUserAuth(t, regPayload)
	assert.Equal(t, http.StatusBadRequest, regRR.Code, "Expected 400 for invalid email format")
	var errorResp map[string]string
	err := json.Unmarshal(regRR.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "Invalid email format")
}

func TestDeleteNonExistentUser_Returns404(t *testing.T) {
	clearTables(t)

	// 1. Register a user
	regPayload := createUserDTO("deletecheck@example.com", "password123", false)
	regRR, _ := registerTestUserAuth(t, regPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "User registration failed")

	// Get user ID after registration (needed for direct deletion if not inferable from login)
	var registeredUser models.User
	err := testDB.QueryRowContext(context.Background(), "SELECT id, email FROM users WHERE email = $1", regPayload.Email).Scan(&registeredUser.ID, &registeredUser.Email)
	require.NoError(t, err, "Failed to query registered user ID")
	require.NotEmpty(t, registeredUser.ID, "Registered user ID is empty")

	// 2. Log in to get a token
	loginRR, authData, err := loginUserAndGetTokenAuth(t, regPayload.Email, regPayload.Password)
	require.Equal(t, http.StatusOK, loginRR.Code, "User login failed")
	require.NoError(t, err)
	require.NotEmpty(t, authData.JWTToken, "JWT token not found in login response")

	// 3. Delete the user (first time - should succeed)
	deleteRR := deleteUser(authData)
	assert.Equal(t, http.StatusOK, deleteRR.Code, "Expected first delete to be successful")

	// 4. Attempt to delete the same user again (should fail with 500)
	secondDeleteRR := deleteUser(authData)

	// Assert that the second delete attempt returns 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, secondDeleteRR.Code, "Expected second delete attempt to return 500 Internal Server Error")
	var errorResp map[string]string
	err = json.Unmarshal(secondDeleteRR.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "Failed to delete user", "Error message mismatch for 500")
}

func TestUserUpdate_GeneratesNewTokenAndUpdatesDetails(t *testing.T) {
	clearTables(t)

	// 1. Register a user
	regPayload := createUserDTO("updatejwt@example.com", "password123", false)

	regRR, _ := registerTestUserAuth(t, regPayload)
	require.Equal(t, http.StatusCreated, regRR.Code, "User registration failed")

	// Get registered user's ID
	var registeredUser models.User
	err := testDB.QueryRowContext(context.Background(), "SELECT id FROM users WHERE email = $1", regPayload.Email).Scan(&registeredUser.ID)
	require.NoError(t, err, "Failed to query registered user ID")

	// 2. Log in to get an initial JWT
	loginRR, authData, err := loginUserAndGetTokenAuth(t, regPayload.Email, regPayload.Password)
	require.Equal(t, http.StatusOK, loginRR.Code, "User login failed")
	require.NoError(t, err)
	require.NotEmpty(t, authData.JWTToken, "Initial JWT token not found")

	// 3. Prepare an update request
	updatedFirstName := "UpdatedName"
	updatedLastName := "UpdatedLast"
	updatePayload := models.UserDTO{
		FirstName: updatedFirstName,
		LastName:  updatedLastName,
	}

	// 4. Call the update user endpoint
	updateRR := updateUser(authData, updatePayload)

	// 5. Assert response from update
	require.Equal(t, http.StatusOK, updateRR.Code, "User update failed. Body: "+updateRR.Body.String())
	var updateResp struct {
		Message string      `json:"message"`
		User    models.User `json:"user"`
		Token   string      `json:"token"`
	}
	err = json.Unmarshal(updateRR.Body.Bytes(), &updateResp)
	require.NoError(t, err, "Failed to unmarshal update response. Body: "+updateRR.Body.String())

	assert.Equal(t, "User updated successfully", updateResp.Message)
	assert.Equal(t, updatedFirstName, updateResp.User.FirstName, "First name not updated in response")
	assert.Equal(t, updatedLastName, updateResp.User.LastName, "Last name not updated in response")
	assert.Equal(t, regPayload.Email, updateResp.User.Email, "Email should not have changed if not in update payload")
	assert.NotEmpty(t, updateResp.Token, "New token not found in update response")
	assert.NotEqual(t, authData.JWTToken, updateResp.Token, "New token should be different from the initial token")

	// 6. Create new auth data with the updated token
	newAuthData := &TestAuthData{
		JWTToken:  updateResp.Token,
		CSRFToken: authData.CSRFToken, // Keep the same CSRF token
		Cookies:   authData.Cookies,   // Keep the same cookies structure
	}

	// 7. Use the new JWT to get user by ID
	getRR := getUserById(newAuthData)

	// 8. Assert that the subsequent request is successful and returns updated info
	require.Equal(t, http.StatusOK, getRR.Code, "GetUserByID with new token failed. Body: "+getRR.Body.String())
	var fetchedUser models.UserDTO
	err = json.Unmarshal(getRR.Body.Bytes(), &fetchedUser)
	require.NoError(t, err)

	assert.Equal(t, updatedFirstName, fetchedUser.FirstName, "Fetched user first name is not the updated one")
	assert.Equal(t, updatedLastName, fetchedUser.LastName, "Fetched user last name is not the updated one")
	assert.Equal(t, regPayload.Email, fetchedUser.Email, "Fetched user email mismatch")
}
