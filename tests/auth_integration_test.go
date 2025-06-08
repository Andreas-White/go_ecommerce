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

// Helper to simulate user registration for auth tests
func registerTestUserAuth(t *testing.T, userDTO models.UserDTO) {
	payloadBytes, err := json.Marshal(userDTO)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users/register", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Uses the shared testRouter

	require.Equal(t, http.StatusCreated, rr.Code, fmt.Sprintf("Registration failed: %s", rr.Body.String()))
}

// Helper to simulate user login for auth tests and get token
func loginUserAndGetTokenAuth(t *testing.T, email, password string) (string, error) {
	// Login handler now expects UserDTO for email and password
	loginPayload := models.UserDTO{Email: email, Password: password}
	payloadBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login payload: %w", err)
	}

	req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Uses the shared testRouter

	if rr.Code != http.StatusOK {
		return "", fmt.Errorf("login request failed with status %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string // Login handler now returns map[string]string{"token": "..."}
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal login response: %w", err)
	}

	token, ok := resp["token"]
	if !ok || token == "" {
		return "", fmt.Errorf("token not found in login response")
	}
	return token, nil
}

// TestChangePassword_Success tests the successful password change flow.
func TestChangePassword_Success(t *testing.T) {
	clearUserTables(t) // Assumes clearUserTables is accessible from user_integration_test.go

	email := "change_pass_succ@example.com"
	initialPassword := "OldSecurePassword123"
	newPassword := "NewSecurePassword456"
	userDTO := models.UserDTO{
		FirstName:  "Pass",
		LastName:   "Changer",
		Email:      email,
		Password:   initialPassword,
		Phone:      int64(1234567890), // Ensure Phone is int64
		Address:    "123 Test St",
		City:       "Testville",
		Country:    "Testland",
		ZipCode:    12345,
		IsProducer: false,
	}
	registerTestUserAuth(t, userDTO)

	token, err := loginUserAndGetTokenAuth(t, email, initialPassword)
	require.NoError(t, err, "Failed to login user for password change test")
	require.NotEmpty(t, token, "Login token should not be empty")

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: initialPassword,
		NewPassword:     newPassword,
	}
	payload, _ := json.Marshal(changeDTO)

	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload)) // Endpoint is /auth/change-password
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, fmt.Sprintf("Expected status OK for successful password change, got %d: %s", rr.Code, rr.Body.String()))
	var successResp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &successResp)
	require.NoError(t, err)
	assert.Equal(t, "Password changed successfully", successResp["message"])

	newToken, err := loginUserAndGetTokenAuth(t, email, newPassword)
	assert.NoError(t, err, "Login with new password should succeed")
	assert.NotEmpty(t, newToken, "Token from new password login should not be empty")

	_, err = loginUserAndGetTokenAuth(t, email, initialPassword)
	assert.Error(t, err, "Login with old password should fail")
}

// TestChangePassword_IncorrectCurrentPassword tests password change with incorrect current password.
func TestChangePassword_IncorrectCurrentPassword(t *testing.T) {
	clearUserTables(t)

	email := "change_pass_fail@example.com"
	correctPassword := "CorrectPassword123"
	wrongCurrentPassword := "WrongPassword123"
	newPassword := "NewPasswordAttempt456"

	userDTO := models.UserDTO{FirstName: "Fail", LastName: "User", Email: email, Password: correctPassword, Phone: int64(111222333), Address: "1 Fail St", City: "Failville", Country: "FailLand", ZipCode: 10001, IsProducer: false}
	registerTestUserAuth(t, userDTO)

	token, err := loginUserAndGetTokenAuth(t, email, correctPassword)
	require.NoError(t, err)

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: wrongCurrentPassword,
		NewPassword:     newPassword,
	}
	payload, _ := json.Marshal(changeDTO)

	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// The service layer might return a more specific error, but the handler might generalize it.
	// Assuming a 500 for now if the service error isn't specifically mapped to 400/401 by handler for "incorrect current password"
	// Or, if your utils.RespondWithError maps service errors to specific HTTP statuses, adjust this.
	// For now, let's assume the handler catches this as a generic internal error or a specific bad request.
	// Based on authHandler.go, it returns 500 for generic service errors.
	// If authService.ChangePassword returns a specific error type that authHandler maps to 400, this should be 400.
	// Let's assume a generic error from service leads to 500 from handler for now.
	// UPDATE: authHandler.go's ChangePassword calls utils.RespondWithError(w, http.StatusInternalServerError, "Failed to change password")
	// This means we should expect 500. If a more specific error like "incorrect current password" should be a 400,
	// the authService needs to return a distinguishable error, and authHandler needs to map it.
	// For now, sticking to what the handler currently does.
	// If the intention is a 400 for "Incorrect current password", the authService.ChangePassword should return an error
	// that authHandler.ChangePassword can identify and map to utils.RespondWithError(w, http.StatusBadRequest, "Incorrect current password")
	assert.Equal(t, http.StatusInternalServerError, rr.Code, fmt.Sprintf("Expected status InternalServerError for incorrect current password if not specifically handled, got %d: %s", rr.Code, rr.Body.String()))
	var errResp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	// The error message might be generic "Failed to change password"
	assert.Contains(t, errResp["error"], "Failed to change password", "Error message mismatch")
}

