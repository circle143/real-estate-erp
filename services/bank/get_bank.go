package bank

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type hGetAllSocietyBankAccounts struct{}

func (h *hGetAllSocietyBankAccounts) execute(db *gorm.DB, orgId, society, cursor string) (*custom.PaginatedData, error) {
	var bankAccountData []models.Bank

	query := db.Where("org_id = ? and society_id = ?", orgId, society).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	err := query.Find(&bankAccountData).Error
	if err != nil {
		return nil, err
	}
	return common.CreatePaginatedResponse(&bankAccountData), nil
}

func (s *bankService) getAllSocietyBankAccounts(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")

	bank := hGetAllSocietyBankAccounts{}
	res, err := bank.execute(s.db, orgId, societyRera, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hGetBankReport struct {
	RecordsFrom time.Time
	RecordsTill time.Time
}

func (h *hGetBankReport) validate(db *gorm.DB, orgId, society, bankId string) error {
	bankSocietyInfo := CreateBankSocietyInfoService(db, uuid.MustParse(bankId))
	return common.IsSameSociety(bankSocietyInfo, orgId, society)
}

func (h *hGetBankReport) execute(db *gorm.DB, orgId, society, bankId string) (*models.Bank, error) {
	err := h.validate(db, orgId, society, bankId)
	if err != nil {
		return nil, err
	}

	bankModel := models.Bank{
		Id: uuid.MustParse(bankId),
	}
	err = db.Preload("ClearedReceipts", func(db *gorm.DB) *gorm.DB {
		if h.RecordsTill.IsZero() && h.RecordsFrom.IsZero() {
			return db.Order("created_at DESC")
		}

		if h.RecordsTill.IsZero() {
			return db.
				Where("created_at <= ?", h.RecordsFrom).
				Order("created_at DESC")
		}

		if h.RecordsFrom.IsZero() {
			return db.
				Where("created_at >= ?", h.RecordsTill).
				Order("created_at DESC")
		}

		return db.
			Where("created_at <= ? AND created_at >= ?", h.RecordsFrom, h.RecordsTill).
			Order("created_at DESC")
	}).Preload("ClearedReceipts.Receipt").
		Preload("ClearedReceipts.Receipt.Sale").
		Preload("ClearedReceipts.Receipt.Sale.Flat").
		Preload("ClearedReceipts.Receipt.Sale.Customers").
		Preload("ClearedReceipts.Receipt.Sale.CompanyCustomer").
		First(&bankModel).Error
	return &bankModel, err
}

func (s *bankService) getBankReport(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	bankId := chi.URLParam(r, "bankId")

	reqBody := payload.ValidateAndDecodeRequest[hGetBankReport](w, r)
	if reqBody == nil {
		return
	}

	report, err := reqBody.execute(s.db, orgId, societyRera, bankId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = report

	payload.EncodeJSON(w, http.StatusOK, response)
}
