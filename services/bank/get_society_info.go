package bank

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bankSocietyInfoService struct {
	db     *gorm.DB
	bankId uuid.UUID
}

func (s *bankSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	// fetch from db and return
	bank := models.Bank{
		Id: s.bankId,
	}

	err := s.db.First(&bank).Error
	if err != nil {
		return nil, err
	}

	return &common.SocietyInfo{
		OrgId:       bank.OrgId,
		SocietyRera: bank.SocietyId,
	}, nil
}

func CreateBankSocietyInfoService(db *gorm.DB, bankId uuid.UUID) common.ISocietyInfo {
	return &bankSocietyInfoService{
		db:     db,
		bankId: bankId,
	}
}
