package organization

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

// hCreateOrganization is createOrganization handler
type hCreateOrganization struct {
	Name  string `validate:"required,min=3"`
	Email string `validate:"required,email"`
}

func (h *hCreateOrganization) execute(db *gorm.DB, cognito *cognitoidentityprovider.Client, userPool string) (*models.Organization, error) {
	organization := models.Organization{
		Status: custom.ACTIVE,
		Name:   h.Name,
	}

	// perform db transaction for atomicity
	err := db.Transaction(func(tx *gorm.DB) error {
		// check if user exists or not
		if userExists(cognito, h.Email, userPool) {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "User already exists.",
			}
		}

		// create organization
		result := tx.Create(&organization)
		if result.Error != nil {
			return result.Error
		}

		// create user
		result = tx.Create(&models.User{
			Name:  h.Email,
			Email: h.Email,
			OrgId: organization.Id,
			Role:  custom.ORGADMIN,
		})
		if result.Error != nil {
			return result.Error
		}

		// create user credentials
		err := createUserInCognito(cognito, h.Email, organization.Id.String(), userPool)
		if err != nil {
			return err
		}

		// add user to a group
		err = addUserToGroup(cognito, h.Email, string(custom.ORGADMIN), userPool)
		if err != nil {
			// clean up from cognito
			go func() {
				err := deleteUserFromCognito(cognito, h.Email, userPool)
				if err != nil {
					return
				}
			}()
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &organization, nil
}

func (s *organizationService) createOrganization(w http.ResponseWriter, r *http.Request) {
	reqBody := payload.ValidateAndDecodeRequest[hCreateOrganization](w, r)
	if reqBody == nil {
		return
	}

	org, err := reqBody.execute(s.db, s.cognito, s.userPool)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new organization."
	response.Data = org

	payload.EncodeJSON(w, http.StatusCreated, response)
}

// hAddUserToOrganization is addUserToOrganization handler
type hAddUserToOrganization struct {
	Email string `validate:"required,email"`
}

func (h *hAddUserToOrganization) execute(db *gorm.DB, cognito *cognitoidentityprovider.Client, orgId, userPool string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// check if user exists or not
		if userExists(cognito, h.Email, userPool) {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "User already exists.",
			}
		}

		// create user
		result := db.Create(&models.User{
			OrgId: uuid.MustParse(orgId),
			Name:  h.Email,
			Email: h.Email,
			Role:  custom.ORGUSER,
		})
		if result.Error != nil {
			return result.Error
		}

		// create user credentials
		err := createUserInCognito(cognito, h.Email, orgId, userPool)
		if err != nil {
			return err
		}

		// add user to a group
		err = addUserToGroup(cognito, h.Email, string(custom.ORGUSER), userPool)
		if err != nil {
			// clean up from cognito
			go func() {
				err := deleteUserFromCognito(cognito, h.Email, userPool)
				if err != nil {
					return
				}
			}()
			return err
		}

		return nil
	})
}

func (s *organizationService) addUserToOrganization(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	reqBody := payload.ValidateAndDecodeRequest[hAddUserToOrganization](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, s.cognito, orgId, s.userPool)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully added user to organization."

	payload.EncodeJSON(w, http.StatusCreated, response)
}
