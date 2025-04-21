package organization

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (os *organizationService) GetBasePath() string {
	return "/organization"
}

func (os *organizationService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// admin role routes
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.AdminAuthorization)

		router.Post("/", os.createOrganization)
		router.Patch("/{orgId}/status", os.updateOrganizationStatus)
		router.Get("/", os.getAllOrganizations)
	})

	// organization admin routes
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/user", os.addUserToOrganization)
		router.Patch("/details", os.updateOrganizationDetails)
		router.Patch("/user/{userEmail}", os.updateOrganizationUserRole)
		router.Get("/users", os.getAllOrganizationUsers)
		router.Delete("/user/{userEmail}", os.removeUserFromOrganization)
	})

	// organization user and admin route
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/self", os.getAllOrganizationUsers)
	})

	return mux
}