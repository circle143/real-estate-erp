package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

// FlatType model
// builtUpArea is calculated by adding reraCarpetArea and Balcony Area
// all the model prices are per sq ft of super-area
type FlatType struct {
	Id             uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId      string          `gorm:"not null;index" json:"societyId"`
	OrgId          uuid.UUID       `gorm:"not null;index" json:"orgId"`
	Society        *Society        `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Name           string          `gorm:"not null" json:"name"`
	Accommodation  string          `gorm:"not null" json:"accommodation"`
	ReraCarpetArea decimal.Decimal `gorm:"not null;type:numeric" json:"reraCarpetArea"`
	BalconyArea    decimal.Decimal `gorm:"not null;type:numeric" json:"balconyArea"`
	BuiltUpArea    decimal.Decimal `gorm:"not null;type:numeric" json:"builtUpArea"`
	SuperArea      decimal.Decimal `gorm:"not null;type:numeric" json:"superArea"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	//DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (u FlatType) GetCreatedAt() time.Time {
	return u.CreatedAt
}
