package charges

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type chargesService struct {
	db *gorm.DB
}

func CreateChargesService(app common.IApp) common.IService {
	return &chargesService{
		db: app.GetDBClient(),
	}
}
