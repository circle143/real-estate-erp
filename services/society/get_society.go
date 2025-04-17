package society

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type hGetAllSocieties struct{}

func (gas *hGetAllSocieties) execute(db *gorm.DB, orgId, cursor string) (*custom.PaginatedData, error) {
	var societyData []models.Society
	query := db.Where("org_id = ?", orgId).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	result := query.Find(&societyData)
	if result.Error != nil {
		return nil, result.Error
	}

	return common.CreatePaginatedResponse(&societyData), nil
}

func (ss *societyService) getAllSocieties(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")

	society := hGetAllSocieties{}
	res, err := society.execute(ss.db, orgId, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}