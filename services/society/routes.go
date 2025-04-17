package society

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (ss *societyService) GetBasePath() string {
	return "/society"
}

func (ss *societyService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// org admin authorization
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", ss.createSociety)
		router.Patch("/{society}", ss.updateSocietyDetails)
		router.Delete("/{society}", ss.deleteSociety)
	})

	// org admin and user
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", ss.getAllSocieties)
	})

	return mux
}