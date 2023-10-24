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

package rest

import (
	"encoding/json"
	"lms/core"
	"lms/core/model"
	"net/http"

	Def "lms/driver/web/docs/gen"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app    *core.Application
	config *model.Config
}

// CreateNudge creates nudge
func (h AdminApisHandler) CreateNudge(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData Def.AdminReqCreateNudge
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "", nil, err, http.StatusBadRequest, true)
	}
	var req model.UsersSource
	var usersSource []model.UsersSource
	for _, u := range *requestData.UsersSources {
		req = model.UsersSource{Params: *u.Params, Type: *u.Type}
	}
	usersSource = append(usersSource, req)

	err = h.app.Administration.CreateNudge(requestData.Id, requestData.Name, requestData.Body, requestData.DeepLink, requestData.Params, requestData.Active, usersSource)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// UpdateNudge updates nudge
func (h AdminApisHandler) UpdateNudge(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var requestData Def.AdminReqUpdateNudge
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "", nil, err, http.StatusBadRequest, true)
	}

	var req model.UsersSources
	var usersSource []model.UsersSources
	if requestData.UsersSources != nil {
		for _, u := range *requestData.UsersSources {
			req = model.UsersSources{Type: u.Type, Params: u.Params}
		}
		usersSource = append(usersSource, req)
	}

	err = h.app.Administration.UpdateNudge(ID, requestData.Name, requestData.Body, requestData.DeepLink, requestData.Params, requestData.Active, usersSource)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// NewAdminApisHandler creates new rest Handler instance
func NewAdminApisHandler(app *core.Application, config *model.Config) AdminApisHandler {
	return AdminApisHandler{app: app, config: config}
}
