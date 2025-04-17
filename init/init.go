package init

import (
	"github.com/MicahParks/keyfunc/v3"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// package init handles all the application initialization

// IService is implemented by all services and return routes exposed by the service
type IService interface {
	GetRoutes() *chi.Mux
}

// IApp is an application interface with all the configurations
type IApp interface {
	GetRouter() *chi.Mux

	GetDBClient() *gorm.DB
	GetAWSConfig() *awsConfig
}

// app is implementation of App interface
type app struct {
	mux      *chi.Mux
	dbClient *gorm.DB
	aws      *awsConfig
	jwtKey   keyfunc.Keyfunc
}

func (a *app) GetRouter() *chi.Mux {
	return a.mux
}

func (a *app) GetDBClient() *gorm.DB {
	return a.dbClient
}

func (a *app) GetAWSConfig() *awsConfig {
	return a.aws
}

// initApplication configures all the objects required for startup of application
func (a *app) initApplication() {
	a.dbClient = a.createDBClient()
	a.aws = a.createAWSConfig()
	a.jwtKey = a.createJWTKeyFunc()

	// route multiplexer at end inorder to get all the fields required by the services
	a.mux = a.createRouter()
}

func GetApplication() IApp {
	appObj := &app{}
	appObj.initApplication()

	return appObj
}