package organization

import (
	"circledigital.in/real-state-erp/init"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"gorm.io/gorm"
)

type organizationService struct {
	db       *gorm.DB
	cognito  *cognitoidentityprovider.Client
	userPool string
}

// CreateOrganizationService is an abstract factory to create organization service
func CreateOrganizationService(dbConn *gorm.DB, cognitoClient *cognitoidentityprovider.Client, userPoolId string) init.IService {
	return &organizationService{
		db:       dbConn,
		cognito:  cognitoClient,
		userPool: userPoolId,
	}
}

// deleteUserFromCognito deletes user from cognito
func deleteUserFromCognito(cognito *cognitoidentityprovider.Client, username, userPool string) error {
	_, err := cognito.AdminDeleteUser(context.TODO(), &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(userPool),
		Username:   aws.String(username),
	})
	return err
}