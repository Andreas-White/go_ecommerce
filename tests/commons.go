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
var testCartHandler *handlers.CartHandler
var testAuthenticator middleware.TokenGenerator
var testRouter *http.ServeMux

// TestAuthData holds authentication data for tests
type TestAuthData struct {
	JWTToken  string
	CSRFToken string
	Cookies   []*http.Cookie
}

func getCSRFTokenForEndpoint(t *testing.T, endpoint string) *TestAuthData {
	getReq, err := http.NewRequest("GET", endpoint, nil)
	require.NoError(t, err)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	return extractAuthDataFromResponse(getRR)
}

func registerTestUserAuth(t *testing.T, userDTO models.UserDTO) (*httptest.ResponseRecorder, *TestAuthData) {
	// First, get CSRF token by making a GET request
	csrfAuthData := getCSRFTokenForEndpoint(t, "/users/register")

	payloadBytes, err := json.Marshal(userDTO)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users/register", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	// Add CSRF token header and cookie
	if csrfAuthData != nil {
		for _, cookie := range csrfAuthData.Cookies {
			req.AddCookie(cookie)
		}
		if csrfAuthData.CSRFToken != "" {
			req.Header.Set("X-CSRF-Token", csrfAuthData.CSRFToken)
		}
	}

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	// Extract cookies and CSRF token from response
	authData := extractAuthDataFromResponse(rr)

	return rr, authData
}

func loginUserAndGetTokenAuth(t *testing.T, email, password string) (*httptest.ResponseRecorder, *TestAuthData, error) {
	// First, get CSRF token by making a GET request
	csrfAuthData := getCSRFTokenForEndpoint(t, "/users/login")

	loginPayload := models.UserDTO{Email: email, Password: password}
	payloadBytes, err := json.Marshal(loginPayload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users/login", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	// Add CSRF token header and cookie
	if csrfAuthData != nil {
		for _, cookie := range csrfAuthData.Cookies {
			req.AddCookie(cookie)
		}
		if csrfAuthData.CSRFToken != "" {
			req.Header.Set("X-CSRF-Token", csrfAuthData.CSRFToken)
		}
	}

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusCreated {
		return rr, nil, fmt.Errorf("login request failed with status %d: %s", rr.Code, rr.Body.String())
	}

	// Extract cookies and CSRF token from response
	authData := extractAuthDataFromResponse(rr)

	return rr, authData, nil
}

// extractAuthDataFromResponse extracts JWT and CSRF tokens from response cookies and body
func extractAuthDataFromResponse(rr *httptest.ResponseRecorder) *TestAuthData {
	authData := &TestAuthData{
		Cookies: rr.Result().Cookies(),
	}

	// Extract JWT from cookie
	for _, cookie := range authData.Cookies {
		if cookie.Name == "jwt" {
			authData.JWTToken = cookie.Value
		}
		if cookie.Name == "csrf_token" {
			authData.CSRFToken = cookie.Value
		}
	}

	// Also try to get CSRF token from response body
	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err == nil {
		if csrfToken, ok := resp["csrf_token"]; ok {
			authData.CSRFToken = csrfToken
		}
	}

	return authData
}

// addAuthHeaders adds authentication headers and cookies to a request
func addAuthHeaders(req *http.Request, authData *TestAuthData) {
	if authData != nil {
		// Add CSRF token header for state-changing requests
		if req.Method != "GET" && req.Method != "HEAD" && req.Method != "OPTIONS" {
			req.Header.Set("X-CSRF-Token", authData.CSRFToken)
		}

		// Add cookies
		for _, cookie := range authData.Cookies {
			req.AddCookie(cookie)
		}

		// Fallback: Add Authorization header for backward compatibility
		if authData.JWTToken != "" {
			req.Header.Set("Authorization", "Bearer "+authData.JWTToken)
		}
	}
}

func getUserById(authData *TestAuthData) *httptest.ResponseRecorder {
	getReq, _ := http.NewRequest(http.MethodGet, "/users/get-by-id", nil)
	addAuthHeaders(getReq, authData)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	return getRR
}

func getUserByName(authData *TestAuthData) *httptest.ResponseRecorder {
	getReq, _ := http.NewRequest(http.MethodGet, "/users/get-by-name", nil)
	addAuthHeaders(getReq, authData)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	return getRR
}

func getUserByEmail(authData *TestAuthData) *httptest.ResponseRecorder {
	getReq, _ := http.NewRequest(http.MethodGet, "/users/get-by-email", nil)
	addAuthHeaders(getReq, authData)
	getRR := httptest.NewRecorder()
	testRouter.ServeHTTP(getRR, getReq)
	return getRR
}

func updateUser(authData *TestAuthData, payload models.UserDTO) *httptest.ResponseRecorder {
	updatePayloadBytes, _ := json.Marshal(payload)
	updateReq, _ := http.NewRequest(http.MethodPut, "/users/update", bytes.NewBuffer(updatePayloadBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	addAuthHeaders(updateReq, authData)
	updateRR := httptest.NewRecorder()
	testRouter.ServeHTTP(updateRR, updateReq)
	return updateRR
}

func deleteUser(authData *TestAuthData) *httptest.ResponseRecorder {
	deleteReq, _ := http.NewRequest(http.MethodDelete, "/users/delete", nil)
	addAuthHeaders(deleteReq, authData)
	deleteRR := httptest.NewRecorder()
	testRouter.ServeHTTP(deleteRR, deleteReq)
	return deleteRR
}

func createUserDTO(email string, password string, isProducer bool) models.UserDTO {
	return models.UserDTO{
		FirstName:  "Login",
		LastName:   "User",
		MiddleName: "",
		Email:      email,
		Phone:      9876543210,
		Password:   password,
		IsProducer: isProducer,
		Address:    "456 Login Ave",
		City:       "Loginton",
		Country:    "Testland",
		ZipCode:    54321,
	}
}
