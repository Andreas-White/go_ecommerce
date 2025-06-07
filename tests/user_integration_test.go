package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go_ecommerce/internal/config"
	"go_ecommerce/pkg/database"
	"go_ecommerce/pkg/handlers"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/services"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB
var testUserHandler *handlers.UserHandler
var testAuthenticator middleware.TokenGenerator

func TestMain(m *testing.M) {
	tCfg := config.LoadTestConfig()
	cfg := &config.Config{
		DBHost:    tCfg.DBHost,
		DBPort:    tCfg.DBPort,
		DBUser:    tCfg.DBUser,
		DBPass:    tCfg.DBPass,
		DBName:    tCfg.DBName,
		DBSslMode: tCfg.DBSslMode,
		JWTKey:    tCfg.JWTKey,
	}

	// Run migrations for the test database
	migrationDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSslMode)

	// Migrations are in the parent directory's 'migrations' folder
	migrationsPath := "file://../migrations"

	log.Printf("Attempting to run migrations for test DB from: %s on database: %s", migrationsPath, cfg.DBName)
	mig, err := migrate.New(migrationsPath, migrationDSN)
	if err != nil {
		log.Fatalf("Failed to create new migrate instance for test DB: %v", err)
	}

	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations for test DB: %v", err)
	}
	log.Println("Test database migrations applied successfully (or no changes needed).")

	testDB, err = database.Init(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize test database: %v", err)
	}
	defer testDB.Close()

	err = testDB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping test database: %v", err)
	}

	auth, err := middleware.NewAuthenticator(cfg.JWTKey)
	if err != nil {
		log.Fatalf("Failed to create authenticator for tests: %v", err)
	}
	testAuthenticator = auth

	userRepo := repositories.NewUserRepository(testDB)
	userService := services.NewUserService(userRepo)
	testUserHandler = handlers.NewUserHandler(userService, testAuthenticator)

	code := m.Run()
	os.Exit(code)
}

func clearUserTables(t *testing.T) {
	t.Helper()
	// IMPORTANT: Adjust table names if they are different in your schema.
	_, err := testDB.Exec("TRUNCATE TABLE users, auths RESTART IDENTITY CASCADE")
	require.NoError(t, err, "Failed to truncate user tables")
}

func TestUserRegistration_Success(t *testing.T) {
	clearUserTables(t)

	registrationPayload := models.UserRegister{
		FirstName:  "Test",
		LastName:   "User",
		Email:      "testuser@example.com",
		Password:   "password123",
		Phone:      1234567890,
		IsProducer: false,
		Address:    "123 Test St",
		City:       "Testville",
		Country:    "Testland",
		ZipCode:    12345,
	}

	payloadBytes, err := json.Marshal(registrationPayload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201 Created")

	var responseBody map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "User created successfully", responseBody["message"])

	var email string
	err = testDB.QueryRow("SELECT email FROM users WHERE email = $1", registrationPayload.Email).Scan(&email)
	require.NoError(t, err, "User should exist in database after registration")
	assert.Equal(t, registrationPayload.Email, email)
}

func TestUserLogin_Success(t *testing.T) {
	clearUserTables(t)

	registeredEmail := "loginuser@example.com"
	registeredPassword := "securepassword123"
	registrationPayload := models.UserRegister{
		FirstName: "Login", LastName: "User", Email: registeredEmail, Password: registeredPassword,
		Phone: 9876543210, IsProducer: false, Address: "456 Login Ave", City: "Loginton", Country: "Testland", ZipCode: 54321,
	}

	regPayloadBytes, err := json.Marshal(registrationPayload)
	require.NoError(t, err)
	regReq, err := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(regPayloadBytes))
	require.NoError(t, err)
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(regRR, regReq)
	require.Equal(t, http.StatusCreated, regRR.Code, "User registration failed")

	loginPayload := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    registeredEmail,
		Password: registeredPassword,
	}

	loginPayloadBytes, err := json.Marshal(loginPayload)
	require.NoError(t, err)
	loginReq, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(loginPayloadBytes))
	require.NoError(t, err)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Login).ServeHTTP(loginRR, loginReq)

	assert.Equal(t, http.StatusOK, loginRR.Code, "Expected status code 200 OK for login")
	var loginResponseBody map[string]string
	err = json.Unmarshal(loginRR.Body.Bytes(), &loginResponseBody)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResponseBody["token"], "Token should not be empty on successful login")
}

