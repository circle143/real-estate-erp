package common

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IPriceHistoryUtil interface {
	AddInitialPrice() error // handles adding initial price
	AddNewPrice() error     // handles populating active till and adding new price
}

func CreatePriceUtil(db *gorm.DB, chargeId uuid.UUID, chargeType custom.PriceChargeType, price float64) IPriceHistoryUtil {
	if !chargeType.IsValid() {
		return nil
	}

	return &priceHistoryUtil{
		db:         db,
		chargeId:   chargeId,
		chargeType: string(chargeType),
		price:      price,
	}
}

type priceHistoryUtil struct {
	db         *gorm.DB
	chargeId   uuid.UUID
	chargeType string
	price      float64
}

func (p *priceHistoryUtil) AddInitialPrice() error {
	priceHistory := models.PriceHistory{
		ChargeId:   p.chargeId,
		ChargeType: p.chargeType,
		Price:      p.price,
	}
	return p.db.Create(&priceHistory).Error
}

func (p *priceHistoryUtil) AddNewPrice() error {
	var activePrice models.PriceHistory
	err := p.db.
		Where("charge_id = ? AND charge_type = ?", p.chargeId, p.chargeType).
		Order("active_from DESC").
		First(&activePrice).Error
	if err != nil {
		return err
	}

	// add new price record
	priceHistory := models.PriceHistory{
		ChargeId:   p.chargeId,
		ChargeType: p.chargeType,
		Price:      p.price,
	}
	err = p.db.Create(&priceHistory).Error
	if err != nil {
		return err
	}

	// update previous active record to update active till property
	return p.db.Model(&activePrice).Update("active_till", priceHistory.ActiveFrom).Error
}
