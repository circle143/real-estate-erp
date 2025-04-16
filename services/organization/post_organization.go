package organization

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"gorm.io/gorm"
	"net/http"
)

// hCreateOrganization is createOrganization handler
type hCreateOrganization struct {
	Name  string `validate:"required,min=3"`
	Email string `validate:"required,email"`
}

func (co *hCreateOrganization) execute(db *gorm.DB, cognito *cognitoidentityprovider.Client, userPool string) (*models.Organization, error) {
	organization := models.Organization{
		Status: custom.ACTIVE,
		Name:   co.Name,
	}

	// perform db transaction for atomicity
	err := db.Transaction(func(tx *gorm.DB) error {
		// check if user exists or not
		_, err := cognito.AdminGetUser(context.TODO(), &cognitoidentityprovider.AdminGetUserInput{
			UserPoolId: aws.String(userPool),
			Username:   aws.String(co.Email),
		})
		if err == nil {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "User already part of an organization.",
			}
		}

		// create organization
		result := tx.Create(&organization)
		if result.Error != nil {
			return result.Error
		}

		// create user
		result = tx.Create(&models.User{
			Name:  co.Email,
			Email: co.Email,
			OrgId: organization.Id,
			Role:  custom.ORGADMIN,
		})
		if result.Error != nil {
			return result.Error
		}

		// create user credentials
		_, err = cognito.AdminCreateUser(context.TODO(), &cognitoidentityprovider.AdminCreateUserInput{
			UserPoolId: aws.String(userPool),
			Username:   aws.String(co.Email),
			UserAttributes: []types.AttributeType{
				{
					Name:  aws.String(custom.OrgIdCustomAttribute),
					Value: aws.String(organization.Id.String()),
				},
			},
		})
		if err != nil {
			return err
		}

		// add user to a group
		_, err = cognito.AdminAddUserToGroup(context.TODO(), &cognitoidentityprovider.AdminAddUserToGroupInput{
			UserPoolId: aws.String(userPool),
			GroupName:  aws.String(string(custom.ORGADMIN)),
			Username:   aws.String(co.Email),
		})
		if err != nil {
			// clean up from cognito
			go func() {
				err := deleteUserFromCognito(cognito, co.Email, userPool)
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

func (os *organizationService) createOrganization(w http.ResponseWriter, r *http.Request) {
	reqBody := payload.ValidateAndDecodeRequest[hCreateOrganization](w, r)
	if reqBody == nil {
		return
	}

	org, err := reqBody.execute(os.db, os.cognito, os.userPool)
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

func (os *organizationService) addUserToOrganization(w http.ResponseWriter, r *http.Request) {

}