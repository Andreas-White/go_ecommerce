package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	Name          string     `json:"name" db:"name"`
	Address       *string    `json:"address,omitempty" db:"address"`
	City          *string    `json:"city,omitempty" db:"city"`
	Country       *string    `json:"country,omitempty" db:"country"`
	ZipCode       *string    `json:"zip_code,omitempty" db:"zip_code"`
	ReviewAverage float64    `json:"review_average" db:"review_average"`
	ReviewCount   int        `json:"review_count" db:"review_count"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
