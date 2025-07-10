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

type ICompanyHandler interface {
	CreateCompany(w http.ResponseWriter, r *http.Request)
	GetCompanyByCompanyID(w http.ResponseWriter, r *http.Request)
	GetCompanyByUserID(w http.ResponseWriter, r *http.Request)
	UpdateCompany(w http.ResponseWriter, r *http.Request)
	DeleteCompany(w http.ResponseWriter, r *http.Request)
}

type CompanyHandler struct {
	companyService services.ICompanyService
}

func NewCompanyHandler(companyService services.ICompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: companyService}
}

func (h *CompanyHandler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/CreateCompany", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	var companyDTO models.CompanyDTO
	if err := json.NewDecoder(r.Body).Decode(&companyDTO); err != nil {
		utils.HandleAPIErrors(err, w, "handler/CreateCompany", http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.companyService.CreateCompany(ctx, &companyDTO, authUser.ID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/CreateCompany", http.StatusInternalServerError, "Failed to create company")
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, companyDTO)
}

func (h *CompanyHandler) GetCompanyByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/GetCompanyByUserID", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	company, err := h.companyService.GetCompanyByUserID(ctx, authUser.ID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetCompanyByUserID", http.StatusInternalServerError, "Failed to get company")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) GetCompanyByCompanyID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		utils.HandleAPIErrors(nil, w, "handler/GetCompanyByCompanyID", http.StatusBadRequest, "Company ID is required")
		return
	}

	parsedCompanyID, err := uuid.Parse(companyID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetCompanyByCompanyID", http.StatusBadRequest, "Invalid company ID format")
		return
	}

	company, err := h.companyService.GetCompanyByCompanyID(ctx, parsedCompanyID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/GetCompanyByCompanyID", http.StatusInternalServerError, "Failed to get company")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/UpdateCompany", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	var companyDTO models.CompanyDTO
	if err := json.NewDecoder(r.Body).Decode(&companyDTO); err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateCompany", http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.companyService.UpdateCompany(ctx, &companyDTO)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/UpdateCompany", http.StatusInternalServerError, "Failed to update company")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, companyDTO)
}

func (h *CompanyHandler) DeleteCompany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authUser := middleware.GetUserFromContext(r, w)
	if authUser == nil {
		utils.HandleAPIErrors(nil, w, "handler/DeleteCompany", http.StatusUnauthorized, "User must be authenticated")
		return
	}

	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		utils.HandleAPIErrors(nil, w, "handler/DeleteCompany", http.StatusBadRequest, "Company ID is required")
		return
	}

	parsedCompanyID, err := uuid.Parse(companyID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteCompany", http.StatusBadRequest, "Invalid company ID format")
		return
	}

	err = h.companyService.DeleteCompany(ctx, parsedCompanyID)
	if err != nil {
		utils.HandleAPIErrors(err, w, "handler/DeleteCompany", http.StatusInternalServerError, "Failed to delete company")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, "Company deleted successfully")
}
