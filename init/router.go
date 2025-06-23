package init

import (
	"circledigital.in/real-state-erp/services/bank"
	"circledigital.in/real-state-erp/services/broker"
	"circledigital.in/real-state-erp/services/charges"
	"circledigital.in/real-state-erp/services/flat"
	"circledigital.in/real-state-erp/services/organization"
	paymentPlan "circledigital.in/real-state-erp/services/payment-plan"
	"circledigital.in/real-state-erp/services/receipt"
	"circledigital.in/real-state-erp/services/sale"
	"circledigital.in/real-state-erp/services/society"
	"circledigital.in/real-state-erp/services/tower"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	appMiddleware "circledigital.in/real-state-erp/utils/middleware"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"net/http"
)

// serviceFactory defines type for getting a service
type serviceFactory func(app common.IApp) common.IService

var services = []serviceFactory{
	organization.CreateOrganizationService,
	society.CreateSocietyService,
	//flatType.CreateFlatTypeService,
	tower.CreateTowerService,
	flat.CreateFlatService,
	sale.CreateSaleService,
	charges.CreateChargesService,
	paymentPlan.CreatePaymentPlanService,
	broker.CreateBrokerService,
	bank.CreateBankService,
	receipt.CreateReceiptService,
}

// handle400 returns custom responses for not found routes and not allowed methods
func (a *app) handle400(router *chi.Mux) {
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		err := &custom.RequestError{
			Status:  http.StatusNotFound,
			Message: http.StatusText(http.StatusNotFound),
		}
		payload.HandleError(w, err)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		err := &custom.RequestError{
			Status:  http.StatusMethodNotAllowed,
			Message: http.StatusText(http.StatusMethodNotAllowed),
		}
		payload.HandleError(w, err)
	})
}

func (a *app) createRouter() *chi.Mux {
	mux := chi.NewMux()

	// application middlewares
	mux.Use(middleware.Heartbeat("/"))
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.StripSlashes)
	mux.Use(middleware.AllowContentType("application/json", "multipart/form-data"))

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	a.handle400(mux)

	// add authentication middleware
	authenticationMiddleware := appMiddleware.AuthenticationMiddleware{
		JWKS: a.jwtKey,
	}
	mux.Use(authenticationMiddleware.AuthenticateRequest)

	// add services routes
	for _, factory := range services {
		service := factory(a)
		mux.Mount(service.GetBasePath(), service.GetRoutes())
	}

	return mux
}
