package sale

import (
	"circledigital.in/real-state-erp/utils/common"
	"gorm.io/gorm"
)

type saleBuyerType string

const (
	company saleBuyerType = "company"
	user    saleBuyerType = "user"
)

func (s saleBuyerType) IsValid() bool {
	switch s {
	case company, user:
		return true
	default:
		return false
	}
}

type saleService struct {
	db *gorm.DB
}

func CreateSaleService(app common.IApp) common.IService {
	return &saleService{
		db: app.GetDBClient(),
	}
}
