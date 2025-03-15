package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	MiddleName string    `json:"middle_name"`
	Email      string    `json:"email"`
	Phone      int64     `json:"phone"`
	IsProducer bool      `json:"is_producer"`
}

type UserRegister struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Email      string `json:"email"`
	Phone      int64  `json:"phone"`
	Password   string `json:"password"`
	IsProducer bool   `json:"is_producer"`
}

type AuthedUser struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	MiddleName string    `json:"middle_name"`
	Email      string    `json:"email"`
	Phone      int64     `json:"phone"`
	IsProducer bool      `json:"is_producer"`
	Auth       Auth      `json:"auth"`
}

type Auth struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
	Password  string    `json:"password"`
}
