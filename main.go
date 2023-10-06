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

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/envloader"
	"github.com/rokwire/core-auth-library-go/v3/keys"
	"github.com/rokwire/core-auth-library-go/v3/sigauth"
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
	loggerOpts := logs.LoggerOpts{SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version")}
	logger := logs.NewLogger(serviceID, &loggerOpts)
	envLoader := envloader.NewEnvLoader(Version, logger)

	envPrefix := strings.ReplaceAll(strings.ToUpper(serviceID), "-", "_") + "_"
	port := envLoader.GetAndLogEnvVar(envPrefix+"PORT", false, false)
	if len(port) == 0 {
		port = "80"
	}

	internalAPIKey := getEnvKey("LMS_INTERNAL_API_KEY", true)

	// mongoDB adapter
	mongoDBAuth := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_AUTH", true, true)
	mongoDBName := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_DATABASE", true, false)
	mongoTimeout := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_TIMEOUT", false, false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, logger)
	err := storageAdapter.Start()
	if err != nil {
		logger.Fatalf("Cannot start the mongoDB adapter: %v", err)
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

	// Service registration
	baseURL := envLoader.GetAndLogEnvVar(envPrefix+"BASE_URL", true, false)
	coreBBBaseURL := envLoader.GetAndLogEnvVar(envPrefix+"CORE_BB_BASE_URL", true, false)

	authService := authservice.AuthService{
		ServiceID:   serviceID,
		ServiceHost: baseURL,
		FirstParty:  true,
		AuthBaseURL: coreBBBaseURL,
	}

	serviceRegLoader, err := authservice.NewRemoteServiceRegLoader(&authService, []string{})
	if err != nil {
		logger.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := authservice.NewServiceRegManager(&authService, serviceRegLoader, !strings.HasPrefix(baseURL, "http://localhost"))
	if err != nil {
		logger.Fatalf("Error initializing service registration manager: %v", err)
	}

	//core adapter
	cHost, cServiceAccountManager := getCoreBBAdapterValues(logger, serviceID, serviceRegManager, envLoader, envPrefix)
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
	webAdapter := driver.NewWebAdapter(port, application, &config, serviceRegManager, logger)
	webAdapter.Start()
}

func getCoreBBAdapterValues(logger *logs.Logger, serviceID string, serviceRegManager *authservice.ServiceRegManager, envLoader envloader.EnvLoader, envPrefix string) (string, *authservice.ServiceAccountManager) {
	host := getEnvKey("LMS_CORE_BB_CURRENT_HOST", true)
	coreBBHost := getEnvKey("LMS_CORE_BB_CORE_HOST", true)

	authService := authservice.AuthService{
		ServiceID:   serviceID,
		ServiceHost: host,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceAccountID := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_ACCOUNT_ID", false, false)
	privKeyRaw := envLoader.GetAndLogEnvVar(envPrefix+"PRIV_KEY", false, true)
	privKeyRaw = strings.ReplaceAll(privKeyRaw, "\\n", "\n")
	privKey, err := keys.NewPrivKey(keys.PS256, privKeyRaw)
	if err != nil {
		logger.Errorf("Error parsing priv key: %v", err)
	} else if serviceAccountID == "" {
		logger.Errorf("Missing service account id")
	}
	signatureAuth, err := sigauth.NewSignatureAuth(privKey, serviceRegManager, false, false)
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
