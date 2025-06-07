package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	FirstName  string     `gorm:"type:text;not null" json:"first_name"`
	LastName   string     `gorm:"type:text;not null" json:"last_name"`
	MiddleName string     `gorm:"type:text" json:"middle_name"`
	Email      string     `gorm:"type:text;not null;uniqueIndex:idx_unique_email" json:"email"`
	Phone      int64      `gorm:"type:integer" json:"phone"`
	IsProducer bool       `gorm:"type:boolean;default:false" json:"is_producer"`
	Address    string     `gorm:"type:text" json:"addess"`
	City       string     `gorm:"type:text" json:"city"`
	Country    string     `gorm:"type:text" json:"country"`
	ZipCode    int32      `gorm:"type:integer" json:"zip_code"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"type:timestamp" json:"updated_at"`
}

type UserRegister struct {
	FirstName  string `gorm:"type:text;not null" json:"first_name"`
	LastName   string `gorm:"type:text;not null" json:"last_name"`
	MiddleName string `gorm:"type:text" json:"middle_name"`
	Email      string `gorm:"type:text;not null;uniqueIndex:idx_unique_email" json:"email"`
	Phone      int64  `gorm:"type:integer" json:"phone"`
	Password   string `gorm:"type:text;not null" json:"password"`
	IsProducer bool   `gorm:"type:boolean;default:false" json:"is_producer"`
	Address    string `gorm:"type:text" json:"addess"`
	City       string `gorm:"type:text" json:"city"`
	Country    string `gorm:"type:text" json:"country"`
	ZipCode    int32  `gorm:"type:integer" json:"zip_code"`
}

type AuthedUser struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	FirstName  string     `gorm:"type:text;not null" json:"first_name"`
	LastName   string     `gorm:"type:text;not null" json:"last_name"`
	MiddleName string     `gorm:"type:text" json:"middle_name"`
	Email      string     `gorm:"type:text;not null;uniqueIndex:idx_unique_email" json:"email"`
	Phone      int64      `gorm:"type:integer" json:"phone"`
	IsProducer bool       `gorm:"type:boolean;default:false" json:"is_producer"`
	Address    string     `gorm:"type:text" json:"addess"`
	City       string     `gorm:"type:text" json:"city"`
	Country    string     `gorm:"type:text" json:"country"`
	ZipCode    int32      `gorm:"type:integer" json:"zip_code"`
	CreatedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"type:timestamp" json:"updated_at"`
	Auth       Auth       `gorm:"foreignKey:UserID" json:"auth"`
}

type UpdateUser struct {
	FirstName  string `gorm:"type:text" json:"first_name"`
	LastName   string `gorm:"type:text" json:"last_name"`
	MiddleName string `gorm:"type:text" json:"middle_name"`
	Email      string `gorm:"type:text;uniqueIndex:idx_unique_email" json:"email"`
	Phone      int64  `gorm:"type:integer" json:"phone"`
	IsProducer bool   `gorm:"type:boolean;default:false" json:"is_producer"`
	Address    string `gorm:"type:text" json:"addess"`
	City       string `gorm:"type:text" json:"city"`
	Country    string `gorm:"type:text" json:"country"`
	ZipCode    int32  `gorm:"type:integer" json:"zip_code"`
}

type UserLogin struct {
	Email    string `gorm:"type:text;not null;uniqueIndex:idx_unique_email" json:"email"`
	Password string `gorm:"type:text;not null" json:"password"`
}

type Auth struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	Active    bool       `gorm:"type:boolean;default:false" json:"active"`
	Password  string     `gorm:"type:text;not null" json:"password"`
	UpdatedAt *time.Time `gorm:"type:timestamp" json:"updated_at"`
}
