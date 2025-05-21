package flat

import (
	"circledigital.in/real-state-erp/models"
	flatType "circledigital.in/real-state-erp/services/flat-type"
	"circledigital.in/real-state-erp/services/tower"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hCreateFlat struct {
	Tower       string `validate:"required,uuid"`
	FlatType    string `validate:"required,uuid"`
	Name        string `validate:"required"`
	FloorNumber int    `validate:"gte=0"`
	Facing      string `validate:"required"`
}

func (h *hCreateFlat) validate(db *gorm.DB, orgId, society string) error {
	// validate facing
	facing := custom.Facing(h.Facing)
	if !facing.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid flat facing value.",
		}
	}

	// validate correct flat type
	flatTypeSocietyInfo := flatType.CreateFlatTypeSocietyInfoService(db, uuid.MustParse(h.FlatType))
	err := common.IsSameSociety(flatTypeSocietyInfo, orgId, society)
	if err != nil {
		return err
	}

	// validate tower belongs to correct society and organization
	towerSocietyInfoService := tower.CreateTowerSocietyInfoService(db, uuid.MustParse(h.Tower))
	err = common.IsSameSociety(towerSocietyInfoService, orgId, society)
	if err != nil {
		return err
	}

	var towerModel models.Tower
	err = db.Where(&models.Tower{
		Id:        uuid.MustParse(h.Tower),
		OrgId:     uuid.MustParse(orgId),
		SocietyId: society,
	}).First(&towerModel).Error
	if err != nil {
		return err
	}

	// validate floor number
	if h.FloorNumber > towerModel.FloorCount {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid floor number.",
		}
	}

	return nil
}

func (h *hCreateFlat) execute(db *gorm.DB, orgId, society string) (*models.Flat, error) {
	err := h.validate(db, orgId, society)
	if err != nil {
		return nil, err
	}

	flat := models.Flat{
		TowerId:     uuid.MustParse(h.Tower),
		FlatTypeId:  uuid.MustParse(h.FlatType),
		Name:        h.Name,
		FloorNumber: h.FloorNumber,
		Facing:      custom.Facing(h.Facing),
	}

	result := db.Create(&flat)
	if result.Error != nil {
		return nil, result.Error
	}

	return &flat, nil

}

func (s *flatService) createNewFlat(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hCreateFlat](w, r)
	if reqBody == nil {
		return
	}

	flat, err := reqBody.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new flat."
	response.Data = flat

	payload.EncodeJSON(w, http.StatusCreated, response)
}
