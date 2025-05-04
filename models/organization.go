package models

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"time"
)

// Organization model
type Organization struct {
	Id        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	//DeletedAt gorm.DeletedAt            `gorm:"index"`
	Status custom.OrganizationStatus `gorm:"not null" json:"status"`
	Logo   string                    `json:"logo"`
	Gst    string                    `json:"gst"`
}

func (u Organization) GetCreatedAt() time.Time {
	return u.CreatedAt
}
