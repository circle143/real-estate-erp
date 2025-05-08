package payment_plan

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type hGetSocietyPaymentPlans struct{}

func (h *hGetSocietyPaymentPlans) execute(db *gorm.DB, orgId, society, cursor string) (*custom.PaginatedData, error) {
	var paymentPlans []models.PaymentPlan

	query := db.Where("org_id = ? and society_id = ?", orgId, society).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	result := query.Find(&paymentPlans)
	if result.Error != nil {
		return nil, result.Error
	}

	return common.CreatePaginatedResponse(&paymentPlans), nil
}

func (s *paymentPlanService) getSocietyPaymentPlans(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")

	paymentPlans := hGetSocietyPaymentPlans{}
	res, err := paymentPlans.execute(s.db, orgId, societyRera, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hGetTowerPaymentPlans struct{}

func (h *hGetTowerPaymentPlans) execute(db *gorm.DB, orgId, society, towerId, cursor string) (*custom.PaginatedData, error) {
	var paymentPlans []models.PaymentPlan

	query := db.Where("org_id = ? and society_id = ?", orgId, society).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	result := query.Find(&paymentPlans)
	if result.Error != nil {
		return nil, result.Error
	}

	// get payment plan ids
	paymentPlanIDs := make([]uuid.UUID, 0, len(paymentPlans))
	for _, plan := range paymentPlans {
		paymentPlanIDs = append(paymentPlanIDs, plan.Id)
	}

	// get active status
	var statuses []models.TowerPaymentStatus
	err := db.Preload("PaymentPlan").
		Where("tower_id = ? AND payment_id IN ?", towerId, paymentPlanIDs).
		Find(&statuses).Error
	if err != nil {
		return nil, err
	}

	// all active payment plans for this tower
	activePaymentMap := make(map[uuid.UUID]bool)
	for _, status := range statuses {
		activePaymentMap[status.PaymentId] = true
	}

	// Mark `Active: true` on matching PaymentPlans
	for i := range paymentPlans {
		val := false
		if activePaymentMap[paymentPlans[i].Id] {
			val = true
		}
		paymentPlans[i].Active = &val
	}

	return common.CreatePaginatedResponse(&paymentPlans), nil
}

func (s *paymentPlanService) getTowerPaymentPlans(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")
	towerId := chi.URLParam(r, "towerId")

	paymentPlans := hGetTowerPaymentPlans{}
	res, err := paymentPlans.execute(s.db, orgId, societyRera, towerId, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
