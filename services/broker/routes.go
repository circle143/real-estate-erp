package broker

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *brokerService) GetBasePath() string {
	return "/society/{society}/broker"
}

func (s *brokerService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", s.addBrokerToSociety)
		router.Patch("/{brokerId}", s.updateBrokerDetails)
	})

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", s.getAllSocietyBrokers)
		//router.Post("/{brokerId}/report", s.getBrokerReport)
	})

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAndViewerAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/{brokerId}/report", s.getBrokerReport)

	})

	return mux
}
