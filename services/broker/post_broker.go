package broker

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hAddBrokerToSociety struct {
	Name         string `validate:"required"`
	PanNumber    string `validate:"required,pan"`
	AadharNumber string `validate:"required,aadhar"`
}

func (h *hAddBrokerToSociety) execute(db *gorm.DB, orgId, society string) (*models.Broker, error) {
	brokerModel := models.Broker{
		OrgId:        uuid.MustParse(orgId),
		SocietyId:    society,
		Name:         h.Name,
		PanNumber:    h.PanNumber,
		AadharNumber: h.AadharNumber,
	}

	err := db.Create(&brokerModel).Error
	return &brokerModel, err
}

func (s *brokerService) addBrokerToSociety(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hAddBrokerToSociety](w, r)
	if reqBody == nil {
		return
	}

	broker, err := reqBody.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully added broker."
	response.Data = broker

	payload.EncodeJSON(w, http.StatusCreated, response)
}
