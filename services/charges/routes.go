package charges

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *chargesService) GetBasePath() string {
	return "/society/{society}/charges"
}

func (s *chargesService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// preference location charges
	location := chi.NewMux()
	location.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", s.addNewPreferenceLocationCharge)
		router.Patch("/{chargeId}/price", s.updatePreferenceLocationChargePrice)
		router.Patch("/{chargeId}/details", s.updatePreferenceLocationChargeDetails)
	})
	location.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", s.getAllPreferenceLocationCharges)
	})

	// other charges
	other := chi.NewMux()
	other.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", s.addNewOtherCharge)
		router.Patch("/{chargeId}/price", s.updateOtherChargePrice)
		router.Patch("/{chargeId}/details", s.updateOtherChargeDetails)
	})
	other.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", s.getAllOtherCharges)
		router.Get("/optional", s.getAllOtherOptionalCharges)
	})

	mux.Mount("/preference-location", location)
	mux.Mount("/other", other)

	return mux
}
