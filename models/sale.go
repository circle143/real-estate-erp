package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type SalePaid struct {
	SaleId          uuid.UUID
	TotalPaidAmount decimal.Decimal
}

type Sale struct {
	Id              uuid.UUID             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FlatId          uuid.UUID             `gorm:"not null" json:"flatId"`
	SocietyId       string                `gorm:"not null;index" json:"societyId"`
	OrgId           uuid.UUID             `gorm:"not null;index" json:"orgId"`
	Society         *Society              `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	TotalPrice      decimal.Decimal       `gorm:"not null;type:numeric" json:"totalPrice"`
	Paid            decimal.Decimal       `gorm:"-" json:"paid"` // used to compute paid amount during req lifecycle
	Remaining       decimal.Decimal       `gorm:"-" json:"remaining"`
	PriceBreakdown  PriceBreakdownDetails `gorm:"not null;type:jsonb" json:"priceBreakdown"`
	CreatedAt       time.Time             `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time             `gorm:"autoUpdateTime" json:"updatedAt"`
	Customers       []Customer            `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"owners,omitempty"`
	CompanyCustomer CompanyCustomer       `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"companyCustomer,omitempty"`
	//PaymentStatus  []SalePaymentStatus   `gorm:"foreignKey:SaleId" json:"paymentStatus,omitempty"`
	//DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (u Sale) GetCreatedAt() time.Time {
	return u.CreatedAt
}
