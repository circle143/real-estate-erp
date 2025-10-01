package models

import (
	"time"

	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentPlanGroup struct {
	Id        uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SocietyId string             `gorm:"not null;index;uniqueIndex:idx_society_payment_plan_name;uniqueIndex:idx_society_payment_plan_abbr" json:"societyId"`
	OrgId     uuid.UUID          `gorm:"not null;index" json:"orgId"`
	Society   *Society           `gorm:"foreignKey:SocietyId,OrgId;references:ReraNumber,OrgId;not null;constraint:OnUpdate:CASCADE" json:"society,omitempty"`
	Name      string             `gorm:"not null;uniqueIndex:idx_society_payment_plan_name" json:"name"`
	Abbr      string             `gorm:"not null;uniqueIndex:idx_society_payment_plan_abbr" json:"abbr"`
	Ratios    []PaymentPlanRatio `gorm:"foreignKey:PaymentPlanGroupId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"ratios,omitempty"`
	CreatedAt time.Time          `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time          `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (u PaymentPlanGroup) GetCreatedAt() time.Time {
	return u.CreatedAt
}

type PaymentPlanRatio struct {
	Id                 uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentPlanGroupId uuid.UUID              `gorm:"not null" json:"paymentPlanGroupId"`
	PaymentPlanGroup   *PaymentPlanGroup      `gorm:"foreignKey:PaymentPlanGroupId" json:"PaymentPlanGroup"`
	Ratio              string                 `gorm:"not null" json:"ratio"`
	Ratios             []PaymentPlanRatioItem `gorm:"foreignKey:PaymentPlanRatioId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"items"`
	CreatedAt          time.Time              `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time              `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (p PaymentPlanRatio) GetRatioAmountDetail(ratioID uuid.UUID, totalPayableAmount, remaining decimal.Decimal, activeFlatPaymentPlans []FlatPaymentStatus, activeTowerPaymentPlans []TowerPaymentStatus) (*Finance, *time.Time) {
	for _, item := range p.Ratios {
		if item.Id == ratioID && item.IsActive(activeFlatPaymentPlans, activeTowerPaymentPlans) {
			// Found the matching item, calculate finance
			return item.GetAmountDetails(totalPayableAmount, remaining), &item.CreatedAt
		}
	}

	// If not found, return zeroed Finance
	return nil, nil
}

type PaymentPlanRatioItem struct {
	Id                 uuid.UUID                   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentPlanRatioId uuid.UUID                   `gorm:"not null" json:"paymentPlanRatioId"`
	PaymentPlanRatio   *PaymentPlanRatio           `gorm:"foreignKey:PaymentPlanRatioId" json:"PaymentPlanRatio"`
	Description        string                      `gorm:"not null;default:'payment plan item description'" json:"description"`
	Ratio              string                      `gorm:"not null" json:"ratio"`
	Scope              custom.PaymentPlanItemScope `gorm:"not null" json:"scope"`
	ConditionType      custom.PaymentPlanCondition `gorm:"not null" json:"conditionType"`
	ConditionValue     int                         `json:"conditionValue,omitempty"`
	Active             *bool                       `gorm:"-" json:"active,omitempty"`
	CreatedAt          time.Time                   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time                   `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (p PaymentPlanRatioItem) GetAmountDetails(totalPayableAmount, remaining decimal.Decimal) *Finance {
	// Convert integer percent to decimal
	ratio, ratioErr := decimal.NewFromString(p.Ratio)
	if ratioErr != nil {
		return nil
	}
	ratioDec := ratio.Div(decimal.NewFromInt(100))

	// Calculate total amount based on ratio
	total := totalPayableAmount.Mul(ratioDec)

	// Ensure paid does not exceed total
	paid := decimal.Min(remaining, total)
	pending := total.Sub(paid)

	return &Finance{
		Total:     total,
		Paid:      paid,
		Remaining: pending,
	}
}

func (p PaymentPlanRatioItem) IsActive(
	activeFlatPaymentPlans []FlatPaymentStatus,
	activeTowerPaymentPlans []TowerPaymentStatus,
) bool {
	now := time.Now()

	switch p.Scope {
	case custom.SCOPE_SALE:
		switch p.ConditionType {
		case custom.ONBOOKING:
			return true
		case custom.WITHINDAYS:
			target := p.CreatedAt.AddDate(0, 0, p.ConditionValue)
			return !now.Before(target) // active if current date >= created + days
		}
	case custom.SCOPE_FLAT:
		for _, plan := range activeFlatPaymentPlans {
			if plan.PaymentId == p.Id {
				return true
			}
		}
	case custom.SCOPE_TOWER:
		for _, plan := range activeTowerPaymentPlans {
			if plan.PaymentId == p.Id {
				return true
			}
		}
	}

	return false
}
