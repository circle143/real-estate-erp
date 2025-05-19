package init

import (
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
	"log"
)

// package init handles all the application initialization

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

func (a *app) GetAWSConfig() common.IAWSConfig {
	return a.aws
}

// initApplication configures all the objects required for startup of application
func (a *app) initApplication() {
	// register validators
	err := payload.RegisterValidators()
	if err != nil {
		log.Fatalf("Error registering validators: %v\n", err)
	}

	a.dbClient = a.createDBClient()
	a.aws = a.createAWSConfig()
	a.jwtKey = a.createJWTKeyFunc()

	// route multiplexer at end inorder to get all the fields required by the services
	a.mux = a.createRouter()
}

func GetApplication() common.IApp {
	appObj := &app{}
	appObj.initApplication()

	return appObj
}
