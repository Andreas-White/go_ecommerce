package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	OrderID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_unique_payment_order" json:"order_id"`
	Amount        float64   `gorm:"type:numeric;not null" json:"amount"`
	PaymentMethod string    `gorm:"type:text;not null" json:"payment_method"`
	Status        string    `gorm:"type:text;not null;default:'pending'" json:"status"`
	TransactionID *string   `gorm:"type:text;uniqueIndex:idx_unique_transaction_id" json:"transaction_id,omitempty"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}
