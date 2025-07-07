package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"go_ecommerce/pkg/models"
	"go_ecommerce/pkg/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

type IProductRepository interface {
	CreateProduct(ctx context.Context, product *models.ProductDTO, userID string) (*models.Product, error)
	GetProductByID(ctx context.Context, id string) (*models.Product, error)
	GetProductsByCategory(ctx context.Context, category string) ([]models.Product, error)
	GetProductsByUserID(ctx context.Context, userID string) ([]models.Product, error)
	SearchProductsByNameAndDescription(ctx context.Context, searchTerm string) ([]models.Product, error)
	GetAllProducts(ctx context.Context, sortBy, sortOrder string) ([]models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	DeleteProduct(ctx context.Context, id string) error
}

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		DB: db,
	}
}

func (r *ProductRepository) CreateProduct(ctx context.Context, productDTO *models.ProductDTO, userID string) (*models.Product, error) {
	productID := uuid.New()
	now := time.Now()
	userIDUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/CreateProduct.ParseUserID", userID)
	}

	query := `
		INSERT INTO products (id, name, description, price, stock, category, image_url, created_at, updated_at, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = r.DB.ExecContext(ctx, query, productID, productDTO.Name, productDTO.Description, productDTO.Price, productDTO.Stock, productDTO.Category, productDTO.ImageUrl, now, nil, userIDUUID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/CreateProduct.ExecContext", userID)
	}

	createdProduct := &models.Product{
		ID:          productID,
		Name:        productDTO.Name,
		Description: productDTO.Description,
		Price:       productDTO.Price,
		Stock:       productDTO.Stock,
		Category:    productDTO.Category,
		ImageUrl:    productDTO.ImageUrl,
		CreatedAt:   now,
		UpdatedAt:   nil,
		UserID:      userIDUUID,
	}

	return createdProduct, nil
}

func (r *ProductRepository) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at, user_id FROM products WHERE id = $1
	`
	var product models.Product
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Category, &product.ImageUrl, &product.CreatedAt, &product.UpdatedAt, &product.UserID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetProductByID", id)
	}

	return &product, nil
}

func (r *ProductRepository) GetProductsByCategory(ctx context.Context, category string) ([]models.Product, error) {
	querry := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at, user_id FROM products WHERE category = $1
	`
	products, err := r.getListOfProductsByValue(ctx, querry, category)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetProductsByCategory", category)
	}

	return products, nil
}

func (r *ProductRepository) GetProductsByUserID(ctx context.Context, userID string) ([]models.Product, error) {
	querry := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at, user_id FROM products WHERE user_id = $1
	`
	products, err := r.getListOfProductsByValue(ctx, querry, userID)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetProductsByUserID", userID)
	}

	return products, nil
}

func (r *ProductRepository) SearchProductsByNameAndDescription(ctx context.Context, searchTerm string) ([]models.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category, image_url, created_at, updated_at, user_id FROM products WHERE name ILIKE $1 OR description ILIKE $1
	`
	searchValue := "%%" + searchTerm + "%%"

	rows, err := r.DB.QueryContext(ctx, query, searchValue)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/SearchProductsByNameAndDescription", searchTerm)
	}
	defer rows.Close()

	products, err := r.scanProducts(rows)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/SearchProductsByNameAndDescription", searchTerm)
	}

	return products, nil
}

func (r *ProductRepository) GetAllProducts(ctx context.Context, sortBy, sortOrder string) ([]models.Product, error) {
	query := `SELECT id, name, description, price, stock, category, image_url, created_at, updated_at, user_id FROM products`

	allowedSortBy := map[string]bool{
		"price":      true,
		"category":   true,
		"name":       true,
		"created_at": true,
	}
	if sortBy != "" && allowedSortBy[sortBy] {
		if strings.ToUpper(sortOrder) != "DESC" {
			sortOrder = "ASC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", sortBy, strings.ToUpper(sortOrder))
	}

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetAllProducts", "all")
	}
	defer rows.Close()

	products, err := r.scanProducts(rows)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/GetAllProducts", "all")
	}

	return products, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	updated_at := time.Now()
	query := `
		UPDATE products SET name = $2, description = $3, price = $4, stock = $5, category = $6, image_url = $7, updated_at = $8 WHERE id = $1
	`
	_, err := r.DB.ExecContext(ctx, query, product.ID, product.Name, product.Description, product.Price, product.Stock, product.Category, product.ImageUrl, updated_at)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/UpdateProduct", product.ID.String())
	}
	return nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, id string) error {
	query := `
		DELETE FROM products WHERE id = $1
	`
	_, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return utils.HandleRepositoryErrors(ctx, err, "repository/DeleteProduct", id)
	}
	return nil
}

func (r *ProductRepository) scanProducts(rows *sql.Rows) ([]models.Product, error) {
	var products []models.Product

	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Stock,
			&product.Category,
			&product.ImageUrl,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.UserID)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepository) getListOfProductsByValue(ctx context.Context, query string, value string) ([]models.Product, error) {
	rows, err := r.DB.QueryContext(ctx, query, value)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/getListOfProductsByValue", value)
	}
	defer rows.Close()

	products, err := r.scanProducts(rows)
	if err != nil {
		return nil, utils.HandleRepositoryErrors(ctx, err, "repository/getListOfProductsByValue", value)
	}

	return products, nil
}
