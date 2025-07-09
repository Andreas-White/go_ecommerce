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
	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)

	var product models.ProductDTO
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		utils.HandleAPIErrors(err, w, "handler/CreateProduct", http.StatusBadRequest, "Invalid request payload")
		return
	}

	createdProduct, err := h.Service.CreateProduct(ctx, &product, authUser.ID.String())
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/CreateProduct", http.StatusInternalServerError, "Failed to create product")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, createdProduct)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.HandleAPIErrors(nil, w, "handler/GetProduct", http.StatusBadRequest, "Product ID is required")
		return
	}

	product, err := h.Service.GetProductByID(r.Context(), id)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetProduct", http.StatusNotFound, "Product not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) GetProductsByCategory(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		utils.HandleAPIErrors(nil, w, "handler/GetProductsByCategory", http.StatusBadRequest, "Category is required")
		return
	}

	products, err := h.Service.GetProductsByCategory(r.Context(), category)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetProductsByCategory", http.StatusNotFound, "Products not found")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetProductsByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.HandleAPIErrors(nil, w, "handler/GetProductsByUserID", http.StatusBadRequest, "User ID is required")
		return
	}

	products, err := h.Service.GetProductsByUserID(r.Context(), userID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetProductsByUserID", http.StatusNotFound, "Products not found")
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
		utils.HandleAPIErrors(err, w, "handler/GetProducts", http.StatusInternalServerError, "Failed to get products")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.HandleAPIErrors(nil, w, "handler/UpdateProduct", http.StatusBadRequest, "Product ID is required")
		return
	}

	product, err := h.Service.GetProductByID(r.Context(), id)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateProduct", http.StatusNotFound, "Product not found")
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateProduct", http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.Service.UpdateProduct(r.Context(), product)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateProduct", http.StatusInternalServerError, "Failed to update product")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		utils.HandleAPIErrors(nil, w, "handler/DeleteProduct", http.StatusBadRequest, "Product ID is required")
		return
	}

	err := h.Service.DeleteProduct(r.Context(), id)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteProduct", http.StatusInternalServerError, "Failed to delete product")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}
