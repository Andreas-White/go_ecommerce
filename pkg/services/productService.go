package services

import (
	"context"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
)

type IProductService interface {
	CreateProduct(ctx context.Context, product *models.ProductDTO, userID string) (*models.Product, error)
	GetProductByID(ctx context.Context, id string) (*models.ProductDTO, error)
	GetProductByIDForAuth(ctx context.Context, id string) (*models.Product, error)
	GetProductsByCategory(ctx context.Context, category string) ([]models.ProductDTO, error)
	GetProductsByUserID(ctx context.Context, userID string) ([]models.Product, error)
	GetProducts(ctx context.Context, sortBy, sortOrder, searchTerm string, limit int) ([]models.ProductDTO, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

type ProductService struct {
	Repo        repositories.IProductRepository
	CompanyRepo repositories.ICompanyRepository
	ReviewRepo  repositories.IReviewRepository
}

func NewProductService(repo repositories.IProductRepository, companyRepo repositories.ICompanyRepository, reviewRepo repositories.IReviewRepository) IProductService {
	return &ProductService{
		Repo:        repo,
		CompanyRepo: companyRepo,
		ReviewRepo:  reviewRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.ProductDTO, userID string) (*models.Product, error) {
	return s.Repo.CreateProduct(ctx, product, userID)
}

func (s *ProductService) GetProductsByCategory(ctx context.Context, category string) ([]models.ProductDTO, error) {
	products, err := s.Repo.GetProductsByCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	return s.toProductsDTO(ctx, products)
}

func (s *ProductService) GetProductsByUserID(ctx context.Context, userID string) ([]models.Product, error) {
	return s.Repo.GetProductsByUserID(ctx, userID)
}

func (s *ProductService) GetProductByID(ctx context.Context, id string) (*models.ProductDTO, error) {
	product, err := s.Repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toProductDTO(ctx, product)
}

func (s *ProductService) GetProductByIDForAuth(ctx context.Context, id string) (*models.Product, error) {
	return s.Repo.GetProductByID(ctx, id)
}

func (s *ProductService) GetProducts(ctx context.Context, sortBy, sortOrder, searchTerm string, limit int) ([]models.ProductDTO, error) {
	var products []models.Product
	var err error

	if limit > 0 && searchTerm != "" {
		products, err = s.Repo.SearchProductsByName(ctx, searchTerm, limit)
	} else if searchTerm != "" {
		products, err = s.Repo.SearchProductsByNameAndDescription(ctx, searchTerm)
	} else {
		products, err = s.Repo.GetAllProducts(ctx, sortBy, sortOrder)
	}

	if err != nil {
		return nil, err
	}

	return s.toProductsDTO(ctx, products)
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	return s.Repo.UpdateProduct(ctx, product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.Repo.DeleteProduct(ctx, id)
}

// toProductDTO converts a Product to ProductDTO with company information
func (s *ProductService) toProductDTO(ctx context.Context, product *models.Product) (*models.ProductDTO, error) {
	company, err := s.CompanyRepo.GetCompanyByUserID(ctx, product.UserID)
	if err != nil {
		companyDTO := models.CompanyDTO{}
		rating, reviewCount, _ := s.ReviewRepo.GetProductReviewStats(ctx, product.ID)
		return &models.ProductDTO{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			Category:    product.Category,
			ImageUrl:    product.ImageUrl,
			Company:     companyDTO,
			Rating:      rating,
			ReviewCount: reviewCount,
		}, nil
	}

	companyDTO := models.CompanyDTO{
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

	rating, reviewCount, _ := s.ReviewRepo.GetProductReviewStats(ctx, product.ID)

	return &models.ProductDTO{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Category:    product.Category,
		ImageUrl:    product.ImageUrl,
		Company:     companyDTO,
		Rating:      rating,
		ReviewCount: reviewCount,
	}, nil
}

// toProductsDTO converts a slice of Products to ProductDTOs with company information
func (s *ProductService) toProductsDTO(ctx context.Context, products []models.Product) ([]models.ProductDTO, error) {
	var productDTOs []models.ProductDTO

	for _, product := range products {
		dto, err := s.toProductDTO(ctx, &product)
		if err != nil {
			return nil, err
		}
		productDTOs = append(productDTOs, *dto)
	}

	return productDTOs, nil
}
