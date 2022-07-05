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

package main

import (
	"lms/core"
	"lms/core/model"
	cacheadapter "lms/driven/cache"
	"lms/driven/provider"
	storage "lms/driven/storage"
	driver "lms/driver/web"
	"log"
	"os"
	"strings"

	"github.com/rokwire/logging-library-go/logs"
)

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}

	loggerOpts := logs.LoggerOpts{SuppressRequests: []logs.HttpRequestProperties{logs.NewAwsHealthCheckHttpRequestProperties("/lms/version")}}
	logger := logs.NewLogger("core", &loggerOpts)

	port := getEnvKey("PORT", true)

	internalAPIKey := getEnvKey("INTERNAL_API_KEY", true)

	//mongoDB adapter
	mongoDBAuth := getEnvKey("MONGO_AUTH", true)
	mongoDBName := getEnvKey("MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("MONGO_TIMEOUT", false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout)
	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	defaultCacheExpirationSeconds := getEnvKey("DEFAULT_CACHE_EXPIRATION_SECONDS", false)
	cacheAdapter := cacheadapter.NewCacheAdapter(defaultCacheExpirationSeconds)

	//provider adapter
	canvasBaseURL := getEnvKey("CANVAS_BASE_URL", true)
	canvasTokenType := getEnvKey("CANVAS_TOKEN_TYPE", true)
	canvasToken := getEnvKey("CANVAS_TOKEN", true)
	providerAdapter := provider.NewProviderAdapter(canvasBaseURL, canvasToken, canvasTokenType)

	// application
	application := core.NewApplication(Version, Build, storageAdapter, providerAdapter, cacheAdapter)
	application.Start()

	// web adapter
	coreBBHost := getEnvKey("CORE_BB_HOST", true)
	lmsServiceURL := getEnvKey("LMS_SERVICE_URL", true)

	config := model.Config{
		InternalAPIKey:  internalAPIKey,
		CoreBBHost:      coreBBHost,
		LmsServiceURL:   lmsServiceURL,
		CanvasBaseURL:   canvasBaseURL,
		CanvasTokenType: canvasTokenType,
		CanvasToken:     canvasToken,
	}

	webAdapter := driver.NewWebAdapter(port, application, &config, logger)

	webAdapter.Start()
}

func getEnvKeyAsList(key string, required bool) []string {
	stringValue := getEnvKey(key, required)

	// it is comma separated format
	stringListValue := strings.Split(stringValue, ",")
	if len(stringListValue) == 0 && required {
		log.Fatalf("missing or empty env var: %s", key)
	}

	return stringListValue
}

func getEnvKey(key string, required bool) string {
	// get from the environment
	value, exist := os.LookupEnv(key)
	if !exist {
		if required {
			log.Fatal("No provided environment variable for " + key)
		} else {
			log.Printf("No provided environment variable for " + key)
		}
	}
	return value
}
