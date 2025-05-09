package sale

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/flat"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hCreateSale struct {
	Details         []customerDetails `validate:"required,min=1,dive"`
	BasicCost       float64           `validate:"required"`
	OptionalCharges []string
}

func (ac *hCreateSale) validate(db *gorm.DB, orgId, society, flatId string) error {
	societyInfoService := flat.CreateFlatSocietyInfoService(db, uuid.MustParse(flatId))
	err := common.IsSameSociety(societyInfoService, orgId, society)
	if err != nil {
		return err
	}

	for _, detail := range ac.Details {
		err = detail.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (ac *hCreateSale) execute(db *gorm.DB, orgId, society, flatId string) error {
	err := ac.validate(db, orgId, society, flatId)
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {

		flatModel := models.Flat{
			Id: uuid.MustParse(flatId),
		}
		err := tx.Preload("FlatType").First(&flatModel).Error
		if err != nil {
			return err
		}
		superArea := flatModel.FlatType.SuperArea

		// get required preference location charges
		var locationCharges []models.PreferenceLocationCharge
		locationChargeQuery := tx.
			Where("org_id = ? and society_id = ? and disable = false", orgId, society).
			Where("type = ? and floor = ?", custom.FLOOR, flatModel.FloorNumber)
		if flatModel.Facing == custom.SPECIAL {
			locationChargeQuery = locationChargeQuery.Or("type = ?", custom.FACING)
		}

		err = locationChargeQuery.Find(&locationCharges).Error
		if err != nil {
			return err
		}

		// other charges
		var otherCharges []models.OtherCharge
		err = tx.
			Where("org_id = ? and society_id = ? and disable = false and optional = false", orgId, society).
			Find(&otherCharges).Error
		if err != nil {
			return err
		}

		// optional charges
		var optionalCharges []models.OtherCharge
		err = tx.
			Where("org_id = ? and society_id = ? and disable = false and optional = true", orgId, society).
			Where("id in ?", ac.OptionalCharges).
			Find(&optionalCharges).Error
		if err != nil {
			return err
		}

		// price calculation
		var priceBreakdowns []models.PriceBreakdownDetail
		var totalPrice float64

		// basic cost
		basicCostDetail := models.PriceBreakdownDetail{
			Type:      "basic-cost",
			Price:     ac.BasicCost,
			Summary:   "Basic flat cost",
			Total:     superArea * ac.BasicCost,
			SuperArea: superArea,
		}
		totalPrice += basicCostDetail.Total
		priceBreakdowns = append(priceBreakdowns, basicCostDetail)

		// Add location charges
		for _, charge := range locationCharges {
			detail := models.PriceBreakdownDetail{
				Type:      "preference-location",
				Price:     charge.Price,
				Summary:   charge.Summary,
				Total:     superArea * charge.Price,
				SuperArea: superArea,
			}
			totalPrice += detail.Total
			priceBreakdowns = append(priceBreakdowns, detail)
		}

		// Helper to process other/optional charges
		processOtherCharges := func(charges []models.OtherCharge) {
			for _, charge := range charges {
				detail := models.PriceBreakdownDetail{
					Type:      "other",
					Price:     charge.Price,
					Summary:   charge.Summary,
					SuperArea: superArea,
				}

				if charge.Recurring && charge.AdvanceMonths >= 1 {
					if charge.Fixed {
						detail.Total = charge.Price * float64(charge.AdvanceMonths)
					} else {
						detail.Total = superArea * charge.Price * float64(charge.AdvanceMonths)
					}
				} else {
					if charge.Fixed {
						detail.Total = charge.Price
					} else {
						detail.Total = superArea * charge.Price
					}
				}

				totalPrice += detail.Total
				priceBreakdowns = append(priceBreakdowns, detail)
			}
		}

		processOtherCharges(otherCharges)
		processOtherCharges(optionalCharges)

		//for i, pb := range priceBreakdowns {
		//	fmt.Printf("Item %d:\n", i+1)
		//	fmt.Printf("  Type:    %s\n", pb.Type)
		//	fmt.Printf("  Price:   %.2f\n", pb.Price)
		//	fmt.Printf("  Summary: %s\n", pb.Summary)
		//	fmt.Printf("  SuperArea:   %.2f\n", pb.SuperArea)
		//	fmt.Printf("  Total:   %.2f\n", pb.Total)
		//}

		saleModel := models.Sale{
			FlatId:         uuid.MustParse(flatId),
			SocietyId:      society,
			OrgId:          uuid.MustParse(orgId),
			TotalPrice:     totalPrice,
			PriceBreakdown: priceBreakdowns,
		}
		err = tx.Create(&saleModel).Error
		if err != nil {
			return err
		}

		customers := make([]*models.Customer, 0, len(ac.Details))
		for _, d := range ac.Details {
			customer := &models.Customer{
				SaleId:           saleModel.Id,
				Level:            d.Level,
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
