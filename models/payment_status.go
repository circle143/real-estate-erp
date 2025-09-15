package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentReport struct {
	Total   decimal.Decimal `json:"total"`
	Paid    decimal.Decimal `json:"paid"`
	Pending decimal.Decimal `json:"pending"`
}

type TowerPaymentStatus struct {
	PaymentId            uuid.UUID             `gorm:"primaryKey" json:"paymentId"`
	TowerId              uuid.UUID             `gorm:"primaryKey" json:"towerId"`
	PaymentPlanRatioItem *PaymentPlanRatioItem `gorm:"foreignKey:PaymentId;not null" json:"paymentPlan,omitempty"`
	Tower                *Tower                `gorm:"foreignKey:TowerId;not null" json:"tower,omitempty"`
	CreatedAt            time.Time             `gorm:"autoCreateTime" json:"createdAt"`
}

type FlatPaymentStatus struct {
	PaymentId            uuid.UUID             `gorm:"primaryKey" json:"paymentId"`
	FlatId               uuid.UUID             `gorm:"primaryKey" json:"flatId"`
	PaymentPlanRatioItem *PaymentPlanRatioItem `gorm:"foreignKey:PaymentId;not null" json:"paymentPlan,omitempty"`
	Flat                 *Flat                 `gorm:"foreignKey:FlatId;not null" json:"flat,omitempty"`
	CreatedAt            time.Time             `gorm:"autoCreateTime" json:"createdAt"`
}

//type SalePaymentStatus struct {
//	PaymentId   uuid.UUID       `gorm:"primaryKey" json:"paymentId"`
//	SaleId      uuid.UUID       `gorm:"primaryKey" json:"saleId"`
//	PaymentPlan *PaymentPlan    `gorm:"foreignKey:PaymentId;not null" json:"paymentPlan,omitempty"`
//	Sale        *Sale           `gorm:"foreignKey:SaleId;not null" json:"sale,omitempty"`
//	Amount      decimal.Decimal `gorm:"not null;type:numeric" json:"amount"`
//	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"createdAt"`
//}
