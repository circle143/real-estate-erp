package broker

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type brokerService struct {
	db *gorm.DB
}

func CreateBrokerService(app common.IApp) common.IService {
	return &brokerService{
		db: app.GetDBClient(),
	}
}
