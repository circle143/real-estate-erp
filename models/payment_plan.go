package models

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type PaymentPlan struct {
	Id             uuid.UUID                   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId      string                      `gorm:"not null;index" json:"societyId"`
	OrgId          uuid.UUID                   `gorm:"not null;index" json:"orgId"`
	Society        *Society                    `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Scope          custom.PaymentPlanScope     `gorm:"not null" json:"scope"`
	Summary        string                      `gorm:"not null" json:"summary"`
	ConditionType  custom.PaymentPlanCondition `gorm:"not null" json:"conditionType"`
	ConditionValue int                         `json:"conditionValue,omitempty"`
	Amount         int                         `gorm:"not null" json:"amount"`
	Active         *bool                       `gorm:"-" json:"active,omitempty"`
	TotalAmount    *decimal.Decimal            `gorm:"-" json:"totalAmount,omitempty"`
	AmountPaid     *decimal.Decimal            `gorm:"-" json:"amountPaid,omitempty"`
	Remaining      *decimal.Decimal            `gorm:"-" json:"remaining,omitempty"` // used by flat payment breakdown
	Due            *time.Time                  `gorm:"-" json:"due,omitempty"`
	CreatedAt      time.Time                   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time                   `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (u PaymentPlan) GetCreatedAt() time.Time {
	return u.CreatedAt
}

type PaymentPlanSaleBreakDown struct {
	TotalAmount decimal.Decimal `json:"totalAmount"`
	PaidAmount  decimal.Decimal `json:"paidAmount"`
	Remaining   decimal.Decimal `json:"remaining"`
	Details     []PaymentPlan   `json:"details"`
}
