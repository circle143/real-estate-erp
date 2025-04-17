package common

import (
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// package common handles all the methods required by multiple services

// IService is implemented by all services and return routes exposed by the service
type IService interface {
	GetRoutes() *chi.Mux
}

type IAWSConfig interface {
	GetCognitoClient() *cognitoidentityprovider.Client
}

// IApp is an application interface with all the configurations
type IApp interface {
	GetRouter() *chi.Mux

	GetDBClient() *gorm.DB
	GetAWSConfig() IAWSConfig
}