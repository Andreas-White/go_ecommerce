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

type ReviewHandler struct {
	Service services.IReviewService
}

func NewReviewHandler(service services.IReviewService) *ReviewHandler {
	return &ReviewHandler{Service: service}
}

// AddReview handles POST /reviews
func (h *ReviewHandler) AddReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/AddReview", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	var dto models.ReviewDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.HandleAPIErrors(err, w, "handler/AddReview", http.StatusBadRequest, "Invalid request payload")
		return
	}

	if dto.Rating < 1 || dto.Rating > 5 {
		utils.HandleAPIErrors(nil, w, "handler/AddReview", http.StatusBadRequest, "Rating must be between 1 and 5")
		return
	}

	if dto.ProductID == uuid.Nil {
		utils.HandleAPIErrors(nil, w, "handler/AddReview", http.StatusBadRequest, "Product ID is required")
		return
	}

	err := h.Service.AddReview(ctx, authUser.ID, &dto)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/AddReview", http.StatusBadRequest, "Failed to add review")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Review submitted successfully"})
}

// GetReviewsByProductID handles GET /reviews?product_id=...
func (h *ReviewHandler) GetReviewsByProductID(w http.ResponseWriter, r *http.Request) {
	productIDStr := r.URL.Query().Get("product_id")
	if productIDStr == "" {
		utils.HandleAPIErrors(nil, w, "handler/GetReviewsByProductID", http.StatusBadRequest, "Product ID is required")
		return
	}
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetReviewsByProductID", http.StatusBadRequest, "Invalid product ID format")
		return
	}

	reviews, err := h.Service.GetReviewsByProductID(r.Context(), productID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetReviewsByProductID", http.StatusInternalServerError, "Failed to get reviews")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, reviews)
}

// UpdateReview handles PUT /reviews/update?id=...
func (h *ReviewHandler) UpdateReview(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/UpdateReview", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	// Get review ID from query param
	reviewIDStr := r.URL.Query().Get("id")
	if reviewIDStr == "" {
		utils.HandleAPIErrors(nil, w, "handler/UpdateReview", http.StatusBadRequest, "Review ID is required in query param")
		return
	}
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateReview", http.StatusBadRequest, "Invalid review ID format")
		return
	}

	var payload struct {
		Rating  int     `json:"rating"`
		Comment *string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateReview", http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.Service.UpdateReview(r.Context(), authUser.ID, reviewID, payload.Rating, payload.Comment)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateReview", http.StatusBadRequest, "Failed to update review")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Review updated successfully"})
}

// DeleteReview handles DELETE /reviews/delete?id=...
func (h *ReviewHandler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/DeleteReview", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	// Get review ID from query param
	reviewIDStr := r.URL.Query().Get("id")
	if reviewIDStr == "" {
		utils.HandleAPIErrors(nil, w, "handler/DeleteReview", http.StatusBadRequest, "Review ID is required in query param")
		return
	}
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteReview", http.StatusBadRequest, "Invalid review ID format")
		return
	}

	err = h.Service.DeleteReview(r.Context(), authUser.ID, reviewID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteReview", http.StatusBadRequest, "Failed to delete review")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Review deleted successfully"})
}
