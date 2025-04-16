package models

import (
	"github.com/google/uuid"
	"time"
)

// FlatType model
type FlatType struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();constraint:OnUpdate:CASCADE" json:"id"`
	SocietyId string    `gorm:"not null;index" json:"societyId"`
	OrgId     uuid.UUID `gorm:"not null;index" json:"orgId"`
	Society   *Society  `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null" json:"society,omitempty"`
	Name      string    `gorm:"not null" json:"name"`
	Type      string    `gorm:"not null" json:"type"`
	Price     float64   `gorm:"not null" json:"price"`
	Area      float64   `gorm:"not null" json:"area"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (u FlatType) GetCreatedAt() time.Time {
	return u.CreatedAt
}