package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID            uuid.UUID `json:"id" db:"id"`
	OrderID       uuid.UUID `json:"order_id" db:"order_id"`
	Amount        float64   `json:"amount" db:"amount"`
	PaymentMethod string    `json:"payment_method" db:"payment_method"`
	Status        string    `json:"status" db:"status"`
	TransactionID *string   `json:"transaction_id,omitempty" db:"transaction_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
