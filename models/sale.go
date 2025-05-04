package models

import (
	"github.com/google/uuid"
	"time"
)

type Sale struct {
	Id             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FlatId         uuid.UUID  `gorm:"not null" json:"flatId"`
	SocietyId      string     `gorm:"not null;index" json:"societyId"`
	OrgId          uuid.UUID  `gorm:"not null;index" json:"orgId"`
	Society        *Society   `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	TotalPrice     float64    `gorm:"not null" json:"totalPrice"`
	PriceBreakdown any        `gorm:"not null;type:JSONB" json:"priceBreakdown"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	Customers      []Customer `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"owners,omitempty"`
}

// add total price and price breakdown

func (u *Sale) GetCreatedAt() time.Time {
	return u.CreatedAt
}
