package tower

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hUpdateTower struct {
	FloorCount int `validate:"required"`
}

func (ut *hUpdateTower) execute(db *gorm.DB, orgId, society, tower string) error {
	return db.
		Model(&models.Tower{
			Id: uuid.MustParse(tower),
		}).
		Where("org_id = ? and society_id = ?", orgId, society).
		Updates(models.Tower{
			FloorCount: ut.FloorCount,
		}).Error
}

func (ts *towerService) updateTower(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	towerId := chi.URLParam(r, "tower")
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateTower](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(ts.db, orgId, societyRera, towerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated tower."

	payload.EncodeJSON(w, http.StatusOK, response)
}
