package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Sale struct {
	Id                 uuid.UUID             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SaleNumber         string                `gorm:"not null;uniqueIndex" json:"saleNumber"`
	FlatId             uuid.UUID             `gorm:"not null" json:"flatId"`
	Flat               *Flat                 `gorm:"foreignKey:FlatId" json:"flat,omitempty"`
	SocietyId          string                `gorm:"not null;index" json:"societyId"`
	OrgId              uuid.UUID             `gorm:"not null;index" json:"orgId"`
	Society            *Society              `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	BrokerId           uuid.UUID             `gorm:"not null;index" json:"brokerId"`
	Broker             *Broker               `gorm:"foreignKey:BrokerId;not null;constraint:OnUpdate:CASCADE" json:"broker,omitempty"`
	PaymentPlanRatioId uuid.UUID             `json:"paymentPlanRatioId"`
	PaymentPlanRatio   *PaymentPlanRatio     `gorm:"foreignKey:PaymentPlanRatioId" json:"PaymentPlanRatio"`
	TotalPrice         decimal.Decimal       `gorm:"not null;type:numeric" json:"totalPrice"`
	Paid               *decimal.Decimal      `gorm:"-" json:"paid,omitempty"` // used to compute paid amount during req lifecycle
	Remaining          *decimal.Decimal      `gorm:"-" json:"remaining,omitempty"`
	TotalPayableAmount *decimal.Decimal      `gorm:"-" json:"totalPayableAmount,omitempty"` // used to compute paid amount during req lifecycle
	PriceBreakdown     PriceBreakdownDetails `gorm:"not null;type:jsonb" json:"priceBreakdown"`
	CreatedAt          time.Time             `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time             `gorm:"autoUpdateTime" json:"updatedAt"`
	Customers          []Customer            `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"owners,omitempty"`
	CompanyCustomer    *CompanyCustomer      `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"companyCustomer,omitempty"`
	Receipts           []Receipt             `gorm:"foreignKey:SaleId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"receipts,omitempty"`
	//PaymentStatus  []SalePaymentStatus   `gorm:"foreignKey:SaleId" json:"paymentStatus,omitempty"`
	//DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (u Sale) GetCreatedAt() time.Time {
	return u.CreatedAt
}
