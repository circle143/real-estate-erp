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

type hDeleteTower struct{}

func (dt *hDeleteTower) execute(db *gorm.DB, tower string) error {
	towerModel := models.Tower{
		Id: uuid.MustParse(tower),
	}

	return db.Delete(&towerModel).Error
}

func (ts *towerService) deleteTower(w http.ResponseWriter, r *http.Request) {
	towerId := chi.URLParam(r, "tower")

	tower := hDeleteTower{}
	err := tower.execute(ts.db, towerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully deleted tower."

	payload.EncodeJSON(w, http.StatusOK, response)
}