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
	FloorCount        int
	PerFloorFlatCount int
}

func (ut *hUpdateTower) validate() error {
	if ut.FloorCount == 0 && ut.PerFloorFlatCount == 0 {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid field values to update.",
		}
	}
	return nil
}

func (ut *hUpdateTower) execute(db *gorm.DB, tower string) error {
	err := ut.validate()
	if err != nil {
		return err
	}

	towerModel := models.Tower{
		Id: uuid.MustParse(tower),
	}

	return db.Model(&towerModel).Updates(models.Tower{
		FloorCount:        ut.FloorCount,
		PerFloorFlatCount: ut.PerFloorFlatCount,
	}).Error
}

func (ts *towerService) updateTower(w http.ResponseWriter, r *http.Request) {
	towerId := chi.URLParam(r, "tower")
	reqBody := payload.ValidateAndDecodeRequest[hUpdateTower](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(ts.db, towerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated tower."

	payload.EncodeJSON(w, http.StatusOK, response)
}