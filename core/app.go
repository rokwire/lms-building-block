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

package core

import (
	"lms/core/interfaces"
	cacheadapter "lms/driven/cache"
	"lms/driven/corebb"

	"github.com/rokwire/logging-library-go/v2/logs"
)

// Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Default interfaces.Default
	Client  interfaces.Client
	Admin   interfaces.Admin

	provider        interfaces.Provider
	groupsBB        interfaces.GroupsBB
	notificationsBB interfaces.NotificationsBB

	storage      interfaces.Storage
	cacheAdapter *cacheadapter.CacheAdapter
	core         *corebb.Adapter

	logger *logs.Logger

	//nudges logic
	nudgesLogic nudgesLogic
}

// Start starts the core part of the application
func (app *Application) Start() {
	app.storage.SetListener(app)

	app.nudgesLogic.start()
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage interfaces.Storage, provider interfaces.Provider, groupsBB interfaces.GroupsBB,
	notificationsBB interfaces.NotificationsBB, cacheadapter *cacheadapter.CacheAdapter, coreBB *corebb.Adapter, logger *logs.Logger) *Application {

	timerDone := make(chan bool)
	nudgesLogic := nudgesLogic{
		provider:        provider,
		groupsBB:        groupsBB,
		notificationsBB: notificationsBB,
		storage:         storage,
		logger:          logger,
		timerDone:       timerDone,
		core:            coreBB,
	}

	application := Application{
		version:         version,
		build:           build,
		provider:        provider,
		groupsBB:        groupsBB,
		notificationsBB: notificationsBB,
		storage:         storage,
		cacheAdapter:    cacheadapter,
		logger:          logger,
		nudgesLogic:     nudgesLogic,
		core:            coreBB,
	}

	// add the drivers ports/interfaces
	application.Default = &defaultImpl{app: &application}
	application.Client = &clientImpl{app: &application}
	application.Admin = &adminImpl{app: &application}

	return &application
}

// OnConfigsUpdated is called when the config collection is updates
func (app *Application) OnConfigsUpdated() {
	config, err := app.storage.FindNudgesConfig()
	if err != nil {
		app.logger.Error("error finding nudge configs on configs changed")
	}

	oldConfig := app.nudgesLogic.config
	app.nudgesLogic.config = config
	if config.ProcessTime != nil {
		if oldConfig == nil || oldConfig.ProcessTime == nil || *oldConfig.ProcessTime != *config.ProcessTime {
			go app.nudgesLogic.setupNudgesTimer()
		}
	}
}
