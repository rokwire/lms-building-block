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
	cacheadapter "lms/driven/cache"
	"lms/driven/corebb"
	"lms/driven/groups"
	"lms/driven/notifications"
	"lms/driven/provider"
	storage "lms/driven/storage"
	driver "lms/driver/web"
	"log"
	"strings"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/envloader"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/keys"
	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/sigauth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
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

	internalAPIKey := envLoader.GetAndLogEnvVar(envPrefix+"INTERNAL_API_KEY", true, true)

	// mongoDB adapter
	mongoDBAuth := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_AUTH", true, true)
	mongoDBName := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_DATABASE", true, false)
	mongoTimeout := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_TIMEOUT", false, false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, logger)
	err := storageAdapter.Start()
	if err != nil {
		logger.Fatalf("Cannot start the mongoDB adapter: %v", err)
	}

	defaultCacheExpirationSeconds := envLoader.GetAndLogEnvVar(envPrefix+"DEFAULT_CACHE_EXPIRATION_SECONDS", false, false)
	cacheAdapter := cacheadapter.NewCacheAdapter(defaultCacheExpirationSeconds)

	//provider adapter
	canvasBaseURL := envLoader.GetAndLogEnvVar(envPrefix+"CANVAS_BASE_URL", true, false)
	canvasTokenType := envLoader.GetAndLogEnvVar(envPrefix+"CANVAS_TOKEN_TYPE", true, false)
	canvasToken := envLoader.GetAndLogEnvVar(envPrefix+"CANVAS_TOKEN", true, true)
	providerAdapter := provider.NewProviderAdapter(canvasBaseURL, canvasToken, canvasTokenType, storageAdapter, logger)

	//groups BB adapter
	groupsHost := envLoader.GetAndLogEnvVar(envPrefix+"GROUPS_BB_HOST", true, false)
	groupsBBAdapter := groups.NewGroupsAdapter(groupsHost, internalAPIKey)

	//notifications BB adapter
	app := envLoader.GetAndLogEnvVar(envPrefix+"APP_ID", true, false)
	org := envLoader.GetAndLogEnvVar(envPrefix+"ORG_ID", true, false)
	notificationHost := envLoader.GetAndLogEnvVar(envPrefix+"NOTIFICATIONS_BB_HOST", true, false)
	notificationsBBAdapter := notifications.NewNotificationsAdapter(notificationHost, internalAPIKey, app, org)

	// Service registration
	baseURL := envLoader.GetAndLogEnvVar(envPrefix+"BASE_URL", true, false)
	coreBBBaseURL := envLoader.GetAndLogEnvVar(envPrefix+"CORE_BB_BASE_URL", true, false)

	authService := auth.Service{
		ServiceID:   serviceID,
		ServiceHost: baseURL,
		FirstParty:  true,
		AuthBaseURL: coreBBBaseURL,
	}

	serviceRegLoader, err := auth.NewRemoteServiceRegLoader(&authService, []string{"groups", "notifications"})
	if err != nil {
		logger.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := auth.NewServiceRegManager(&authService, serviceRegLoader, !strings.HasPrefix(baseURL, "http://localhost"))
	if err != nil {
		logger.Fatalf("Error initializing service registration manager: %v", err)
	}

	//core adapter
	cHost, cServiceAccountManager := getCoreBBAdapterValues(logger, serviceID, serviceRegManager, envLoader, envPrefix)
	coreAdapter := corebb.NewCoreAdapter(cHost, cServiceAccountManager, org, app)

	// application
	application := core.NewApplication(Version, Build, storageAdapter, providerAdapter,
		groupsBBAdapter, notificationsBBAdapter, cacheAdapter, coreAdapter, serviceID, logger)
	application.Start()

	// web adapter
	lmsServiceURL := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_URL", true, false)
	webAdapter := driver.NewWebAdapter(lmsServiceURL, port, serviceID, application, serviceRegManager, logger)
	webAdapter.Start()
}

func getCoreBBAdapterValues(logger *logs.Logger, serviceID string, serviceRegManager *auth.ServiceRegManager, envLoader envloader.EnvLoader, envPrefix string) (string, *auth.ServiceAccountManager) {
	host := envLoader.GetAndLogEnvVar(envPrefix+"CORE_BB_CURRENT_HOST", true, false)
	coreBBHost := envLoader.GetAndLogEnvVar(envPrefix+"CORE_BB_CORE_HOST", true, false)

	authService := auth.Service{
		ServiceID:   serviceID,
		ServiceHost: host,
		FirstParty:  true,
		AuthBaseURL: coreBBHost,
	}

	serviceAccountID := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_ACCOUNT_ID", false, false)
	privKeyRaw := envLoader.GetAndLogEnvVar(envPrefix+"PRIV_KEY", false, true)
	privKeyRaw = strings.ReplaceAll(privKeyRaw, "\\n", "\n")
	privKey, err := keys.NewPrivKey(keys.RS256, privKeyRaw)
	if err != nil {
		logger.Errorf("Error parsing priv key: %v", err)
	} else if serviceAccountID == "" {
		logger.Errorf("Missing service account id")
	}
	signatureAuth, err := sigauth.NewSignatureAuth(privKey, serviceRegManager, false, false)
	if err != nil {
		log.Fatalf("Error initializing signature auth: %v", err)
	}

	serviceAccountLoader, err := auth.NewRemoteServiceAccountLoader(&authService, serviceAccountID, signatureAuth)
	if err != nil {
		log.Fatalf("Error initializing remote service account loader: %v", err)
	}

	serviceAccountManager, err := auth.NewServiceAccountManager(&authService, serviceAccountLoader)
	if err != nil {
		log.Fatalf("Error initializing service account manager: %v", err)
	}
	return coreBBHost, serviceAccountManager
}
