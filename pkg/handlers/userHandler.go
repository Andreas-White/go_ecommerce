package handlers

import (
	"encoding/json"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/services"
	"go_ecommerce/pkg/utils"
	"log"
	"net/http"
	"strings"
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
	var user models.UserDTO
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.UserService.CreateUser(ctx, &user)
	if err != nil {
		if strings.Contains(err.Error(), "password must be at least 8 characters long") {
			utils.HandleAPIErrors(err, w, "handler/Register", http.StatusBadRequest, "Password must be at least 8 characters long")
			return
		}
		if strings.Contains(err.Error(), "invalid email format") {
			utils.HandleAPIErrors(err, w, "handler/Register", http.StatusBadRequest, "Invalid email format")
			return
		}
		utils.HandleAPIErrors(err, w, "handler/Register", http.StatusInternalServerError, "Failed to create user")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "User created successfully"})
}

// GetUserByID handles retrieving a user by their ID
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := middleware.GetUserFromContext(r, w)

	user, err := h.UserService.GetUserByID(ctx, authUser.ID.String())
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetUserByID", http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// GetUserByName handles retrieving a user by their name
func (h *UserHandler) GetUserByName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := middleware.GetUserFromContext(r, w)

	user, err := h.UserService.GetUserByName(ctx, authUser.FirstName, authUser.LastName, authUser.MiddleName)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetUserByName", http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// GetUserByEmail handles retrieving a user by their email
func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := middleware.GetUserFromContext(r, w)

	user, err := h.UserService.GetUserByEmail(ctx, authUser.Email)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetUserByEmail", http.StatusInternalServerError, "Failed to retrieve user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

// UpdateUser handles updating user details
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := *middleware.GetUserFromContext(r, w)

	var userPayload models.UserDTO
	if err := json.NewDecoder(r.Body).Decode(&userPayload); err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateUser", http.StatusBadRequest, "Invalid request payload")
		return
	}

	updatedUser, err := h.UserService.UpdateUser(ctx, &authUser, &userPayload)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateUser", http.StatusInternalServerError, "Failed to update user")
		return
	}

	newToken, err := h.TokenGenerator.GenerateJWT(updatedUser)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateUser", http.StatusInternalServerError, "Failed to generate new token after update")
		return
	}

	response := map[string]interface{}{
		"message": "User updated successfully",
		"user":    updatedUser,
		"token":   newToken,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// DeleteUser handles deleting a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := *middleware.GetUserFromContext(r, w)

	err := h.UserService.DeleteUser(ctx, authUser.ID.String())
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteUser", http.StatusInternalServerError, "Failed to delete user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// Login handles user login and generates a JWT token
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest models.UserDTO

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		utils.HandleAPIErrors(err, w, "handler/Login", http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx := r.Context()
	user, err := h.UserService.AuthenticateUser(ctx, loginRequest.Email, loginRequest.Password)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/Login", http.StatusUnauthorized, "Failed to authenticate user")
		return
	}

	// Generate JWT token for authenticated user
	token, err := h.TokenGenerator.GenerateJWT(user)
	if err != nil {
		log.Printf("{handler/Login - Error generating JWT token: %v}", err)
		utils.HandleAPIErrors(err, w, "handler/Login", http.StatusInternalServerError, "Internal server error")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}
