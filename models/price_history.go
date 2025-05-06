package models

import (
	"github.com/google/uuid"
	"time"
)

// PriceHistory is append-only table
type PriceHistory struct {
	Id         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ChargeId   uuid.UUID `gorm:"not null" json:"-"`
	ChargeType string    `gorm:"not null" json:"-"`
	Price      float64   `gorm:"not null" json:"price"`
	ActiveFrom time.Time `gorm:"autoCreateTime;not null" json:"activeFrom"`
	ActiveTill time.Time `json:"activeTill,omitempty"`
	//DeletedAt  gorm.DeletedAt `gorm:"index"`
}
