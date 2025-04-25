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

type hCreateTower struct {
	FloorCount int    `validate:"required"`
	Name       string `validate:"required"`
}

func (ct *hCreateTower) execute(db *gorm.DB, orgId, society string) (*models.Tower, error) {
	tower := models.Tower{
		OrgId:      uuid.MustParse(orgId),
		SocietyId:  society,
		FloorCount: ct.FloorCount,
		Name:       ct.Name,
	}

	result := db.Create(&tower)
	if result.Error != nil {
		return nil, result.Error
	}

	return &tower, nil
}

func (ts *towerService) createTower(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hCreateTower](w, r)
	if reqBody == nil {
		return
	}

	tower, err := reqBody.execute(ts.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new tower."
	response.Data = tower

	payload.EncodeJSON(w, http.StatusCreated, response)
}