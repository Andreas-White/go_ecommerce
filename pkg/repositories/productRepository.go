package repositories

import (
	"database/sql"
	"fmt"
	"go_ecommerce/pkg/database"
	"go_ecommerce/pkg/models"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		DB: database.DB,
	}
}

func (r *ProductRepository) CreateProduct(product models.Product) error {
	query := `
		INSERT INTO products (id, name, description, price, stock, category, image_url, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.DB.Exec(query, product.ID, product.Name, product.Description, product.Price, product.Stock, product.Category, product.ImageUrl, product.CreatedAt, product.UpdatedAt, product.UserID)
	if err != nil {
		return fmt.Errorf("error inserting product: %w", err)
	}

	return nil
}

func (r *ProductRepository) GetProductByID(id string) (*models.Product, error) {
	query := `
		SELECT * FROM products WHERE id = $1
	`
	var product models.Product
	err := r.DB.QueryRow(query, id).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.Category, &product.ImageUrl, &product.CreatedAt, &product.UpdatedAt, &product.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("error retrieving product: %v", err)
	}

	return &product, nil
}

func (r *ProductRepository) GetProductsByCategory(category string) ([]models.Product, error) {
	querry := `
		SELECT * FROM products WHERE category = $1
	`
	products, err := r.getListOfProductsByValue(querry, category)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("error retrieving product: %v", err)
	}

	return products, nil
}

func (r *ProductRepository) GetProductsByUserID(userID string) ([]models.Product, error) {
	querry := `
		SELECT * FROM products WHERE user_id = $1
	`
	products, err := r.getListOfProductsByValue(querry, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("error retrieving product: %v", err)
	}

	return products, nil
}

func (r *ProductRepository) getListOfProductsByValue(query string, value string) ([]models.Product, error) {
	rows, err := r.DB.Query(query, value)
	if err != nil {
		// Wrap error for context
		return nil, fmt.Errorf("failed to execute product query: %w", err)
	}

	defer rows.Close()

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
			// Wrap error for context
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
}
