package payment_plan_group

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/flat"
	"circledigital.in/real-state-erp/services/tower"
	"circledigital.in/real-state-erp/utils/common"
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
				Description:    item.Description,
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

type hMarkPaymentPlanActiveForTower struct{}

func (h *hMarkPaymentPlanActiveForTower) validate(db *gorm.DB, orgId, society, paymentId, towerId string) error {
	paymentUUID := uuid.MustParse(paymentId)
	towerUUID := uuid.MustParse(towerId)

	// validate payment permission (society match)
	// paymentSocietyInfoService := CreatePaymentPlanSocietyInfoService(db, paymentUUID)
	// if err := common.IsSameSociety(paymentSocietyInfoService, orgId, society); err != nil {
	// 	return err
	// }

	// validate tower permission (society match)
	towerSocietyInfoService := tower.CreateTowerSocietyInfoService(db, towerUUID)
	if err := common.IsSameSociety(towerSocietyInfoService, orgId, society); err != nil {
		return err
	}

	// validate that this payment item is scoped for TOWER
	var item models.PaymentPlanRatioItem
	if err := db.First(&item, "id = ?", paymentUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &custom.RequestError{
				Status:  http.StatusNotFound,
				Message: "Invalid payment plan item.",
			}
		}
		return err
	}

	if item.Scope != custom.SCOPE_TOWER {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Selected payment plan item can't be associated with a tower.",
		}
	}

	return nil
}

func (h *hMarkPaymentPlanActiveForTower) execute(db *gorm.DB, orgId, society, paymentId, towerId string) error {
	err := h.validate(db, orgId, society, paymentId, towerId)
	if err != nil {
		return err
	}

	// insert TowerPaymentStatus (idempotent)
	status := models.TowerPaymentStatus{
		TowerId:   uuid.MustParse(towerId),
		PaymentId: uuid.MustParse(paymentId),
	}
	if err := db.FirstOrCreate(&status, status).Error; err != nil {
		return err
	}
	return nil
}

func (s *paymentPlanService) markPaymentPlanItemActiveForTower(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	paymentId := chi.URLParam(r, "paymentPlanItemId")
	towerId := chi.URLParam(r, "towerId")

	towerPayment := hMarkPaymentPlanActiveForTower{}
	err := towerPayment.execute(s.db, orgId, societyRera, paymentId, towerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Payment plan is now active."

	payload.EncodeJSON(w, http.StatusCreated, response)

}

type hMarkPaymentPlanActiveForFlat struct{}

func (h *hMarkPaymentPlanActiveForFlat) validate(db *gorm.DB, orgId, society, paymentId, flatId string) error {
	paymentUUID := uuid.MustParse(paymentId)
	flatUUID := uuid.MustParse(flatId)

	// validate payment permission (society match)
	// paymentSocietyInfoService := CreatePaymentPlanSocietyInfoService(db, paymentUUID)
	// if err := common.IsSameSociety(paymentSocietyInfoService, orgId, society); err != nil {
	// 	return err
	// }

	// validate flat permission (society match)
	flatSocietyInfoService := flat.CreateFlatSocietyInfoService(db, flatUUID)
	if err := common.IsSameSociety(flatSocietyInfoService, orgId, society); err != nil {
		return err
	}

	// validate that this payment item is scoped for FLAT
	var item models.PaymentPlanRatioItem
	if err := db.First(&item, "id = ?", paymentUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &custom.RequestError{
				Status:  http.StatusNotFound,
				Message: "Invalid payment plan item.",
			}
		}
		return err
	}

	if item.Scope != custom.SCOPE_FLAT {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Selected payment plan item can't be associated with a flat.",
		}
	}

	return nil
}

func (h *hMarkPaymentPlanActiveForFlat) execute(db *gorm.DB, orgId, society, paymentId, flatId string) error {
	if err := h.validate(db, orgId, society, paymentId, flatId); err != nil {
		return err
	}

	// insert FlatPaymentStatus (idempotent)
	status := models.FlatPaymentStatus{
		FlatId:    uuid.MustParse(flatId),
		PaymentId: uuid.MustParse(paymentId),
	}
	if err := db.FirstOrCreate(&status, status).Error; err != nil {
		return err
	}

	return nil
}

func (s *paymentPlanService) markPaymentPlanItemActiveForFlat(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	paymentId := chi.URLParam(r, "paymentPlanItemId")
	flatId := chi.URLParam(r, "flatId")

	towerPayment := hMarkPaymentPlanActiveForFlat{}
	err := towerPayment.execute(s.db, orgId, societyRera, paymentId, flatId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Payment plan is now active."

	payload.EncodeJSON(w, http.StatusCreated, response)

}
