package broker

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type brokerSocietyInfoService struct {
	db       *gorm.DB
	brokerId uuid.UUID
}

func (s *brokerSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	// fetch from db and return
	broker := models.Broker{
		Id: s.brokerId,
	}

	err := s.db.First(&broker).Error
	if err != nil {
		return nil, err
	}

	return &common.SocietyInfo{
		OrgId:       broker.OrgId,
		SocietyRera: broker.SocietyId,
	}, nil
}

func CreateBrokerSocietyInfoService(db *gorm.DB, brokerId uuid.UUID) common.ISocietyInfo {
	return &brokerSocietyInfoService{
		db:       db,
		brokerId: brokerId,
	}
}
