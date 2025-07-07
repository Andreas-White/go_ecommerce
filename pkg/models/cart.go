package models

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
	SessionID *string    `json:"session_id,omitempty" db:"session_id"`
}

type CartItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CartID    uuid.UUID `json:"cart_id" db:"cart_id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CartItemProductDetails struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	CartID             uuid.UUID `json:"cart_id" db:"cart_id"`
	ProductID          uuid.UUID `json:"product_id" db:"product_id"`
	Quantity           int       `json:"quantity" db:"quantity"`
	Price              float64   `json:"price" db:"price"`
	ProductName        string    `json:"product_name" db:"product_name"`
	ProductDescription string    `json:"product_description,omitempty" db:"product_description"`
	ProductStock       int32     `json:"product_stock" db:"product_stock"`
	ProductCategory    string    `json:"product_category_id" db:"product_category_id"`
	ProductImageUrl    string    `json:"product_image_url,omitempty" db:"product_image_url"`
}

type CartItemDTO struct {
	CartID    uuid.UUID `json:"cart_id" `
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
}

type CartDeleteRequest struct {
	Cart  Cart          `json:"cart"`
	Items []CartItemDTO `json:"items"`
}
