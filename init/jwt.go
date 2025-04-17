package init

import (
	"fmt"
	"github.com/MicahParks/keyfunc/v3"
	"log"
	"os"
)

func (a *app) createJWTKeyFunc() keyfunc.Keyfunc {
	userPoolID := os.Getenv("USER_POOL_ID")
	region := os.Getenv("AWS_REGION")
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)

	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		log.Fatalf("Failed to create JWK Set from resource at the given URL.\nError: %s", err)
	}

	return jwks
}