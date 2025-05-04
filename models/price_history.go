package models

import (
	"github.com/google/uuid"
	"time"
)

type PriceHistory struct {
	Id         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();constraint:OnUpdate:CASCADE" json:"id"`
	ChargeId   uuid.UUID `gorm:"not null" json:"chargeId"`
	ChargeType string    `gorm:"not null" json:"chargeType"`
	Price      float64   `gorm:"not null" json:"price"`
	ActiveFrom time.Time `gorm:"autoCreateTime" json:"activeFrom"`
}