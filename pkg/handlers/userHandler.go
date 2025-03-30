package handlers

import (
	"encoding/json"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/services"
	"go_ecommerce/pkg/utils"
	"net/http"
)

// UserHandler struct
type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

// CreateUser handles the creation of a new user
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.UserRegister

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.UserService.CreateUser(&user)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "User created successfully"})
}

// GetUserByID handles retrieving a user by their ID
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.GetUserFromContext(r)

	user, err := h.UserService.GetUserByID(authUser.ID.String())
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// GetUserByName handles retrieving a user by their name
func (h *UserHandler) GetUserByName(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.GetUserFromContext(r)

	user, err := h.UserService.GetUserByName(authUser.FirstName, authUser.LastName, authUser.MiddleName)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// GetUserByEmail handles retrieving a user by their email
func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.GetUserFromContext(r)

	user, err := h.UserService.GetUserByEmail(authUser.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// UpdateUser handles updating user details
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.GetUserFromContext(r)

	if (authUser == &models.User{}) {
		utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

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

	err := h.UserService.UpdateUser(updatedUser)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User updated successfully"})
}

// DeleteUser handles deleting a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.GetUserFromContext(r)

	err := h.UserService.DeleteUser(authUser.ID.String())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
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

	// Authenticate user using the service
	user, err := h.UserService.AuthenticateUser(loginRequest.Email, loginRequest.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token for authenticated user
	token, err := middleware.GenerateJWT(user)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
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
