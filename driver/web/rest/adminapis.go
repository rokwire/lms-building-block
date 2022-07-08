package rest

import (
	"encoding/json"
	"io/ioutil"
	"lms/core"
	"lms/core/model"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/tokenauth"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"
)

// AdminApisHandler handles the rest Admin APIs implementation
type AdminApisHandler struct {
	app    *core.Application
	config *model.Config
}

//GetNudges gets all the nudges
func (h AdminApisHandler) GetNudges(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {

	nudges, err := h.app.Administration.GetNudges()
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, "nudge", nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(nudges)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionMarshal, "nudge", nil, err, http.StatusInternalServerError, false)
	}

	return l.HttpResponseSuccessJSON(data)
}

type createNudge struct {
	ID     string                 `json:"id" bson:"id"`
	Name   string                 `json:"name" bson:"name"`
	Body   string                 `json:"body" bson:"body"`
	Params map[string]interface{} `json:"params" bson:"params"`
}

//CreateNudge creates nudge
func (h AdminApisHandler) CreateNudge(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionRead, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}
	var requestData createNudge
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionUnmarshal, "", nil, err, http.StatusBadRequest, true)
	}

	err = h.app.Administration.CreateNudge(l, requestData.ID, requestData.Name, requestData.Body, &requestData.Params)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HttpResponseSuccess()
}

//UpdateNudge updates nudge
func (h AdminApisHandler) UpdateNudge(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {
	params := mux.Vars(r)
	ID := params["id"]
	if len(ID) <= 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionRead, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, false)
	}
	var requestData model.Nudge
	err = json.Unmarshal(data, &requestData)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionUnmarshal, "", nil, err, http.StatusBadRequest, true)
	}

	err = h.app.Administration.UpdateNudge(l, ID, requestData.Name, requestData.Body, &requestData.Params)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionGet, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HttpResponseSuccess()
}

func (h AdminApisHandler) DeleteNudge(l *logs.Log, claims *tokenauth.Claims, w http.ResponseWriter, r *http.Request) logs.HttpResponse {
	params := mux.Vars(r)
	nudgeID := params["id"]
	if len(nudgeID) <= 0 {
		return l.HttpResponseErrorData(logutils.StatusMissing, logutils.TypeQueryParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Administration.DeleteNudge(l, nudgeID)
	if err != nil {
		return l.HttpResponseErrorAction(logutils.ActionDelete, "", nil, err, http.StatusInternalServerError, true)
	}
	return l.HttpResponseSuccess()
}
