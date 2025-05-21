package charges

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
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
	price := decimal.NewFromFloat(h.Price)

	return db.Transaction(func(tx *gorm.DB) error {
		// update price in db
		err := tx.Model(&chargeModel).
			Where("id = ? AND org_id = ? AND society_id = ?", chargeModel.Id, orgId, society).
			Update("price", price).Error
		if err != nil {
			return err
		}

		priceHistoryUtil := createPriceUtil(tx, chargeModel.Id, custom.PREFERENCELOCATIONCHARGE, price)
		return priceHistoryUtil.addNewPrice()
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

type hUpdatePreferenceLocationChargeDetails struct {
	Summary string `validate:"required"`
	Disable bool
}

func (h *hUpdatePreferenceLocationChargeDetails) execute(db *gorm.DB, orgId, society, chargeId string) error {
	chargeModel := models.PreferenceLocationCharge{
		Id: uuid.MustParse(chargeId),
	}

	updates := map[string]interface{}{
		"disable": h.Disable,
		"summary": h.Summary,
	}

	return db.Model(&chargeModel).
		Where("id = ? AND org_id = ? AND society_id = ?", chargeModel.Id, orgId, society).
		Updates(updates).Error
}

func (s *chargesService) updatePreferenceLocationChargeDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	chargeId := chi.URLParam(r, "chargeId")

	reqBody := payload.ValidateAndDecodeRequest[hUpdatePreferenceLocationChargeDetails](w, r)
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
	response.Message = "Successfully updated charge details."

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hUpdateOtherChargePrice struct {
	Price float64 `validate:"required"`
}

func (h *hUpdateOtherChargePrice) execute(db *gorm.DB, orgId, society, chargeId string) error {
	chargeModel := models.OtherCharge{
		Id: uuid.MustParse(chargeId),
	}
	price := decimal.NewFromFloat(h.Price)

	return db.Transaction(func(tx *gorm.DB) error {
		// update price in db
		err := tx.Model(&chargeModel).
			Where("id = ? AND org_id = ? AND society_id = ?", chargeModel.Id, orgId, society).
			Update("price", price).Error
		if err != nil {
			return err
		}

		priceHistoryUtil := createPriceUtil(tx, chargeModel.Id, custom.OTHERCHARGE, price)
		return priceHistoryUtil.addNewPrice()
	})
}

func (s *chargesService) updateOtherChargePrice(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	chargeId := chi.URLParam(r, "chargeId")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateOtherChargePrice](w, r)
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

type hUpdateOtherChargeDetails struct {
	Summary       string `validate:"required"`
	Disable       bool
	Recurring     bool
	Optional      bool
	Fixed         bool
	AdvanceMonths int
}

func (h *hUpdateOtherChargeDetails) validate() error {
	if (h.Recurring && h.AdvanceMonths >= 1) || (!h.Recurring && h.AdvanceMonths == 0) {
		return nil
	}

	return &custom.RequestError{
		Status:  http.StatusBadRequest,
		Message: "Invalid value for field advanceMonths",
	}
}

func (h *hUpdateOtherChargeDetails) execute(db *gorm.DB, orgId, society, chargeId string) error {
	chargeModel := models.OtherCharge{
		Id: uuid.MustParse(chargeId),
	}

	updates := map[string]interface{}{
		"disable":        h.Disable,
		"recurring":      h.Recurring,
		"optional":       h.Optional,
		"summary":        h.Summary,
		"advance_months": h.AdvanceMonths,
		"fixed":          h.Fixed,
	}

	return db.Model(&chargeModel).
		Where("id = ? AND org_id = ? AND society_id = ?", chargeModel.Id, orgId, society).
		Updates(updates).Error
}

func (s *chargesService) updateOtherChargeDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	chargeId := chi.URLParam(r, "chargeId")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateOtherChargeDetails](w, r)
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
	response.Message = "Successfully updated charge details."

	payload.EncodeJSON(w, http.StatusOK, response)
}
