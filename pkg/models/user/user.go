package models

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Phone int64     `json:"phone"`
}

type UserRegister struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    int64  `json:"phone"`
	Password string `json:"password"`
}

type AuthedUser struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Phone int64     `json:"phone"`
	Auth  Auth      `json:"auth"`
}

type Auth struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	CreationDate string    `json:"creation_date"`
	Active       bool      `json:"active"`
	Password     string    `json:"password"`
}
