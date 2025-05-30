package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type Bank struct {
	Id              uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId       string         `gorm:"not null;index;uniqueIndex:idx_society_bank_account_number" json:"societyId"`
	OrgId           uuid.UUID      `gorm:"not null;index;uniqueIndex:idx_society_bank_account_number" json:"orgId"`
	Society         *Society       `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Name            string         `gorm:"not null" json:"name"`
	AccountNumber   string         `gorm:"not null;uniqueIndex:idx_society_bank_account_number" json:"accountNumber"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	ClearedReceipts []ReceiptClear `gorm:"foreignKey:BankId" json:"clearedReceipts,omitempty"`
}

func (u Bank) GetCreatedAt() time.Time {
	return u.CreatedAt
}

type BankReport struct {
	TotalAmount decimal.Decimal `json:"totalAmount"`
	Details     Bank
}
