package flat_type

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type flatTypeSocietyInfoService struct {
	db         *gorm.DB
	flatTypeId uuid.UUID
}

func (fts *flatTypeSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	flatType := models.FlatType{
		Id: fts.flatTypeId,
	}

	err := fts.db.First(&flatType).Error
	if err != nil {
		return nil, err
	}

	return &common.SocietyInfo{
		OrgId:       flatType.OrgId,
		SocietyRera: flatType.SocietyId,
	}, nil
}

func CreateFlatTypeSocietyInfoService(db *gorm.DB, flatTypeId uuid.UUID) common.ISocietyInfo {
	return &flatTypeSocietyInfoService{
		db:         db,
		flatTypeId: flatTypeId,
	}
}
