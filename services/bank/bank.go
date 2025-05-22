package bank

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type bankService struct {
	db *gorm.DB
}

func CreateBankService(app common.IApp) common.IService {
	return &bankService{
		db: app.GetDBClient(),
	}
}
