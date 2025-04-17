package society

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hDeleteSociety struct{}

func (ds *hDeleteSociety) execute(db *gorm.DB, society, orgId string) error {
	societyModel := models.Society{
		ReraNumber: society,
		OrgId:      uuid.MustParse(orgId),
	}

	return db.Delete(&societyModel).Error
}

func (ss *societyService) deleteSociety(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	society := hDeleteSociety{}
	err := society.execute(ss.db, societyRera, orgId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully deleted society."

	payload.EncodeJSON(w, http.StatusOK, response)
}