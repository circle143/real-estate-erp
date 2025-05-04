package models

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"time"
)

// PreferenceLocationCharge defines charges for society flats location
// if type is floor than floor (non-nullable) defines the floor which will have this charge
type PreferenceLocationCharge struct {
	Id        uuid.UUID                            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId string                               `gorm:"not null;index" json:"societyId"`
	OrgId     uuid.UUID                            `gorm:"not null;index" json:"orgId"`
	Society   *Society                             `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Summary   string                               `gorm:"not null" json:"summary"`
	Type      custom.PreferenceLocationChargesType `gorm:"not null" json:"type"`
	Floor     int                                  `json:"floor"`
	Price     float64                              `gorm:"not null" json:"price"`
	Disable   bool                                 `gorm:"not null;default:false" json:"disable"`
	CreatedAt time.Time                            `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time                            `gorm:"autoUpdateTime" json:"updatedAt"`
	//DeletedAt    gorm.DeletedAt                       `gorm:"index"`
	PriceHistory []PriceHistory `gorm:"polymorphicType:ChargeType;polymorphicId:ChargeId;polymorphicValue:location" json:"priceHistory;omitempty"`
}

func (u *PreferenceLocationCharge) GetCreatedAt() time.Time {
	return u.CreatedAt
}

type OtherCharge struct {
	Id            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId     string    `gorm:"not null;index" json:"societyId"`
	OrgId         uuid.UUID `gorm:"not null;index" json:"orgId"`
	Society       *Society  `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Summary       string    `gorm:"not null" json:"summary"`
	Recurring     bool      `gorm:"not null;default:false" json:"recurring"`
	Optional      bool      `gorm:"not null;default:false" json:"optional"`
	AdvanceMonths int       `json:"advanceMonths"` // in case of recurring charge defines advance required in months
	Price         float64   `gorm:"not null" json:"price"`
	Disable       bool      `gorm:"not null;default:false" json:"disable"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	//DeletedAt     gorm.DeletedAt `gorm:"index"`
	PriceHistory []PriceHistory `gorm:"polymorphicType:ChargeType;polymorphicId:ChargeId;polymorphicValue:other" json:"priceHistory;omitempty"`
}

func (u *OtherCharge) GetCreatedAt() time.Time {
	return u.CreatedAt
}
