package sale

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/flat"
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type saleSocietyInfoService struct {
	db     *gorm.DB
	saleId uuid.UUID
}

func (s *saleSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	sale := models.Sale{
		Id: s.saleId,
	}

	err := s.db.First(&sale).Error
	if err != nil {
		return nil, err
	}

	flatSocietyInfo := flat.CreateFlatSocietyInfoService(s.db, sale.FlatId)
	return flatSocietyInfo.GetSocietyInfo()
}

func CreateSaleSocietyInfoService(db *gorm.DB, saleId uuid.UUID) common.ISocietyInfo {
	return &saleSocietyInfoService{
		db:     db,
		saleId: saleId,
	}
}
