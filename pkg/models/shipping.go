package models

import (
	"time"

	"github.com/google/uuid"
)

type Shipping struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OrderID      uuid.UUID  `json:"order_id" db:"order_id"`
	Method       *string    `json:"method,omitempty" db:"method"`
	TrackingCode *string    `json:"tracking_code,omitempty" db:"tracking_code"`
	Cost         *float64   `json:"cost,omitempty" db:"cost"`
	Address      string     `json:"address" db:"address"`
	City         string     `json:"city" db:"city"`
	Country      string     `json:"country" db:"country"`
	ZipCode      string     `json:"zip_code" db:"zip_code"`
	ShippedAt    *time.Time `json:"shipped_at,omitempty" db:"shipped_at"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
