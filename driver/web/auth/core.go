package web

import (
	"lms/core"
	"lms/core/model"
	"log"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v2/authorization"
	"github.com/rokwire/core-auth-library-go/v2/authservice"
	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logutils"
)

// CoreAuth implementation
type CoreAuth struct {
	app       *core.Application
	tokenAuth *tokenauth.TokenAuth
}

// NewCoreAuth creates new CoreAuth
func NewCoreAuth(app *core.Application, config *model.Config) *CoreAuth {
	authService := authservice.AuthService{
		ServiceID:   "lms",
		ServiceHost: config.LmsServiceURL,
		FirstParty:  true,
		AuthBaseURL: config.CoreBBHost,
	}

	serviceRegLoader, err := authservice.NewRemoteServiceRegLoader(&authService, []string{"auth"})
	if err != nil {
		log.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := authservice.NewServiceRegManager(&authService, serviceRegLoader)
	if err != nil {
		log.Fatalf("Error initializing service registration manager: %v", err)
	}

	permissionAuth := authorization.NewCasbinStringAuthorization("driver/web/authorization_admin_policy.csv")
	tokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, permissionAuth, nil)
	if err != nil {
		log.Fatalf("Error intitializing token auth: %v", err)
	}

	auth := CoreAuth{app: app, tokenAuth: tokenAuth}
	return &auth
}

// Check checks the request contains a valid Core access token
func (ca CoreAuth) Check(r *http.Request) (*tokenauth.Claims, error) {
	claims, err := ca.tokenAuth.CheckRequestTokens(r)
	if err != nil || claims == nil {
		log.Printf("error validating token: %s", err)
		return nil, err
	}

	if claims.Anonymous {
		err = errors.ErrorData(logutils.StatusInvalid, logutils.TypeClaim, logutils.StringArgs("anonymous"))
		log.Println(err)
		return claims, err
	}

	return claims, nil
}

// AdminCheck checks the request contains a valid admin Core access token with the appropriate permissions
func (ca CoreAuth) AdminCheck(r *http.Request) (*tokenauth.Claims, error) {
	claims, err := ca.tokenAuth.CheckRequestTokens(r)
	if err != nil || claims == nil {
		log.Printf("error validate token: %s", err)
		return nil, err
	}

	if !claims.Admin {
		err = errors.ErrorData(logutils.StatusInvalid, logutils.TypeClaim, logutils.StringArgs("admin"))
		log.Println(err)
		return claims, err
	}

	err = ca.tokenAuth.AuthorizeRequestPermissions(claims, r)
	if err != nil {
		log.Println("invalid permissions:", err)
		return claims, err
	}

	return claims, err
}
