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

type hDeleteFlatType struct{}

func (dft *hDeleteFlatType) execute(db *gorm.DB, orgId, society, flatType string) error {
	return db.
		Model(&models.FlatType{
			Id: uuid.MustParse(flatType),
		}).
		Where("org_id = ? and society_id = ?", orgId, society).
		Delete(&models.FlatType{}).Error
}

func (fts *flatTypeService) deleteFlatType(w http.ResponseWriter, r *http.Request) {
	flatTypeId := chi.URLParam(r, "flatType")
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	flatType := hDeleteFlatType{}
	err := flatType.execute(fts.db, orgId, societyRera, flatTypeId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully deleted flat type."

	payload.EncodeJSON(w, http.StatusOK, response)
}
