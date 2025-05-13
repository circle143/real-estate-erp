package tower

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type hGetAllTowers struct{}

func (gat *hGetAllTowers) execute(db *gorm.DB, orgId, society, cursor string) (*custom.PaginatedData, error) {
	var towerData []models.Tower

	query := db.Where("org_id = ? and society_id = ?", orgId, society).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	result := query.Find(&towerData)
	if result.Error != nil {
		return nil, result.Error
	}

	// create tower slice
	towerIDs := make([]uuid.UUID, 0, len(towerData))
	for _, t := range towerData {
		towerIDs = append(towerIDs, t.Id)
	}

	type TowerFinance struct {
		TowerID     uuid.UUID       `json:"towerId"`
		TotalAmount decimal.Decimal `json:"totalAmount"`
		PaidAmount  decimal.Decimal `json:"paidAmount"`
	}
	var towerFinance []TowerFinance
	financeQuery := `
		WITH total_sales AS (
			SELECT
				f.tower_id,
				SUM(s.total_price) AS total_amount
			FROM sales s
			JOIN flats f ON f.id = s.flat_id
			WHERE f.tower_id IN ?
			GROUP BY f.tower_id
		),
		total_payments AS (
			SELECT
				f.tower_id,
				SUM(p.amount) AS paid_amount
			FROM sales s
			JOIN flats f ON f.id = s.flat_id
			JOIN sale_payment_statuses p ON p.sale_id = s.id
			WHERE f.tower_id IN ?
			GROUP BY f.tower_id
		)
		SELECT
			COALESCE(ts.tower_id, tp.tower_id) AS tower_id,
			COALESCE(ts.total_amount, 0) AS total_amount,
			COALESCE(tp.paid_amount, 0) AS paid_amount
		FROM total_sales ts
		FULL OUTER JOIN total_payments tp ON ts.tower_id = tp.tower_id;
`
	db.Raw(financeQuery, towerIDs, towerIDs).Scan(&towerFinance)

	// merge to tower
	financeMap := make(map[uuid.UUID]TowerFinance)
	for _, f := range towerFinance {
		financeMap[f.TowerID] = f
	}
	for i := range towerData {
		if finance, ok := financeMap[towerData[i].Id]; ok {
			towerData[i].TotalAmount = finance.TotalAmount
			towerData[i].PaidAmount = finance.PaidAmount
			towerData[i].Remaining = finance.TotalAmount.Sub(finance.PaidAmount)
		}
	}

	return common.CreatePaginatedResponse(&towerData), nil
}

func (ts *towerService) getAllTowers(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")

	tower := hGetAllTowers{}
	res, err := tower.execute(ts.db, orgId, societyRera, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
