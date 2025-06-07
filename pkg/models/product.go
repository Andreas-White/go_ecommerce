package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string     `gorm:"type:text;not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	Price       float64    `gorm:"type:numeric;not null" json:"price"`
	Stock       int32      `gorm:"type:integer;not null" json:"stock"`
	Category    string     `gorm:"type:text;not null" json:"category_id"`
	ImageUrl    string     `gorm:"type:text" json:"image_url"`
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"type:timestamp" json:"updated_at"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
}
