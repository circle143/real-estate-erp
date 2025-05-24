package receipt

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/sale"
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type receiptSocietyInfoService struct {
	db        *gorm.DB
	receiptId uuid.UUID
}

func (s *receiptSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	receipt := models.Receipt{
		Id: s.receiptId,
	}

	err := s.db.First(&receipt).Error
	if err != nil {
		return nil, err
	}

	flatSocietyInfo := sale.CreateSaleSocietyInfoService(s.db, receipt.SaleId)
	return flatSocietyInfo.GetSocietyInfo()
}

func CreateReceiptSocietyInfoService(db *gorm.DB, receiptId uuid.UUID) common.ISocietyInfo {
	return &receiptSocietyInfoService{
		db:        db,
		receiptId: receiptId,
	}
}