func TestUserLogin_Failure(t *testing.T) {
	clearUserTables(t)

	// Test with non-existent user
	loginPayloadNonExistent := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{"nonexistent@example.com", "anypassword"}
	payloadBytes, _ := json.Marshal(loginPayloadNonExistent)
	req, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Login).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected 401 for non-existent user")

	// Register a user for the wrong password test
	regEmail := "wrongpass@example.com"
	regPassword := "correctpassword"
	registrationPayload := models.UserRegister{
		FirstName: "Wrong", LastName: "Pass", Email: regEmail, Password: regPassword,
		Phone: 123, IsProducer: false, Address: "Addr", City: "City", Country: "Country", ZipCode: 123,
	}
	regPayloadBytes, _ := json.Marshal(registrationPayload)
	regReq, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(regPayloadBytes))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(regRR, regReq)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration for wrong password test failed")

	// Test with wrong password
	loginPayloadWrongPass := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{regEmail, "incorrectpassword"}
	payloadBytesWrong, _ := json.Marshal(loginPayloadWrongPass)
	reqWrong, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(payloadBytesWrong))
	reqWrong.Header.Set("Content-Type", "application/json")
	rrWrong := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Login).ServeHTTP(rrWrong, reqWrong)
	assert.Equal(t, http.StatusUnauthorized, rrWrong.Code, "Expected 401 for wrong password")
}

func TestGetUserByID_Success(t *testing.T) {
	clearUserTables(t)
	regEmail := "getuser@example.com"
	regPassword := "password123"
	var registeredUser models.User

	regPayload := models.UserRegister{
		FirstName: "Get", LastName: "User", Email: regEmail, Password: regPassword,
		Phone: 1112223333, IsProducer: false, Address: "789 Get St", City: "Getville", Country: "Testland", ZipCode: 67890,
	}
	regPayloadBytes, _ := json.Marshal(regPayload)
	regReq, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(regPayloadBytes))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(regRR, regReq)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	err := testDB.QueryRowContext(context.Background(), "SELECT id, first_name, last_name, email, phone, is_producer, address, city, country, zip_code, created_at FROM users WHERE email = $1", regEmail).
		Scan(&registeredUser.ID, &registeredUser.FirstName, &registeredUser.LastName, &registeredUser.Email, &registeredUser.Phone, &registeredUser.IsProducer, &registeredUser.Address, &registeredUser.City, &registeredUser.Country, &registeredUser.ZipCode, &registeredUser.CreatedAt)
	require.NoError(t, err, "Failed to retrieve registered user from DB")

	loginPayload := map[string]string{"email": regEmail, "password": regPassword}
	loginPayloadBytes, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(loginPayloadBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Login).ServeHTTP(loginRR, loginReq)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	var loginRespBody map[string]string
	_ = json.Unmarshal(loginRR.Body.Bytes(), &loginRespBody)
	token := loginRespBody["token"]
	require.NotEmpty(t, token, "Token not found")

	getReq, _ := http.NewRequest(http.MethodGet, "/users/get_by_id", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRR := httptest.NewRecorder()
	authedHandler := testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByID))
	authedHandler.ServeHTTP(getRR, getReq)

	assert.Equal(t, http.StatusOK, getRR.Code, "Expected 200 OK for GetUserByID")
	var fetchedUser models.User
	err = json.Unmarshal(getRR.Body.Bytes(), &fetchedUser)
	require.NoError(t, err)
	assert.Equal(t, registeredUser.ID, fetchedUser.ID)
	assert.Equal(t, registeredUser.FirstName, fetchedUser.FirstName)
	assert.Equal(t, registeredUser.Email, fetchedUser.Email)
}

func TestGetUserByID_Unauthorized(t *testing.T) {
	clearUserTables(t)
	getReqUnauth, _ := http.NewRequest(http.MethodGet, "/users/get_by_id", nil)
	getRRUnauth := httptest.NewRecorder()
	authedHandler := testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByID))
	authedHandler.ServeHTTP(getRRUnauth, getReqUnauth)
	assert.Equal(t, http.StatusUnauthorized, getRRUnauth.Code, "Expected 401 Unauthorized without token")

	getReqInvalidToken, _ := http.NewRequest(http.MethodGet, "/users/get_by_id", nil)
	getReqInvalidToken.Header.Set("Authorization", "Bearer invalidtoken123")
	getRRInvalidToken := httptest.NewRecorder()
	authedHandler.ServeHTTP(getRRInvalidToken, getReqInvalidToken)
	assert.Equal(t, http.StatusUnauthorized, getRRInvalidToken.Code, "Expected 401 Unauthorized with invalid token")
}

