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

type hUpdateFlatDetails struct {
	Tower       string `validate:"required,uuid"`
	FlatType    string `validate:"required,uuid"`
	Name        string `validate:"required"`
	FloorNumber int    `validate:"gte=0"`
	Facing      string `validate:"required"`
}

func (h *hUpdateFlatDetails) validate(db *gorm.DB, orgId, society, flatId string) error {
	// validate facing
	facing := custom.Facing(h.Facing)
	if !facing.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid flat facing value.",
		}
	}

	// validate flat
	flatSocietyInfo := CreateFlatSocietyInfoService(db, uuid.MustParse(flatId))
	err := common.IsSameSociety(flatSocietyInfo, orgId, society)
	if err != nil {
		return err
	}

	// validate correct flat type
	flatTypeSocietyInfo := flatType.CreateFlatTypeSocietyInfoService(db, uuid.MustParse(h.FlatType))
	err = common.IsSameSociety(flatTypeSocietyInfo, orgId, society)
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

func (h *hUpdateFlatDetails) execute(db *gorm.DB, orgId, society, flatId string) error {
	err := h.validate(db, orgId, society, flatId)
	if err != nil {
		return err
	}

	return db.Model(&models.Flat{
		Id: uuid.MustParse(flatId),
	}).Updates(models.Flat{
		TowerId:     uuid.MustParse(h.Tower),
		FlatTypeId:  uuid.MustParse(h.FlatType),
		Name:        h.Name,
		FloorNumber: h.FloorNumber,
		Facing:      custom.Facing(h.Facing),
	}).Error
}

func (s *flatService) updateFlatDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	flatId := chi.URLParam(r, "flatId")
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateFlatDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId, societyRera, flatId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated flat details."

	payload.EncodeJSON(w, http.StatusOK, response)
}
