package flat

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *flatService) GetBasePath() string {
	return "/society/{society}/flat"
}

func (s *flatService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// org admin authorization
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", s.createNewFlat)
		router.Post("/tower/{towerId}/bulk", s.createBulkFlats)
		router.Delete("/{flat}", s.deleteFlat)
		router.Patch("/{flatId}", s.updateFlatDetails)
	})

	// org admin and user
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", s.getAllSocietyFlats)
		router.Get("/tower/{tower}", s.getAllTowerFlats)
		router.Get("/search", s.getSocietyFlatByName)
	})

	return mux
}
