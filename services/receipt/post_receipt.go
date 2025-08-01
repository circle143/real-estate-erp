package receipt

import (
	"net/http"
	"strings"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/bank"
	"circledigital.in/real-state-erp/services/sale"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TODO: later handle automatic generation for receipt number
type hCreateSaleReceipt struct {
	ReceiptNumber     string          `validate:"required"`
	TotalAmount       float64         `validate:"required"`
	Mode              string          `validate:"required"`
	DateIssued        custom.DateOnly `validate:"required"`
	BankName          string
	TransactionNumber string
	GstRate           int
}

func (h *hCreateSaleReceipt) validate(db *gorm.DB, orgId, society, saleId string) error {
	if h.GstRate != 5 && h.GstRate != 1 {
		h.GstRate = 5
		// return &custom.RequestError{
		// 	Status:  http.StatusBadRequest,
		// 	Message: "Invalid gst rate. Rate should be either 1 or 5.",
		// }
	}

	mode := custom.ReceiptMode(h.Mode)
	if !mode.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid mode value.",
		}
	}

	if mode.RequireBankDetails() {
		// check details
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
		TransactionNumber: h.TransactionNumber,
		BankName:          h.BankName,
		Mode:              custom.ReceiptMode(h.Mode),
		DateIssued:        h.DateIssued,
	}

	gstInfo := receiptModel.CalcGST(h.GstRate)
	receiptModel.Amount = gstInfo.Amount
	receiptModel.SGST = gstInfo.SGST
	receiptModel.CGST = gstInfo.CGST

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
