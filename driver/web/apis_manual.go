// Copyright 2024 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package web

import (
	"encoding/json"
	"lms/core"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// ManualAPIsHandler handles the client rest APIs implementation
type ManualAPIsHandler struct {
	app *core.Application
}

func (h ManualAPIsHandler) getUserData(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	userData, err := h.app.Manual.GetUserData(claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)

	}
	response, err := json.Marshal(userData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(response)
}

// NewManualAPIsHandler creates new client API handler instance
func NewManualAPIsHandler(app *core.Application) ManualAPIsHandler {
	return ManualAPIsHandler{app: app}
}
