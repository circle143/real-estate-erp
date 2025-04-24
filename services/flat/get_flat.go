package flat

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type hGetAllSocietyFlats struct{}

func (gsf *hGetAllSocietyFlats) execute(db *gorm.DB, orgId, societyRera, cursor, filter string) (*custom.PaginatedData, error) {
	var flatData []models.Flat
	query := db.
		Joins("JOIN towers ON towers.id = flats.tower_id").
		Where("towers.society_id = ? AND towers.org_id = ?", societyRera, orgId).
		Preload("Owners").
		Order("flats.id").
		Limit(custom.LIMIT + 1)

	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("flats.created_at < ?", decodedCursor)
		}
	}

	if filter == "1" || filter == "2" {
		// 1 -> sold // 2 -> unsold
		if filter == "1" {
			query = query.Where("flats.sold_by = ?", custom.UNSOLD)
		} else {
			query = query.Where("flats.sold_by != ?", custom.UNSOLD)
		}
	}

	result := query.Find(&flatData)
	if result.Error != nil {
		return nil, result.Error
	}

	return common.CreatePaginatedResponse(&flatData), nil
}

func (fs *flatService) getAllSocietyFlats(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	filter := r.URL.Query().Get("filter")
	societyRera := chi.URLParam(r, "society")

	flat := hGetAllSocietyFlats{}
	res, err := flat.execute(fs.db, orgId, societyRera, cursor, filter)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hGetAllTowerFlats struct{}

func (gtf *hGetAllTowerFlats) execute(db *gorm.DB, orgId, societyRera, towerId, cursor, filter string) (*custom.PaginatedData, error) {
	var flatData []models.Flat
	query := db.
		Joins("JOIN towers ON towers.id = flats.tower_id").
		Where("flat.tower_id = ? AND towers.society_id = ? AND towers.org_id = ?", towerId, societyRera, orgId).
		Preload("Owners").
		Order("flats.id").
		Limit(custom.LIMIT + 1)

	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("flats.created_at < ?", decodedCursor)
		}
	}

	if filter == "1" || filter == "2" {
		// 1 -> sold // 2 -> unsold
		if filter == "1" {
			query = query.Where("flats.sold_by = ?", custom.UNSOLD)
		} else {
			query = query.Where("flats.sold_by != ?", custom.UNSOLD)
		}
	}

	result := query.Find(&flatData)
	if result.Error != nil {
		return nil, result.Error
	}

	return common.CreatePaginatedResponse(&flatData), nil
}

func (fs *flatService) getAllTowerFlats(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	filter := r.URL.Query().Get("filter")
	societyRera := chi.URLParam(r, "society")
	towerId := chi.URLParam(r, "tower")

	flat := hGetAllTowerFlats{}
	res, err := flat.execute(fs.db, orgId, societyRera, towerId, cursor, filter)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
