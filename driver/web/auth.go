/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package web

import (
	"fmt"
	"lms/core"
	"lms/core/model"
	web "lms/driver/web/auth"
	"log"
	"net/http"

	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"

	"github.com/rokwire/core-auth-library-go/tokenauth"
)

// Auth handler
type Auth struct {
	admin        *TokenAuthHandlers
	internalAuth *InternalAuth
	coreAuth     *web.CoreAuth
	logger       *logs.Logger
}

//Authorization is an interface for auth types
type Authorization interface {
	check(req *http.Request) (int, *tokenauth.Claims, error)
	start()
}

//TokenAuthorization is an interface for auth types
type TokenAuthorization interface {
	Authorization
	getTokenAuth() *tokenauth.TokenAuth
}

func (auth *Auth) clientIDCheck(w http.ResponseWriter, r *http.Request) bool {
	clientID := r.Header.Get("APP")
	if len(clientID) == 0 {
		clientID = "edu.illinois.rokwire"
	}

	log.Println(fmt.Sprintf("400 - Bad Request"))
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad Request"))
	return false
}

// NewAuth creates new auth handler
func NewAuth(app *core.Application, config *model.Config) *Auth {
	coreAuth := web.NewCoreAuth(app, config)
	internalAuth := newInternalAuth(config)
	auth := Auth{coreAuth: coreAuth, internalAuth: internalAuth}
	return &auth
}

// InternalAuth handling the internal calls fromother BBs
type InternalAuth struct {
	internalAPIKey string
}

func newInternalAuth(config *model.Config) *InternalAuth {
	auth := InternalAuth{internalAPIKey: config.InternalAPIKey}
	return &auth
}

func (auth *InternalAuth) check(w http.ResponseWriter, r *http.Request) bool {
	apiKey := r.Header.Get("INTERNAL-API-KEY")
	//check if there is api key in the header
	if len(apiKey) == 0 {
		//no key, so return 400
		log.Println(fmt.Sprintf("400 - Bad Request"))

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return false
	}

	exist := auth.internalAPIKey == apiKey

	if !exist {
		//not exist, so return 401
		log.Println(fmt.Sprintf("401 - Unauthorized for key %s", apiKey))

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return false
	}
	return true
}

//PermissionsAuth entity
//This enforces that the user has permissions matching the policy
type PermissionsAuth struct {
	auth TokenAuthorization
}

func (a *PermissionsAuth) start() {}

func (a *PermissionsAuth) check(req *http.Request) (int, *tokenauth.Claims, error) {
	status, claims, err := a.auth.check(req)

	if err == nil && claims != nil {
		err = a.auth.getTokenAuth().AuthorizeRequestPermissions(claims, req)
		if err != nil {
			return http.StatusForbidden, nil, errors.WrapErrorAction("", logutils.TypeRequest, nil, err)
		}
	}

	return status, claims, err
}

func newPermissionsAuth(auth TokenAuthorization) *PermissionsAuth {
	permissionsAuth := PermissionsAuth{auth: auth}
	return &permissionsAuth
}

//TokenAuthHandlers represents token auth handlers
type TokenAuthHandlers struct {
	standard    TokenAuthorization
	permissions *PermissionsAuth
}

func (auth *TokenAuthHandlers) start() {
	auth.standard.start()
	auth.permissions.start()
}

//newTokenAuthHandlers creates new auth handlers for a
func newTokenAuthHandlers(auth TokenAuthorization) (*TokenAuthHandlers, error) {
	permissionsAuth := newPermissionsAuth(auth)

	authWrappers := TokenAuthHandlers{standard: auth, permissions: permissionsAuth}
	return &authWrappers, nil
}
