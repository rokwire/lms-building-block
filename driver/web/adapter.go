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

	apisHandler       APIsHandler
	clientAPIsHandler ClientAPIsHandler

	paths map[string]*openapi3.PathItem

	logger *logs.Logger
}

// Start starts the module
func (a *Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)
	router.Use()

	subrouter := router.PathPrefix("/" + a.serviceID).Subrouter()
	subrouter.PathPrefix("/doc/ui").Handler(a.serveDocUI())
	subrouter.HandleFunc("/doc", a.serveDoc)

	clientRouter := router.PathPrefix("/api").Subrouter()
	clientRouter.HandleFunc("/user-data", a.wrapFunc(a.clientAPIsHandler.getUserData, a.auth.client.User)).Methods("GET")

	err := a.routeAPIs(router)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":"+a.port, router))
}

// routeAPIs calls registerHandler for every path specified as auto-generated in docs
func (a *Adapter) routeAPIs(router *mux.Router) error {
	for path, pathItem := range a.paths {
		operations := map[string]*openapi3.Operation{
			http.MethodGet:    pathItem.Get,
			http.MethodPost:   pathItem.Post,
			http.MethodPut:    pathItem.Put,
			http.MethodDelete: pathItem.Delete,
		}

		for method, operation := range operations {
			if operation == nil || operation.Extensions[XCoreFunction] == nil || operation.Extensions[XDataType] == nil {
				continue
			}

			var requestBody interface{}
			tag := operation.Tags[0]
			convFunc := operation.Extensions[XConversionFunction]
			if operation.RequestBody != nil {
				// allow a panic to occur if something goes wrong
				// the service should be stopped anyway and the stack trace is logged without needing to recover and import runtime/debug to get the stack trace
				requestBody = operation.RequestBody.Value.Content.Get("application/json").Schema.Ref
			}
			err := a.registerHandler(router, path, method, tag, operation.Extensions[XCoreFunction].(string), operation.Extensions[XDataType].(string),
				operation.Extensions[XAuthType], requestBody, convFunc)
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

func (a *Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")
	http.ServeFile(w, r, "./driver/web/docs/gen/def.yaml")
}

func (a *Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/doc", a.lmsServiceURL)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

// func (a *Adapter) validateRequest(req *http.Request) (*openapi3filter.RequestValidationInput, error) {
// 	route, pathParams, err := a.openAPIRouter.FindRoute(req)
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

type handlerFunc = func(*logs.Log, *http.Request, *tokenauth.Claims) logs.HTTPResponse

func (a *Adapter) wrapFunc(handler handlerFunc, authorization tokenauth.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logObj := a.logger.NewRequestLog(req)

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

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(baseURL string, port string, serviceID string, app *core.Application, serviceRegManager *authservice.ServiceRegManager, logger *logs.Logger) Adapter {
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
	paths := make(map[string]*openapi3.PathItem, doc.Paths.Len())
	for path, obj := range doc.Paths.Map() {
		paths["/"+serviceID+path] = obj
	}

	auth, err := NewAuth(serviceRegManager, app)
	if err != nil {
		logger.Fatalf("error creating auth - %s", err.Error())
	}

	apisHandler := NewAPIsHandler(app)
	return Adapter{
		lmsServiceURL: baseURL,
		port:          port,
		serviceID:     serviceID,
		auth:          auth,
		paths:         paths,
		apisHandler:   apisHandler,
		logger:        logger,
	}
}

// AppListener implements core.ApplicationListener interface
type AppListener struct {
	adapter *Adapter
}
