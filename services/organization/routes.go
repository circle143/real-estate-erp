package organization

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *organizationService) GetBasePath() string {
	return "/organization"
}

func (s *organizationService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// admin role routes
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.AdminAuthorization)

		router.Post("/", s.createOrganization)
		router.Patch("/{orgId}/status", s.updateOrganizationStatus)
		router.Get("/", s.getAllOrganizations)
	})

	// organization admin routes
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/user", s.addUserToOrganization)
		router.Patch("/details", s.updateOrganizationDetails)
		router.Patch("/user/{userEmail}", s.updateOrganizationUserRole)
		router.Get("/users", s.getAllOrganizationUsers)
		router.Delete("/user/{userEmail}", s.removeUserFromOrganization)
	})

	// organization user and admin route
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/self", s.getCurrentUserOrganization)
	})

	return mux
}
