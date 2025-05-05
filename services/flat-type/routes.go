package flat_type

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (fts *flatTypeService) GetBasePath() string {
	return "/society/{society}/flat-type"
}

func (fts *flatTypeService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// org admin authorization
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", fts.createFlatType)
		//router.Patch("/{flatType}", fts.updateFlatType)
		router.Delete("/{flatType}", fts.deleteFlatType)
	})

	// org admin and user
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", fts.getAllFlatTypes)
	})

	return mux
}
