package payment_plan_group

import (
	"errors"
	"net/http"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type paymentPlanSocietyInfoService struct {
	db        *gorm.DB
	paymentId uuid.UUID
}

func (s *paymentPlanSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	// fetch from db and return
	paymentPlan := models.PaymentPlanRatio{
		Id: s.paymentId,
	}

	err := s.db.First(&paymentPlan).Preload("PaymentPlanGroup").Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Payment id record not found",
			}
		}
		return nil, err
	}

	return &common.SocietyInfo{
		OrgId:       paymentPlan.PaymentPlanGroup.OrgId,
		SocietyRera: paymentPlan.PaymentPlanGroup.SocietyId,
	}, nil
}

func CreatePaymentPlanSocietyInfoService(db *gorm.DB, paymentId uuid.UUID) common.ISocietyInfo {
	return &paymentPlanSocietyInfoService{
		db:        db,
		paymentId: paymentId,
	}
}
