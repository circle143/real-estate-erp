package customer

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/flat"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type customerSocietyInfoService struct {
	db         *gorm.DB
	customerId uuid.UUID
	flatId     uuid.UUID
}

func (cs *customerSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	customer := models.Customer{
		Id: cs.customerId,
	}

	err := cs.db.First(&customer).Error
	if err != nil {
		return nil, err
	}

	if cs.flatId != customer.FlatId {
		return nil, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid flat customer.",
		}
	}
	flatSocietyInfo := flat.CreateFlatSocietyInfoService(cs.db, customer.FlatId)
	return flatSocietyInfo.GetSocietyInfo()
}

func CreateCustomerSocietyInfoService(db *gorm.DB, flatId, customerId uuid.UUID) common.ISocietyInfo {
	return &customerSocietyInfoService{
		db:         db,
		customerId: customerId,
		flatId:     flatId,
	}
}