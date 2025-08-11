package models

import (
	"time"

	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type AmountWithGSTInclusive struct {
	CGST   decimal.Decimal
	SGST   decimal.Decimal
	Amount decimal.Decimal
}

type Receipt struct {
	Id                uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReceiptNumber     string             `gorm:"not null; uniqueIndex:idx_sale_receipt_number" json:"receiptNumber"`
	SaleId            uuid.UUID          `gorm:"not null; uniqueIndex:idx_sale_receipt_number" json:"saleId"`
	Sale              *Sale              `gorm:"foreignKey:SaleId;constraint:OnDelete:CASCADE" json:"sale,omitempty"`
	TotalAmount       decimal.Decimal    `gorm:"not null;type:numeric" json:"totalAmount"`
	Mode              custom.ReceiptMode `gorm:"not null" json:"mode"`
	DateIssued        pgtype.Date        `gorm:"not null" json:"dateIssued"`
	BankName          string             `json:"bankName"`
	TransactionNumber string             `json:"transactionNumber"`
	Failed            bool               `gorm:"not null;default:false" json:"failed"`
	Amount            decimal.Decimal    `gorm:"not null;type:numeric" json:"amount"`
	CGST              *decimal.Decimal   `gorm:"type:numeric" json:"cgst,omitempty"`
	SGST              *decimal.Decimal   `gorm:"type:numeric" json:"sgst,omitempty"`
	ServiceTax        *decimal.Decimal   `gorm:"type:numeric" json:"serviceTax,omitempty"`
	SwathchBharatCess *decimal.Decimal   `gorm:"type:numeric" json:"swatchBharatCess,omitempty"`
	KrishiKalyanCess  *decimal.Decimal   `gorm:"type:numeric" json:"krishiKalyanCess,omitempty"`
	Cleared           *ReceiptClear      `gorm:"foreignKey:ReceiptId;constraint:OnDelete:CASCADE" json:"cleared,omitempty"`
	CreatedAt         time.Time          `gorm:"autoCreateTime" json:"createdAt"`
}

func (r Receipt) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func calcGST(totalAmount decimal.Decimal, rate int) *AmountWithGSTInclusive {
	if rate <= 0 {
		return &AmountWithGSTInclusive{
			Amount: totalAmount,
			CGST:   decimal.Zero,
			SGST:   decimal.Zero,
		}
	}

	decimalHundred := decimal.NewFromInt(100)
	decimalRate := decimal.NewFromInt(int64(rate))
	gstAmount := totalAmount.Sub(totalAmount.Mul(decimalHundred.Div(decimalHundred.Add(decimalRate)))).Round(2)

	cgst := gstAmount.Div(decimal.NewFromInt(2))
	return &AmountWithGSTInclusive{
		Amount: totalAmount.Sub(gstAmount),
		CGST:   cgst,
		SGST:   cgst,
	}
}

func (r Receipt) CalcGST(rate int) *AmountWithGSTInclusive {
	return calcGST(r.TotalAmount, rate)
	// amount := r.TotalAmount.Div(decimal.NewFromFloat(1.05)).Round(2)
	// gstAmount := r.TotalAmount.Sub(amount)
	// decimalHundred := decimal.NewFromInt(100)
	// decimalRate := decimal.NewFromInt(int64(rate))
	// gstAmount := r.TotalAmount.Sub(r.TotalAmount.Mul(decimalHundred.Div(decimalHundred.Add(decimalRate))))
	//
	// cgst := gstAmount.Div(decimal.NewFromInt(2))
	// return &AmountWithGSTInclusive{
	// 	Amount: r.TotalAmount.Sub(gstAmount),
	// 	CGST:   cgst,
	// 	SGST:   cgst,
	// }
}

type ReceiptClear struct {
	ReceiptId uuid.UUID `gorm:"not null;uniqueIndex" json:"receiptId"`
	BankId    uuid.UUID `gorm:"not null" json:"bankId"`
	Bank      *Bank     `gorm:"foreignKey:BankId" json:"bank,omitempty"`
	Receipt   *Receipt  `gorm:"foreignKey:ReceiptId" json:"receipt,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}
