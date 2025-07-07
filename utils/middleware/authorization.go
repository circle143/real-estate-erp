package middleware

import (
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"net/http"
	"strings"
)

// unauthorizedError defined error to send when authorization fails
var unauthorizedError = &custom.RequestError{
	Status:  http.StatusForbidden,
	Message: http.StatusText(http.StatusForbidden),
}

// AuthorizationMiddleware handles user authorization
type AuthorizationMiddleware struct{}

// AdminAuthorization protects admin routes
func (am *AuthorizationMiddleware) AdminAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(custom.UserRoleKey).(custom.UserRole)
		if userRole != custom.CIRCLEADMIN {
			payload.HandleError(w, unauthorizedError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OrganizationAdminAuthorization protects admin routes
func (am *AuthorizationMiddleware) OrganizationAdminAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(custom.UserRoleKey).(custom.UserRole)
		if userRole != custom.ORGADMIN {
			payload.HandleError(w, unauthorizedError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OrganizationAdminAndUserAuthorization protects organization user routes
func (am *AuthorizationMiddleware) OrganizationAdminAndUserAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(custom.UserRoleKey).(custom.UserRole)
		if userRole != custom.ORGADMIN && userRole != custom.ORGUSER {
			payload.HandleError(w, unauthorizedError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OrganizationAdminAndUserAndViewerAuthorization protects organization user routes
func (am *AuthorizationMiddleware) OrganizationAdminAndUserAndViewerAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(custom.UserRoleKey).(custom.UserRole)
		if userRole != custom.ORGADMIN && userRole != custom.ORGUSER && userRole != custom.ORGVIEWER {
			payload.HandleError(w, unauthorizedError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OrganizationAuthorization checks organization id for user
func (am *AuthorizationMiddleware) OrganizationAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orgId := r.Context().Value(custom.OrganizationIDKey).(string)
		if strings.TrimSpace(orgId) == "" {
			payload.HandleError(w, unauthorizedError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
