package services

import (
	"context"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/repositories"
)

type IProductService interface {
	CreateProduct(ctx context.Context, product *models.ProductDTO, userID string) (*models.Product, error)
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	GetProductsByCategory(ctx context.Context, category string) ([]models.Product, error)
	GetProductsByUserID(ctx context.Context, userID string) ([]models.Product, error)
	GetProducts(ctx context.Context, sortBy, sortOrder, searchTerm string) ([]models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

type ProductService struct {
	Repo repositories.IProductRepository
}

func NewProductService(repo repositories.IProductRepository) IProductService {
	return &ProductService{Repo: repo}
}

func (s *ProductService) CreateProduct(ctx context.Context, product *models.ProductDTO, userID string) (*models.Product, error) {
	return s.Repo.CreateProduct(ctx, product, userID)
}

func (s *ProductService) GetProductsByCategory(ctx context.Context, category string) ([]models.Product, error) {
	return s.Repo.GetProductsByCategory(ctx, category)
}

func (s *ProductService) GetProductsByUserID(ctx context.Context, userID string) ([]models.Product, error) {
	return s.Repo.GetProductsByUserID(ctx, userID)
}

func (s *ProductService) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	return s.Repo.GetProductByID(ctx, id)
}

func (s *ProductService) GetProducts(ctx context.Context, sortBy, sortOrder, searchTerm string) ([]models.Product, error) {
	if searchTerm != "" {
		return s.Repo.SearchProductsByNameAndDescription(ctx, searchTerm)
	}
	return s.Repo.GetAllProducts(ctx, sortBy, sortOrder)
}

func (s *ProductService) UpdateProduct(ctx context.Context, product *models.Product) error {
	return s.Repo.UpdateProduct(ctx, product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	return s.Repo.DeleteProduct(ctx, id)
}
