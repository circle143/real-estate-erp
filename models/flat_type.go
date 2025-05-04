package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// FlatType model
// builtUpArea is calculated by adding reraCarpetArea and Balcony Area
// all the model prices are per sq ft of super-area
type FlatType struct {
	Id             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId      string         `gorm:"not null;index" json:"societyId"`
	OrgId          uuid.UUID      `gorm:"not null;index" json:"orgId"`
	Society        *Society       `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Name           string         `gorm:"not null" json:"name"`
	Accommodation  string         `gorm:"not null" json:"accommodation"`
	ReraCarpetArea float64        `gorm:"not null" json:"reraCarpetArea"`
	BalconyArea    float64        `gorm:"not null" json:"balconyArea"`
	BuiltUpArea    float64        `gorm:"not null" json:"builtUpArea"`
	SuperArea      float64        `gorm:"not null" json:"superArea"`
	Price          float64        `gorm:"not null" json:"price"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	PriceHistory   []PriceHistory `gorm:"polymorphicType:ChargeType;polymorphicId:ChargeId;polymorphicValue:flat-type" json:"priceHistory;omitempty"`
}

func (u FlatType) GetCreatedAt() time.Time {
	return u.CreatedAt
}
