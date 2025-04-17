package middleware

import (
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"context"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

// tokenPayload represents user related info from JWT
type tokenPayload struct {
	UserRole custom.UserRole
	OrgId    string
}

// AuthenticationMiddleware authenticates the incoming http request for JWT authentication
type AuthenticationMiddleware struct {
	JWKS keyfunc.Keyfunc
}

// AuthenticateRequest authenticates and adds values to request context
func (am *AuthenticationMiddleware) AuthenticateRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenPayloadObj, err := am.parseAuthHeader(r)
		if err != nil {
			payload.HandleError(w, err)
			return
		}

		// add context values to request
		reqContext := r.Context()
		reqContext = context.WithValue(reqContext, custom.UserRoleKey, tokenPayloadObj.UserRole)

		if tokenPayloadObj.OrgId != "" {
			reqContext = context.WithValue(reqContext, custom.OrganizationIDKey, tokenPayloadObj.OrgId)
		}
		reqWithValues := r.WithContext(reqContext)
		*r = *reqWithValues

		next.ServeHTTP(w, r)
	})
}

// parseAuthHeader parses the JWT and returns token payload relevant information
func (am *AuthenticationMiddleware) parseAuthHeader(r *http.Request) (*tokenPayload, error) {
	// get auth header
	authHeader := r.Header.Get("Authorization")
	if strings.TrimSpace(authHeader) == "" {
		return nil, &custom.RequestError{
			Status:  http.StatusUnauthorized,
			Message: "Request lacks authorization header.",
		}
	}

	// validate authorization scheme
	authArray := strings.Split(authHeader, " ")
	if len(authArray) != 2 {
		return nil, &custom.RequestError{
			Status:  http.StatusUnauthorized,
			Message: "Unsupported authorization scheme.",
		}
	}
	if bearer := authArray[0]; strings.ToLower(bearer) != "bearer" {
		return nil, &custom.RequestError{
			Status:  http.StatusUnauthorized,
			Message: "Unsupported authorization scheme.",
		}
	}

	invalidTokenError := &custom.RequestError{
		Status:  http.StatusUnauthorized,
		Message: "Invalid token.",
	}

	// validate JWT
	tokenString := authArray[1]
	token, err := jwt.Parse(tokenString, am.JWKS.Keyfunc)
	if err != nil || !token.Valid {
		return nil, invalidTokenError
	}

	// get claims from token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, invalidTokenError
	}

	groups, ok := claims["cognito:groups"].([]interface{})
	if !ok || len(groups) == 0 {
		return nil, invalidTokenError
	}

	role, ok := groups[0].(string)
	if !ok {
		return nil, invalidTokenError
	}

	orgID := ""
	if val, ok := claims[custom.OrgIdCustomAttribute]; ok {
		if str, ok := val.(string); ok && str != "" {
			orgID = str
		}
	}

	return &tokenPayload{
		UserRole: custom.UserRole(role),
		OrgId:    orgID,
	}, nil
}