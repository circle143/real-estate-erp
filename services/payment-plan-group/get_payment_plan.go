package payment_plan_group

import (
	"net/http"
	"strings"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
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

func (s *paymentPlanService) getTowerPaymentPlan(w http.ResponseWriter, r *http.Request) {}

func (s *paymentPlanService) getFlatPaymentPlan(w http.ResponseWriter, r *http.Request) {}
