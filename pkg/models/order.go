package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	TotalAmount   float64    `json:"total_amount" db:"total_amount"`
	Status        string     `json:"status" db:"status"`
	PaymentStatus string     `json:"payment_status" db:"payment_status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
