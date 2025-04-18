package flat_type

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hCreateFlatType struct {
	Name  string  `validate:"required"`
	Type  string  `validate:"required"`
	Price float64 `validate:"required"`
	Area  float64 `validate:"required"`
}

func (cft *hCreateFlatType) execute(db *gorm.DB, orgId, society string) (*models.FlatType, error) {
	flatType := models.FlatType{
		OrgId:     uuid.MustParse(orgId),
		SocietyId: society,
		Name:      cft.Name,
		Type:      cft.Type,
		Price:     cft.Price,
		Area:      cft.Area,
	}

	result := db.Create(&flatType)
	if result.Error != nil {
		return nil, result.Error
	}

	return &flatType, nil
}

func (fts *flatTypeService) createFlatType(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hCreateFlatType](w, r)
	if reqBody == nil {
		return
	}

	flatType, err := reqBody.execute(fts.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new flat type."
	response.Data = flatType

	payload.EncodeJSON(w, http.StatusCreated, response)
}