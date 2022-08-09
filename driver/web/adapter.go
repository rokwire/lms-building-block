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
	"context"
	"fmt"
	"lms/core"
	"lms/core/model"
	"lms/driver/web/rest"
	"lms/utils"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"

	"github.com/rokwire/core-auth-library-go/v2/tokenauth"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

//Adapter entity
type Adapter struct {
	env           string
	lmsServiceURL string
	port          string
	auth          *Auth
	openAPIRouter routers.Router

	apisHandler         rest.ApisHandler
	adminApisHandler    rest.AdminApisHandler
	internalApisHandler rest.InternalApisHandler

	app *core.Application

	logger *logs.Logger
}

// @title Rewards Building Block API
// @description RoRewards Building Block API Documentation.
// @version 1.0.2
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost
// @BasePath /content
// @schemes https

// @securityDefinitions.apikey InternalApiAuth
// @in header (add INTERNAL-API-KEY with correct value as a header)
// @name Authorization

// @securityDefinitions.apikey AdminUserAuth
// @in header (add Bearer prefix to the Authorization value)
// @name Authorization

// @securityDefinitions.apikey AdminGroupAuth
// @in header
// @name GROUP

//Start starts the module
func (we Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)
	router.Use()

	subrouter := router.PathPrefix("/lms").Subrouter()
	subrouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	subrouter.HandleFunc("/doc", we.serveDoc)
	subrouter.HandleFunc("/version", we.wrapFunc(we.apisHandler.Version)).Methods("GET")

	// handle apis
	apiRouter := subrouter.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("/courses", we.userAuthWrapFunc(we.apisHandler.GetCourses)).Methods("GET")
	apiRouter.HandleFunc("/courses/{id}", we.userAuthWrapFunc(we.apisHandler.GetCourse)).Methods("GET")
	apiRouter.HandleFunc("/courses/{id}/assignment-groups", we.userAuthWrapFunc(we.apisHandler.GetAssignemntGroups)).Methods("GET")
	apiRouter.HandleFunc("/courses/{id}/users", we.userAuthWrapFunc(we.apisHandler.GetUsers)).Methods("GET")
	apiRouter.HandleFunc("/users/self", we.userAuthWrapFunc(we.apisHandler.GetCurrentUser)).Methods("GET")

	///admin ///
	adminRouter := subrouter.PathPrefix("/admin").Subrouter()

	adminRouter.HandleFunc("/nudges-config", we.adminAuthWrapFunc(we.adminApisHandler.GetNudgesConfig)).Methods("GET")
	adminRouter.HandleFunc("/nudges-config", we.adminAuthWrapFunc(we.adminApisHandler.UpdateNudgesConfig)).Methods("PUT")
	adminRouter.HandleFunc("/nudges", we.adminAuthWrapFunc(we.adminApisHandler.GetNudges)).Methods("GET")
	adminRouter.HandleFunc("/nudges", we.adminAuthWrapFunc(we.adminApisHandler.CreateNudge)).Methods("POST")
	adminRouter.HandleFunc("/nudges/{id}", we.adminAuthWrapFunc(we.adminApisHandler.UpdateNudge)).Methods("PUT")
	adminRouter.HandleFunc("/nudges/{id}", we.adminAuthWrapFunc(we.adminApisHandler.DeleteNudge)).Methods("DELETE")
	adminRouter.HandleFunc("/sent-nudges", we.adminAuthWrapFunc(we.adminApisHandler.FindSentNudges)).Methods("GET")
	adminRouter.HandleFunc("/sent-nudges", we.adminAuthWrapFunc(we.adminApisHandler.DeleteSentNudges)).Methods("DELETE")
	adminRouter.HandleFunc("/test-sent-nudges", we.adminAuthWrapFunc(we.adminApisHandler.ClearTestSentNudges)).Methods("DELETE")
	adminRouter.HandleFunc("/nudges-process", we.adminAuthWrapFunc(we.adminApisHandler.FindNudgesProcess)).Methods("GET")

	log.Fatal(http.ListenAndServe(":"+we.port, router))
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./driver/web/docs/gen/def.yaml")
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/doc", we.lmsServiceURL)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func (we Adapter) wrapFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		handler(w, req)
	}
}

type userAuthFunc = func(*logs.Log, *tokenauth.Claims, http.ResponseWriter, *http.Request) logs.HttpResponse

