package payment_plan

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/tower"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hCreatePaymentPlan struct {
	Summary        string `validate:"required"`
	Scope          string `validate:"required"`
	ConditionType  string `validate:"required"`
	Amount         int    `validate:"required,gt=0,lte=100"`
	ConditionValue int
}

func (h *hCreatePaymentPlan) validate(db *gorm.DB, orgId, society string) error {
	scope := custom.PaymentPlanScope(h.Scope)
	if !scope.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid scope value for payment plan.",
		}
	}

	conditionType := custom.PaymentPlanCondition(h.ConditionType)
	if !conditionType.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid condition-type value for payment plan.",
		}
	}

	if conditionType == custom.AFTERDAYS && h.ConditionValue <= 0 {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid condition value for payment plan.",
		}
	}

	// check amount total
	var total int
	err := db.Model(&models.PaymentPlan{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("society_id = ? AND org_id = ?", society, orgId).
		Scan(&total).Error
	if err != nil {
		return err
	}

	if total+h.Amount > 100 {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Amount exceeds 100%",
		}
	}

	return nil
}

func (h *hCreatePaymentPlan) execute(db *gorm.DB, orgId, society string) (*models.PaymentPlan, error) {
	err := h.validate(db, orgId, society)
	if err != nil {
		return nil, err
	}

	paymentPlanModel := models.PaymentPlan{
		SocietyId:      society,
		OrgId:          uuid.MustParse(orgId),
		Scope:          custom.PaymentPlanScope(h.Scope),
		ConditionType:  custom.PaymentPlanCondition(h.ConditionType),
		ConditionValue: h.ConditionValue,
		Amount:         h.Amount,
		Summary:        h.Summary,
	}

	err = db.Create(&paymentPlanModel).Error
	return &paymentPlanModel, err
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

type hMarkPaymentPlanActiveForTower struct{}

func (h *hMarkPaymentPlanActiveForTower) validate(db *gorm.DB, orgId, society, paymentId, towerId string) error {
	paymentUUID := uuid.MustParse(paymentId)
	towerUUID := uuid.MustParse(towerId)

	// validate payment permission
	paymentSocietyInfoService := CreatePaymentPlanSocietyInfoService(db, paymentUUID)
	err := common.IsSameSociety(paymentSocietyInfoService, orgId, society)
	if err != nil {
		return err
	}

	// validate tower permission
	towerSocietyInfoService := tower.CreateTowerSocietyInfoService(db, towerUUID)
	return common.IsSameSociety(towerSocietyInfoService, orgId, society)
}

func (h *hMarkPaymentPlanActiveForTower) execute(db *gorm.DB, orgId, society, paymentId, towerId string) error {
	err := h.validate(db, orgId, society, paymentId, towerId)
	if err != nil {
		return err
	}

	towerPaymentModel := models.TowerPaymentStatus{
		TowerId:   uuid.MustParse(towerId),
		PaymentId: uuid.MustParse(paymentId),
	}
	return db.Create(&towerPaymentModel).Error
}

func (s *paymentPlanService) markPaymentPlanActiveForTower(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	paymentId := chi.URLParam(r, "paymentId")
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
