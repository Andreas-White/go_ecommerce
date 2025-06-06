package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/services"
	"go_ecommerce/pkg/utils"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// UserHandler struct
type UserHandler struct {
	UserService    services.IUserService
	TokenGenerator middleware.TokenGenerator
}

func NewUserHandler(userService services.IUserService, tokenGenerator middleware.TokenGenerator) *UserHandler {
	return &UserHandler{
		UserService:    userService,
		TokenGenerator: tokenGenerator,
	}
}

// CreateUser handles the creation of a new user
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.UserRegister
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.UserService.CreateUser(ctx, &user)
	if err != nil {
		h.handleErrors(err, w, "handler/CreateUser", "Failed to create user")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "User created successfully"})
}

// GetUserByID handles retrieving a user by their ID
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := checkUserFromContext(w, r)

	user, err := h.UserService.GetUserByID(ctx, authUser.ID.String())
	if err != nil {
		h.handleErrors(err, w, "handler/GetUserByID", "Failed to retrieve user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// GetUserByName handles retrieving a user by their name
func (h *UserHandler) GetUserByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := checkUserFromContext(w, r)

	user, err := h.UserService.GetUserByName(ctx, authUser.FirstName, authUser.LastName, authUser.MiddleName)
	if err != nil {
		h.handleErrors(err, w, "handler/GetUserByName", "Failed to retrieve user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// GetUserByEmail handles retrieving a user by their email
func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := checkUserFromContext(w, r)

	user, err := h.UserService.GetUserByEmail(ctx, authUser.Email)
	if err != nil {
		h.handleErrors(err, w, "handler/GetUserByEmail", "Failed to retrieve user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// UpdateUser handles updating user details
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := checkUserFromContext(w, r)

	var userPayload struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		MiddleName string `json:"middle_name"`
		Email      string `json:"email"`
		Phone      int64  `json:"phone"`
		IsProducer bool   `json:"is_producer"`
		Address    string `json:"address"`
		City       string `json:"city"`
		Country    string `json:"country"`
		ZipCode    int32  `json:"zip_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&userPayload); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	updatedUser := &models.User{
		ID:         authUser.ID,
		FirstName:  stringOrDefault(userPayload.FirstName, authUser.FirstName),
		LastName:   stringOrDefault(userPayload.LastName, authUser.LastName),
		MiddleName: stringOrDefault(userPayload.MiddleName, authUser.MiddleName),
		Email:      stringOrDefault(userPayload.Email, authUser.Email),
		Phone:      int64OrDefault(userPayload.Phone, authUser.Phone),
		IsProducer: boolOrDefault(userPayload.IsProducer, authUser.IsProducer),
		Address:    stringOrDefault(userPayload.Address, authUser.Address),
		City:       stringOrDefault(userPayload.City, authUser.City),
		Country:    stringOrDefault(userPayload.Country, authUser.Country),
		ZipCode:    int32OrDefault(userPayload.ZipCode, authUser.ZipCode),
	}

	err := h.UserService.UpdateUser(ctx, updatedUser)
	if err != nil {
		h.handleErrors(err, w, "handler/UpdateUser", "Failed to update user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

// DeleteUser handles deleting a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := checkUserFromContext(w, r)

	err := h.UserService.DeleteUser(ctx, authUser.ID.String())
	if err != nil {
		h.handleErrors(err, w, "handler/DeleteUser", "Failed to delete user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// Login handles user login and generates a JWT token
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx := r.Context()
	user, err := h.UserService.AuthenticateUser(ctx, loginRequest.Email, loginRequest.Password)
	if err != nil {
		h.handleErrors(err, w, "handler/Login", "Failed to authenticate user")
		return
	}

	// Generate JWT token for authenticated user
	token, err := h.TokenGenerator.GenerateJWT(user)
	if err != nil {
		log.Printf("{handler/Login - Error generating JWT token: %v}", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

func checkUserFromContext(w http.ResponseWriter, r *http.Request) models.User {
	authUser := middleware.GetUserFromContext(r)
	if authUser == nil || authUser.ID == uuid.Nil {
		log.Println("{handler/DeleteUser - User not found in context}")
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized user")
	}

	return *authUser
}

func stringOrDefault(newValue string, oldValue string) string {
	if newValue != "" {
		return newValue
	}

	return oldValue
}

func int64OrDefault(newValue int64, oldValue int64) int64 {
	if newValue != 0 {
		return newValue
	}

	return oldValue
}

func int32OrDefault(newValue int32, oldValue int32) int32 {
	if newValue != 0 {
		return newValue
	}

	return oldValue
}

func boolOrDefault(newValue bool, oldValue bool) bool {
	if newValue {
		return newValue
	}

	return oldValue
}

func (h *UserHandler) handleErrors(err error, w http.ResponseWriter, sourceFuncName string, genericErrorMessage string) {
	if errors.Is(err, context.Canceled) {
		log.Printf("{handler/Login - Request cancelled: %v}", err)
		utils.RespondWithError(w, http.StatusRequestTimeout, "Request cancelled")
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		log.Printf("{handler/Login - Request timed out: %v}", err)
		utils.RespondWithError(w, http.StatusGatewayTimeout, "Operation timed out")
		return
	}
	log.Printf("{%v - error: %v}", sourceFuncName, err)
	utils.RespondWithError(w, http.StatusUnauthorized, genericErrorMessage)
}
