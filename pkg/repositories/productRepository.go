package repositories

import (
	"database/sql"
	"go_ecommerce/pkg/database"
	// "go_ecommerce/pkg/models"
)

type ProductRepository struct {
	DB *sql.DB
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		DB: database.DB,
	}
}

// func (r *ProductRepository) createProduct(product models.Product) error {

// }
