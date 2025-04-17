package organization

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (os *organizationService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// admin role routes
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.AdminAuthorization)

		router.Post("/organization", os.createOrganization)
		router.Patch("/organization/{orgId}/status", os.updateOrganizationStatus)
		router.Get("/organizations", os.getAllOrganizations)
	})

	// organization admin routes
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/organization/user", os.addUserToOrganization)
		router.Patch("/organization/details", os.updateOrganizationDetails)
		router.Patch("/organization/user/{userEmail}", os.updateOrganizationUserRole)
		router.Get("/organization/users", os.getAllOrganizationUsers)
		router.Delete("/organization/user/{userEmail}", os.removeUserFromOrganization)
	})

	return mux
}