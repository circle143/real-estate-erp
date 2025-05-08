package models

import "github.com/google/uuid"

type TowerPaymentStatus struct {
	PaymentId   uuid.UUID    `gorm:"primaryKey" json:"paymentId"`
	TowerId     uuid.UUID    `gorm:"primaryKey" json:"towerId"`
	PaymentPlan *PaymentPlan `gorm:"foreignKey:PaymentId;not null" json:"paymentPlan,omitempty"`
	Tower       *Tower       `gorm:"foreignKey:TowerId;not null" json:"tower,omitempty"`
}

type SalePaymentStatus struct {
	PaymentId   uuid.UUID    `gorm:"primaryKey" json:"paymentId"`
	SaleId      uuid.UUID    `gorm:"primaryKey" json:"saleId"`
	PaymentPlan *PaymentPlan `gorm:"foreignKey:PaymentId;not null" json:"paymentPlan,omitempty"`
	Sale        *Sale        `gorm:"foreignKey:SaleId;not null" json:"sale,omitempty"`
	Amount      float64      `gorm:"not null" json:"amount"`
}
