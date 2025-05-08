package payment_plan

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
