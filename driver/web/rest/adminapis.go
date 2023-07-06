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
	"io/ioutil"
	"lms/core"
	"lms/core/model"
	"net/http"
	"strconv"
	"strings"

	Def "lms/driver/web/docs/gen"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app    *core.Application
	config *model.Config
}

// GetNudgesConfig gets the nudges config
func (h AdminApisHandler) GetNudgesConfig(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {

	nudges, err := h.app.Administration.GetNudgesConfig(l)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "nudges config", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(nudges)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "nudge config", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// UpdateNudgesConfig updates the nudges config
func (h AdminApisHandler) UpdateNudgesConfig(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionRead, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}

	var requestData Def.NudgesConfig
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "nudges config", nil, err, http.StatusBadRequest, true)
	}

	err = h.app.Administration.UpdateNudgesConfig(l, requestData.Active, requestData.GroupName,
		requestData.TestGroupName, string(requestData.Mode), requestData.ProcessTime, requestData.BlockSize)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "nudges config", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// GetNudges gets all the nudges
func (h AdminApisHandler) GetNudges(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {

	nudges, err := h.app.Administration.GetNudges()
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "nudge", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(nudges)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "nudge", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// CreateNudge creates nudge
func (h AdminApisHandler) CreateNudge(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionRead, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}
	var requestData Def.AdminReqCreateNudge
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, "", nil, err, http.StatusBadRequest, true)
	}
	var req model.UsersSource
	var usersSource []model.UsersSource
	for _, u := range *requestData.UsersSources {
		req = model.UsersSource{Params: *u.Params, Type: *u.Type}
	}
	usersSource = append(usersSource, req)

	err = h.app.Administration.CreateNudge(l, requestData.Id, requestData.Name, requestData.Body, requestData.DeepLink, requestData.Params, requestData.Active, usersSource)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// AdminReqUpdateNudge defines model for _admin_req_update_nudge.
type adminReqUpdateNudge struct {
	Active       bool                   `json:"active"`
	Body         string                 `json:"body"`
	DeepLink     string                 `json:"deep_link"`
	Name         string                 `json:"name"`
	Params       map[string]interface{} `json:"params"`
	UsersSources *[]UsersSources        `json:"users_sources"`
}

// UsersSources defines model for UsersSources.
type UsersSources struct {
	Params *map[string]interface{} `json:"params,omitempty"`
	Type   *string                 `json:"type,omitempty"`
}

// UpdateNudge updates nudge
func (h AdminApisHandler) UpdateNudge(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {
	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionRead, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}
	var requestData adminReqUpdateNudge
	err = json.Unmarshal(data, &requestData)
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

	err = h.app.Administration.UpdateNudge(l, ID, requestData.Name, requestData.Body, requestData.DeepLink, requestData.Params, requestData.Active, usersSource)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// DeleteNudge deletes nudge
func (h AdminApisHandler) DeleteNudge(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {
	params := mux.Vars(r)
	nudgeID := params["id"]
	if len(nudgeID) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Administration.DeleteNudge(l, nudgeID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// FindSentNudges gets all the sent_nudges
func (h AdminApisHandler) FindSentNudges(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {
	//nudgeID
	var nudgeID *string
	nudgeIDParam := r.URL.Query().Get("nudge-id")
	if len(nudgeIDParam) > 0 {
		nudgeID = &nudgeIDParam
	}
	//userID
	var userID *string
	userIDParam := r.URL.Query().Get("user-id")
	if len(userIDParam) > 0 {
		userID = &userIDParam
	}
	//netID
	var netID *string
	netIDParam := r.URL.Query().Get("net-id")
	if len(netIDParam) > 0 {
		netID = &netIDParam
	}

	//mode
	var mode *string
	modeIDParam := r.URL.Query().Get("mode")
	if len(modeIDParam) > 0 {
		mode = &modeIDParam
	}

	sentNudges, err := h.app.Administration.FindSentNudges(l, nudgeID, userID, netID, mode)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "sent_nudges", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(sentNudges)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "sent_nudges", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// DeleteSentNudges deletes sent nudge
func (h AdminApisHandler) DeleteSentNudges(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {
	//sent nudge ID
	sentNudgesIDsParam := r.URL.Query().Get("ids")
	if sentNudgesIDsParam == "" {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("ids"), nil, http.StatusBadRequest, false)
	}

	ids := strings.Split(sentNudgesIDsParam, ",")

	err := h.app.Administration.DeleteSentNudges(l, ids)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// ClearTestSentNudges clears all sent nudges with the test mode
func (h AdminApisHandler) ClearTestSentNudges(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {
	err := h.app.Administration.ClearTestSentNudges(l)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, "test sent nudges", nil, err, http.StatusInternalServerError, true)
	}
	return l.HTTPResponseSuccess()
}

// FindNudgesProcesses gets all the nudges-process
func (h AdminApisHandler) FindNudgesProcesses(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HTTPResponse {
	var err error

	//limit and offset
	limit := 5
	limitArg := r.URL.Query().Get("limit")
	if limitArg != "" {
		limit, err = strconv.Atoi(limitArg)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionParse, logutils.TypeArg, logutils.StringArgs("limit"), err, http.StatusBadRequest, false)
		}
	}
	offset := 0
	offsetArg := r.URL.Query().Get("offset")
	if offsetArg != "" {
		offset, err = strconv.Atoi(offsetArg)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionParse, logutils.TypeArg, logutils.StringArgs("offset"), err, http.StatusBadRequest, false)
		}
	}
	nudgesProcesses, err := h.app.Administration.FindNudgesProcesses(l, limit, offset)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, "nudges_processes", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(nudgesProcesses)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, "sent_nudges", nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}
