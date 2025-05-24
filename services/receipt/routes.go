package receipt

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *receiptService) GetBasePath() string {
	return "/society/{society}/sale/{saleId}/receipt"
}

func (s *receiptService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", s.createSaleReceipt)
		router.Post("/{receiptId}/clear", s.clearSaleReceipt)
	})

	return mux
}
