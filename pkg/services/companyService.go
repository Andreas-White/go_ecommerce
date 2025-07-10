package services

import (
	"context"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
	"go_ecommerce/pkg/utils"

	"github.com/google/uuid"
)

type ICompanyService interface {
	CreateCompany(ctx context.Context, company *models.CompanyDTO, userID uuid.UUID) error
	GetCompanyByCompanyID(ctx context.Context, companyID uuid.UUID) (*models.CompanyDTO, error)
	GetCompanyByUserID(ctx context.Context, userID uuid.UUID) (*models.CompanyDTO, error)
	UpdateCompany(ctx context.Context, company *models.CompanyDTO) error
	DeleteCompany(ctx context.Context, companyID uuid.UUID) error
}

type CompanyService struct {
	companyRepository repositories.ICompanyRepository
}

func NewCompanyService(companyRepository repositories.ICompanyRepository) ICompanyService {
	return &CompanyService{companyRepository: companyRepository}
}

func (s *CompanyService) CreateCompany(ctx context.Context, company *models.CompanyDTO, userID uuid.UUID) error {
	return s.companyRepository.CreateCompany(ctx, company, userID)
}

func (s *CompanyService) GetCompanyByCompanyID(ctx context.Context, companyID uuid.UUID) (*models.CompanyDTO, error) {
	company, err := s.companyRepository.GetCompanyByCompanyID(ctx, companyID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetCompanyByCompanyID")
	}

	return toCompanyDTO(company), nil
}

func (s *CompanyService) GetCompanyByUserID(ctx context.Context, userID uuid.UUID) (*models.CompanyDTO, error) {
	company, err := s.companyRepository.GetCompanyByUserID(ctx, userID)
	if err != nil {
		return nil, utils.HandleServiceErrors(ctx, err, "service/GetCompanyByUserID")
	}

	return toCompanyDTO(company), nil
}

func (s *CompanyService) UpdateCompany(ctx context.Context, company *models.CompanyDTO) error {
	return s.companyRepository.UpdateCompany(ctx, company)
}

func (s *CompanyService) DeleteCompany(ctx context.Context, companyID uuid.UUID) error {
	return s.companyRepository.DeleteCompany(ctx, companyID)
}

func toCompanyDTO(company *models.Company) *models.CompanyDTO {
	return &models.CompanyDTO{
		ID:            company.ID,
		Name:          company.Name,
		Address:       company.Address,
		City:          company.City,
		Country:       company.Country,
		ZipCode:       company.ZipCode,
		ReviewAverage: company.ReviewAverage,
		ReviewCount:   company.ReviewCount,
		CreatedAt:     company.CreatedAt,
		UpdatedAt:     company.UpdatedAt,
	}
}
