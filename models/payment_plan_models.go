package models

import (
	"time"

	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
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

type PaymentPlanRatioItem struct {
	Id                 uuid.UUID                   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentPlanRatioId uuid.UUID                   `gorm:"not null" json:"paymentPlanRatioId"`
	PaymentPlanRatio   *PaymentPlanRatio           `gorm:"foreignKey:PaymentPlanRatioId" json:"PaymentPlanRatio"`
	Ratio              string                      `gorm:"not null" json:"ratio"`
	Scope              custom.PaymentPlanItemScope `gorm:"not null" json:"scope"`
	ConditionType      custom.PaymentPlanCondition `gorm:"not null" json:"conditionType"`
	ConditionValue     int                         `json:"conditionValue,omitempty"`
	Active             *bool                       `gorm:"-" json:"active,omitempty"`
	CreatedAt          time.Time                   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time                   `gorm:"autoUpdateTime" json:"updatedAt"`
}
