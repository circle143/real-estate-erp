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
	"regexp"
	"strings"
)

// hUpdateOrganizationStatus is updateOrganizationStatus handler
type hUpdateOrganizationStatus struct {
	Status string `validate:"required"`
}

func (uos *hUpdateOrganizationStatus) validate() error {
	status := custom.OrganizationStatus(uos.Status)
	if status.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid status for organization",
		}
	}

	return nil
}

func (uos *hUpdateOrganizationStatus) execute(db *gorm.DB, orgId string) error {
	err := uos.validate()
	if err != nil {
		return err
	}
	org := models.Organization{
		Id: uuid.MustParse(orgId),
	}
	return db.Model(&org).Update("status", custom.OrganizationStatus(uos.Status)).Error
}

func (os *organizationService) updateOrganizationStatus(w http.ResponseWriter, r *http.Request) {
	orgId := chi.URLParam(r, "orgId")
	reqBody := payload.ValidateAndDecodeRequest[hUpdateOrganizationStatus](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(os.db, orgId)
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
	GST  string
}

func (uod *hUpdateOrganizationDetails) validate() error {
	// check if at least one of the value is present or not
	if strings.TrimSpace(uod.Name) == "" && strings.TrimSpace(uod.Logo) == "" && strings.TrimSpace(uod.GST) == "" {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid field values to update.",
		}
	}

	// check if gst is present and is valid
	if strings.TrimSpace(uod.GST) != "" {
		gstRegex := `^[0-9]{2}[A-Z]{3}[ABCFGHLJPTF]{1}[A-Z]{1}[0-9]{4}[A-Z]{1}[1-9A-Z]{1}Z[0-9A-Z]{1}$`
		valid, _ := regexp.MatchString(gstRegex, uod.GST)

		if !valid {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Invalid GST provided.",
			}
		}

	}
	return nil
}

func (uod *hUpdateOrganizationDetails) execute(db *gorm.DB, orgId string) error {
	err := uod.validate()
	if err != nil {
		return err
	}

	org := models.Organization{
		Id: uuid.MustParse(orgId),
	}

	return db.Model(&org).Updates(models.Organization{
		Name: uod.Name,
		Logo: uod.Logo,
		Gst:  uod.GST,
	}).Error

}

func (os *organizationService) updateOrganizationDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	reqBody := payload.ValidateAndDecodeRequest[hUpdateOrganizationDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(os.db, orgId)
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

func (uou *hUpdateOrganizationUserRole) validate() error {
	role := custom.UserRole(uou.Role)
	if !role.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid user role.",
		}
	}

	return nil
}

func (uou *hUpdateOrganizationUserRole) execute(db *gorm.DB, cognito *cognitoidentityprovider.Client, user, orgId, userPool string) error {
	err := uou.validate()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		userModel := models.User{
			Email: user,
		}
		// update user
		result := tx.Model(&userModel).Where("org_id = ?", orgId).Update("role", custom.UserRole(uou.Role))
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
		err = addUserToGroup(cognito, user, uou.Role, userPool)
		return err
	})
}

func (os *organizationService) updateOrganizationUserRole(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	user := chi.URLParam(r, "userEmail")
	reqBody := payload.ValidateAndDecodeRequest[hUpdateOrganizationUserRole](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(os.db, os.cognito, user, orgId, os.userPool)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated user role."

	payload.EncodeJSON(w, http.StatusOK, response)
}