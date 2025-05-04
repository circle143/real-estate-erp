package models

import (
	"github.com/google/uuid"
	"time"
)

// FlatType model
// builtUpArea is calculated by adding reraCarpetArea and Balcony Area
type FlatType struct {
	Id             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid();constraint:OnUpdate:CASCADE" json:"id"`
	SocietyId      string         `gorm:"not null;index" json:"societyId"`
	OrgId          uuid.UUID      `gorm:"not null;index" json:"orgId"`
	Society        *Society       `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null" json:"society,omitempty"`
	Accommodation  string         `gorm:"not null" json:"accommodation"`
	ReraCarpetArea float64        `gorm:"not null" json:"reraCarpetArea"`
	BalconyArea    float64        `gorm:"not null" json:"balconyArea"`
	BuiltUpArea    float64        `gorm:"-" json:"builtUpArea"`
	SuperArea      float64        `gorm:"not null" json:"superArea"`
	Price          float64        `gorm:"not null" json:"price"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	PriceHistory   []PriceHistory `gorm:"polymorphicType:ChargeType;polymorphicId:ChargeId;polymorphicValue:flat-type" json:"priceHistory;omitempty"`
}

func (u *FlatType) GetCreatedAt() time.Time {
	return u.CreatedAt
}

func (u *FlatType) calcBuiltUpArea() {
	u.BuiltUpArea = u.ReraCarpetArea + u.BalconyArea
}