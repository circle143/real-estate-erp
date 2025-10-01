package models

import (
	"time"

	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Flat model
// todo method to generate flat name from floor number, flat count in floor, and tower name
type Flat struct {
	Id      uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TowerId uuid.UUID `gorm:"not null;index;uniqueIndex:tower_flat_unique" json:"towerId"`
	Tower   *Tower    `gorm:"foreignKey:TowerId;not null" json:"tower,omitempty"`
	//FlatTypeId  uuid.UUID       `gorm:"not null;index" json:"flatTypeId"`
	//FlatType    *FlatType       `gorm:"foreignKey:FlatTypeId;not null" json:"flatType,omitempty"`
	Name                        string              `gorm:"not null;uniqueIndex:tower_flat_unique" json:"name"`
	FloorNumber                 int                 `gorm:"not null" json:"floorNumber"`
	Facing                      custom.Facing       `gorm:"not null;default:Default" json:"facing"`
	SaleableArea                decimal.Decimal     `gorm:"not null;type:numeric" json:"salableArea"`
	UnitType                    string              `gorm:"not null" json:"unitType"`
	CreatedAt                   time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt                   time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
	SaleDetail                  *Sale               `gorm:"foreignKey:FlatId" json:"saleDetail,omitempty"`
	ActivePaymentPlanRatioItems []FlatPaymentStatus `gorm:"foreignKey:FlatId" json:"-"`
	//DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (u Flat) GetCreatedAt() time.Time {
	return u.CreatedAt
}
