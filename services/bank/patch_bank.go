package bank

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hUpdateBankAccountDetails struct {
	Name          string `validate:"required"`
	AccountNumber string `validate:"required,bank-account-number"`
}

func (h *hUpdateBankAccountDetails) execute(db *gorm.DB, orgId, societyRera, bankId string) error {
	return db.
		Model(&models.Bank{
			Id: uuid.MustParse(bankId),
		}).
		Where("org_id = ? and society_id = ?", orgId, societyRera).
		Updates(models.Bank{
			Name:          h.Name,
			AccountNumber: h.AccountNumber,
		}).Error
}

func (s *bankService) updateBankAccountDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	bankId := chi.URLParam(r, "bankId")
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateBankAccountDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId, societyRera, bankId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated bank account details."

	payload.EncodeJSON(w, http.StatusOK, response)
}
