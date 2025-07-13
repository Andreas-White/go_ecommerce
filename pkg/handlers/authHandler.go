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

type IAuthHandler interface {
	ChangePassword(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type AuthHandler struct {
	AuthService services.IAuthService
}

func NewAuthHandler(authService services.IAuthService) IAuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser := middleware.GetUserFromContext(r, w)

	var changePasswordDTO models.ChangePasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&changePasswordDTO); err != nil {
		utils.HandleAPIErrors(err, w, "handler/ChangePassword", http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.AuthService.ChangePassword(ctx, authUser.ID.String(), &changePasswordDTO)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/ChangePassword", http.StatusInternalServerError, "Failed to change password")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	isDevelopment := r.Header.Get("X-Development-Mode") == "true" || strings.Contains(r.Host, "localhost")

	// Clear JWT cookie by setting it to expire immediately
	jwtCookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   !isDevelopment,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // This makes the cookie expire immediately
	}
	http.SetCookie(w, jwtCookie)

	// Clear CSRF cookie
	csrfCookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   !isDevelopment,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1, // This makes the cookie expire immediately
	}
	http.SetCookie(w, csrfCookie)

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
