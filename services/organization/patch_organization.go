package organization

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

// hUpdateOrganizationStatus is updateOrganizationStatus handler
type hUpdateOrganizationStatus struct {
	Status string `validate:"required"`
}

func (h *hUpdateOrganizationStatus) validate() error {
	status := custom.OrganizationStatus(h.Status)
	if !status.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid status for organization",
		}
	}

	return nil
}

func (h *hUpdateOrganizationStatus) execute(db *gorm.DB, orgId string) error {
	err := h.validate()
	if err != nil {
		return err
	}
	org := models.Organization{
		Id: uuid.MustParse(orgId),
	}
	return db.Model(&org).Update("status", custom.OrganizationStatus(h.Status)).Error
}

func (s *organizationService) updateOrganizationStatus(w http.ResponseWriter, r *http.Request) {
	orgId := chi.URLParam(r, "orgId")
	reqBody := payload.ValidateAndDecodeRequest[hUpdateOrganizationStatus](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated organization status."

	payload.EncodeJSON(w, http.StatusOK, response)
}

// hUpdateOrganizationDetails is updateOrganizationDetails handler
type hUpdateOrganizationDetails struct {
	Name string
	Logo string
	GST  string `validate:"omitempty,gst"`
}

func (h *hUpdateOrganizationDetails) validate() error {
	// check if at least one of the value is present or not
	if strings.TrimSpace(h.Name) == "" && strings.TrimSpace(h.Logo) == "" && strings.TrimSpace(h.GST) == "" {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid field values to update.",
		}
	}

	return nil
}

func (h *hUpdateOrganizationDetails) execute(db *gorm.DB, orgId string) error {
	err := h.validate()
	if err != nil {
		return err
	}

	org := models.Organization{
		Id: uuid.MustParse(orgId),
	}

	return db.Model(&org).Updates(models.Organization{
		Name: h.Name,
		Logo: h.Logo,
		Gst:  h.GST,
	}).Error

}

func (s *organizationService) updateOrganizationDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	reqBody := payload.ValidateAndDecodeRequest[hUpdateOrganizationDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated organization details."

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hUpdateOrganizationUserRole struct {
	Role string
}

func (h *hUpdateOrganizationUserRole) validate() error {
	role := custom.UserRole(h.Role)
	if !role.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid user role.",
		}
	}

	return nil
}

func (h *hUpdateOrganizationUserRole) execute(db *gorm.DB, cognito *cognitoidentityprovider.Client, user, orgId, userPool string) error {
	err := h.validate()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		userModel := models.User{
			Email: user,
		}
		// update user
		result := tx.Model(&userModel).Where("org_id = ?", orgId).Update("role", custom.UserRole(h.Role))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "user not found",
			}
		}

		// update cognito
		err := removeUserFromGroup(cognito, user, userPool)
		if err != nil {
			return err
		}

		// add user to a new group
		err = addUserToGroup(cognito, user, h.Role, userPool)
		return err
	})
}

func (s *organizationService) updateOrganizationUserRole(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	user := chi.URLParam(r, "userEmail")
	reqBody := payload.ValidateAndDecodeRequest[hUpdateOrganizationUserRole](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, s.cognito, user, orgId, s.userPool)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated user role."

	payload.EncodeJSON(w, http.StatusOK, response)
}
