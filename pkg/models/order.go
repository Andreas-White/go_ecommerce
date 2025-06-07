package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	TotalAmount   float64    `gorm:"type:numeric;not null" json:"total_amount"`
	Status        string     `gorm:"type:text;not null;default:'pending'" json:"status"`
	PaymentStatus string     `gorm:"type:text;not null;default:'pending'" json:"payment_status"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"type:timestamp" json:"updated_at,omitempty"`
}
