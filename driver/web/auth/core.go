package web

import (
	"lms/core"
	"lms/core/model"
	"log"
	"net/http"

	"github.com/rokwire/core-auth-library-go/authservice"
	"github.com/rokwire/core-auth-library-go/tokenauth"
	"github.com/rokwire/logging-library-go/logs"
)

// CoreAuth implementation
type CoreAuth struct {
	app       *core.Application
	tokenAuth *tokenauth.TokenAuth
}

// NewCoreAuth creates new CoreAuth
func NewCoreAuth(app *core.Application, config *model.Config) *CoreAuth {

	remoteConfig := authservice.RemoteAuthDataLoaderConfig{
		AuthServicesHost: config.CoreBBHost,
	}

	serviceLoader, err := authservice.NewRemoteAuthDataLoader(remoteConfig, []string{"core"}, logs.NewLogger("lms", &logs.LoggerOpts{}))
	authService, err := authservice.NewAuthService("lms", config.LmsServiceURL, serviceLoader)
	if err != nil {
		log.Fatalf("Error initializing auth service: %v", err)
	}
	tokenAuth, err := tokenauth.NewTokenAuth(true, authService, nil, nil)
	if err != nil {
		log.Fatalf("Error intitializing token auth: %v", err)
	}

	auth := CoreAuth{app: app, tokenAuth: tokenAuth}
	return &auth
}

// Check checks the request contains a valid Core access token
func (ca CoreAuth) Check(r *http.Request) (bool, *tokenauth.Claims) {
	claims, err := ca.tokenAuth.CheckRequestTokens(r)
	if err != nil {
		log.Printf("error validate token: %s", err)
		return false, nil
	}

	if claims != nil {
		if claims.Valid() == nil {
			return true, claims
		}
	}

	return false, nil
}
