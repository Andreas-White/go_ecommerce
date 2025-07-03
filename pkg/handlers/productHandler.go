package handlers

import (
	"encoding/json"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/services"
	"go_ecommerce/pkg/utils"
	"net/http"
)

type ProductHandler struct {
	Service services.IProductService
}

func NewProductHandler(service services.IProductService) *ProductHandler {
	return &ProductHandler{Service: service}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)
	err := h.Service.CreateProduct(ctx, &product, authUser.ID.String())
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/CreateProduct", http.StatusInternalServerError, "Failed to create product")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "Product created successfully"})
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	product, err := h.Service.GetProductByID(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) GetProductsByCategory(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Category is required")
		return
	}

	products, err := h.Service.GetProductsByCategory(r.Context(), category)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetProductsByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	products, err := h.Service.GetProductsByUserID(r.Context(), userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	sortBy := queryParams.Get("sortBy")
	sortOrder := queryParams.Get("sortOrder")
	searchTerm := queryParams.Get("search")

	products, err := h.Service.GetProducts(r.Context(), sortBy, sortOrder, searchTerm)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	product, err := h.Service.GetProductByID(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.Service.UpdateProduct(r.Context(), product)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	err := h.Service.DeleteProduct(r.Context(), id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}
