package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"go_ecommerce/pkg/handlers"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

var testDB *sql.DB
var testUserHandler *handlers.UserHandler
var testAuthHandler handlers.IAuthHandler
var testProductHandler *handlers.ProductHandler
var testAuthenticator middleware.TokenGenerator
var testRouter *http.ServeMux

func registerTestUserAuth(t *testing.T, userDTO models.UserDTO) *httptest.ResponseRecorder {
	payloadBytes, err := json.Marshal(userDTO)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users/register", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req) // Uses the shared testRouter

	return rr
}

func loginUserAndGetTokenAuth(t *testing.T, email, password string) (*httptest.ResponseRecorder, string, error) {
	loginPayload := models.UserDTO{Email: email, Password: password}
	payloadBytes, err := json.Marshal(loginPayload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		return rr, "", fmt.Errorf("login request failed with status %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	token, ok := resp["token"]
	if !ok || token == "" {
		return rr, "", fmt.Errorf("token not found in login response")
	}
	return rr, token, nil
}

func getUserById(token string) *httptest.ResponseRecorder {
	getReq, _ := http.NewRequest(http.MethodGet, "/users/get-by-id", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	return getRR
}

func getUserByName(token string) *httptest.ResponseRecorder {
	getReq, _ := http.NewRequest(http.MethodGet, "/users/get-by-name", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	return getRR
}

func getUserByEmail(token string) *httptest.ResponseRecorder {
	getReq, _ := http.NewRequest(http.MethodGet, "/users/get-by-email", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	return getRR
}

func updateUser(token string, payload models.UserDTO) *httptest.ResponseRecorder {
	updatePayloadBytes, _ := json.Marshal(payload)
	updateReq, _ := http.NewRequest(http.MethodPut, "/users/update", bytes.NewBuffer(updatePayloadBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateRR, updateReq)
	return updateRR
}

func deleteUser(token string) *httptest.ResponseRecorder {
	deleteReq, _ := http.NewRequest(http.MethodDelete, "/users/delete", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteRR, deleteReq)
	return deleteRR
}

func createUserDTO(email string, password string) models.UserDTO {
	return models.UserDTO{
		FirstName:  "Login",
		LastName:   "User",
		MiddleName: "",
		Email:      email,
		Phone:      9876543210,
		Password:   password,
		IsProducer: false,
		Address:    "456 Login Ave",
		City:       "Loginton",
		Country:    "Testland",
		ZipCode:    54321,
	}
}
