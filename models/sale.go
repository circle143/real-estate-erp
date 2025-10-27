package models

import (
	"time"

	"circledigital.in/real-state-erp/utils/custom"
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
	PaymentPlanRatio   *PaymentPlanRatio     `gorm:"foreignKey:PaymentPlanRatioId;constraint:OnUpdate:CASCADE" json:"PaymentPlanRatio"`
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

func (u Sale) Pending() decimal.Decimal {
	return u.GetTotalPayableAmount().Sub(u.PaidAmount())
}

func (u Sale) PaidAmount() decimal.Decimal {
	sum := decimal.Zero

	for _, receipt := range u.Receipts {
		if receipt.Mode != custom.ADJUSTMENT && receipt.Cleared != nil {
			sum = sum.Add(receipt.TotalAmount)
		}
	}

	return sum
}

func (u Sale) GetTotalPayableAmount() decimal.Decimal {
	payableAmount := u.TotalPrice

	for _, receipt := range u.Receipts {
		if receipt.Mode == custom.ADJUSTMENT {
			payableAmount = payableAmount.Add(receipt.TotalAmount)
		}
	}

	return payableAmount
}

func (u Sale) GetValidReceiptsCount() int {
	return len(u.Receipts)
}

// GetTotalCGST returns the total CGST as a string ("" if no value)
func (s Sale) GetTotalCGST() string {
	total := decimal.Zero
	for _, r := range s.Receipts {
		if r.CGST != nil {
			total = total.Add(*r.CGST)
		}
	}
	if total.Equal(decimal.Zero) {
		return ""
	}
	return total.String()
}

// GetTotalSGST returns the total SGST as a string ("" if no value)
func (s Sale) GetTotalSGST() string {
	total := decimal.Zero
	for _, r := range s.Receipts {
		if r.SGST != nil {
			total = total.Add(*r.SGST)
		}
	}
	if total.Equal(decimal.Zero) {
		return ""
	}
	return total.String()
}

// GetTotalServiceTax returns the total Service Tax as a string ("" if no value)
func (s Sale) GetTotalServiceTax() string {
	total := decimal.Zero
	for _, r := range s.Receipts {
		if r.ServiceTax != nil {
			total = total.Add(*r.ServiceTax)
		}
	}
	if total.Equal(decimal.Zero) {
		return ""
	}
	return total.String()
}

// GetTotalSwachhBharatCess returns the total Swachh Bharat Cess as a string ("" if no value)
func (s Sale) GetTotalSwachhBharatCess() string {
	total := decimal.Zero
	for _, r := range s.Receipts {
		if r.SwathchBharatCess != nil {
			total = total.Add(*r.SwathchBharatCess)
		}
	}
	if total.Equal(decimal.Zero) {
		return ""
	}
	return total.String()
}

// GetTotalKrishiKalyanCess returns the total Krishi Kalyan Cess as a string ("" if no value)
func (s Sale) GetTotalKrishiKalyanCess() string {
	total := decimal.Zero
	for _, r := range s.Receipts {
		if r.KrishiKalyanCess != nil {
			total = total.Add(*r.KrishiKalyanCess)
		}
	}
	if total.Equal(decimal.Zero) {
		return ""
	}
	return total.String()
}
