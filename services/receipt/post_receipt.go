package receipt

import (
	"net/http"
	"strings"
	"time"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/bank"
	"circledigital.in/real-state-erp/services/sale"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var gstDate = time.Date(2017, time.July, 1, 0, 0, 0, 0, time.UTC)

// TODO: later handle automatic generation for receipt number
type hCreateSaleReceipt struct {
	ReceiptNumber     string      `validate:"required"`
	TotalAmount       float64     `validate:"required"`
	Mode              string      `validate:"required"`
	DateIssued        pgtype.Date `validate:"required"`
	BankName          string
	TransactionNumber string
	GstRate           int
	ServiceTax        float64
	SwatchBharatCess  float64
	KrishiKalyanCess  float64
}

func (h *hCreateSaleReceipt) validate(db *gorm.DB, orgId, society, saleId string) error {
	if h.GstRate != 5 && h.GstRate != 1 {
		h.GstRate = 5
	}

	if h.DateIssued.Time.Before(gstDate) {
		h.GstRate = 0
		if h.ServiceTax == 0 && h.SwatchBharatCess == 0 && h.KrishiKalyanCess == 0 {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Required missing values for 'serviceTax', 'swatchBharatCess' or 'krishiKalyanCess' for entries before 1 July 2017.",
			}
		}
	}

	mode := custom.ReceiptMode(h.Mode)
	if !mode.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid mode value.",
		}
	}

	if mode != custom.ADJUSTMENT && h.TotalAmount < 0 {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Negative receipt amount are only allowed for Adjustment.",
		}
	}

	if mode.RequireBankDetails() {
		if strings.TrimSpace(h.BankName) == "" || strings.TrimSpace(h.TransactionNumber) == "" {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Required missing values: Bank Name or Transaction Number",
			}
		}
	}

	societyInfoService := sale.CreateSaleSocietyInfoService(db, uuid.MustParse(saleId))
	return common.IsSameSociety(societyInfoService, orgId, society)
}

func (h *hCreateSaleReceipt) execute(db *gorm.DB, orgId, society, saleId string) (*models.Receipt, error) {
	err := h.validate(db, orgId, society, saleId)
	if err != nil {
		return nil, err
	}

	receiptModel := models.Receipt{
		ReceiptNumber:     h.ReceiptNumber,
		SaleId:            uuid.MustParse(saleId),
		TotalAmount:       decimal.NewFromFloat(h.TotalAmount),
		Amount:            decimal.NewFromFloat(h.TotalAmount),
		TransactionNumber: h.TransactionNumber,
		BankName:          h.BankName,
		Mode:              custom.ReceiptMode(h.Mode),
		DateIssued:        h.DateIssued,
	}

	mode := custom.ReceiptMode(h.Mode)
	if mode != custom.ADJUSTMENT {
		if !h.DateIssued.Time.Before(gstDate) {
			gstInfo := receiptModel.CalcGST(h.GstRate)
			receiptModel.Amount = gstInfo.Amount
			receiptModel.SGST = &gstInfo.SGST
			receiptModel.CGST = &gstInfo.CGST
		} else {
			if h.ServiceTax > 0 {
				tax := decimal.NewFromFloat(h.ServiceTax)
				receiptModel.Amount = receiptModel.Amount.Sub(tax)
				receiptModel.ServiceTax = &tax
			}

			if h.SwatchBharatCess > 0 {
				tax := decimal.NewFromFloat(h.SwatchBharatCess)
				receiptModel.Amount = receiptModel.Amount.Sub(tax)
				receiptModel.SwathchBharatCess = &tax
			}

			if h.KrishiKalyanCess > 0 {
				tax := decimal.NewFromFloat(h.KrishiKalyanCess)
				receiptModel.Amount = receiptModel.Amount.Sub(tax)
				receiptModel.KrishiKalyanCess = &tax
			}
		}
	}

	err = db.Create(&receiptModel).Error
	return &receiptModel, err
}

