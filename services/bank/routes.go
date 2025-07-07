package bank

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *bankService) GetBasePath() string {
	return "/society/{society}/bank"
}

func (s *bankService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", s.addBankAccountToSociety)
		router.Patch("/{bankId}", s.updateBankAccountDetails)
	})

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", s.getAllSocietyBankAccounts)
		//router.Post("/{bankId}/report", s.getBankReport)
	})

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAndViewerAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/{bankId}/report", s.getBankReport)

	})
	return mux
}
