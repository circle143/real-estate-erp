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

type hUpdateBrokerDetails struct {
	Name string `validate:"required"`
}

//func (h *hUpdateBrokerDetails) validate(db *gorm.DB, orgId, societyRera, brokerId string) error {
//	societyInfoService := CreateBrokerSocietyInfoService(db, uuid.MustParse(brokerId))
//	return common.IsSameSociety(societyInfoService, orgId, societyRera)
//}

func (h *hUpdateBrokerDetails) execute(db *gorm.DB, orgId, societyRera, brokerId string) error {
	return db.
		Model(&models.Broker{
			Id: uuid.MustParse(brokerId),
		}).
		Where("org_id = ? and society_id = ?", orgId, societyRera).
		Updates(models.Broker{
			Name: h.Name,
		}).Error
}

func (s *brokerService) updateBrokerDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	brokerId := chi.URLParam(r, "brokerId")
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateBrokerDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId, societyRera, brokerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated broker details."

	payload.EncodeJSON(w, http.StatusOK, response)
}
