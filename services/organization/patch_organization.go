package organization

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
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
	status := uos.Status
	if status == string(custom.ACTIVE) || status == string(custom.INACTIVE) || status == string(custom.ARCHIVE) {
		return nil
	}

	return &custom.RequestError{
		Status:  http.StatusBadRequest,
		Message: "Invalid status for organization",
	}
}

func (uos *hUpdateOrganizationStatus) execute(db *gorm.DB, orgId string) error {
	err := uos.validate()
	if err != nil {
		return err
	}
	org := models.Organization{
		Id: uuid.MustParse(orgId),
	}
	return db.Model(&org).Update("status", uos.Status).Error
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

func (os *organizationService) updateOrganizationUserRole(w http.ResponseWriter, r *http.Request) {

}