package organization

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

// hGetAllOrganizations is getAllOrganizations handler
type hGetAllOrganizations struct{}

func (gao *hGetAllOrganizations) execute(db *gorm.DB, cursor string) (*custom.PaginatedData, error) {
	var orgData []models.Organization
	query := db.Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}

	}

	tx := query.Find(&orgData)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return common.CreatePaginatedResponse(&orgData), nil
}

func (os *organizationService) getAllOrganizations(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")

	organizations := hGetAllOrganizations{}
	res, err := organizations.execute(os.db, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hGetAllOrganizationUsers struct{}

func (gou *hGetAllOrganizationUsers) execute(db *gorm.DB, orgId, cursor string) (*custom.PaginatedData, error) {
	var userData []models.User
	query := db.Where("org_id = ?", orgId).Order("created_at DESC").Limit(custom.LIMIT + 1)
	if strings.TrimSpace(cursor) != "" {
		decodedCursor, err := common.DecodeCursor(cursor)
		if err == nil {
			query = query.Where("created_at < ?", decodedCursor)
		}

	}

	tx := query.Find(&userData)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return common.CreatePaginatedResponse(&userData), nil
}

func (os *organizationService) getAllOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	cursor := r.URL.Query().Get("cursor")

	users := hGetAllOrganizationUsers{}
	res, err := users.execute(os.db, orgId, cursor)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hGetCurrentUserOrganization struct{}

func (gou *hGetCurrentUserOrganization) execute(db *gorm.DB, orgId string) (*models.Organization, error) {
	organization := models.Organization{
		Id: uuid.MustParse(orgId),
	}

	err := db.First(&organization).Error
	return &organization, err
}

func (os *organizationService) getCurrentUserOrganization(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)

	org := hGetCurrentUserOrganization{}
	res, err := org.execute(os.db, orgId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Data = res

	payload.EncodeJSON(w, http.StatusOK, response)
}