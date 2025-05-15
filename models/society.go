package models

import (
	"github.com/google/uuid"
	"time"
)

// Society model
type Society struct {
	ReraNumber   string        `gorm:"primaryKey" json:"reraNumber"`
	OrgId        uuid.UUID     `gorm:"primaryKey" json:"orgId"`
	Organization *Organization `gorm:"foreignKey:OrgId;not null" json:"organization,omitempty"`
	Name         string        `gorm:"not null" json:"name"`
	Address      string        `gorm:"not null" json:"address"`
	CoverPhoto   string        `json:"coverPhoto"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`
	TotalFlats   int64         `gorm:"-" json:"totalFlats"`
	SoldFlats    int64         `gorm:"-" json:"soldFlats"`
	UnsoldFlats  int64         `gorm:"-" json:"unsoldFlats"`
	//DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (u Society) GetCreatedAt() time.Time {
	return u.CreatedAt
}
