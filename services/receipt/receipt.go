package receipt

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type receiptService struct {
	db *gorm.DB
}

func CreateReceiptService(app common.IApp) common.IService {
	return &receiptService{
		db: app.GetDBClient(),
	}
}
