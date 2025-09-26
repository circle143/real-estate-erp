package reports

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type reportService struct {
	db *gorm.DB
}

func NewReportService(app common.IApp) common.IService {
	return &reportService{
		db: app.GetDBClient(),
	}
}
