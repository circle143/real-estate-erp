package charges

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

type hGetAllPreferenceLocationCharges struct{}

func (h *hGetAllPreferenceLocationCharges) execute(db *gorm.DB, orgId, society, cursor string) (*custom.PaginatedData, error) {
	var charges []models.PreferenceLocationCharge
	query := db.Where("org_id = ? and society_id = ?", orgId, society).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	result := query.Find(&charges)
	if result.Error != nil {
		return nil, result.Error
	}
	return common.CreatePaginatedResponse(&charges), nil
}

func (s *chargesService) getAllPreferenceLocationCharges(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")

	tower := hGetAllPreferenceLocationCharges{}
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

type hGetAllOtherCharges struct{}

func (h *hGetAllOtherCharges) execute(db *gorm.DB, orgId, society, cursor string) (*custom.PaginatedData, error) {
	var charges []models.OtherCharge
	query := db.Where("org_id = ? and society_id = ?", orgId, society).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}
	}

	result := query.Find(&charges)
	if result.Error != nil {
		return nil, result.Error
	}
	return common.CreatePaginatedResponse(&charges), nil
}

func (s *chargesService) getAllOtherCharges(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")
	societyRera := chi.URLParam(r, "society")

	tower := hGetAllOtherCharges{}
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

type hGetAllOtherOptionalCharges struct{}

func (h *hGetAllOtherOptionalCharges) execute(db *gorm.DB, orgId, society string) (*[]models.OtherCharge, error) {
	var charges []models.OtherCharge
	query := db.Where("org_id = ? and society_id = ? and optional = true", orgId, society).Order("created_at DESC")

	result := query.Find(&charges)
	if result.Error != nil {
		return nil, result.Error
	}
	return &charges, nil
}

func (s *chargesService) getAllOtherOptionalCharges(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	tower := hGetAllOtherOptionalCharges{}
	res, err := tower.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}
