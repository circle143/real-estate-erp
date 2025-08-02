package payment_plan

import (
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type paymentPlanSocietyInfoService struct {
	db        *gorm.DB
	paymentId uuid.UUID
}

func (s *paymentPlanSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	return nil, nil
	// fetch from db and return
	// paymentPlan := models.PaymentPlan{
	// 	Id: s.paymentId,
	// }
	//
	// err := s.db.First(&paymentPlan).Error
	// if err != nil {
	// 	return nil, err
	// }
	//
	// return &common.SocietyInfo{
	// 	OrgId:       paymentPlan.OrgId,
	// 	SocietyRera: paymentPlan.SocietyId,
	// }, nil
}

func CreatePaymentPlanSocietyInfoService(db *gorm.DB, paymentId uuid.UUID) common.ISocietyInfo {
	return &paymentPlanSocietyInfoService{
		db:        db,
		paymentId: paymentId,
	}
}
