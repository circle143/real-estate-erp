package organization

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"net/http"
)

type hRemoveUserFromOrganization struct{}

func (h *hRemoveUserFromOrganization) execute(db *gorm.DB, cognito *cognitoidentityprovider.Client, user, orgId, userPool string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		userModel := models.User{
			Email: user,
		}
		result := tx.Where("org_id = ?", orgId).Delete(&userModel)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "user not found",
			}
		}
		return deleteUserFromCognito(cognito, user, userPool)
	})
}

func (s *organizationService) removeUserFromOrganization(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	user := chi.URLParam(r, "userEmail")

	organization := hRemoveUserFromOrganization{}
	err := organization.execute(s.db, s.cognito, user, orgId, s.userPool)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully removed user from organization."

	payload.EncodeJSON(w, http.StatusOK, response)
}
