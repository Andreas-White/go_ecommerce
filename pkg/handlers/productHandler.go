package handlers

import (
	"encoding/json"
	"go_ecommerce/pkg/middleware"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/services"
	"go_ecommerce/pkg/utils"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type IProductHandler interface {
	CreateProduct(w http.ResponseWriter, r *http.Request)
	GetProduct(w http.ResponseWriter, r *http.Request)
	GetProductsByCategory(w http.ResponseWriter, r *http.Request)
	GetProductsByUserID(w http.ResponseWriter, r *http.Request)
	GetProducts(w http.ResponseWriter, r *http.Request)
	UpdateProduct(w http.ResponseWriter, r *http.Request)
	DeleteProduct(w http.ResponseWriter, r *http.Request)
}

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
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil || authUser.ID == uuid.Nil {
		utils.HandleAPIErrors(nil, w, "handler/GetProductsByUserID", http.StatusBadRequest, "User ID is required")
		return
	}

	products, err := h.Service.GetProductsByUserID(r.Context(), authUser.ID.String())
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

	limit := 0
	if queryParams.Get("limit") != "" {
		if l, err := strconv.Atoi(queryParams.Get("limit")); err == nil {
			limit = l
		}
	}

	products, err := h.Service.GetProducts(r.Context(), sortBy, sortOrder, searchTerm, limit)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetProducts", http.StatusInternalServerError, "Failed to get products")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/UpdateProduct", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.HandleAPIErrors(nil, w, "handler/UpdateProduct", http.StatusBadRequest, "Product ID is required")
		return
	}

	product, err := h.Service.GetProductByIDForAuth(ctx, id)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateProduct", http.StatusNotFound, "Product not found")
		return
	}

	// Check ownership
	if product.UserID != authUser.ID {
		utils.HandleAPIErrors(nil, w, "handler/UpdateProduct", http.StatusForbidden, "You can only update your own products")
		return
	}

	// Parse the update payload
	var updatePayload models.ProductDTO
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateProduct", http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Update the product fields
	product.Name = updatePayload.Name
	product.Description = updatePayload.Description
	product.Price = updatePayload.Price
	product.Stock = updatePayload.Stock
	product.Category = updatePayload.Category
	product.ImageUrl = updatePayload.ImageUrl

	err = h.Service.UpdateProduct(ctx, product)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateProduct", http.StatusInternalServerError, "Failed to update product")
		return
	}

	// Return the updated product as DTO
	updatedProductDTO, err := h.Service.GetProductByID(ctx, id)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateProduct", http.StatusInternalServerError, "Failed to get updated product")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, updatedProductDTO)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/DeleteProduct", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		utils.HandleAPIErrors(nil, w, "handler/DeleteProduct", http.StatusBadRequest, "Product ID is required")
		return
	}

	// Check ownership by getting the product first
	product, err := h.Service.GetProductByIDForAuth(ctx, id)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteProduct", http.StatusNotFound, "Product not found")
		return
	}

	// Check ownership
	if product.UserID != authUser.ID {
		utils.HandleAPIErrors(nil, w, "handler/DeleteProduct", http.StatusForbidden, "You can only delete your own products")
		return
	}

	err = h.Service.DeleteProduct(ctx, id)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteProduct", http.StatusInternalServerError, "Failed to delete product")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}
