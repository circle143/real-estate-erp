package tower

import (
	"circledigital.in/real-state-erp/utils/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *towerService) GetBasePath() string {
	return "/society/{society}/tower"
}

func (s *towerService) GetRoutes() *chi.Mux {
	mux := chi.NewMux()
	authorizationMiddleware := &middleware.AuthorizationMiddleware{}

	// org admin authorization
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Post("/", s.createTower)
		router.Post("/bulk", s.bulkCreateTower)
		router.Patch("/{tower}", s.updateTower)
		router.Delete("/{tower}", s.deleteTower)
	})

	// org admin and user
	mux.Group(func(router chi.Router) {
		router.Use(authorizationMiddleware.OrganizationAdminAndUserAuthorization)
		router.Use(authorizationMiddleware.OrganizationAuthorization)

		router.Get("/", s.getAllTowers)
		router.Get("/{towerId}", s.getTowerById)
	})

	return mux
}
