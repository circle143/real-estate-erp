package sale

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (cs *customerService) GetBasePath() string {
	return "/society/{society}/flat/{flat}/sale"
}

func (cs *customerService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", cs.addCustomerToFlat)
		//router.Delete("/{customer}", cs.addCustomerToFlat)
		//router.Patch("/{customer}", cs.updateCustomerDetails)
	})

	return mux
}