func TestUpdateUser_Success(t *testing.T) {
	clearUserTables(t)
	regEmail := "updateuser@example.com"
	regPassword := "password123"
	originalFirstName := "OriginalFirst"

	regPayload := models.UserRegister{
		FirstName: originalFirstName, LastName: "User", Email: regEmail, Password: regPassword,
		Phone: 2223334444, IsProducer: true, Address: "111 Update St", City: "Updateville", Country: "Testland", ZipCode: 11111,
	}
	regPayloadBytes, _ := json.Marshal(regPayload)
	regReq, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(regPayloadBytes))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(regRR, regReq)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	loginPayload := map[string]string{"email": regEmail, "password": regPassword}
	loginPayloadBytes, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(loginPayloadBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Login).ServeHTTP(loginRR, loginReq)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	var loginRespBody map[string]string
	_ = json.Unmarshal(loginRR.Body.Bytes(), &loginRespBody)
	token := loginRespBody["token"]
	require.NotEmpty(t, token, "Token not found")

	updatedFirstName := "UpdatedFirst"
	updatePayload := struct {
		FirstName string `json:"first_name"`
		City      string `json:"city"`
	}{
		FirstName: updatedFirstName, City: "NewCity",
	}
	updatePayloadBytes, _ := json.Marshal(updatePayload)
	updateReq, _ := http.NewRequest(http.MethodPut, "/users/update", bytes.NewBuffer(updatePayloadBytes))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.Header.Set("Authorization", "Bearer "+token)

	updateRR := httptest.NewRecorder()
	authedUpdateHandler := testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.UpdateUser))
	authedUpdateHandler.ServeHTTP(updateRR, updateReq)

	assert.Equal(t, http.StatusOK, updateRR.Code, "Expected 200 OK for UpdateUser")
	var updateRespBody map[string]string
	_ = json.Unmarshal(updateRR.Body.Bytes(), &updateRespBody)
	assert.Equal(t, "User updated successfully", updateRespBody["message"])

	var dbFirstName, dbCity string
	err := testDB.QueryRow("SELECT first_name, city FROM users WHERE email = $1", regEmail).Scan(&dbFirstName, &dbCity)
	require.NoError(t, err, "Failed to query updated user from DB")
	assert.Equal(t, updatedFirstName, dbFirstName)
	assert.Equal(t, updatePayload.City, dbCity)
}

func TestUpdateUser_Unauthorized(t *testing.T) {
	clearUserTables(t)
	updatePayload := map[string]string{"first_name": "AttemptUpdate"}
	updatePayloadBytes, _ := json.Marshal(updatePayload)

	updateReqUnauth, _ := http.NewRequest(http.MethodPut, "/users/update", bytes.NewBuffer(updatePayloadBytes))
	updateReqUnauth.Header.Set("Content-Type", "application/json")
	updateRRUnauth := httptest.NewRecorder()
	authedUpdateHandler := testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.UpdateUser))
	authedUpdateHandler.ServeHTTP(updateRRUnauth, updateReqUnauth)
	assert.Equal(t, http.StatusUnauthorized, updateRRUnauth.Code, "Expected 401 Unauthorized without token")

	updateReqInvalidToken, _ := http.NewRequest(http.MethodPut, "/users/update", bytes.NewBuffer(updatePayloadBytes))
	updateReqInvalidToken.Header.Set("Content-Type", "application/json")
	updateReqInvalidToken.Header.Set("Authorization", "Bearer invalidtoken123")
	updateRRInvalidToken := httptest.NewRecorder()
	authedUpdateHandler.ServeHTTP(updateRRInvalidToken, updateReqInvalidToken)
	assert.Equal(t, http.StatusUnauthorized, updateRRInvalidToken.Code, "Expected 401 Unauthorized with invalid token")
}

