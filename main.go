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

package main

import (
	"lms/core"
	"lms/core/model"
	cacheadapter "lms/driven/cache"
	"lms/driven/corebb"
	"lms/driven/groups"
	"lms/driven/notifications"
	"lms/driven/provider"
	storage "lms/driven/storage"
	driver "lms/driver/web"
	"log"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/rokwire/core-auth-library-go/v2/authservice"
	"github.com/rokwire/core-auth-library-go/v2/sigauth"
	"github.com/rokwire/logging-library-go/v2/logs"
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

	serviceID := "lms"

	loggerOpts := logs.LoggerOpts{
		SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version"),
		SensitiveHeaders: []string{"Internal-Api-Key"},
	}
	logger := logs.NewLogger(serviceID, &loggerOpts)

	port := getEnvKey("LMS_PORT", true)

	internalAPIKey := getEnvKey("LMS_INTERNAL_API_KEY", true)

	//mongoDB adapter
	mongoDBAuth := getEnvKey("LMS_MONGO_AUTH", true)
	mongoDBName := getEnvKey("LMS_MONGO_DATABASE", true)
	mongoTimeout := getEnvKey("LMS_MONGO_TIMEOUT", false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, logger)
	err := storageAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the mongoDB adapter - " + err.Error())
	}

	defaultCacheExpirationSeconds := getEnvKey("LMS_DEFAULT_CACHE_EXPIRATION_SECONDS", false)
	cacheAdapter := cacheadapter.NewCacheAdapter(defaultCacheExpirationSeconds)

	//provider adapter
	canvasBaseURL := getEnvKey("LMS_CANVAS_BASE_URL", true)
	canvasTokenType := getEnvKey("LMS_CANVAS_TOKEN_TYPE", true)
	canvasToken := getEnvKey("LMS_CANVAS_TOKEN", true)
	providerAdapter := provider.NewProviderAdapter(canvasBaseURL, canvasToken, canvasTokenType, mongoDBAuth, mongoDBName, mongoTimeout, logger)
	err = providerAdapter.Start()
	if err != nil {
		log.Fatal("Cannot start the provider adapter - " + err.Error())
	}

	//groups BB adapter
	groupsHost := getEnvKey("LMS_GROUPS_BB_HOST", true)
	groupsBBAdapter := groups.NewGroupsAdapter(groupsHost, internalAPIKey)

	//notifications BB adapter
	app := getEnvKey("LMS_APP_ID", true)
	org := getEnvKey("LMS_ORG_ID", true)
	notificationHost := getEnvKey("LMS_NOTIFICATIONS_BB_HOST", true)
	notificationsBBAdapter := notifications.NewNotificationsAdapter(notificationHost, internalAPIKey, app, org)

	//core adapter
	cHost, cServiceAccountManager := getCoreBBAdapterValues(logger, serviceID)
	coreAdapter := corebb.NewCoreAdapter(cHost, cServiceAccountManager, org, app)

	// application
	application := core.NewApplication(Version, Build, storageAdapter, providerAdapter,
		groupsBBAdapter, notificationsBBAdapter, cacheAdapter, coreAdapter, logger)
	application.Start()

	// web adapter
	coreBBHost := getEnvKey("LMS_CORE_BB_HOST", true)
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

func getCoreBBAdapterValues(logger *logs.Logger, serviceID string) (string, *authservice.ServiceAccountManager) {
	host := getEnvKey("LMS_CORE_BB_CURRENT_HOST", true)
	coreBBHost := getEnvKey("LMS_CORE_BB_CORE_HOST", true)

	authService := authservice.AuthService{
		ServiceID:   serviceID,
		ServiceHost: host,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceRegLoader, err := authservice.NewRemoteServiceRegLoader(&authService, []string{})
	if err != nil {
		logger.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := authservice.NewServiceRegManager(&authService, serviceRegLoader)
	if err != nil {
		logger.Fatalf("Error initializing service registration manager: %v", err)
	}

	serviceAccountID := getEnvKey("LMS_SERVICE_ACCOUNT_ID", true)
	privKeyRaw := getEnvKey("LMS_PRIV_KEY", true)
	privKeyRaw = strings.ReplaceAll(privKeyRaw, "\\n", "\n")
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privKeyRaw))
	if err != nil {
		log.Fatalf("Error parsing priv key: %v", err)
	}

	signatureAuth, err := sigauth.NewSignatureAuth(privKey, serviceRegManager, false)
	if err != nil {
		log.Fatalf("Error initializing signature auth: %v", err)
	}

	serviceAccountLoader, err := authservice.NewRemoteServiceAccountLoader(&authService, serviceAccountID, signatureAuth)
	if err != nil {
		log.Fatalf("Error initializing remote service account loader: %v", err)
	}

	serviceAccountManager, err := authservice.NewServiceAccountManager(&authService, serviceAccountLoader)
	if err != nil {
		log.Fatalf("Error initializing service account manager: %v", err)
	}
	return coreBBHost, serviceAccountManager
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
