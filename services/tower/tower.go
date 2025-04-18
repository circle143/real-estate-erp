package tower

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type towerService struct {
	db *gorm.DB
}

func CreateTowerService(app common.IApp) common.IService {
	return &towerService{
		db: app.GetDBClient(),
	}
}