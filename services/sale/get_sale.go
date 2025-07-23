package sale

import (
	"net/http"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/tower"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

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
		if plan.ConditionType == custom.WITHINDAYS {
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
	paymentPlans = common.SortDbModels(paymentPlans)

	// total amount paid
	var totalPaid decimal.Decimal
	err = db.Table("receipts AS r").
		Select("COALESCE(SUM(r.total_amount), 0)").
		Joins("JOIN receipt_clears c ON c.receipt_id = r.id").
		Where("r.sale_id = ?", saleId).
		Group("r.sale_id").
		Scan(&totalPaid).Error
	if err != nil {
		return nil, err
	}

	totalPaidCpy := totalPaid
	// total amount according to active payment plans
	var total = decimal.Zero
	for i, plan := range paymentPlans {
		percent := decimal.NewFromInt(int64(plan.Amount)) // Convert int to decimal
		amount := sale.TotalPrice.Mul(percent).Div(decimal.NewFromInt(100))

		paymentPlans[i].TotalAmount = &amount
		total = total.Add(amount)

		// Distribute totalPaidCpy
		if totalPaidCpy.GreaterThanOrEqual(amount) {
			paymentPlans[i].AmountPaid = &amount
			paymentPlans[i].Remaining = decimalPtr(decimal.Zero)
			totalPaidCpy = totalPaidCpy.Sub(amount)
		} else if totalPaidCpy.GreaterThan(decimal.Zero) {
			amountPaid := totalPaidCpy
			paymentPlans[i].AmountPaid = &amountPaid
			remaining := amount.Sub(totalPaidCpy)
			paymentPlans[i].Remaining = decimalPtr(remaining)
			totalPaidCpy = decimal.Zero
		} else {
			paymentPlans[i].AmountPaid = decimalPtr(decimal.Zero)
			paymentPlans[i].Remaining = decimalPtr(amount)
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
	var total decimal.Decimal
	err := db.Model(&models.Sale{}).
		Where("org_id = ? AND society_id = ?", orgId, society).
		Select("COALESCE(SUM(total_price), 0)"). // Use COALESCE to avoid null
		Scan(&total).Error

	if err != nil {
		return nil, err
	}

	var paid decimal.Decimal
	err = db.
		Table("receipts AS r").
		Select("COALESCE(SUM(r.total_amount), 0) AS total_paid_amount").
		Joins("JOIN receipt_clears c ON c.receipt_id = r.id").
		Joins("JOIN sales s ON s.id = r.sale_id").
		Where("s.society_id = ? AND s.org_id = ?", society, orgId).
		Scan(&paid).Error
	if err != nil {
		return nil, err
	}

	return &models.PaymentReport{
		Total:   total,
		Paid:    paid,
		Pending: total.Sub(paid),
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
	err = db.
		//Preload("FlatType").
		Preload("SaleDetail").
		Preload("SaleDetail.Customers").
		Preload("SaleDetail.CompanyCustomer").
		Preload("SaleDetail.Broker").
		Preload("SaleDetail.Receipts").
		Preload("SaleDetail.Receipts.Cleared").
		Preload("SaleDetail.Receipts.Cleared.Bank").
		Joins("JOIN sales s ON s.flat_id = flats.id").
		Where("flats.tower_id = ?", towerId).
		Find(&soldFlats).Error
	if err != nil {
		return nil, err
	}

	// 2 -> get total sale amount
	totalAmountTower := decimal.Zero
	totalAmountTowerPaymentPlan := decimal.Zero
	totalTowerPaid := decimal.Zero

	// 3 -> create map flatId -> flatStatsInfo
	type flatStatInfo struct {
		total decimal.Decimal
		paid  decimal.Decimal
	}
	flatsMap := make(map[uuid.UUID]flatStatInfo)

	// populate map and tower amount
	for _, flat := range soldFlats {
		flatId := flat.Id
		total := flat.SaleDetail.TotalPrice

		paid := decimal.Zero
		for _, receipt := range flat.SaleDetail.Receipts {
			if receipt.Cleared != nil {
				paid = paid.Add(receipt.TotalAmount)
			}
		}

		flatsMap[flatId] = flatStatInfo{
			total: total,
			paid:  paid,
		}
		totalAmountTower = totalAmountTower.Add(total)
	}

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

	// all payment plans combined and // sort them
	paymentPlans = append(directPlans, paymentPlans...)
	paymentPlans = common.SortDbModels(paymentPlans)

	// create tower payment breakdown
	var towerReportPaymentBreakdown []models.TowerReportPaymentBreakdown
	for _, plan := range paymentPlans {
		totalPlanAmount := decimal.Zero
		totalPlanPaid := decimal.Zero

		percent := decimal.NewFromInt(int64(plan.Amount)).Div(decimal.NewFromInt(100))

		var paidItems []models.TowerReportPaymentBreakdownItem
		var unpaidItems []models.TowerReportPaymentBreakdownItem

		for _, flat := range soldFlats {
			flatId := flat.Id
			flatPaidRem := flatsMap[flatId].paid
			isPaid := false

			flatTotalPaymentPlan := flatsMap[flatId].total.Mul(percent)
			flatPaid := decimal.Zero
			flatRemaining := decimal.Zero

			if flatPaidRem.GreaterThanOrEqual(flatTotalPaymentPlan) {
				isPaid = true

				flatPaid = flatTotalPaymentPlan
				flatPaidRem = flatPaidRem.Sub(flatTotalPaymentPlan)
			} else if flatPaidRem.GreaterThan(decimal.Zero) {
				flatPaid = flatPaidRem
				flatPaidRem = decimal.Zero
				flatRemaining = flatTotalPaymentPlan.Sub(flatPaid)
			} else {
				flatRemaining = flatTotalPaymentPlan
			}

			totalPlanAmount = totalPlanAmount.Add(flatTotalPaymentPlan)
			totalPlanPaid = totalPlanPaid.Add(flatPaid)

			// update flat
			flat := flatsMap[flatId]
			flat.paid = flatPaidRem
			flatsMap[flatId] = flat

			// create payment plan item
			paymentPlanItem := models.TowerReportPaymentBreakdownItem{
				FlatId:    flatId,
				Total:     flatTotalPaymentPlan,
				Paid:      flatPaid,
				Remaining: flatRemaining,
			}

			// add to correct slice
			if isPaid {
				paidItems = append(paidItems, paymentPlanItem)
			} else {
				unpaidItems = append(unpaidItems, paymentPlanItem)
			}
		}

		totalAmountTowerPaymentPlan = totalAmountTowerPaymentPlan.Add(totalPlanAmount)
		totalTowerPaid = totalTowerPaid.Add(totalPlanPaid)

		towerPaymentPlanItem := models.TowerReportPaymentBreakdown{
			PaymentPlan: plan,
			Total:       totalPlanAmount,
			Paid:        totalPlanPaid,
			Remaining:   totalPlanAmount.Sub(totalPlanPaid),
			PaidItems:   paidItems,
			UnpaidItems: unpaidItems,
		}
		towerReportPaymentBreakdown = append(towerReportPaymentBreakdown, towerPaymentPlanItem)
	}

	return &models.TowerReport{
		Flats: soldFlats,
		Overall: models.Finance{
			Total:     totalAmountTower,
			Paid:      totalTowerPaid,
			Remaining: totalAmountTower.Sub(totalTowerPaid),
		},
		PaymentPlan: models.Finance{
			Total:     totalAmountTowerPaymentPlan,
			Paid:      totalTowerPaid,
			Remaining: totalAmountTowerPaymentPlan.Sub(totalTowerPaid),
		},
		PaymentBreakdown: towerReportPaymentBreakdown,
	}, nil
}

// todo fix this
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
