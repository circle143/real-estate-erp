package sale

import (
	"net/http"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/broker"
	"circledigital.in/real-state-erp/services/flat"
	payment_plan_group "circledigital.in/real-state-erp/services/payment-plan-group"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type optionalChargesDetails struct {
	Name      string  `validate:"required"`
	TotalCost float64 `validate:"required"`
}

type hCreateSale struct {
	SaleNumber string            `validate:"required"`
	Type       string            `validate:"required"`
	Details    []customerDetails `validate:"omitempty,dive"`
	BasicCost  float64           `validate:"required"`
	PaymentId  string            `validate:"required,uuid"`
	//OptionalCharges []string
	OtherCharges []optionalChargesDetails `validate:"omitempty,dive"`
	CompanyBuyer companyCustomerDetails   `validate:"omitempty"`
	BrokerId     string                   `validate:"required,uuid"`
}

func (h *hCreateSale) validate(db *gorm.DB, orgId, society, flatId, paymentId string) error {
	// check type and validate
	buyerType := saleBuyerType(h.Type)
	if !buyerType.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid sale buyer type.",
		}
	}

	if buyerType == user {
		if len(h.Details) == 0 {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Missing buyer details.",
			}
		}
		for _, detail := range h.Details {
			err := detail.validate()
			if err != nil {
				return err
			}
		}
	} else {
		err := h.CompanyBuyer.validate()
		if err != nil {
			return err
		}
	}

	societyInfoService := flat.CreateFlatSocietyInfoService(db, uuid.MustParse(flatId))
	err := common.IsSameSociety(societyInfoService, orgId, society)
	if err != nil {
		return err
	}

	paymentInfoService := payment_plan_group.CreatePaymentPlanSocietyInfoService(db, uuid.MustParse(paymentId))
	err = common.IsSameSociety(paymentInfoService, orgId, society)
	if err != nil {
		return err
	}

	brokerSocietyInfoService := broker.CreateBrokerSocietyInfoService(db, uuid.MustParse(h.BrokerId))
	return common.IsSameSociety(brokerSocietyInfoService, orgId, society)
}

