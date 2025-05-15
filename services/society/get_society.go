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

	// count flats
	societyIDs := make([]string, 0, len(societyData))
	for _, s := range societyData {
		societyIDs = append(societyIDs, s.ReraNumber)
	}

	type SocietyFlatStats struct {
		SocietyID  string `gorm:"column:society_id"`
		TotalFlats int64  `gorm:"column:total_flats"`
		SoldFlats  int64  `gorm:"column:sold_flats"`
	}
	var stats []SocietyFlatStats
	err := db.Raw(`
		WITH total_flats_cte AS (
			SELECT t.society_id, COUNT(f.id) AS total_flats
			FROM towers t
			JOIN flats f ON f.tower_id = t.id
			WHERE t.society_id IN ?
			GROUP BY t.society_id
		),
		sold_flats_cte AS (
			SELECT t.society_id, COUNT(s.id) AS sold_flats
			FROM sales s
			JOIN flats f ON f.id = s.flat_id
			JOIN towers t ON t.id = f.tower_id
			WHERE t.society_id IN ?
			GROUP BY t.society_id
		)
		SELECT
			tf.society_id,
			tf.total_flats,
			COALESCE(sf.sold_flats, 0) AS sold_flats
		FROM total_flats_cte tf
		LEFT JOIN sold_flats_cte sf ON tf.society_id = sf.society_id
	`, societyIDs, societyIDs).Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	statMap := make(map[string]SocietyFlatStats)
	for _, s := range stats {
		statMap[s.SocietyID] = s
	}

	for i, s := range societyData {
		if stat, ok := statMap[s.ReraNumber]; ok {
			societyData[i].TotalFlats = stat.TotalFlats
			societyData[i].SoldFlats = stat.SoldFlats
			societyData[i].UnsoldFlats = stat.TotalFlats - stat.SoldFlats
		}
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
