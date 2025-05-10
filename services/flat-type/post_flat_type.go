package flat_type

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
)

type hCreateFlatType struct {
	Name           string  `validate:"required"`
	Accommodation  string  `validate:"required"`
	ReraCarpetArea float64 `validate:"required"`
	BalconyArea    float64 `validate:"required"`
	SuperArea      float64 `validate:"required"`
}

func (cft *hCreateFlatType) getBuiltUpArea() decimal.Decimal {
	return decimal.NewFromFloat(cft.ReraCarpetArea).Add(decimal.NewFromFloat(cft.BalconyArea))
}

func (cft *hCreateFlatType) execute(db *gorm.DB, orgId, society string) (*models.FlatType, error) {
	flatType := models.FlatType{
		OrgId:          uuid.MustParse(orgId),
		SocietyId:      society,
		Name:           cft.Name,
		Accommodation:  cft.Accommodation,
		ReraCarpetArea: decimal.NewFromFloat(cft.ReraCarpetArea),
		BalconyArea:    decimal.NewFromFloat(cft.BalconyArea),
		BuiltUpArea:    cft.getBuiltUpArea(),
		SuperArea:      decimal.NewFromFloat(cft.SuperArea),
	}

	err := db.Create(&flatType).Error
	return &flatType, err
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