func TestDeleteUser_Success(t *testing.T) {
	clearUserTables(t)
	regEmail := "deleteuser@example.com"
	regPassword := "password123"

	regPayload := models.UserRegister{
		FirstName: "Delete", LastName: "Me", Email: regEmail, Password: regPassword,
		Phone: 3334445555, IsProducer: false, Address: "777 Delete Ln", City: "Deleteville", Country: "Testland", ZipCode: 22222,
	}
	regPayloadBytes, _ := json.Marshal(regPayload)
	regReq, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(regPayloadBytes))
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(regRR, regReq)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	loginPayload := map[string]string{"email": regEmail, "password": regPassword}
	loginPayloadBytes, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(loginPayloadBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Login).ServeHTTP(loginRR, loginReq)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	var loginRespBody map[string]string
	_ = json.Unmarshal(loginRR.Body.Bytes(), &loginRespBody)
	token := loginRespBody["token"]
	require.NotEmpty(t, token, "Token not found")

	deleteReq, _ := http.NewRequest(http.MethodDelete, "/users/delete", nil)
	deleteReq.Header.Set("Authorization", "Bearer "+token)
	deleteRR := httptest.NewRecorder()
	authedDeleteHandler := testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.DeleteUser))
	authedDeleteHandler.ServeHTTP(deleteRR, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRR.Code, "Expected 200 OK for DeleteUser")
	var deleteRespBody map[string]string
	_ = json.Unmarshal(deleteRR.Body.Bytes(), &deleteRespBody)
	assert.Equal(t, "User deleted successfully", deleteRespBody["message"])

	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", regEmail).Scan(&count)
	require.NoError(t, err, "Failed to query user count from DB after delete")
	assert.Equal(t, 0, count, "User should not exist in DB after deletion")
}

func TestDeleteUser_Unauthorized(t *testing.T) {
	clearUserTables(t)
	deleteReqUnauth, _ := http.NewRequest(http.MethodDelete, "/users/delete", nil)
	deleteRRUnauth := httptest.NewRecorder()
	authedDeleteHandler := testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.DeleteUser))
	authedDeleteHandler.ServeHTTP(deleteRRUnauth, deleteReqUnauth)
	assert.Equal(t, http.StatusUnauthorized, deleteRRUnauth.Code, "Expected 401 Unauthorized without token")

	deleteReqInvalidToken, _ := http.NewRequest(http.MethodDelete, "/users/delete", nil)
	deleteReqInvalidToken.Header.Set("Authorization", "Bearer invalidtoken123")
	deleteRRInvalidToken := httptest.NewRecorder()
	authedDeleteHandler.ServeHTTP(deleteRRInvalidToken, deleteReqInvalidToken)
	assert.Equal(t, http.StatusUnauthorized, deleteRRInvalidToken.Code, "Expected 401 Unauthorized with invalid token")
}

func TestGetUserByName(t *testing.T) {
	clearUserTables(t)
	regEmail := "byname@example.com"
	regFirstName := "John"
	regLastName := "Doe"

	regPayload := models.UserRegister{
		FirstName: regFirstName, LastName: regLastName, Email: regEmail, Password: "password123",
		Phone: 1234567890, IsProducer: false, Address: "123 Name St", City: "Nameville", Country: "Testland", ZipCode: 12345,
	}
	regPayloadBytes, err := json.Marshal(regPayload)
	require.NoError(t, err)
	regReq, err := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(regPayloadBytes))
	require.NoError(t, err)
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(regRR, regReq)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	loginPayload := map[string]string{"email": regEmail, "password": "password123"}
	loginPayloadBytes, err := json.Marshal(loginPayload)
	require.NoError(t, err)
	loginReq, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(loginPayloadBytes))
	require.NoError(t, err)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Login).ServeHTTP(loginRR, loginReq)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	var loginRespBody map[string]string
	err = json.Unmarshal(loginRR.Body.Bytes(), &loginRespBody)
	require.NoError(t, err)
	token := loginRespBody["token"]
	require.NotEmpty(t, token, "Token not found")

	getReq, err := http.NewRequest(http.MethodGet, "/users/get_by_name?name=John%20Doe", nil)
	require.NoError(t, err)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRR := httptest.NewRecorder()
	authedHandler := testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByName))
	authedHandler.ServeHTTP(getRR, getReq)

	assert.Equal(t, http.StatusOK, getRR.Code, "Expected 200 OK for GetUserByName")
	var user models.User
	err = json.Unmarshal(getRR.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, regFirstName, user.FirstName)
	assert.Equal(t, regLastName, user.LastName)
}