func (h *hCreateSale) execute(db *gorm.DB, orgId, society, flatId string) error {
	err := h.validate(db, orgId, society, flatId, h.PaymentId)
	if err != nil {
		return err
	}
	basicCost := decimal.NewFromFloat(h.BasicCost)
	// if err != nil {
	// 	return &custom.RequestError{
	// 		Status:  http.StatusBadRequest,
	// 		Message: "Invalid basic cost provided.",
	// 	}
	// }
	buyerType := saleBuyerType(h.Type)

	return db.Transaction(func(tx *gorm.DB) error {
		flatModel := models.Flat{
			Id: uuid.MustParse(flatId),
		}
		err := tx.
			//Preload("FlatType").
			First(&flatModel).Error
		if err != nil {
			return err
		}
		//superArea := flatModel.FlatType.SuperArea
		salableArea := flatModel.SaleableArea

		// get required preference location charges
		//var locationCharges []models.PreferenceLocationCharge
		//locationChargeQuery := tx.
		//	Where("org_id = ? and society_id = ? and disable = false", orgId, society)
		//if flatModel.Facing == custom.SPECIAL {
		//	locationChargeQuery = locationChargeQuery.Where(
		//		"(type = ? AND floor = ?) OR type = ?",
		//		custom.FLOOR, flatModel.FloorNumber, custom.FACING,
		//	)
		//} else {
		//	locationChargeQuery = locationChargeQuery.Where("type = ? AND floor = ?", custom.FLOOR, flatModel.FloorNumber)
		//}
		//
		//err = locationChargeQuery.Find(&locationCharges).Error
		//if err != nil {
		//	return err
		//}

		// other charges
		//var otherCharges []models.OtherCharge
		//err = tx.
		//	Where("org_id = ? and society_id = ? and disable = false and optional = false", orgId, society).
		//	Find(&otherCharges).Error
		//if err != nil {
		//	return err
		//}

		// optional charges
		//var optionalCharges []models.OtherCharge
		//err = tx.
		//	Where("org_id = ? and society_id = ? and disable = false and optional = true", orgId, society).
		//	Where("id in ?", h.OptionalCharges).
		//	Find(&optionalCharges).Error
		//if err != nil {
		//	return err
		//}

		// price calculation
		var priceBreakdowns []models.PriceBreakdownDetail
		totalPrice := decimal.NewFromInt(0)

		// basic cost
		basicCostDetail := models.PriceBreakdownDetail{
			Type:    "basic-cost",
			Price:   basicCost,
			Summary: "Basic flat cost",
			// Total:     salableArea.Mul(basicCost),
			Total:     basicCost,
			SuperArea: salableArea,
		}
		totalPrice = totalPrice.Add(basicCostDetail.Total)
		//log.Printf("total price: %v\n", totalPrice.String())
		//log.Printf("basic cost detail: %v\n\n", basicCostDetail.Total.String())

		priceBreakdowns = append(priceBreakdowns, basicCostDetail)

		// Add location charges
		//for _, charge := range locationCharges {
		//	detail := models.PriceBreakdownDetail{
		//		Type:      "preference-location",
		//		Price:     charge.Price,
		//		Summary:   charge.Summary,
		//		Total:     salableArea.Mul(charge.Price),
		//		SuperArea: salableArea,
		//	}
		//	totalPrice = totalPrice.Add(detail.Total)
		//	//log.Printf("total price: %v\n", totalPrice.String())
		//	//log.Printf("%v: %v\n\n", detail.Summary, detail.Total.String())
		//	priceBreakdowns = append(priceBreakdowns, detail)
		//}

		// add other charges
		for _, chargeDetails := range h.OtherCharges {
			priceDecimal := decimal.NewFromFloat(chargeDetails.TotalCost)
			detail := models.PriceBreakdownDetail{
				Type:      "other-charges",
				Price:     priceDecimal,
				Summary:   chargeDetails.Name,
				SuperArea: salableArea,
				Total:     priceDecimal,
			}
			totalPrice = totalPrice.Add(priceDecimal)
			priceBreakdowns = append(priceBreakdowns, detail)
		}

		// Helper to process other/optional charges
		//processOtherCharges := func(charges []models.OtherCharge) {
		//	for _, charge := range charges {
		//		detail := models.PriceBreakdownDetail{
		//			Type:      "other",
		//			Price:     charge.Price,
		//			Summary:   charge.Summary,
		//			SuperArea: salableArea,
		//		}
		//
		//		if charge.Recurring && charge.AdvanceMonths >= 1 {
		//			advanceMonths := decimal.NewFromInt(int64(charge.AdvanceMonths))
		//			if charge.Fixed {
		//				detail.Total = charge.Price.Mul(advanceMonths)
		//			} else {
		//				detail.Total = salableArea.Mul(charge.Price).Mul(advanceMonths)
		//			}
		//		} else {
		//			if charge.Fixed {
		//				detail.Total = charge.Price
		//			} else {
		//				detail.Total = salableArea.Mul(charge.Price)
		//			}
		//		}
		//
		//		totalPrice = totalPrice.Add(detail.Total)
		//		//log.Printf("total price: %v\n", totalPrice.String())
		//		//log.Printf("%v: %v\n\n", detail.Summary, detail.Total.String())
		//		priceBreakdowns = append(priceBreakdowns, detail)
		//	}
		//}
		//
		//processOtherCharges(otherCharges)
		//processOtherCharges(optionalCharges)

		//log.Printf("total price: %v\n", totalPrice.String())
		//
		//for i, pb := range priceBreakdowns {
		//	fmt.Printf("Item %d:\n", i+1)
		//	fmt.Printf("  Type:    %s\n", pb.Type)
		//	fmt.Printf("  Price:   %v\n", pb.Price.String())
		//	fmt.Printf("  Summary: %s\n", pb.Summary)
		//	fmt.Printf("  SuperArea:   %v\n", pb.SuperArea.String())
		//	fmt.Printf("  Total:   %v\n", pb.Total.String())
		//}

		//return &custom.RequestError{
		//	Status:  http.StatusBadRequest,
		//	Message: "trial",
		//}

		saleModel := models.Sale{
			SaleNumber:         h.SaleNumber,
			FlatId:             uuid.MustParse(flatId),
			SocietyId:          society,
			OrgId:              uuid.MustParse(orgId),
			TotalPrice:         totalPrice,
			PriceBreakdown:     priceBreakdowns,
			BrokerId:           uuid.MustParse(h.BrokerId),
			PaymentPlanRatioId: uuid.MustParse(h.PaymentId),
		}
		err = tx.Create(&saleModel).Error
		if err != nil {
			return err
		}

		if buyerType == user {
			customers := make([]*models.Customer, 0, len(h.Details))
			for _, d := range h.Details {
				customer := &models.Customer{
					SaleId:           saleModel.Id,
					Salutation:       custom.Salutation(d.Salutation),
					FirstName:        d.FirstName,
					LastName:         d.LastName,
					DateOfBirth:      d.DateOfBirth,
					Gender:           custom.Gender(d.Gender),
					Photo:            d.Photo,
					MaritalStatus:    custom.MaritalStatus(d.MaritalStatus),
					Nationality:      custom.Nationality(d.Nationality),
					Email:            d.Email,
					PhoneNumber:      d.PhoneNumber,
					MiddleName:       d.MiddleName,
					NumberOfChildren: d.NumberOfChildren,
					AnniversaryDate:  d.AnniversaryDate,
					AadharNumber:     d.AadharNumber,
					PanNumber:        d.PanNumber,
					PassportNumber:   d.PassportNumber,
					Profession:       d.Profession,
					Designation:      d.Designation,
					CompanyName:      d.CompanyName,
				}
				customers = append(customers, customer)
			}
			return tx.Create(customers).Error
		} else {
			companyBuyer := models.CompanyCustomer{
				SaleId:       saleModel.Id,
				Name:         h.CompanyBuyer.Name,
				CompanyPan:   h.CompanyBuyer.CompanyPan,
				CompanyGst:   h.CompanyBuyer.CompanyGst,
				AadharNumber: h.CompanyBuyer.AadharNumber,
				PanNumber:    h.CompanyBuyer.PanNumber,
			}

			return tx.Create(&companyBuyer).Error
		}
	})
}

