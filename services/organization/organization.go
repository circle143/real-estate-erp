package organization

import (
	"circledigital.in/real-state-erp/init"
	"circledigital.in/real-state-erp/utils/custom"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"gorm.io/gorm"
	"os"
)

type organizationService struct {
	db       *gorm.DB
	cognito  *cognitoidentityprovider.Client
	userPool string
}

// CreateOrganizationService is an abstract factory to create organization service
func CreateOrganizationService(dbConn *gorm.DB, cognitoClient *cognitoidentityprovider.Client) init.IService {
	userPoolId := os.Getenv("USER_POOL_ID")
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

// userExists check if the given user already exists in cognito user pool or not
func userExists(cognito *cognitoidentityprovider.Client, username, userPool string) bool {
	_, err := cognito.AdminGetUser(context.TODO(), &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(userPool),
		Username:   aws.String(username),
	})
	return err == nil
}

// createUserInCognito creates new user in cognitoUserPool
func createUserInCognito(cognito *cognitoidentityprovider.Client, username, orgId, userPool string) error {
	_, err := cognito.AdminCreateUser(context.TODO(), &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId: aws.String(userPool),
		Username:   aws.String(username),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String(custom.OrgIdCustomAttribute),
				Value: aws.String(orgId),
			},
		},
	})
	return err
}

// addUserToGroup adds the given user to the provided group
func addUserToGroup(cognito *cognitoidentityprovider.Client, username, group, userPool string) error {
	_, err := cognito.AdminAddUserToGroup(context.TODO(), &cognitoidentityprovider.AdminAddUserToGroupInput{
		UserPoolId: aws.String(userPool),
		GroupName:  aws.String(group),
		Username:   aws.String(username),
	})
	return err
}

// removeUserFromGroup removes user from his existing group
func removeUserFromGroup(cognito *cognitoidentityprovider.Client, username, userPool string) error {
	listOut, err := cognito.AdminListGroupsForUser(context.TODO(), &cognitoidentityprovider.AdminListGroupsForUserInput{
		UserPoolId: aws.String(userPool),
		Username:   aws.String(username),
	})
	if err != nil {
		return err
	}

	// 3. Remove user from all current groups
	for _, group := range listOut.Groups {
		_, err := cognito.AdminRemoveUserFromGroup(context.TODO(), &cognitoidentityprovider.AdminRemoveUserFromGroupInput{
			UserPoolId: aws.String(userPool),
			Username:   aws.String(username),
			GroupName:  group.GroupName,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
