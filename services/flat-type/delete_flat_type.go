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

func (dft *hDeleteFlatType) execute(db *gorm.DB, flatType string) error {
	flatTypeModel := models.FlatType{
		Id: uuid.MustParse(flatType),
	}

	return db.Delete(&flatTypeModel).Error
}

func (fts *flatTypeService) deleteFlatType(w http.ResponseWriter, r *http.Request) {
	flatTypeId := chi.URLParam(r, "flatType")

	flatType := hDeleteFlatType{}
	err := flatType.execute(fts.db, flatTypeId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully deleted flat type."

	payload.EncodeJSON(w, http.StatusOK, response)
}