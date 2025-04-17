package society

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type societyService struct {
	db *gorm.DB
}

// CreateSocietyService is an abstract factory to create organization service
func CreateSocietyService(app common.IApp) common.IService {
	return &societyService{
		db: app.GetDBClient(),
	}
}