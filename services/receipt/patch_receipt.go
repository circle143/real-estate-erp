package receipt

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

type hMarkReceiptAsFailed struct{}

func (h *hMarkReceiptAsFailed) validate(db *gorm.DB, orgId, society, receiptId string) error {
	receiptSocietyInfo := CreateReceiptSocietyInfoService(db, uuid.MustParse(receiptId))
	return common.IsSameSociety(receiptSocietyInfo, orgId, society)
}

func (h *hMarkReceiptAsFailed) execute(db *gorm.DB, orgId, society, receiptId string) error {
	err := h.validate(db, orgId, society, receiptId)
	if err != nil {
		return err
	}

	return db.Model(&models.Receipt{
		Id: uuid.MustParse(receiptId),
	}).Updates(models.Receipt{
		Failed: true,
	}).Error
}

func (s *receiptService) markReceiptAsFailed(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	receiptId := chi.URLParam(r, "receiptId")

	receipt := hMarkReceiptAsFailed{}
	err := receipt.execute(s.db, orgId, societyRera, receiptId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Receipt marked as failed."

	payload.EncodeJSON(w, http.StatusCreated, response)
}
