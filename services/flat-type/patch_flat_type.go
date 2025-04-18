package flat_type

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type hUpdateFlatType struct {
	Name  string
	Type  string
	Price float64
	Area  float64
}

func (uft *hUpdateFlatType) validate() error {
	if strings.TrimSpace(uft.Name) == "" && strings.TrimSpace(uft.Type) == "" && uft.Price == 0.0 && uft.Area == 0.0 {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid field values to update.",
		}
	}
	return nil
}

func (uft *hUpdateFlatType) execute(db *gorm.DB, flatType string) error {
	err := uft.validate()
	if err != nil {
		return err
	}

	flatTypeModel := models.FlatType{
		Id: uuid.MustParse(flatType),
	}

	return db.Model(&flatTypeModel).Updates(models.FlatType{
		Name:  uft.Name,
		Type:  uft.Type,
		Price: uft.Price,
		Area:  uft.Area,
	}).Error
}

func (fts *flatTypeService) updateFlatType(w http.ResponseWriter, r *http.Request) {
	flatTypeId := chi.URLParam(r, "flatType")
	reqBody := payload.ValidateAndDecodeRequest[hUpdateFlatType](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(fts.db, flatTypeId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated flat type."

	payload.EncodeJSON(w, http.StatusOK, response)
}