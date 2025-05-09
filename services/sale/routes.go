package sale

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *saleService) GetBasePath() string {
	return "/society/{society}/sale"
}

func (s *saleService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/flat/{flat}", s.createSale)
		router.Get("/payment-breakdown/{saleId}", s.getSalePaymentBreakDown)
		//router.Delete("/{customer}", cs.addCustomerToFlat)
		//router.Patch("/{customer}", cs.updateCustomerDetails)
	})

	return mux
}
