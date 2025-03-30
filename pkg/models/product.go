package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Stock       int32      `json:"stock"`
	Category    string     `json:"category_id"`
	ImageUrl    string     `json:"image_url"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}
