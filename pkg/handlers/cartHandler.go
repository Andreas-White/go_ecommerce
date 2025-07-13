package handlers

import (
	"encoding/json"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/services"
	"go_ecommerce/pkg/utils"
	"net/http"

	"github.com/google/uuid"
)

const cartSessionCookie = "cart_session_id"

type ICartHandler interface {
	AddToCart(w http.ResponseWriter, r *http.Request)
	RemoveFromCart(w http.ResponseWriter, r *http.Request)
}

type CartHandler struct {
	cartService services.ICartService
}

func NewCartHandler(cartService services.ICartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var cartItemsDTO []models.CartItemDTO
	if err := json.NewDecoder(r.Body).Decode(&cartItemsDTO); err != nil {
		utils.HandleAPIErrors(err, w, "handler/AddToCart", http.StatusBadRequest, "Invalid request payload")
		return
	}

	if len(cartItemsDTO) == 0 {
		utils.HandleAPIErrors(nil, w, "handler/AddToCart", http.StatusBadRequest, "Request payload must contain at least one cart item")
		return
	}

	var userID *uuid.UUID
	var sessionID *string

	user := middleware.GetUserFromContext(r, w)
	if user != nil {
		userID = &user.ID
	} else {
		// Guest user - use session-based cart
		cookie, err := r.Cookie(cartSessionCookie)
		if err != nil || cookie.Value == "" {
			newSessionID := uuid.NewString()
			sessionID = &newSessionID
			http.SetCookie(w, &http.Cookie{
				Name:     cartSessionCookie,
				Value:    newSessionID,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		} else {
			existingSessionID := cookie.Value
			sessionID = &existingSessionID
		}
	}

	cart, err := h.cartService.AddProductsToCart(ctx, userID, sessionID, cartItemsDTO)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/AddToCart", http.StatusInternalServerError, "Failed to add products to cart")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, cart)
}

func (h *CartHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var cartItemDTO []models.CartItemDTO
	if err := json.NewDecoder(r.Body).Decode(&cartItemDTO); err != nil {
		utils.HandleAPIErrors(err, w, "handler/RemoveFromCart", http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get cart ID for the items
	var cartID uuid.UUID
	user := middleware.GetUserFromContext(r, w)
	if user != nil {
		// Authenticated user - get cart by user ID
		cart, err := h.cartService.GetCartByUserID(ctx, user.ID)
		if err != nil {
			utils.HandleAPIErrors(err, w, "handler/RemoveFromCart", http.StatusInternalServerError, "Failed to get user cart")
			return
		}
		cartID = cart.ID
	} else {
		// Guest user - get cart by session ID
		cookie, err := r.Cookie(cartSessionCookie)
		if err != nil || cookie.Value == "" {
			utils.HandleAPIErrors(nil, w, "handler/RemoveFromCart", http.StatusBadRequest, "No cart session found")
			return
		}
		cart, err := h.cartService.GetCartBySessionID(ctx, cookie.Value)
		if err != nil {
			utils.HandleAPIErrors(err, w, "handler/RemoveFromCart", http.StatusInternalServerError, "Failed to get guest cart")
			return
		}
		cartID = cart.ID
	}

	// Set cart ID for all items
	for i := range cartItemDTO {
		cartItemDTO[i].CartID = cartID
	}

	err := h.cartService.RemoveProductsFromCart(ctx, cartItemDTO)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/RemoveFromCart", http.StatusInternalServerError, "Failed to remove products from cart")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product removed from cart successfully"})
}

func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var cartID *uuid.UUID
	user := middleware.GetUserFromContext(r, w)
	if user != nil {
		// Authenticated user - get cart by user ID
		cart, err := h.cartService.GetCartByUserID(ctx, user.ID)
		if err != nil {
			utils.HandleAPIErrors(err, w, "handler/ClearCart", http.StatusInternalServerError, "Failed to get user cart")
			return
		}
		cartID = &cart.ID
	} else {
		// Guest user - get cart by session ID
		cookie, err := r.Cookie(cartSessionCookie)
		if err != nil || cookie.Value == "" {
			utils.HandleAPIErrors(nil, w, "handler/ClearCart", http.StatusBadRequest, "No cart session found")
			return
		}
		cart, err := h.cartService.GetCartBySessionID(ctx, cookie.Value)
		if err != nil {
			utils.HandleAPIErrors(err, w, "handler/ClearCart", http.StatusInternalServerError, "Failed to get guest cart")
			return
		}
		cartID = &cart.ID
	}

	err := h.cartService.ClearCart(ctx, cartID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/ClearCart", http.StatusInternalServerError, "Failed to clear cart")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Cart cleared successfully"})
}

func (h *CartHandler) GetCartItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var cartID uuid.UUID
	user := middleware.GetUserFromContext(r, w)
	if user != nil {
		// Authenticated user - get cart by user ID
		cart, err := h.cartService.GetCartByUserID(ctx, user.ID)
		if err != nil {
			utils.HandleAPIErrors(err, w, "handler/GetCartItems", http.StatusInternalServerError, "Failed to get user cart")
			return
		}
		cartID = cart.ID
	} else {
		// Guest user - get cart by session ID
		cookie, err := r.Cookie(cartSessionCookie)
		if err != nil || cookie.Value == "" {
			// Return empty cart for guest users without session
			utils.RespondWithJSON(w, http.StatusOK, []models.CartItemProductDetails{})
			return
		}
		cart, err := h.cartService.GetCartBySessionID(ctx, cookie.Value)
		if err != nil {
			// Return empty cart if session cart doesn't exist
			utils.RespondWithJSON(w, http.StatusOK, []models.CartItemProductDetails{})
			return
		}
		cartID = cart.ID
	}

	cartItems, err := h.cartService.GetAllCartItemsByCartID(ctx, cartID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetCartItems", http.StatusInternalServerError, "Failed to get cart items")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, cartItems)
}

func (h *CartHandler) UpdateCartItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var cartItemDTO []models.CartItemDTO
	if err := json.NewDecoder(r.Body).Decode(&cartItemDTO); err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateCartItems", http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get cart ID for the items
	var cartID uuid.UUID
	user := middleware.GetUserFromContext(r, w)
	if user != nil {
		// Authenticated user - get cart by user ID
		cart, err := h.cartService.GetCartByUserID(ctx, user.ID)
		if err != nil {
			utils.HandleAPIErrors(err, w, "handler/UpdateCartItems", http.StatusInternalServerError, "Failed to get user cart")
			return
		}
		cartID = cart.ID
	} else {
		// Guest user - get cart by session ID
		cookie, err := r.Cookie(cartSessionCookie)
		if err != nil || cookie.Value == "" {
			utils.HandleAPIErrors(nil, w, "handler/UpdateCartItems", http.StatusBadRequest, "No cart session found")
			return
		}
		cart, err := h.cartService.GetCartBySessionID(ctx, cookie.Value)
		if err != nil {
			utils.HandleAPIErrors(err, w, "handler/UpdateCartItems", http.StatusInternalServerError, "Failed to get guest cart")
			return
		}
		cartID = cart.ID
	}

	// Set cart ID for all items
	for i := range cartItemDTO {
		cartItemDTO[i].CartID = cartID
	}

	err := h.cartService.UpdateCartItems(ctx, cartItemDTO)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateCartItems", http.StatusInternalServerError, "Failed to update cart items")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Cart items updated successfully"})
}
