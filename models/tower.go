package models

import (
	"github.com/google/uuid"
	"time"
)

// Tower model
type Tower struct {
	Id         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();constraint:OnUpdate:CASCADE" json:"id"`
	SocietyId  string    `gorm:"not null;index" json:"societyId"`
	OrgId      uuid.UUID `gorm:"not null;index" json:"orgId"`
	Society    *Society  `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null" json:"society,omitempty"`
	FloorCount int       `gorm:"not null" json:"floorCount"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (u Tower) GetCreatedAt() time.Time {
	return u.CreatedAt
}
