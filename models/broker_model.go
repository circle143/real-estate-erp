package models

import (
	"github.com/google/uuid"
	"time"
)

type Broker struct {
	Id           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId    string    `gorm:"not null;index;uniqueIndex:idx_society_aadhar;uniqueIndex:idx_society_pan" json:"societyId"`
	OrgId        uuid.UUID `gorm:"not null;index" json:"orgId"`
	Society      *Society  `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Name         string    `gorm:"not null" json:"name"`
	AadharNumber string    `gorm:"not null;uniqueIndex:idx_society_aadhar" json:"aadharNumber"`
	PanNumber    string    `gorm:"not null;uniqueIndex:idx_society_pan" json:"panNumber"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (u Broker) GetCreatedAt() time.Time {
	return u.CreatedAt
}
