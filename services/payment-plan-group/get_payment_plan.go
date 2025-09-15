package payment_plan_group

import (
	"net/http"
	"strings"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type hGetPaymentPlans struct{}

func (h *hGetPaymentPlans) execute(db *gorm.DB, orgId, society, cursor string) (*custom.PaginatedData, error) {
	var paymentPlans []models.PaymentPlanGroup

	//	query := db.Where("org_id = ? and society_id = ?", orgId, society).
	//
	// Preload("Ratios").
	//
	//	Preload("Ratios.Ratios").
	//	.Order("created_at DESC").Limit(custom.LIMIT + 1)
	query := db.
		Preload("Ratios").
		Preload("Ratios.Ratios"). // preload nested PaymentPlanRatioItem
		Where("org_id = ? AND society_id = ?", orgId, society).
		Order("created_at DESC").
		Limit(custom.LIMIT + 1)

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

func (s *paymentPlanService) getPaymentPlan(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")

	paymentPlans := hGetPaymentPlans{}
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
	var paymentPlans []models.PaymentPlanGroup

	// Load groups with ratios and only tower-scoped items
	query := db.Preload("Ratios.Ratios", "scope = ?", custom.SCOPE_TOWER).
		Where("org_id = ? AND society_id = ?", orgId, society).
		Order("created_at DESC").
		Limit(custom.LIMIT + 1)

	if strings.TrimSpace(cursor) != "" {
		if decodedCursor, err := common.DecodeCursor(cursor); err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	if err := query.Find(&paymentPlans).Error; err != nil {
		return nil, err
	}

	// Collect all item IDs (tower-scoped only, thanks to preload condition)
	var itemIDs []uuid.UUID
	for _, group := range paymentPlans {
		for _, ratio := range group.Ratios {
			for _, item := range ratio.Ratios {
				itemIDs = append(itemIDs, item.Id)
			}
		}
	}

	if len(itemIDs) == 0 {
		return common.CreatePaginatedResponse(&paymentPlans), nil
	}

	// Fetch tower statuses
	var statuses []models.TowerPaymentStatus
	if err := db.
		Where("tower_id = ? AND payment_id IN ?", towerId, itemIDs).
		Find(&statuses).Error; err != nil {
		return nil, err
	}

	// Map active items
	activeMap := make(map[uuid.UUID]bool)
	for _, st := range statuses {
		activeMap[st.PaymentId] = true
	}

	// Mark active items
	for gi := range paymentPlans {
		for ri := range paymentPlans[gi].Ratios {
			for ii := range paymentPlans[gi].Ratios[ri].Ratios {
				item := &paymentPlans[gi].Ratios[ri].Ratios[ii]
				val := activeMap[item.Id]
				item.Active = &val
			}
		}
	}

	return common.CreatePaginatedResponse(&paymentPlans), nil
}

func (s *paymentPlanService) getTowerPaymentPlan(w http.ResponseWriter, r *http.Request) {
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

type hGetFlatPaymentPlans struct{}

func (h *hGetFlatPaymentPlans) execute(db *gorm.DB, orgId, society, flatID, cursor string) (*custom.PaginatedData, error) {
	var paymentPlans []models.PaymentPlanGroup

	// Load groups with ratios and only flat-scoped items
	query := db.Preload("Ratios.Ratios", "scope = ?", custom.SCOPE_FLAT).
		Where("org_id = ? AND society_id = ?", orgId, society).
		Order("created_at DESC").
		Limit(custom.LIMIT + 1)

	if strings.TrimSpace(cursor) != "" {
		if decodedCursor, err := common.DecodeCursor(cursor); err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	if err := query.Find(&paymentPlans).Error; err != nil {
		return nil, err
	}

	// Collect all item IDs (flat-scoped only, thanks to preload condition)
	var itemIDs []uuid.UUID
	for _, group := range paymentPlans {
		for _, ratio := range group.Ratios {
			for _, item := range ratio.Ratios {
				itemIDs = append(itemIDs, item.Id)
			}
		}
	}

	if len(itemIDs) == 0 {
		return common.CreatePaginatedResponse(&paymentPlans), nil
	}

	// Fetch flat statuses
	var statuses []models.FlatPaymentStatus
	if err := db.
		Where("flat_id = ? AND payment_id IN ?", flatID, itemIDs).
		Find(&statuses).Error; err != nil {
		return nil, err
	}

	// Map active items
	activeMap := make(map[uuid.UUID]bool)
	for _, st := range statuses {
		activeMap[st.PaymentId] = true
	}

	// Mark active items
	for gi := range paymentPlans {
		for ri := range paymentPlans[gi].Ratios {
			for ii := range paymentPlans[gi].Ratios[ri].Ratios {
				item := &paymentPlans[gi].Ratios[ri].Ratios[ii]
				val := activeMap[item.Id]
				item.Active = &val
			}
		}
	}

	return common.CreatePaginatedResponse(&paymentPlans), nil
}

func (s *paymentPlanService) getFlatPaymentPlan(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")
	flatId := chi.URLParam(r, "flatId")

	paymentPlans := hGetFlatPaymentPlans{}
	res, err := paymentPlans.execute(s.db, orgId, societyRera, flatId, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)

}
