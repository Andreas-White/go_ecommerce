package models

import (
	"time"

	"github.com/google/uuid"
)

type UserAuthProvider struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Provider    string    `json:"provider" db:"provider"`
	ProviderID  string    `json:"provider_id" db:"provider_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}