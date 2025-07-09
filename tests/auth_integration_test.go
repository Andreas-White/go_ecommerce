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

func TestChangePassword_Success(t *testing.T) {
	clearTables(t)

	email := "change_pass_succ@example.com"
	initialPassword := "OldSecurePassword123"
	newPassword := "NewSecurePassword456"
	userDTO := createUserDTO(email, initialPassword, false)
	registerTestUserAuth(t, userDTO)

	_, token, err := loginUserAndGetTokenAuth(t, email, initialPassword)
	require.NoError(t, err, "Failed to login user for password change test")
	require.NotEmpty(t, token, "Login token should not be empty")

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: initialPassword,
		NewPassword:     newPassword,
	}
	payload, _ := json.Marshal(changeDTO)

	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, fmt.Sprintf("Expected status OK for successful password change, got %d: %s", rr.Code, rr.Body.String()))
	var successResp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &successResp)
	require.NoError(t, err)
	assert.Equal(t, "Password changed successfully", successResp["message"])

	_, newToken, err := loginUserAndGetTokenAuth(t, email, newPassword)
	assert.NoError(t, err, "Login with new password should succeed")
	assert.NotEmpty(t, newToken, "Token from new password login should not be empty")

	_, _, err = loginUserAndGetTokenAuth(t, email, initialPassword)
	assert.Error(t, err, "Login with old password should fail")
}

func TestChangePassword_IncorrectCurrentPassword(t *testing.T) {
	clearTables(t)

	email := "change_pass_fail@example.com"
	correctPassword := "CorrectPassword123"
	wrongCurrentPassword := "WrongPassword123"
	newPassword := "NewPasswordAttempt456"

	userDTO := createUserDTO(email, correctPassword, false)
	registerTestUserAuth(t, userDTO)

	_, token, err := loginUserAndGetTokenAuth(t, email, correctPassword)
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

	assert.Equal(t, http.StatusInternalServerError, rr.Code, fmt.Sprintf("Expected status InternalServerError for incorrect current password if not specifically handled, got %d: %s", rr.Code, rr.Body.String()))
	var errResp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "Failed to change password", "Error message mismatch")
}

func TestChangePassword_InvalidNewPassword_TooShort(t *testing.T) {
	clearTables(t)
	email := "change_pass_short@example.com"
	currentPassword := "ValidOldPassword123"
	shortNewPassword := "short"

	userDTO := createUserDTO(email, currentPassword, false)
	registerTestUserAuth(t, userDTO)
	_, token, err := loginUserAndGetTokenAuth(t, email, currentPassword)
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

	assert.Equal(t, http.StatusInternalServerError, rr.Code, fmt.Sprintf("Response: %s", rr.Body.String()))
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.Contains(t, errResp["error"], "Failed to change password")
}

func TestChangePassword_NewPasswordSameAsOld(t *testing.T) {
	clearTables(t)
	email := "change_pass_same@example.com"
	currentPassword := "SameOldPassword123"

	userDTO := createUserDTO(email, currentPassword, false)
	registerTestUserAuth(t, userDTO)
	_, token, err := loginUserAndGetTokenAuth(t, email, currentPassword)
	require.NoError(t, err)

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: currentPassword,
		NewPassword:     currentPassword,
	}
	payload, _ := json.Marshal(changeDTO)
	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, fmt.Sprintf("Response: %s", rr.Body.String()))
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	assert.Contains(t, errResp["error"], "Failed to change password")
}

func TestChangePassword_MissingFields(t *testing.T) {
	clearTables(t)
	email := "change_pass_missing@example.com"
	currentPassword := "ValidPassword123"

	userDTO := createUserDTO(email, currentPassword, false)
	registerTestUserAuth(t, userDTO)
	_, token, err := loginUserAndGetTokenAuth(t, email, currentPassword)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		payload models.ChangePasswordDTO
	}{
		{
			name:    "Missing new_password",
			payload: models.ChangePasswordDTO{CurrentPassword: currentPassword},
		},
		{
			name:    "Missing current_password",
			payload: models.ChangePasswordDTO{NewPassword: "NewSecurePassword123"},
		},
		{
			name:    "Missing both passwords",
			payload: models.ChangePasswordDTO{},
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

			assert.Equal(t, http.StatusInternalServerError, rr.Code, fmt.Sprintf("Response for %s: %s", tc.name, rr.Body.String()))
			var errResp map[string]string
			json.Unmarshal(rr.Body.Bytes(), &errResp)
			assert.Contains(t, errResp["error"], "Failed to change password")
		})
	}
}

func TestChangePassword_Unauthorized_NoToken(t *testing.T) {
	clearTables(t)

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: "anyCurrentPassword",
		NewPassword:     "anyNewPassword123",
	}
	payload, _ := json.Marshal(changeDTO)

	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, fmt.Sprintf("Expected status Unauthorized for missing token, got %d: %s", rr.Code, rr.Body.String()))
	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "Authorization header required")
}

func TestChangePassword_Unauthorized_InvalidToken(t *testing.T) {
	clearTables(t)

	changeDTO := models.ChangePasswordDTO{
		CurrentPassword: "anyCurrentPassword",
		NewPassword:     "anyNewPassword123",
	}
	payload, _ := json.Marshal(changeDTO)

	req, _ := http.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalidtoken123")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, fmt.Sprintf("Expected status Unauthorized for invalid token, got %d: %s", rr.Code, rr.Body.String()))
	var errResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "Invalid or expired token")
}
