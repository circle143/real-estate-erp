package flat

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type flatService struct {
	db *gorm.DB
}

func CreateFlatService(app common.IApp) common.IService {
	return &flatService{
		db: app.GetDBClient(),
	}
}