func TestGetUserByEmail(t *testing.T) {
	clearUserTables(t)
	regEmail := "byemail@example.com"
	regFirstName := "Jane"
	regLastName := "Smith"

	regPayload := models.UserRegister{
		FirstName: regFirstName, LastName: regLastName, Email: regEmail, Password: "password123",
		Phone: 1234567890, IsProducer: false, Address: "123 Email St", City: "Emailville", Country: "Testland", ZipCode: 12345,
	}
	regPayloadBytes, err := json.Marshal(regPayload)
	require.NoError(t, err)
	regReq, err := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(regPayloadBytes))
	require.NoError(t, err)
	regReq.Header.Set("Content-Type", "application/json")
	regRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(regRR, regReq)
	require.Equal(t, http.StatusCreated, regRR.Code, "Registration failed")

	loginPayload := map[string]string{"email": regEmail, "password": "password123"}
	loginPayloadBytes, err := json.Marshal(loginPayload)
	require.NoError(t, err)
	loginReq, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(loginPayloadBytes))
	require.NoError(t, err)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Login).ServeHTTP(loginRR, loginReq)
	require.Equal(t, http.StatusOK, loginRR.Code, "Login failed")
	var loginRespBody map[string]string
	err = json.Unmarshal(loginRR.Body.Bytes(), &loginRespBody)
	require.NoError(t, err)
	token := loginRespBody["token"]
	require.NotEmpty(t, token, "Token not found")

	getReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/get_by_email?email=%s", regEmail), nil)
	require.NoError(t, err)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRR := httptest.NewRecorder()
	authedHandler := testAuthenticator.AuthenticateJWT(http.HandlerFunc(testUserHandler.GetUserByEmail))
	authedHandler.ServeHTTP(getRR, getReq)

	assert.Equal(t, http.StatusOK, getRR.Code, "Expected 200 OK for GetUserByEmail")
	var user models.User
	err = json.Unmarshal(getRR.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, regEmail, user.Email)
	assert.Equal(t, regFirstName, user.FirstName)
}

func TestDuplicateRegistration(t *testing.T) {
	clearUserTables(t)
	regEmail := "duplicate@example.com"
	regPayload := models.UserRegister{
		FirstName: "Duplicate", LastName: "User", Email: regEmail, Password: "password123",
		Phone: 123, Address: "Addr", City: "City", Country: "Country", ZipCode: 123,
	}
	regPayloadBytes, _ := json.Marshal(regPayload)
	regReq, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(regPayloadBytes))
	regReq.Header.Set("Content-Type", "application/json")

	regRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(regRR, regReq)
	require.Equal(t, http.StatusCreated, regRR.Code, "First registration failed")

	// Attempt second registration with same email
	secondRegRR := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(secondRegRR, regReq) // Use the same request
	assert.Equal(t, http.StatusBadRequest, secondRegRR.Code, "Expected 400 for duplicate registration")
	var errorResp map[string]string
	err := json.Unmarshal(secondRegRR.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "Invalid request payload")
}

func TestInvalidPasswordRegistration(t *testing.T) {
	clearUserTables(t)
	regPayload := models.UserRegister{
		FirstName: "Short", LastName: "Pass", Email: "shortpass@example.com", Password: "123", // Invalid: too short
		Phone: 123, Address: "Addr", City: "City", Country: "Country", ZipCode: 123,
	}
	payloadBytes, _ := json.Marshal(regPayload)
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected 400 for invalid password")
	var errorResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "Password must be at least 6 characters long")
}

func TestInvalidEmailRegistration(t *testing.T) {
	clearUserTables(t)
	regPayload := models.UserRegister{
		FirstName: "Invalid", LastName: "Email", Email: "invalid-email-format", Password: "password123",
		Phone: 123, Address: "Addr", City: "City", Country: "Country", ZipCode: 123,
	}
	payloadBytes, _ := json.Marshal(regPayload)
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	http.HandlerFunc(testUserHandler.Register).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected 400 for invalid email format")
	var errorResp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Contains(t, errorResp["error"], "Invalid email format")
}
