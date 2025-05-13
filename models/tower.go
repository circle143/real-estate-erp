package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

// Tower model
type Tower struct {
	Id          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId   string          `gorm:"not null;index" json:"societyId"`
	OrgId       uuid.UUID       `gorm:"not null;index" json:"orgId"`
	Society     *Society        `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	FloorCount  int             `gorm:"not null" json:"floorCount"`
	Name        string          `gorm:"not null" json:"name"`
	TotalAmount decimal.Decimal `gorm:"-" json:"totalAmount"`
	PaidAmount  decimal.Decimal `gorm:"-" json:"paidAmount"`
	Remaining   decimal.Decimal `gorm:"-" json:"remaining"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	//DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (u Tower) GetCreatedAt() time.Time {
	return u.CreatedAt
}
