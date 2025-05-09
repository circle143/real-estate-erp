package flat

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (fs *flatService) GetBasePath() string {
	return "/society/{society}/flat"
}

func (fs *flatService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// org admin authorization
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", fs.createNewFlat)
		router.Delete("/{flat}", fs.deleteFlat)
	})

	// org admin and user
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", fs.getAllSocietyFlats)
		router.Get("/tower/{tower}", fs.getAllTowerFlats)
		router.Get("/search", fs.getSocietyFlatByName)
	})

	return mux
}
