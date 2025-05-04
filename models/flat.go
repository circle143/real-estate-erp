package models

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"time"
)

// Flat model
type Flat struct {
	Id          uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid();constraint:OnUpdate:CASCADE" json:"id"`
	TowerId     uuid.UUID     `gorm:"not null;index" json:"towerId"`
	Tower       *Tower        `gorm:"foreignKey:TowerId;not null" json:"tower,omitempty"`
	FlatTypeId  uuid.UUID     `gorm:"not null;index" json:"flatTypeId"`
	FlatType    *FlatType     `gorm:"foreignKey:FlatTypeId;not null" json:"flatType,omitempty"`
	Name        string        `gorm:"not null" json:"name"`
	FloorNumber int           `gorm:"not null" json:"floorNumber"`
	Facing      custom.Facing `gorm:"not null" json:"facing"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updatedAt"`
	SaleDetail  *Sale         `gorm:"foreignKey:FlatId" json:"saleDetail,omitempty"`
}

func (u Flat) GetCreatedAt() time.Time {
	return u.CreatedAt
}