func (we Adapter) userAuthWrapFunc(handler userAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logObj := we.logger.NewRequestLog(req)

		logObj.RequestReceived()

		// validate request
		_, err := we.validateRequest(req)
		if err != nil {
			logObj.RequestErrorAction(w, logutils.ActionValidate, logutils.TypeRequest, nil, err, http.StatusBadRequest, true)
			return
		}

		claims, err := we.auth.coreAuth.Check(req)
		if err != nil {
			if claims == nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// process the request
		response := handler(logObj, claims, w, req)

		/// return response
		// headers
		if len(response.Headers) > 0 {
			for key, values := range response.Headers {
				if len(values) > 0 {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
			}
		}
		// response code
		w.WriteHeader(response.ResponseCode)
		// body
		if len(response.Body) > 0 {
			w.Write(response.Body)
		}

		logObj.RequestComplete()
	}
}

type adminAuthFunc = func(*logs.Log, *tokenauth.Claims, http.ResponseWriter, *http.Request) logs.HttpResponse

func (we Adapter) adminAuthWrapFunc(handler userAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logObj := we.logger.NewRequestLog(req)

		logObj.RequestReceived()

		// validate request
		_, err := we.validateRequest(req)
		if err != nil {
			logObj.RequestErrorAction(w, logutils.ActionValidate, logutils.TypeRequest, nil, err, http.StatusBadRequest, true)
			return
		}

		claims, err := we.auth.coreAuth.AdminCheck(req)
		if err != nil {
			if claims == nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		// process the request
		response := handler(logObj, claims, w, req)

		/// return response
		// headers
		if len(response.Headers) > 0 {
			for key, values := range response.Headers {
				if len(values) > 0 {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
			}
		}
		// response code
		w.WriteHeader(response.ResponseCode)
		// body
		if len(response.Body) > 0 {
			w.Write(response.Body)
		}

		logObj.RequestComplete()
	}
}

type internalAPIKeyAuthFunc = func(http.ResponseWriter, *http.Request)

func (we Adapter) internalAPIKeyAuthWrapFunc(handler internalAPIKeyAuthFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		utils.LogRequest(req)

		apiKeyAuthenticated := we.auth.internalAuth.check(w, req)

		if apiKeyAuthenticated {
			handler(w, req)
		} else {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	}
}

func (we Adapter) validateRequest(req *http.Request) (*openapi3filter.RequestValidationInput, error) {
	route, pathParams, err := we.openAPIRouter.FindRoute(req)
	if err != nil {
		return nil, err
	}

	dummyAuthFunc := func(c context.Context, input *openapi3filter.AuthenticationInput) error {
		return nil
	}
	options := &openapi3filter.Options{AuthenticationFunc: dummyAuthFunc}
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options:    options,
	}

	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
		return nil, err
	}
	return requestValidationInput, nil
}

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(port string, app *core.Application, config *model.Config, logger *logs.Logger) Adapter {
	//openAPI doc
	loader := &openapi3.Loader{Context: context.Background(), IsExternalRefsAllowed: true}
	doc, err := loader.LoadFromFile("driver/web/docs/gen/def.yaml")
	if err != nil {
		logger.Fatalf("error on openapi3 load from file - %s", err.Error())
	}
	err = doc.Validate(loader.Context)
	if err != nil {
		logger.Fatalf("error on openapi3 validate - %s", err.Error())
	}

	//Ignore servers. Validating reqeusts against the documented servers can cause issues when routing traffic through proxies/load-balancers.
	doc.Servers = nil

	//To correctly route traffic to base path, we must add to all paths since servers are ignored
	paths := make(openapi3.Paths, len(doc.Paths))
	for path, obj := range doc.Paths {
		paths["/lms"+path] = obj
	}
	doc.Paths = paths

	openAPIRouter, err := gorillamux.NewRouter(doc)
	if err != nil {
		logger.Fatalf("error on openapi3 gorillamux router - %s", err.Error())
	}

	auth := NewAuth(app, config)

	apisHandler := rest.NewApisHandler(app, config)
	adminApisHandler := rest.NewAdminApisHandler(app, config)
	internalApisHandler := rest.NewInternalApisHandler(app, config)
	return Adapter{
		lmsServiceURL:       config.LmsServiceURL,
		port:                port,
		auth:                auth,
		apisHandler:         apisHandler,
		adminApisHandler:    adminApisHandler,
		internalApisHandler: internalApisHandler,
		app:                 app,
		logger:              logger,
		openAPIRouter:       openAPIRouter,
	}
}

// AppListener implements core.ApplicationListener interface
type AppListener struct {
	adapter *Adapter
}
