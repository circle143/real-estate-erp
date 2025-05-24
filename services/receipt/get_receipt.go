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

type hGetReceiptById struct{}

func (h *hGetReceiptById) validate(db *gorm.DB, orgId, society, receiptId string) error {
	receiptSocietyInfo := CreateReceiptSocietyInfoService(db, uuid.MustParse(receiptId))
	return common.IsSameSociety(receiptSocietyInfo, orgId, society)
}

func (h *hGetReceiptById) execute(db *gorm.DB, orgId, society, receiptId string) (*models.Receipt, error) {
	err := h.validate(db, orgId, society, receiptId)
	if err != nil {
		return nil, err
	}

	var receipt models.Receipt

	err = db.
		Preload("Cleared").
		Preload("Cleared.Bank").
		Preload("Sale").
		Preload("Sale.Customers").
		Preload("Sale.CompanyCustomer").
		Preload("Sale.Broker").
		Preload("Sale.Flat").
		First(&receipt, "id = ?", receiptId).
		Error

	return &receipt, err
}

func (s *receiptService) getReceiptById(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	receiptId := chi.URLParam(r, "receiptId")

	receipt := hGetReceiptById{}
	item, err := receipt.execute(s.db, orgId, societyRera, receiptId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = item

	payload.EncodeJSON(w, http.StatusOK, response)
}
