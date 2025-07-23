package payment_plan_group

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
		router.Post("/{paymentPlanItemId}/tower/{towerId}", s.markPaymentPlanItemActiveForTower)
		router.Post("/{paymentPlanItemId}/flat/{flatId}", s.markPaymentPlanItemActiveForFlat)
		router.Get("/", s.getPaymentPlan)
		router.Get("/tower/{towerId}", s.getTowerPaymentPlan)
		router.Get("/flat/{flatId}", s.getFlatPaymentPlan)
	})

	return mux
}