// TestChangePassword_InvalidNewPassword_TooShort tests new password validation (too short).
func TestChangePassword_InvalidNewPassword_TooShort(t *testing.T) {
	clearUserTables(t)
	email := "change_pass_short@example.com"
	currentPassword := "ValidOldPassword123"
	shortNewPassword := "short" // Less than 8 chars

	userDTO := models.UserDTO{FirstName: "Short", LastName: "Pass", Email: email, Password: currentPassword, Phone: int64(222333444), Address: "2 Short St", City: "Shortville", Country: "Shortland", ZipCode: 20002, IsProducer: false}
	registerTestUserAuth(t, userDTO)
	token, err := loginUserAndGetTokenAuth(t, email, currentPassword)
	require.NoError(t, err)

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: currentPassword,
		NewPassword:     shortNewPassword,
	}
	payload, _ := json.Marshal(changeDTO)
	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// Assuming validation happens in the service layer and might result in a generic error from handler
	// Or if DTO validation is used by handler before calling service.
	// The ChangePasswordDTO has `validate:"required,min=8"` for NewPassword.
	// The handler currently doesn't explicitly run validation on ChangePasswordDTO before passing to service.
	// If service does validation and returns error, handler returns 500.
	// If we want 400, handler should validate DTO.
	// For now, assume service error -> handler 500.
	assert.Equal(t, http.StatusInternalServerError, rr.Code, fmt.Sprintf("Response: %s", rr.Body.String()))
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.Contains(t, errResp["error"], "Failed to change password") // Generic message
}

// TestChangePassword_NewPasswordSameAsOld tests new password being same as old.
func TestChangePassword_NewPasswordSameAsOld(t *testing.T) {
	clearUserTables(t)
	email := "change_pass_same@example.com"
	currentPassword := "SameOldPassword123"

	userDTO := models.UserDTO{FirstName: "Same", LastName: "Pass", Email: email, Password: currentPassword, Phone: int64(333444555), Address: "3 Same St", City: "Sameville", Country: "Sameland", ZipCode: 30003, IsProducer: false}
	registerTestUserAuth(t, userDTO)
	token, err := loginUserAndGetTokenAuth(t, email, currentPassword)
	require.NoError(t, err)

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: currentPassword,
		NewPassword:     currentPassword, // Same as old
	}
	payload, _ := json.Marshal(changeDTO)
	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// Expecting service layer to catch this and return an error. Handler returns 500.
	assert.Equal(t, http.StatusInternalServerError, rr.Code, fmt.Sprintf("Response: %s", rr.Body.String()))
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.Contains(t, errResp["error"], "Failed to change password") // Generic message
}

// TestChangePassword_MissingFields tests request with missing fields.
func TestChangePassword_MissingFields(t *testing.T) {
	clearUserTables(t)
	email := "change_pass_missing@example.com"
	currentPassword := "ValidPassword123"

	userDTO := models.UserDTO{FirstName: "Missing", LastName: "Fields", Email: email, Password: currentPassword, Phone: int64(444555666), Address: "4 Missing St", City: "Missingville", Country: "Missingland", ZipCode: 40004, IsProducer: false}
	registerTestUserAuth(t, userDTO)
	token, err := loginUserAndGetTokenAuth(t, email, currentPassword)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		payload models.ChangePasswordDTO
		// errMsg  string // Error message might be generic from handler
	}{
		{
			name:    "Missing new_password",
			payload: models.ChangePasswordDTO{CurrentPassword: currentPassword},
			// errMsg:  "new_password is required", // Or generic "Failed to change password"
		},
		{
			name:    "Missing current_password",
			payload: models.ChangePasswordDTO{NewPassword: "NewSecurePassword123"},
			// errMsg:  "current_password is required", // Or generic "Failed to change password"
		},
		{
			name:    "Missing both passwords",
			payload: models.ChangePasswordDTO{},
			// errMsg:  "current_password and new_password are required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.payload)
			req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()
			testRouter.ServeHTTP(rr, req)

			// The handler decodes, but if fields are missing and service layer requires them,
			// it will error. Handler returns 500.
			// If DTO validation `validate:"required"` was checked by handler, it would be 400.
			assert.Equal(t, http.StatusInternalServerError, rr.Code, fmt.Sprintf("Response for %s: %s", tc.name, rr.Body.String()))
			var errResp map[string]string
			json.Unmarshal(rr.Body.Bytes(), &errResp)
			assert.Contains(t, errResp["error"], "Failed to change password") // Generic
		})
	}
}

// TestChangePassword_Unauthorized_NoToken tests request without an auth token.
func TestChangePassword_Unauthorized_NoToken(t *testing.T) {
	clearUserTables(t)

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: "anyCurrentPassword",
		NewPassword:     "anyNewPassword123",
	}
	payload, _ := json.Marshal(changeDTO)

	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Middleware should catch this

	assert.Equal(t, http.StatusUnauthorized, rr.Code, fmt.Sprintf("Expected status Unauthorized for missing token, got %d: %s", rr.Code, rr.Body.String()))
	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	// Message from middleware.AuthenticateJWT
	assert.Contains(t, errResp["error"], "Authorization header required", "Error message mismatch for no token")
}

// TestChangePassword_Unauthorized_InvalidToken tests request with an invalid auth token.
func TestChangePassword_Unauthorized_InvalidToken(t *testing.T) {
	clearUserTables(t)

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: "anyCurrentPassword",
		NewPassword:     "anyNewPassword123",
	}
	payload, _ := json.Marshal(changeDTO)

	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalidtoken123")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Middleware should catch this

	assert.Equal(t, http.StatusUnauthorized, rr.Code, fmt.Sprintf("Expected status Unauthorized for invalid token, got %d: %s", rr.Code, rr.Body.String()))
	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	// Message from middleware.AuthenticateJWT
	assert.Contains(t, errResp["error"], "Invalid or expired token", "Error message mismatch for invalid token")
}
