package flat_type

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type flatTypeService struct {
	db *gorm.DB
}

func CreateFlatTypeService(app common.IApp) common.IService {
	return &flatTypeService{
		db: app.GetDBClient(),
	}
}