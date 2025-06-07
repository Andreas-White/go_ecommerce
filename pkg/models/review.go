package models

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ProductID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_unique_user_product_review,priority:2" json:"product_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_unique_user_product_review,priority:1" json:"user_id"`
	Rating    int        `gorm:"type:integer;not null" json:"rating"`
	Comment   *string    `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// Associations
	// User      User    `gorm:"foreignKey:UserID" json:"-"`
	// Product   Product `gorm:"foreignKey:ProductID" json:"-"`
}
