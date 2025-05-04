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

type hUpdateFlatType struct {
	Price float64 `validate:"required"`
}

func (uft *hUpdateFlatType) execute(db *gorm.DB, flatType string) error {
	flatTypeModel := models.FlatType{
		Id: uuid.MustParse(flatType),
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// get current active price
		var activePrice models.PriceHistory
		err := tx.
			Where("charge_id = ? AND charge_type = ?", flatTypeModel.Id, string(custom.FLATTYPECHARGE)).
			Order("active_from DESC").
			First(&activePrice).Error

		if err != nil {
			return err
		}

		// update price in db
		err = tx.Model(&flatTypeModel).Update("price", uft.Price).Error
		if err != nil {
			return err
		}

		// add new price record
		priceHistory := models.PriceHistory{
			ChargeId:   flatTypeModel.Id,
			ChargeType: string(custom.FLATTYPECHARGE),
			Price:      uft.Price,
		}
		err = tx.Create(&priceHistory).Error
		if err != nil {
			return err
		}

		// update previous active record to update active till property
		return tx.Model(&activePrice).Update("active_till", priceHistory.ActiveFrom).Error
	})
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
