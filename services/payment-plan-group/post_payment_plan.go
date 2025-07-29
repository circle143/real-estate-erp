package payment_plan_group

import (
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type hCreatePaymentPlan struct {
	Name   string             `validate:"required"`
	Abbr   string             `validate:"required"`
	Ratios []paymentPlanRatio `validate:"required,dive"`
}

func (h *hCreatePaymentPlan) validate() error {
	for _, ratio := range h.Ratios {
		val := decimal.Zero
		for _, item := range ratio.Items {
			val = val.Add(decimal.NewFromFloat(float64(item.Ratio)))

			scope := custom.PaymentPlanItemScope(item.Scope)
			if !scope.IsValid() {
				return &custom.RequestError{
					Status:  http.StatusBadRequest,
					Message: "Invalid scope value for payment plan ratio item.",
				}
			}

			conditionType := custom.PaymentPlanCondition(item.ConditionType)
			if !conditionType.IsValid() {
				return &custom.RequestError{
					Status:  http.StatusBadRequest,
					Message: "Invalid condition-type value for payment plan item.",
				}
			}

			if !slices.Contains(custom.ValidPaymentPlanScopeCondtion[scope], conditionType) {
				return &custom.RequestError{
					Status:  http.StatusBadRequest,
					Message: "Invalid condition-type value for payment plan scope.",
				}

			}

			if conditionType == custom.WITHINDAYS && item.ConditionValue <= 0 {
				return &custom.RequestError{
					Status:  http.StatusBadRequest,
					Message: "Invalid condition value for payment plan item.",
				}
			}

		}

		if !val.Equal(decimal.NewFromInt(100)) {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("Total ratio is not 100. Required ratio: 100. Got ratio: %s", val),
			}
		}
	}
	return nil
}

func (h *hCreatePaymentPlan) execute(db *gorm.DB, orgId, society string) (*models.PaymentPlanGroup, error) {
	err := h.validate()
	if err != nil {
		return nil, err
	}

	// paymentPlanModel := models.PaymentPlan{
	// 	SocietyId:      society,
	// 	OrgId:          uuid.MustParse(orgId),
	// 	Scope:          custom.PaymentPlanScope(h.Scope),
	// 	ConditionType:  custom.PaymentPlanCondition(h.ConditionType),
	// 	ConditionValue: h.ConditionValue,
	// 	Amount:         h.Amount,
	// 	Summary:        h.Summary,
	// }
	//
	// err = db.Create(&paymentPlanModel).Error

	// create payment plan group

	group := models.PaymentPlanGroup{
		Name:      h.Name,
		Abbr:      h.Abbr,
		OrgId:     uuid.MustParse(orgId),
		SocietyId: society,
		Ratios:    make([]models.PaymentPlanRatio, len(h.Ratios)),
	}

	for i, r := range h.Ratios {
		var ratioStrings []string
		items := make([]models.PaymentPlanRatioItem, len(r.Items))

		for j, item := range r.Items {
			ratioStr := fmt.Sprintf("%.2f", item.Ratio)
			ratioStrings = append(ratioStrings, ratioStr)

			items[j] = models.PaymentPlanRatioItem{
				Ratio:          fmt.Sprintf("%.2f", item.Ratio),
				Scope:          custom.PaymentPlanItemScope(item.Scope),
				ConditionType:  custom.PaymentPlanCondition(item.ConditionType),
				ConditionValue: item.ConditionValue,
			}
		}

		group.Ratios[i] = models.PaymentPlanRatio{
			Ratio:  strings.Join(ratioStrings, ","),
			Ratios: items,
		}
	}

	if err := db.Create(&group).Error; err != nil {
		log.Printf("create group failed: %v", err)
		return nil, err
	}

	return &group, nil
}

func (s *paymentPlanService) createPaymentPlan(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hCreatePaymentPlan](w, r)
	if reqBody == nil {
		return
	}

	paymentPlan, err := reqBody.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new payment plan."
	response.Data = paymentPlan

	payload.EncodeJSON(w, http.StatusCreated, response)

}

func (s *paymentPlanService) addPaymentPlanRatio(w http.ResponseWriter, r *http.Request) {}

func (s *paymentPlanService) markPaymentPlanItemActiveForTower(w http.ResponseWriter, r *http.Request) {

}

func (s *paymentPlanService) markPaymentPlanItemActiveForFlat(w http.ResponseWriter, r *http.Request) {

}
