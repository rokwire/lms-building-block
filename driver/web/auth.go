// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"lms/core"
	"lms/core/model"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v3/authorization"
	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// Auth handler
type Auth struct {
	client tokenauth.Handlers
	admin  tokenauth.Handlers
	logger *logs.Logger
}

// NewAuth creates new auth handler
func NewAuth(serviceRegManager *authservice.ServiceRegManager, app *core.Application, config *model.Config) (*Auth, error) {
	client, err := newClientAuth(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "client auth", nil, err)
	}
	clientHandlers := tokenauth.NewHandlers(client)

	admin, err := newAdminAuth(serviceRegManager)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "admin auth", nil, err)
	}
	adminHandlers := tokenauth.NewHandlers(admin)

	auth := Auth{client: clientHandlers, admin: adminHandlers}
	return &auth, nil
}

func newClientAuth(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	clientPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/client_permission_policy.csv")
	clientScopeAuth := authorization.NewCasbinScopeAuthorization("driver/web/client_scope_policy.csv", serviceRegManager.AuthService.ServiceID)
	clientTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, clientPermissionAuth, clientScopeAuth)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "client token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if claims.Admin {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "admin claim", nil)
		}
		if claims.System {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "system claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewScopeHandler(clientTokenAuth, check)
	return auth, nil
}

func newAdminAuth(serviceRegManager *authservice.ServiceRegManager) (*tokenauth.StandardHandler, error) {
	adminPermissionAuth := authorization.NewCasbinStringAuthorization("driver/web/admin_permission_policy.csv")
	adminTokenAuth, err := tokenauth.NewTokenAuth(true, serviceRegManager, adminPermissionAuth, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, "admin token auth", nil, err)
	}

	check := func(claims *tokenauth.Claims, req *http.Request) (int, error) {
		if !claims.Admin {
			return http.StatusUnauthorized, errors.ErrorData(logutils.StatusInvalid, "admin claim", nil)
		}

		return http.StatusOK, nil
	}

	auth := tokenauth.NewStandardHandler(adminTokenAuth, check)
	return auth, nil
}
