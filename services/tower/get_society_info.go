package tower

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type towerSocietyInfoService struct {
	db      *gorm.DB
	towerId uuid.UUID
}

func (s *towerSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	// fetch from db and return
	tower := models.Tower{
		Id: s.towerId,
	}

	err := s.db.First(&tower).Error
	if err != nil {
		return nil, err
	}

	return &common.SocietyInfo{
		OrgId:       tower.OrgId,
		SocietyRera: tower.SocietyId,
	}, nil
}

func CreateTowerSocietyInfoService(db *gorm.DB, towerId uuid.UUID) common.ISocietyInfo {
	return &towerSocietyInfoService{
		db:      db,
		towerId: towerId,
	}
}
