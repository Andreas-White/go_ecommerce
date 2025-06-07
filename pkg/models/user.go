package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	FirstName  string     `json:"first_name" db:"first_name"`
	LastName   string     `json:"last_name" db:"last_name"`
	MiddleName string     `json:"middle_name,omitempty" db:"middle_name"`
	Email      string     `json:"email" db:"email"`
	Phone      int64      `json:"phone" db:"phone"`
	IsProducer bool       `json:"is_producer" db:"is_producer"`
	Address    string     `json:"address" db:"address"`
	City       string     `json:"city" db:"city"`
	Country    string     `json:"country" db:"country"`
	ZipCode    int32      `json:"zip_code" db:"zip_code"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type UserDTO struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name,omitempty"`
	Email      string `json:"email"`
	Phone      int64  `json:"phone,omitempty"`
	Password   string `json:"password,omitempty"`
	IsProducer bool   `json:"is_producer"`
	Address    string `json:"address,omitempty"`
	City       string `json:"city,omitempty"`
	Country    string `json:"country,omitempty"`
	ZipCode    int32  `json:"zip_code,omitempty"`
}

type AuthedUser struct {
	ID         uuid.UUID  `json:"id"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	MiddleName string     `json:"middle_name,omitempty"`
	Email      string     `json:"email"`
	Phone      int64      `json:"phone"`
	IsProducer bool       `json:"is_producer"`
	Address    string     `json:"address"`
	City       string     `json:"city"`
	Country    string     `json:"country"`
	ZipCode    int32      `json:"zip_code"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	Auth       Auth       `json:"auth"`
}

type Auth struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	Active    bool       `json:"active" db:"active"`
	Password  string     `json:"-" db:"password"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
