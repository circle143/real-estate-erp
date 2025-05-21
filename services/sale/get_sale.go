package sale

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/tower"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
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
		Where("org_id = ? and society_id = ? and scope = ?", orgId, society, custom.DIRECT).
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

type hGetTowerSalesReport struct{}

func (h *hGetTowerSalesReport) validate(db *gorm.DB, orgId, society, towerId string) error {
	societyInfoService := tower.CreateTowerSocietyInfoService(db, uuid.MustParse(towerId))
	return common.IsSameSociety(societyInfoService, orgId, society)
}

func (h *hGetTowerSalesReport) execute(db *gorm.DB, orgId, society, towerId string) (*models.TowerReport, error) {
	err := h.validate(db, orgId, society, towerId)
	if err != nil {
		return nil, err
	}

	// 1 -> get all tower sold flats
	var soldFlats []models.Flat
	err = db.Preload("SaleDetail").
		Preload("SaleDetail.Customers").
		Preload("SaleDetail.CompanyCustomer").
		Preload("SaleDetail.Broker").
		Joins("JOIN sales s ON s.flat_id = flats.id").
		Where("flats.tower_id = ?", towerId).
		Find(&soldFlats).Error
	if err != nil {
		return nil, err
	}

	// 2 -> create map saleId -> models.Flat for easy lookup and creating final response
	flatsMap := make(map[uuid.UUID]models.Flat)
	for _, flat := range soldFlats {
		saleId := flat.SaleDetail.Id
		flatsMap[saleId] = flat
	}

	// 3 -> get total sale amount
	var totalAmountTower decimal.Decimal
	err = db.
		Table("towers t").
		Select("COALESCE(SUM(s.total_price), 0)").
		Joins("JOIN flats f ON f.tower_id = t.id").
		Joins("JOIN sales s ON s.flat_id = f.id").
		Where("t.id = ?", towerId).
		Scan(&totalAmountTower).Error

	// 4 -> get all direct plans
	var directPlans []models.PaymentPlan
	err = db.
		Model(&models.PaymentPlan{}).
		Where("org_id = ? and society_id = ? and scope = ?", orgId, society, custom.DIRECT).
		Find(&directPlans).Error
	if err != nil {
		return nil, err
	}

	// 5 -> get all tower active payment plans
	var paymentPlans []models.PaymentPlan
	err = db.
		Model(&models.PaymentPlan{}).
		Joins("JOIN tower_payment_statuses tps ON tps.payment_id = payment_plans.id").
		Where("tps.tower_id = ?", towerId).
		Select("payment_plans.*").
		Scan(&paymentPlans).Error
	if err != nil {
		return nil, err
	}

	// all payment plans combined
	paymentPlans = append(directPlans, paymentPlans...)
	var paymentPlansIds []uuid.UUID

	var towerReportPaymentBreakdown []models.TowerReportPaymentBreakdown
	towerReportPaymentBreakdownMap := make(map[uuid.UUID]models.TowerReportPaymentBreakdown)

	// 6 -> update each plan amount paid
	for _, plan := range paymentPlans {
		percent := decimal.NewFromInt(int64(plan.Amount))
		paymentReport := models.TowerReportPaymentBreakdown{
			PaymentPlan: plan,
			Total:       totalAmountTower.Mul(percent).Div(decimal.NewFromInt(100)),
		}
		towerReportPaymentBreakdown = append(towerReportPaymentBreakdown, paymentReport)
		towerReportPaymentBreakdownMap[plan.Id] = paymentReport
		paymentPlansIds = append(paymentPlansIds, plan.Id)
	}

	// 7 -> get paid amount for each plan and paid sale items
	type PaymentSummary struct {
		PaymentID uuid.UUID
		TotalPaid decimal.Decimal
		SaleIDs   pq.StringArray `gorm:"type:[]text"`
	}
	var paymentSummaries []PaymentSummary

	err = db.Raw(
		`
		SELECT
			ps.payment_id,
			COALESCE(SUM(ps.amount), 0) AS total_paid,
			ARRAY_AGG(DISTINCT ps.sale_id) AS sale_ids
		FROM sales s
		JOIN flats f ON s.flat_id = f.id
		JOIN towers t ON t.id = f.tower_id
		JOIN sale_payment_statuses ps ON s.id = ps.sale_id
		WHERE
		t.id = ?
		AND ps.payment_id IN ?
		GROUP BY ps.payment_id
	`,
		towerId,
		paymentPlansIds,
	).Scan(&paymentSummaries).Error
	if err != nil {
		return nil, err
	}

	// Convert existing summaries to a map for quick lookup
	summaryMap := make(map[uuid.UUID]PaymentSummary)
	for _, summary := range paymentSummaries {
		summaryMap[summary.PaymentID] = summary
	}

	// Add missing payment plans with default values
	for _, paymentId := range paymentPlansIds {
		if _, found := summaryMap[paymentId]; !found {
			summaryMap[paymentId] = PaymentSummary{
				PaymentID: paymentId,
				TotalPaid: decimal.Zero,
				SaleIDs:   pq.StringArray{},
			}
		}
	}

	// Convert the map back to a slice
	paymentSummaries = make([]PaymentSummary, 0, len(summaryMap))
	for _, summary := range summaryMap {
		paymentSummaries = append(paymentSummaries, summary)
	}

	totalTowerPaid := decimal.Zero
	totalAmountTowerPaymentPlan := decimal.Zero
	// 8 -> populate payment breakdown
	for _, summary := range paymentSummaries {
		paymentId := summary.PaymentID
		saleIds := summary.SaleIDs

		saleIdMap := make(map[string]struct{}, len(saleIds))
		for _, sid := range saleIds {
			saleIdMap[sid] = struct{}{}
		}

		if paymentBreakdown, ok := towerReportPaymentBreakdownMap[paymentId]; ok {
			percent := decimal.NewFromInt(int64(paymentBreakdown.Amount))
			paymentBreakdown.Paid = summary.TotalPaid
			paymentBreakdown.Remaining = paymentBreakdown.Total.Sub(summary.TotalPaid)

			totalTowerPaid = totalTowerPaid.Add(summary.TotalPaid)
			totalAmountTowerPaymentPlan = totalAmountTowerPaymentPlan.Add(paymentBreakdown.Total)

			var paidItems []models.TowerReportPaymentBreakdownItem
			var unpaidItems []models.TowerReportPaymentBreakdownItem

			for _, flat := range soldFlats {
				id := flat.SaleDetail.Id.String()
				totalAmount := flat.SaleDetail.TotalPrice
				paymentPlanAmount := totalAmount.Mul(percent).Div(decimal.NewFromInt(100))
				paymentBreakdownItem := models.TowerReportPaymentBreakdownItem{
					//Flat:   flat,
					FlatId: flat.Id,
					Amount: paymentPlanAmount,
				}

				if _, found := saleIdMap[id]; found {
					paidItems = append(paidItems, paymentBreakdownItem)
				} else {
					unpaidItems = append(unpaidItems, paymentBreakdownItem)
				}
			}

			paymentBreakdown.PaidItems = paidItems
			paymentBreakdown.UnpaidItems = unpaidItems
			towerReportPaymentBreakdownMap[paymentId] = paymentBreakdown
		}
	}

	towerReportPaymentBreakdown = make([]models.TowerReportPaymentBreakdown, 0, len(towerReportPaymentBreakdownMap))
	for _, breakdown := range towerReportPaymentBreakdownMap {
		towerReportPaymentBreakdown = append(towerReportPaymentBreakdown, breakdown)
	}
	return &models.TowerReport{
		Flats: soldFlats,
		Overall: models.TowerFinance{
			Total:     totalAmountTower,
			Paid:      totalTowerPaid,
			Remaining: totalAmountTower.Sub(totalTowerPaid),
		},
		PaymentPlan: models.TowerFinance{
			Total:     totalAmountTowerPaymentPlan,
			Paid:      totalTowerPaid,
			Remaining: totalAmountTowerPaymentPlan.Sub(totalTowerPaid),
		},
		PaymentBreakdown: towerReportPaymentBreakdown,
	}, nil
}

func (s *saleService) getTowerSalesReport(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	towerId := chi.URLParam(r, "towerId")

	report := hGetTowerSalesReport{}
	res, err := report.execute(s.db, orgId, societyRera, towerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