func (s *receiptService) createSaleReceipt(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	saleId := chi.URLParam(r, "saleId")

	reqBody := payload.ValidateAndDecodeRequest[hCreateSaleReceipt](w, r)
	if reqBody == nil {
		return
	}

	receipt, err := reqBody.execute(s.db, orgId, societyRera, saleId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully added new receipt to sale."
	response.Data = receipt

	payload.EncodeJSON(w, http.StatusCreated, response)
}

type hClearSaleReceipt struct {
	BankId string `validate:"required,uuid"`
}

func (h *hClearSaleReceipt) validate(db *gorm.DB, orgId, society, receiptId string) error {
	receiptSocietyInfo := CreateReceiptSocietyInfoService(db, uuid.MustParse(receiptId))
	err := common.IsSameSociety(receiptSocietyInfo, orgId, society)
	if err != nil {
		return err
	}

	bankSocietyInfo := bank.CreateBankSocietyInfoService(db, uuid.MustParse(h.BankId))
	err = common.IsSameSociety(bankSocietyInfo, orgId, society)
	if err != nil {
		return err
	}

	receipt := models.Receipt{
		Id: uuid.MustParse(receiptId),
	}
	err = db.Find(&receipt).Error
	if err != nil {
		return err
	}

	if receipt.Failed {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "This receipt is marked as failed and you can't clear it anymore.",
		}
	}
	return nil
}

func (h *hClearSaleReceipt) execute(db *gorm.DB, orgId, society, receiptId string) (*models.ReceiptClear, error) {
	err := h.validate(db, orgId, society, receiptId)
	if err != nil {
		return nil, err
	}

	receiptClearModel := models.ReceiptClear{
		ReceiptId: uuid.MustParse(receiptId),
		BankId:    uuid.MustParse(h.BankId),
	}

	err = db.Create(&receiptClearModel).Error
	if err != nil {
		return nil, err
	}

	err = db.Preload("Bank").
		First(&receiptClearModel, "receipt_id = ?", receiptClearModel.ReceiptId).
		Error
	return &receiptClearModel, err
}

func (s *receiptService) clearSaleReceipt(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	receiptId := chi.URLParam(r, "receiptId")

	reqBody := payload.ValidateAndDecodeRequest[hClearSaleReceipt](w, r)
	if reqBody == nil {
		return
	}

	receipt, err := reqBody.execute(s.db, orgId, societyRera, receiptId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully cleared receipt."
	response.Data = receipt

	payload.EncodeJSON(w, http.StatusCreated, response)
}

// ==================== UPDATE RECEIPT ====================

type hUpdateSaleReceipt struct {
	ReceiptNumber     string  `validate:"required"`
	TotalAmount       float64 `validate:"required"`
	Mode              string  `validate:"required"`
	BankName          string
	TransactionNumber string
	GstRate           int
	ServiceTax        float64
	SwatchBharatCess  float64
	KrishiKalyanCess  float64
}

func (h *hUpdateSaleReceipt) validate(db *gorm.DB, orgId, society, receiptId string) (*models.Receipt, error) {
	// Fetch existing receipt
	var existingReceipt models.Receipt
	err := db.Preload("Cleared").First(&existingReceipt, "id = ?", receiptId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &custom.RequestError{
				Status:  http.StatusNotFound,
				Message: "Receipt not found.",
			}
		}
		return nil, err
	}

	// Check if receipt is cleared
	if existingReceipt.Cleared != nil {
		return nil, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Cannot edit a cleared receipt.",
		}
	}

	// Check if receipt is failed
	if existingReceipt.Failed {
		return nil, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Cannot edit a failed receipt.",
		}
	}

	// Validate GST rate
	if h.GstRate != 5 && h.GstRate != 1 {
		h.GstRate = 5
	}

	// For pre-GST dates, validate old tax fields (using existing DateIssued since it can't change)
	if existingReceipt.DateIssued.Time.Before(gstDate) {
		h.GstRate = 0
		if h.ServiceTax == 0 && h.SwatchBharatCess == 0 && h.KrishiKalyanCess == 0 {
			return nil, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Required missing values for 'serviceTax', 'swatchBharatCess' or 'krishiKalyanCess' for entries before 1 July 2017.",
			}
		}
	}

	// Validate mode
	mode := custom.ReceiptMode(h.Mode)
	if !mode.IsValid() {
		return nil, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid mode value.",
		}
	}

	// Negative amounts only for ADJUSTMENT
	if mode != custom.ADJUSTMENT && h.TotalAmount < 0 {
		return nil, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Negative receipt amount are only allowed for Adjustment.",
		}
	}

	// Bank details required for certain modes
	if mode.RequireBankDetails() {
		if strings.TrimSpace(h.BankName) == "" || strings.TrimSpace(h.TransactionNumber) == "" {
			return nil, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Required missing values: Bank Name or Transaction Number",
			}
		}
	}

	// Validate same society
	receiptSocietyInfo := CreateReceiptSocietyInfoService(db, uuid.MustParse(receiptId))
	err = common.IsSameSociety(receiptSocietyInfo, orgId, society)
	if err != nil {
		return nil, err
	}

	return &existingReceipt, nil
}

