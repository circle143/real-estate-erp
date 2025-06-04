package tower

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type hGetAllTowers struct{}

func (h *hGetAllTowers) execute(db *gorm.DB, orgId, society, cursor string) (*custom.PaginatedData, error) {
	var towerData []models.Tower
	var towerFull []models.Tower

	query := db.Where("org_id = ? and society_id = ?", orgId, society).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	err := query.Find(&towerData).Error
	if err != nil {
		return nil, err
	}

	err = query.
		Preload("Flats").
		Preload("Flats.SaleDetail").
		Preload("Flats.SaleDetail.Receipts").
		Preload("Flats.SaleDetail.Receipts.Cleared").
		Find(&towerFull).Error
	if err != nil {
		return nil, err
	}

	// calc tower stats
	for i, tower := range towerFull {
		totalSaleAmount := decimal.Zero
		totalPaidAmount := decimal.Zero

		totalSold := int64(0)
		totalFlats := int64(len(tower.Flats))

		for _, flat := range tower.Flats {
			if flat.SaleDetail != nil {
				totalSold++
				totalSaleAmount = totalSaleAmount.Add(flat.SaleDetail.TotalPrice)

				for _, receipt := range flat.SaleDetail.Receipts {
					if receipt.Cleared != nil {
						totalPaidAmount = totalPaidAmount.Add(receipt.TotalAmount)
					}
				}

			}
		}

		towerData[i].TotalFlats = totalFlats
		towerData[i].SoldFlats = totalSold
		towerData[i].UnsoldFlats = totalFlats - totalSold

		towerData[i].TotalAmount = totalSaleAmount
		towerData[i].PaidAmount = totalPaidAmount
		towerData[i].Remaining = totalSaleAmount.Sub(totalPaidAmount)
	}

	return common.CreatePaginatedResponse(&towerData), nil
}

func (s *towerService) getAllTowers(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")

	tower := hGetAllTowers{}
	res, err := tower.execute(s.db, orgId, societyRera, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
