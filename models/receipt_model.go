package models

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type AmountWithGSTInclusive struct {
	CGST   decimal.Decimal
	SGST   decimal.Decimal
	Amount decimal.Decimal
}

type Receipt struct {
	Id                uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SaleId            uuid.UUID          `gorm:"not null" json:"saleId"`
	Sale              *Sale              `gorm:"foreignKey:SaleId;constraint:OnDelete:CASCADE" json:"sale,omitempty"`
	TotalAmount       decimal.Decimal    `gorm:"not null;type:numeric" json:"totalAmount"`
	Mode              custom.ReceiptMode `gorm:"not null" json:"mode"`
	DateIssued        custom.DateOnly    `gorm:"not null" json:"dateIssued"`
	BankName          string             `json:"bankName"`
	TransactionNumber string             `json:"transactionNumber"`
	Failed            bool               `gorm:"not null;default:false" json:"failed"`
	Amount            decimal.Decimal    `gorm:"not null;type:numeric" json:"amount"`
	CGST              decimal.Decimal    `gorm:"not null;type:numeric" json:"cgst"`
	SGST              decimal.Decimal    `gorm:"not null;type:numeric" json:"sgst"`
	Cleared           *ReceiptClear      `gorm:"foreignKey:ReceiptId;constraint:OnDelete:CASCADE" json:"cleared,omitempty"`
	CreatedAt         time.Time          `gorm:"autoCreateTime" json:"createdAt"`
}

func (r Receipt) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func (r Receipt) CalcGST() *AmountWithGSTInclusive {
	amount := r.TotalAmount.Div(decimal.NewFromFloat(1.05)).Round(2)
	gstAmount := r.TotalAmount.Sub(amount)

	cgst := gstAmount.Div(decimal.NewFromInt(2))
	return &AmountWithGSTInclusive{
		Amount: amount,
		CGST:   cgst,
		SGST:   cgst,
	}
}

type ReceiptClear struct {
	ReceiptId uuid.UUID `gorm:"not null;uniqueIndex" json:"receiptId"`
	BankId    uuid.UUID `gorm:"not null" json:"bankId"`
	Bank      *Bank     `gorm:"foreignKey:BankId" json:"bank,omitempty"`
}