func (h *hUpdateSaleReceipt) execute(db *gorm.DB, orgId, society, receiptId string) (*models.Receipt, error) {
	existingReceipt, err := h.validate(db, orgId, society, receiptId)
	if err != nil {
		return nil, err
	}

	// Prepare updated fields
	updates := map[string]interface{}{
		"receipt_number":      h.ReceiptNumber,
		"total_amount":        decimal.NewFromFloat(h.TotalAmount),
		"mode":                custom.ReceiptMode(h.Mode),
		"bank_name":           h.BankName,
		"transaction_number":  h.TransactionNumber,
		"amount":              decimal.NewFromFloat(h.TotalAmount),
		"cgst":                nil,
		"sgst":                nil,
		"service_tax":         nil,
		"swathch_bharat_cess": nil,
		"krishi_kalyan_cess":  nil,
	}

	mode := custom.ReceiptMode(h.Mode)

	if mode != custom.ADJUSTMENT {
		if !existingReceipt.DateIssued.Time.Before(gstDate) {
			// Post-GST: Calculate GST using the exported method
			totalAmount := decimal.NewFromFloat(h.TotalAmount)
			tempReceipt := models.Receipt{TotalAmount: totalAmount}
			gstInfo := tempReceipt.CalcGST(h.GstRate)
			updates["amount"] = gstInfo.Amount
			updates["cgst"] = gstInfo.CGST
			updates["sgst"] = gstInfo.SGST
		} else {
			// Pre-GST: Use old tax system
			amount := decimal.NewFromFloat(h.TotalAmount)

			if h.ServiceTax > 0 {
				tax := decimal.NewFromFloat(h.ServiceTax)
				amount = amount.Sub(tax)
				updates["service_tax"] = tax
			}

			if h.SwatchBharatCess > 0 {
				tax := decimal.NewFromFloat(h.SwatchBharatCess)
				amount = amount.Sub(tax)
				updates["swathch_bharat_cess"] = tax
			}

			if h.KrishiKalyanCess > 0 {
				tax := decimal.NewFromFloat(h.KrishiKalyanCess)
				amount = amount.Sub(tax)
				updates["krishi_kalyan_cess"] = tax
			}

			updates["amount"] = amount
		}
	}

	// Update the receipt
	err = db.Model(&models.Receipt{}).Where("id = ?", receiptId).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	// Fetch and return updated receipt
	var updatedReceipt models.Receipt
	err = db.First(&updatedReceipt, "id = ?", receiptId).Error
	return &updatedReceipt, err
}

func (s *receiptService) updateSaleReceipt(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	receiptId := chi.URLParam(r, "receiptId")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateSaleReceipt](w, r)
	if reqBody == nil {
		return
	}

	receipt, err := reqBody.execute(s.db, orgId, societyRera, receiptId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated receipt."
	response.Data = receipt

	payload.EncodeJSON(w, http.StatusOK, response)
}