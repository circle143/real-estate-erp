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
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Patch("/customer/{customerId}", s.updateSaleCustomerDetails)
		router.Patch("/company-customer/{customerId}", s.updateSaleCompanyCustomerDetails)
		router.Delete("/{saleId}", s.clearSaleRecord)
	})

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/flat/{flat}", s.createSale)
		//router.Post("/{saleId}/add-payment-installment/{paymentId}", s.addPaymentInstallmentForSale)
		router.Get("/{saleId}/payment-breakdown", s.getSalePaymentBreakDown)
		//router.Get("/report", s.getSocietySalesReport)
		//router.Get("/tower/{towerId}/report", s.getTowerSalesReport)

	})

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAndViewerAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/report", s.getSocietySalesReport)
		router.Get("/tower/{towerId}/report", s.getTowerSalesReport)

	})

	return mux
}
