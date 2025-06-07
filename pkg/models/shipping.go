package models

import (
	"time"

	"github.com/google/uuid"
)

type Shipping struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	OrderID      uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_unique_shipping_order" json:"order_id"`
	Method       *string    `gorm:"type:text" json:"method,omitempty"`
	TrackingCode *string    `gorm:"type:text" json:"tracking_code,omitempty"`
	Cost         *float64   `gorm:"type:numeric" json:"cost,omitempty"`
	Address      string     `gorm:"type:text;not null" json:"address"`
	City         string     `gorm:"type:text;not null" json:"city"`
	Country      string     `gorm:"type:text;not null" json:"country"`
	ZipCode      string     `gorm:"type:text;not null" json:"zip_code"`
	ShippedAt    *time.Time `gorm:"type:timestamp" json:"shipped_at,omitempty"`
	DeliveredAt  *time.Time `gorm:"type:timestamp" json:"delivered_at,omitempty"`
	CreatedAt    time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"type:timestamp" json:"updated_at,omitempty"`

	// Associations
	// Order        Order      `gorm:"foreignKey:OrderID" json:"-"`
}
