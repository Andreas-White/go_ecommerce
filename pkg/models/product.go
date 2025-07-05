package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description,omitempty" db:"description"`
	Price       float64    `json:"price" db:"price"`
	Stock       int32      `json:"stock" db:"stock"`
	Category    string     `json:"category_id" db:"category"`
	ImageUrl    string     `json:"image_url,omitempty" db:"image_url"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
}

type ProductDTO struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	Stock       int32   `json:"stock"`
	Category    string  `json:"category_id"`
	ImageUrl    string  `json:"image_url,omitempty"`
}
