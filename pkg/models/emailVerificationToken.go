package models

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerificationToken struct {
	TokenHash string    `json:"-" db:"token_hash"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}