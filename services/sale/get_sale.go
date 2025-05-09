package sale

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

type hGetSalePaymentBreakDown struct{}

func (h *hGetSalePaymentBreakDown) validate(db *gorm.DB, orgId, society, saleId string) error {
	saleSocietyInfo := CreateSaleSocietyInfoService(db, uuid.MustParse(saleId))
	return common.IsSameSociety(saleSocietyInfo, orgId, society)
}

func (h *hGetSalePaymentBreakDown) execute(db *gorm.DB, orgId, society, saleId string) (*[]models.PaymentPlan, error) {
	err := h.validate(db, orgId, society, saleId)
	if err != nil {
		return nil, err
	}

	var paymentPlans []models.PaymentPlan
	err = db.
		Joins("JOIN flats ON flats.id = sales.flat_id").
		Joins("JOIN towers ON towers.id = flats.tower_id").
		Joins("JOIN tower_payment_statuses tps ON tps.tower_id = towers.id").
		Joins("JOIN payment_plans pp ON pp.id = tps.payment_id").
		Model(&models.PaymentPlan{}).
		Where("sales.id = ?", saleId).
		Select("pp.*").
		Scan(&paymentPlans).Error
	if err != nil {
		return nil, err
	}

	var statuses []models.SalePaymentStatus
	err = db.
		Where("sale_id = ?", saleId).
		Find(&statuses).Error
	if err != nil {
		return nil, err
	}

	// Create a lookup map for PaymentId → Amount
	paidAmountMap := make(map[uuid.UUID]float64, len(statuses))
	for _, s := range statuses {
		paidAmountMap[s.PaymentId] = s.Amount
	}

	// Set Paid = true and add amount
	for i, p := range paymentPlans {
		if amt, ok := paidAmountMap[p.Id]; ok {
			value := true
			paymentPlans[i].Paid = &value
			paymentPlans[i].AmountPaid = &amt
		}
	}

	return &paymentPlans, nil
}

func (s *saleService) getSalePaymentBreakDown(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	saleId := chi.URLParam(r, "saleId")
	societyRera := chi.URLParam(r, "society")

	details := hGetSalePaymentBreakDown{}
	res, err := details.execute(s.db, orgId, societyRera, saleId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
