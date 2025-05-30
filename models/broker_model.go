package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type Broker struct {
	Id           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId    string    `gorm:"not null;index;uniqueIndex:idx_society_aadhar;uniqueIndex:idx_society_pan" json:"societyId"`
	OrgId        uuid.UUID `gorm:"not null;index;uniqueIndex:idx_society_aadhar;uniqueIndex:idx_society_pan" json:"orgId"`
	Society      *Society  `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Name         string    `gorm:"not null" json:"name"`
	AadharNumber string    `gorm:"not null;uniqueIndex:idx_society_aadhar" json:"aadharNumber"`
	PanNumber    string    `gorm:"not null;uniqueIndex:idx_society_pan" json:"panNumber"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	Sales        []Sale    `gorm:"foreignKey:BrokerId" json:"sales,omitempty"`
}

func (u Broker) GetCreatedAt() time.Time {
	return u.CreatedAt
}

type BrokerReport struct {
	TotalAmount decimal.Decimal `json:"totalAmount"`
	Details     Broker          `json:"details"`
}
