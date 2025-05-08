package payment_plan

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *paymentPlanService) GetBasePath() string {
	return "/society/{society}/payment-plan"
}

func (s *paymentPlanService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", s.createPaymentPlan)
		router.Post("/{paymentId}/tower/{towerId}", s.markPaymentPlanActiveForTower)
		router.Get("/", s.getSocietyPaymentPlans)
		router.Get("/tower/{towerId}", s.getTowerPaymentPlans)
	})

	return mux
}
