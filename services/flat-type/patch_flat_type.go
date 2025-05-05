package flat_type

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

type hUpdateFlatType struct {
	Price float64 `validate:"required"`
}

func (uft *hUpdateFlatType) execute(db *gorm.DB, orgId, society, flatType string) error {
	flatTypeModel := models.FlatType{
		Id: uuid.MustParse(flatType),
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// update price in db
		err := tx.Model(&flatTypeModel).
			Where("id = ? AND org_id = ? AND society_id = ?", flatTypeModel.Id, orgId, society).
			Update("price", uft.Price).Error
		if err != nil {
			return err
		}

		priceHistoryUtil := common.CreatePriceUtil(tx, flatTypeModel.Id, custom.FLATTYPECHARGE, uft.Price)
		return priceHistoryUtil.AddNewPrice()
	})
}

func (fts *flatTypeService) updateFlatType(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	flatTypeId := chi.URLParam(r, "flatType")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateFlatType](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(fts.db, orgId, societyRera, flatTypeId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated flat type."

	payload.EncodeJSON(w, http.StatusOK, response)
}
