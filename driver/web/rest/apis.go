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
	"lms/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"
)

// ApisHandler handles the rest APIs implementation
type ApisHandler struct {
	app    *core.Application
	config *model.Config
}

// Version gets version
func (h ApisHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.app.Services.GetVersion()))
}

// GetCourses gets courses
func (h ApisHandler) GetCourses(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {
	providerUserID := h.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, "net_id", nil, nil, http.StatusInternalServerError, false)
	}

	courses, err := h.app.Services.GetCourses(l, providerUserID)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, "course", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(courses)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, "course", nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(data)
}

// GetCourse gets a course
func (h ApisHandler) GetCourse(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {
	providerUserID := h.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, "net_id", nil, nil, http.StatusInternalServerError, false)
	}

	//course id
	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}
	courseID, err := strconv.Atoi(ID)
	if err != nil {
		return l.HttpResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	course, err := h.app.Services.GetCourse(l, providerUserID, courseID)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, "course", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(course)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, "course", nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(data)
}

// GetAssignemntGroups gets course assignments
func (h ApisHandler) GetAssignemntGroups(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {
	providerUserID := h.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, "net_id", nil, nil, http.StatusInternalServerError, false)
	}

	//course id
	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}
	courseID, err := strconv.Atoi(ID)
	if err != nil {
		return l.HttpResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	//include
	var include *string
	includeParam := r.URL.Query().Get("include")
	if len(includeParam) > 0 {
		include = &includeParam
	}

	assignmentGroups, err := h.app.Services.GetAssignmentGroups(l, providerUserID, courseID, include)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, "assignment group", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(assignmentGroups)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, "assignment group", nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(data)
}

// GetUsers gets course users
func (h ApisHandler) GetUsers(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {
	providerUserID := h.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, "net_id", nil, nil, http.StatusInternalServerError, false)
	}

	//course id
	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}
	courseID, err := strconv.Atoi(ID)
	if err != nil {
		return l.HttpResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	//include
	include := []string{}
	includeParam := r.URL.Query().Get("include")
	if len(includeParam) > 0 {
		include = strings.Split(includeParam, ",")
	}
	includeEnrolments := utils.Exist(include, "enrollments")
	includeScores := utils.Exist(include, "scores")

	user, err := h.app.Services.GetCourseUser(l, providerUserID, courseID, includeEnrolments, includeScores)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, "user", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(user)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, "user", nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(data)
}

// GetCurrentUser gets the current user
func (h ApisHandler) GetCurrentUser(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {
	providerUserID := h.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, "net_id", nil, nil, http.StatusInternalServerError, false)
	}

	user, err := h.app.Services.GetCurrentUser(l, providerUserID)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, "user", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(user)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, "user", nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(data)
}

func (h ApisHandler) getProviderUserID(claims *tokenauth.Claims) string {
	if claims == nil {
		return ""
	}
	return claims.ExternalIDs["net_id"]
}

// NewApisHandler creates new rest Handler instance
func NewApisHandler(app *core.Application, config *model.Config) ApisHandler {
	return ApisHandler{app: app, config: config}
}

// NewAdminApisHandler creates new rest Handler instance
func NewAdminApisHandler(app *core.Application, config *model.Config) AdminApisHandler {
	return AdminApisHandler{app: app, config: config}
}

// NewInternalApisHandler creates new rest Handler instance
func NewInternalApisHandler(app *core.Application, config *model.Config) InternalApisHandler {
	return InternalApisHandler{app: app, config: config}
}
