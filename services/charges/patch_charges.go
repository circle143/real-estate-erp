package charges

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

type hUpdatePreferenceLocationChargePrice struct {
	Price float64 `validate:"required"`
}

func (h *hUpdatePreferenceLocationChargePrice) execute(db *gorm.DB, orgId, society, chargeId string) error {
	chargeModel := models.PreferenceLocationCharge{
		Id: uuid.MustParse(chargeId),
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// update price in db
		err := tx.Model(&chargeModel).
			Where("id = ? AND org_id = ? AND society_id = ?", chargeModel.Id, orgId, society).
			Update("price", h.Price).Error
		if err != nil {
			return err
		}

		priceHistoryUtil := common.CreatePriceUtil(tx, chargeModel.Id, custom.PREFERENCELOCATIONCHARGE, h.Price)
		return priceHistoryUtil.AddNewPrice()
	})
}

func (s *chargesService) updatePreferenceLocationChargePrice(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	chargeId := chi.URLParam(r, "chargeId")

	reqBody := payload.ValidateAndDecodeRequest[hUpdatePreferenceLocationChargePrice](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId, societyRera, chargeId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated price."

	payload.EncodeJSON(w, http.StatusOK, response)
}

func (s *chargesService) updatePreferenceLocationChargeDetails(w http.ResponseWriter, r *http.Request) {
}

type hUpdateOtherChargePrice struct {
	Price float64 `validate:"required"`
}

func (h *hUpdateOtherChargePrice) execute(db *gorm.DB, orgId, society, chargeId string) error {
	chargeModel := models.OtherCharge{
		Id: uuid.MustParse(chargeId),
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// update price in db
		err := tx.Model(&chargeModel).
			Where("id = ? AND org_id = ? AND society_id = ?", chargeModel.Id, orgId, society).
			Update("price", h.Price).Error
		if err != nil {
			return err
		}

		priceHistoryUtil := common.CreatePriceUtil(tx, chargeModel.Id, custom.OTHERCHARGE, h.Price)
		return priceHistoryUtil.AddNewPrice()
	})
}

func (s *chargesService) updateOtherChargePrice(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	chargeId := chi.URLParam(r, "chargeId")

	reqBody := payload.ValidateAndDecodeRequest[hUpdatePreferenceLocationChargePrice](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId, societyRera, chargeId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated price."

	payload.EncodeJSON(w, http.StatusOK, response)
}

func (s *chargesService) updateOtherChargeDetails(w http.ResponseWriter, r *http.Request) {}
