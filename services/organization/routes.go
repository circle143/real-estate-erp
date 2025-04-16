package organization

import "github.com/go-chi/chi/v5"

func (os *organizationService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()

	// admin role routes
	mux.Group(func(router chi.Router) {
		router.Post("/organization", os.createOrganization)
		router.Patch("/organization/{orgId}/status", os.updateOrganizationStatus)
		router.Get("/organizations", os.getAllOrganizations)
	})

	// organization admin routes
	mux.Group(func(router chi.Router) {
		router.Post("/organization/user", os.addUserToOrganization)
		router.Patch("/organization/details", os.updateOrganizationDetails)
		router.Patch("/organization/user/{userEmail}", os.updateOrganizationUserRole)
		router.Get("/organization/users", os.getAllOrganizationUsers)
		router.Delete("/organization/user/{userEmail}", os.removeUserFromOrganization)
	})

	return mux
}