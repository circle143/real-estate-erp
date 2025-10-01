package payment_plan_group

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type paymentPlanService struct {
	db *gorm.DB
}

func CreatePaymentPlanService(app common.IApp) common.IService {
	return &paymentPlanService{
		db: app.GetDBClient(),
	}
}

type paymentPlanRatioItem struct {
	Description    string  `validate:"required"`
	Ratio          float64 `validate:"required,gt=0,lte=100"`
	Scope          string  `validate:"required"`
	ConditionType  string  `validate:"required"`
	ConditionValue int
}

type paymentPlanRatio struct {
	Items []paymentPlanRatioItem `validate:"required,dive"`
}
