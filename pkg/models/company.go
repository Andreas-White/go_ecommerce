package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	Name          string     `gorm:"type:text;not null" json:"name"`
	Address       *string    `gorm:"type:text" json:"address,omitempty"`
	City          *string    `gorm:"type:text" json:"city,omitempty"`
	Country       *string    `gorm:"type:text" json:"country,omitempty"`
	ZipCode       *string    `gorm:"type:text" json:"zip_code,omitempty"`
	ReviewAverage float64    `gorm:"type:numeric(3,2);default:0.00" json:"review_average"`
	ReviewCount   int        `gorm:"type:integer;default:0" json:"review_count"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"type:timestamp" json:"updated_at,omitempty"`
}
