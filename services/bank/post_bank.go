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

type hAddBankAccountToSociety struct {
	Name          string `validate:"required"`
	AccountNumber string `validate:"required,bank-account-number"`
}

func (h *hAddBankAccountToSociety) execute(db *gorm.DB, orgId, society string) (*models.Bank, error) {
	bankModel := models.Bank{
		OrgId:         uuid.MustParse(orgId),
		SocietyId:     society,
		Name:          h.Name,
		AccountNumber: h.AccountNumber,
	}

	err := db.Create(&bankModel).Error
	return &bankModel, err
}

func (s *bankService) addBankAccountToSociety(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hAddBankAccountToSociety](w, r)
	if reqBody == nil {
		return
	}

	bank, err := reqBody.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully added bank account to society."
	response.Data = bank

	payload.EncodeJSON(w, http.StatusCreated, response)
}
