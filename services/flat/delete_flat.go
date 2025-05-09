package flat

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hDeleteFlat struct{}

func (df *hDeleteFlat) validate(db *gorm.DB, orgId, societyRera, flatId string) error {
	flatSocietyInfo := CreateFlatSocietyInfoService(db, uuid.MustParse(flatId))
	return common.IsSameSociety(flatSocietyInfo, orgId, societyRera)
}

func (df *hDeleteFlat) execute(db *gorm.DB, orgId, societyRera, flatId string) error {
	err := df.validate(db, orgId, societyRera, flatId)
	if err != nil {
		return err
	}

	return db.Model(&models.Flat{
		Id: uuid.MustParse(flatId),
	}).Delete(&models.Flat{}).Error
}

func (s *flatService) deleteFlat(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	flatId := chi.URLParam(r, "flat")
	societyRera := chi.URLParam(r, "society")

	flat := hDeleteFlat{}

	err := flat.execute(s.db, orgId, societyRera, flatId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully deleted flat."

	payload.EncodeJSON(w, http.StatusOK, response)
}
