package models

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CartID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_unique_cart_product,priority:1" json:"cart_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_unique_cart_product,priority:2" json:"product_id"`
	Quantity  int       `gorm:"type:integer;not null" json:"quantity"`
	Price     float64   `gorm:"type:numeric;not null" json:"price"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
