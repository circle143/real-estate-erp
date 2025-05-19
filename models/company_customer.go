package models

import (
	"github.com/google/uuid"
	"time"
)

// CompanyCustomer is modeled for company as a buyer
type CompanyCustomer struct {
	Id           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SaleId       uuid.UUID `gorm:"not null;index" json:"saleId"`
	Name         string    `gorm:"required" json:"name"`
	AadharNumber string    `json:"aadharNumber"`
	PanNumber    string    `json:"panNumber"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (m CompanyCustomer) GetCreatedAt() time.Time { return m.CreatedAt }
