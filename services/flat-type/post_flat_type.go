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

func (h *hCreateFlatType) getBuiltUpArea() decimal.Decimal {
	return decimal.NewFromFloat(h.ReraCarpetArea).Add(decimal.NewFromFloat(h.BalconyArea))
}

func (h *hCreateFlatType) execute(db *gorm.DB, orgId, society string) (*models.FlatType, error) {
	flatType := models.FlatType{
		OrgId:          uuid.MustParse(orgId),
		SocietyId:      society,
		Name:           h.Name,
		Accommodation:  h.Accommodation,
		ReraCarpetArea: decimal.NewFromFloat(h.ReraCarpetArea),
		BalconyArea:    decimal.NewFromFloat(h.BalconyArea),
		BuiltUpArea:    h.getBuiltUpArea(),
		SuperArea:      decimal.NewFromFloat(h.SuperArea),
	}

	err := db.Create(&flatType).Error
	return &flatType, err
}

func (s *flatTypeService) createFlatType(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hCreateFlatType](w, r)
	if reqBody == nil {
		return
	}

	flatType, err := reqBody.execute(s.db, orgId, societyRera)
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
