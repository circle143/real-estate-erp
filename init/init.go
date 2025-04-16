package init

import "github.com/go-chi/chi/v5"

// package init handles all the application initialization

// IService is implemented by all services and return routes exposed by the service
type IService interface {
	GetRoutes() *chi.Mux
}