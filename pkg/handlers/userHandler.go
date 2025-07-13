package handlers

import (
	"encoding/json"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/services"
	"go_ecommerce/pkg/utils"
	"net/http"
	"strings"
)

// IUserHandler interface
type IUserHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	GetUserByID(w http.ResponseWriter, r *http.Request)
	GetUserByName(w http.ResponseWriter, r *http.Request)
	GetUserByEmail(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
}

// UserHandler struct
type UserHandler struct {
	UserService services.IUserService
}

func NewUserHandler(userService services.IUserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

// CreateUser handles the creation of a new user
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	isDevelopment := r.Header.Get("X-Development-Mode") == "true" || strings.Contains(r.Host, "localhost")
	if r.Method == http.MethodGet {
		// Set CSRF cookie for preflight
		csrfToken := utils.GenerateCSRFToken(32)
		csrfCookie := &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			Path:     "/",
			HttpOnly: false,
			Secure:   !isDevelopment, // For local dev; adjust as needed
			SameSite: http.SameSiteStrictMode,
			MaxAge:   60 * 60 * 24,
		}
		http.SetCookie(w, csrfCookie)
		w.WriteHeader(http.StatusNoContent)
		return
	}

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

	// Auto-login: generate JWT for the new user
	token, err := h.UserService.AuthenticateUser(ctx, user.Email, user.Password)
	if err == nil && token != "" {
		cookie := &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   !isDevelopment, // Set to false for local dev
			SameSite: http.SameSiteStrictMode,
			MaxAge:   60 * 60 * 24, // 1 day
		}
		http.SetCookie(w, cookie)

		csrfToken := utils.GenerateCSRFToken(32)
		csrfCookie := &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			Path:     "/",
			HttpOnly: false,
			Secure:   !isDevelopment, // Set to false for local dev
			SameSite: http.SameSiteStrictMode,
			MaxAge:   60 * 60 * 24,
		}
		http.SetCookie(w, csrfCookie)

		utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "User created successfully", "csrf_token": csrfToken})
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

	authUser := middleware.GetUserFromContext(r, w)

	var userPayload models.UserDTO
	if err := json.NewDecoder(r.Body).Decode(&userPayload); err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateUser", http.StatusBadRequest, "Invalid request payload")
		return
	}

	updatedUser, newToken, err := h.UserService.UpdateUser(ctx, authUser, &userPayload)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateUser", http.StatusInternalServerError, "Failed to update user")
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

	authUser := middleware.GetUserFromContext(r, w)

	err := h.UserService.DeleteUser(ctx, authUser.ID.String())
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteUser", http.StatusInternalServerError, "Failed to delete user")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

// Login handles user login and generates a JWT token
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	isDevelopment := r.Header.Get("X-Development-Mode") == "true" || strings.Contains(r.Host, "localhost")
	if r.Method == http.MethodGet {
		// Set CSRF cookie for preflight
		csrfToken := utils.GenerateCSRFToken(32)
		csrfCookie := &http.Cookie{
			Name:     "csrf_token",
			Value:    csrfToken,
			Path:     "/",
			HttpOnly: false,
			Secure:   !isDevelopment, // For local dev; adjust as needed
			SameSite: http.SameSiteStrictMode,
			MaxAge:   60 * 60 * 24,
		}
		http.SetCookie(w, csrfCookie)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var loginRequest models.UserDTO

	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		utils.HandleAPIErrors(err, w, "handler/Login", http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx := r.Context()

	// Generate JWT token for authenticated user
	token, err := h.UserService.AuthenticateUser(ctx, loginRequest.Email, loginRequest.Password)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/Login", http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Set JWT as httpOnly, Secure, SameSite cookie
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !isDevelopment, // Set to false for local dev
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24, // 1 day
	}
	http.SetCookie(w, cookie)

	// Generate CSRF token and set as non-httpOnly cookie
	csrfToken := utils.GenerateCSRFToken(32)
	csrfCookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   !isDevelopment, // Set to false for local dev
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24,
	}
	http.SetCookie(w, csrfCookie)

	// Optionally, return the CSRF token in the response body for initial JS access
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Login successful", "csrf_token": csrfToken})
}
