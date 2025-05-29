package broker

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type hGetAllSocietyBrokers struct{}

func (h *hGetAllSocietyBrokers) execute(db *gorm.DB, orgId, society, cursor string) (*custom.PaginatedData, error) {
	var brokerData []models.Broker

	query := db.Where("org_id = ? and society_id = ?", orgId, society).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	err := query.Find(&brokerData).Error
	if err != nil {
		return nil, err
	}
	return common.CreatePaginatedResponse(&brokerData), nil
}

func (s *brokerService) getAllSocietyBrokers(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")

	broker := hGetAllSocietyBrokers{}
	res, err := broker.execute(s.db, orgId, societyRera, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}

// start time should be greater than end time
// filtering is based on the past
type hGetBrokerReport struct {
	RecordsFrom time.Time
	RecordsTill time.Time
}

func (h *hGetBrokerReport) validate(db *gorm.DB, orgId, society, brokerId string) error {
	brokerSocietyInfo := CreateBrokerSocietyInfoService(db, uuid.MustParse(brokerId))
	return common.IsSameSociety(brokerSocietyInfo, orgId, society)
}

func (h *hGetBrokerReport) execute(db *gorm.DB, orgId, society, brokerId string) (*models.BrokerReport, error) {
	err := h.validate(db, orgId, society, brokerId)
	if err != nil {
		return nil, err
	}

	brokerModel := models.Broker{
		Id: uuid.MustParse(brokerId),
	}
	err = db.Preload("Sales", func(db *gorm.DB) *gorm.DB {
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
	}).Preload("Sales.Flat").
		Preload("Sales.Customers").
		Preload("Sales.CompanyCustomer").
		First(&brokerModel).Error

	totalAmount := decimal.Zero
	totalPaid := decimal.Zero

	for _, sale := range brokerModel.Sales {
		totalAmount = totalAmount.Add(sale.TotalPrice)

		for _, receipt := range sale.Receipts {
			if receipt.Cleared != nil {
				totalPaid = totalPaid.Add(receipt.TotalAmount)
			}
		}
	}

	return &models.BrokerReport{
		Finance: models.Finance{
			Total:     totalAmount,
			Paid:      totalPaid,
			Remaining: totalAmount.Sub(totalPaid),
		},
		Details: brokerModel,
	}, err
}

func (s *brokerService) getBrokerReport(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	brokerId := chi.URLParam(r, "brokerId")

	reqBody := payload.ValidateAndDecodeRequest[hGetBrokerReport](w, r)
	if reqBody == nil {
		return
	}

	report, err := reqBody.execute(s.db, orgId, societyRera, brokerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = report

	payload.EncodeJSON(w, http.StatusOK, response)
}
