package sale

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
)

type hGetSalePaymentBreakDown struct{}

func (h *hGetSalePaymentBreakDown) validate(db *gorm.DB, orgId, society, saleId string) error {
	saleSocietyInfo := CreateSaleSocietyInfoService(db, uuid.MustParse(saleId))
	return common.IsSameSociety(saleSocietyInfo, orgId, society)
}

func (h *hGetSalePaymentBreakDown) execute(db *gorm.DB, orgId, society, saleId string) (*models.PaymentPlanSaleBreakDown, error) {
	err := h.validate(db, orgId, society, saleId)
	if err != nil {
		return nil, err
	}

	sale := models.Sale{
		Id: uuid.MustParse(saleId),
	}
	err = db.Find(&sale).Error
	if err != nil {
		return nil, err
	}

	// direct payment plans
	var directPlans []models.PaymentPlan
	err = db.
		Model(&models.PaymentPlan{}).
		Where("org_id = ? and society_id = ? and scope = ?", orgId, society, custom.DIRECT). // assuming custom.Direct is the correct enum value
		Find(&directPlans).Error
	if err != nil {
		return nil, err
	}

	for i := range directPlans {
		plan := &directPlans[i]
		if plan.ConditionType == custom.AFTERDAYS {
			due := plan.CreatedAt.AddDate(0, 0, plan.ConditionValue)
			plan.Due = &due
		}
	}

	// tower active payment plans
	var paymentPlans []models.PaymentPlan
	err = db.
		Model(&models.PaymentPlan{}).
		Joins("JOIN tower_payment_statuses tps ON tps.payment_id = payment_plans.id").
		Joins("JOIN towers ON towers.id = tps.tower_id").
		Joins("JOIN flats ON flats.tower_id = towers.id").
		Joins("JOIN sales ON sales.flat_id = flats.id").
		Where("sales.id = ?", saleId).
		Select("payment_plans.*").
		Scan(&paymentPlans).Error
	if err != nil {
		return nil, err
	}

	paymentPlans = append(directPlans, paymentPlans...)
	var statuses []models.SalePaymentStatus
	err = db.
		Where("sale_id = ?", saleId).
		Find(&statuses).Error
	if err != nil {
		return nil, err
	}

	// Create a lookup map for PaymentId → Amount
	paidAmountMap := make(map[uuid.UUID]decimal.Decimal, len(statuses))
	for _, s := range statuses {
		paidAmountMap[s.PaymentId] = s.Amount
	}

	// Set Paid = true and add amount
	for i, p := range paymentPlans {
		value := false
		var amount decimal.Decimal
		if amt, ok := paidAmountMap[p.Id]; ok {
			value = true
			amount = amt
		} else {
			percent := decimal.NewFromInt(int64(p.Amount)) // Convert int to decimal
			amount = sale.TotalPrice.Mul(percent).Div(decimal.NewFromInt(100))
		}
		paymentPlans[i].Paid = &value
		paymentPlans[i].AmountPaid = &amount
	}

	// payment
	var totalPaid = decimal.Zero
	var total = decimal.Zero

	for _, plan := range paymentPlans {
		if plan.AmountPaid == nil {
			continue // skip nil amount to avoid panic
		}

		total = total.Add(*plan.AmountPaid)

		if plan.Paid != nil && *plan.Paid {
			totalPaid = totalPaid.Add(*plan.AmountPaid)
		}
	}

	return &models.PaymentPlanSaleBreakDown{
		TotalAmount: total,
		PaidAmount:  totalPaid,
		Remaining:   total.Sub(totalPaid),
		Details:     paymentPlans,
	}, nil
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

type hGetSocietySalesReport struct{}

func (h *hGetSocietySalesReport) execute(db *gorm.DB, orgId, society string) (*models.PaymentReport, error) {
	var total float64
	err := db.Model(&models.Sale{}).
		Where("org_id = ? AND society_id = ?", orgId, society).
		Select("COALESCE(SUM(total_price), 0)"). // Use COALESCE to avoid null
		Scan(&total).Error

	if err != nil {
		return nil, err
	}

	var paid float64
	err = db.
		Joins("JOIN sales ON sales.id = sale_payment_statuses.sale_id").
		Model(&models.SalePaymentStatus{}).
		Where("sales.society_id = ? AND sales.org_id = ?", society, orgId).
		Select("COALESCE(SUM(sale_payment_statuses.amount), 0)").
		Scan(&paid).Error
	if err != nil {
		return nil, err
	}

	return &models.PaymentReport{
		Total:   total,
		Paid:    paid,
		Pending: total - paid,
	}, nil
}

func (s *saleService) getSocietySalesReport(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	report := hGetSocietySalesReport{}
	res, err := report.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
