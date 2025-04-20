package flat

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/tower"
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type flatSocietyInfoService struct {
	db     *gorm.DB
	flatId uuid.UUID
}

func (fs *flatSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	flat := models.Flat{
		Id: fs.flatId,
	}

	err := fs.db.First(&flat).Error
	if err != nil {
		return nil, err
	}

	towerSocietyInfo := tower.CreateTowerSocietyInfoService(fs.db, flat.TowerId)
	return towerSocietyInfo.GetSocietyInfo()
}

func CreateFlatSocietyInfoService(db *gorm.DB, flatId uuid.UUID) common.ISocietyInfo {
	return &flatSocietyInfoService{
		db:     db,
		flatId: flatId,
	}
}