func (s *saleService) createSale(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	flatId := chi.URLParam(r, "flat")

	reqBody := payload.ValidateAndDecodeRequest[hCreateSale](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId, societyRera, flatId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created sale record."

	payload.EncodeJSON(w, http.StatusCreated, response)
}

//type hAddPaymentInstallmentForSale struct{}
//
//func (h *hAddPaymentInstallmentForSale) validate(db *gorm.DB, orgId, society, saleId, paymentId string) error {
//	// validate payment permission
//	paymentSocietyInfoService := paymentPlan.CreatePaymentPlanSocietyInfoService(db, uuid.MustParse(paymentId))
//	err := common.IsSameSociety(paymentSocietyInfoService, orgId, society)
//	if err != nil {
//		return err
//	}
//
//	saleSocietyInfo := CreateSaleSocietyInfoService(db, uuid.MustParse(saleId))
//	err = common.IsSameSociety(saleSocietyInfo, orgId, society)
//	if err != nil {
//		return err
//	}
//
//	// check payment scope
//	paymentModel := models.PaymentPlan{
//		Id: uuid.MustParse(paymentId),
//	}
//	err = db.Find(&paymentModel).Error
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return &custom.RequestError{
//				Status:  http.StatusBadRequest,
//				Message: "Invalid payment plan selected.",
//			}
//		}
//	}
//
//	// if direct just return
//	if paymentModel.Scope == custom.DIRECT {
//		return nil
//	}
//
//	// check if payment is active for the tower
//	var status models.TowerPaymentStatus
//	err = db.
//		Joins("JOIN flats ON flats.tower_id = tower_payment_statuses.tower_id").
//		Joins("JOIN sales ON sales.flat_id = flats.id").
//		Where("sales.id = ?", saleId).
//		First(&status).Error
//
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return &custom.RequestError{
//				Status:  http.StatusBadRequest,
//				Message: "Payment plan is not active for the flat.",
//			}
//		}
//		return err
//	}
//
//	return nil
//}
//
//func (h *hAddPaymentInstallmentForSale) execute(db *gorm.DB, orgId, society, saleId, paymentId string) (*models.SalePaymentStatus, error) {
//	err := h.validate(db, orgId, society, saleId, paymentId)
//	if err != nil {
//		return nil, err
//	}
//
//	// calc payment
//	plan := models.PaymentPlan{
//		Id: uuid.MustParse(paymentId),
//	}
//	err = db.
//		First(&plan).Error
//	if err != nil {
//		return nil, err
//	}
//
//	// get sale record
//	sale := models.Sale{
//		Id: uuid.MustParse(saleId),
//	}
//	err = db.
//		First(&sale).Error
//	if err != nil {
//		return nil, err
//	}
//
//	// amount to be paid for this plan
//	percent := decimal.NewFromInt(int64(plan.Amount)) // e.g., 5 means 5%
//	amount := sale.TotalPrice.Mul(percent).Div(decimal.NewFromInt(100))
//	salePaymentModel := models.SalePaymentStatus{
//		PaymentId: uuid.MustParse(paymentId),
//		SaleId:    uuid.MustParse(saleId),
//		Amount:    amount,
//	}
//
//	err = db.Create(&salePaymentModel).Error
//	return &salePaymentModel, err
//}
//
//func (s *saleService) addPaymentInstallmentForSale(w http.ResponseWriter, r *http.Request) {
//	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
//	societyRera := chi.URLParam(r, "society")
//	saleId := chi.URLParam(r, "saleId")
//	paymentId := chi.URLParam(r, "paymentId")
//
//	addPayment := hAddPaymentInstallmentForSale{}
//	res, err := addPayment.execute(s.db, orgId, societyRera, saleId, paymentId)
//	if err != nil {
//		payload.HandleError(w, err)
//		return
//	}
//
//	var response custom.JSONResponse
//	response.Error = false
//	response.Message = "Successfully added installment for the sale."
//	response.Data = res
//
//	payload.EncodeJSON(w, http.StatusCreated, response)
//}
