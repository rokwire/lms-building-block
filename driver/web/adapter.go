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
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	// XCoreFunction defines the core function from docs
	XCoreFunction = "x-core-function"
	// XDataType defines the data type from docs
	XDataType = "x-data-type"
	// XAuthType defines the auth type from docs
	XAuthType = "x-authentication-type"
	// XRequestBody defines the request body from docs
	XRequestBody = "x-request-body"
	// XConversionFunction defines the conversion function from docs
	XConversionFunction = "x-conversion-function"
)

// Adapter entity
type Adapter struct {
	env           string
	lmsServiceURL string
	port          string
	serviceID     string
	auth          *Auth

	apisHandler      APIsHandler
	adminApisHandler rest.AdminApisHandler

	paths openapi3.Paths

	app *core.Application

	logger *logs.Logger
}

type handlerFunc = func(*logs.Log, *http.Request, *tokenauth.Claims) logs.HTTPResponse

// Start starts the module
func (we Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)
	router.Use()

	subrouter := router.PathPrefix("/" + we.serviceID).Subrouter()
	subrouter.PathPrefix("/doc/ui").Handler(we.serveDocUI())
	subrouter.HandleFunc("/doc", we.serveDoc)

	we.routeAPIs(router)

	log.Fatal(http.ListenAndServe(":"+we.port, router))
}

// routeAPIs calls registerHandler for every path specified as auto-generated in docs
func (we Adapter) routeAPIs(router *mux.Router) error {
	pathStrs := we.paths.InMatchingOrder()
	for _, pathStr := range pathStrs {
		path := we.paths.Find(pathStr)

		operations := map[string]*openapi3.Operation{
			http.MethodGet:    path.Get,
			http.MethodPost:   path.Post,
			http.MethodPut:    path.Put,
			http.MethodDelete: path.Delete,
		}

		for method, operation := range operations {
			if operation == nil || operation.Extensions[XCoreFunction] == nil || operation.Extensions[XDataType] == nil {
				continue
			}

			tag := operation.Tags[0]
			err := we.registerHandler(router, pathStr, method, tag, operation.Extensions[XCoreFunction].(string), operation.Extensions[XDataType].(string),
				operation.Extensions[XAuthType], operation.Extensions[XRequestBody], operation.Extensions[XConversionFunction])
			if err != nil {
				errArgs := logutils.FieldArgs(operation.Extensions)
				errArgs["method"] = method
				errArgs["tag"] = tag
				return errors.WrapErrorAction(logutils.ActionRegister, "api handler", &errArgs, err)
			}
		}
	}

	return nil
}

func (we Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./driver/web/docs/gen/def.yaml")
}

func (we Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/doc", we.lmsServiceURL)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func (we Adapter) wrapFunc(handler handlerFunc, authorization tokenauth.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logObj := we.logger.NewRequestLog(req)

		logObj.RequestReceived()

		var response logs.HTTPResponse
		if authorization != nil {
			responseStatus, claims, err := authorization.Check(req)
			if err != nil {
				logObj.SendHTTPResponse(w, logObj.HTTPResponseErrorAction(logutils.ActionValidate, logutils.TypeRequest, nil, err, responseStatus, true))
				return
			}

			if claims != nil {
				logObj.SetContext("account_id", claims.Subject)
			}
			response = handler(logObj, req, claims)
		} else {
			response = handler(logObj, req, nil)
		}

		logObj.SendHTTPResponse(w, response)
		logObj.RequestComplete()
	}
}

// func (we Adapter) validateRequest(req *http.Request) (*openapi3filter.RequestValidationInput, error) {
// 	route, pathParams, err := we.openAPIRouter.FindRoute(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dummyAuthFunc := func(c context.Context, input *openapi3filter.AuthenticationInput) error {
// 		return nil
// 	}
// 	options := &openapi3filter.Options{AuthenticationFunc: dummyAuthFunc}
// 	requestValidationInput := &openapi3filter.RequestValidationInput{
// 		Request:    req,
// 		PathParams: pathParams,
// 		Route:      route,
// 		Options:    options,
// 	}

// 	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
// 		return nil, err
// 	}
// 	return requestValidationInput, nil
// }

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(port string, serviceID string, app *core.Application, config *model.Config, serviceRegManager *authservice.ServiceRegManager, logger *logs.Logger) Adapter {
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
		paths["/"+serviceID+path] = obj
	}

	auth, err := NewAuth(serviceRegManager, app, config)
	if err != nil {
		logger.Fatalf("error creating auth - %s", err.Error())
	}

	apisHandler := NewAPIsHandler(app)
	adminApisHandler := rest.NewAdminApisHandler(app, config)
	return Adapter{
		lmsServiceURL:    config.LmsServiceURL,
		port:             port,
		serviceID:        serviceID,
		auth:             auth,
		paths:            paths,
		apisHandler:      apisHandler,
		adminApisHandler: adminApisHandler,
		app:              app,
		logger:           logger,
	}
}

// AppListener implements core.ApplicationListener interface
type AppListener struct {
	adapter *Adapter
}
