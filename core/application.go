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

package core

import (
	"lms/core/model"
	cacheadapter "lms/driven/cache"
	"time"

	"github.com/rokwire/logging-library-go/logs"
)

//Application represents the core application code based on hexagonal architecture
type Application struct {
	version string
	build   string

	Services       Services       //expose to the drivers adapters
	Administration Administration //expose to the drivers adapters

	provider        Provider
	groupsBB        GroupsBB
	notificationsBB NotificationsBB

	storage      Storage
	cacheAdapter *cacheadapter.CacheAdapter

	logger *logs.Logger

	//nudges timer
	dailyNudgesTimer *time.Timer
	timerDone        chan bool
}

// Start starts the core part of the application
func (app *Application) Start() {
	app.storage.SetListener(app)

	go app.setupNudgesTimer()
}

func (app *Application) setupNudgesTimer() {
	app.logger.Info("Setup nudges timer")

	//cancel if active
	if app.dailyNudgesTimer != nil {
		app.logger.Info("setupNudgesTimer -> there is active timer, so cancel it")

		app.timerDone <- true
		app.dailyNudgesTimer.Stop()
	}

	//wait until it is the correct moment from the day
	location, err := time.LoadLocation("America/Chicago")
	if err != nil {
		app.logger.Errorf("Error getting location:%s\n", err.Error())
	}
	now := time.Now().In(location)
	app.logger.Infof("setupNudgesTimer -> now - hours:%d minutes:%d seconds:%d\n", now.Hour(), now.Minute(), now.Second())

	nowSecondsInDay := 60*60*now.Hour() + 60*now.Minute() + now.Second()
	desiredMoment := 39600 //desired moment in the day in seconds, i.e. 11:00 AM

	var durationInSeconds int
	app.logger.Infof("setupNudgesTimer -> nowSecondsInDay:%d desiredMoment:%d\n", nowSecondsInDay, desiredMoment)
	if nowSecondsInDay <= desiredMoment {
		app.logger.Info("setupNudgesTimer -> not processed nudges today, so the first nudges process will be today")
		durationInSeconds = desiredMoment - nowSecondsInDay
	} else {
		app.logger.Info("setupNudgesTimer -> the nudges have already been processed today, so the first nudges process will be tomorrow")
		leftToday := 86400 - nowSecondsInDay
		durationInSeconds = leftToday + desiredMoment // the time which left today + desired moment from tomorrow
	}
	//app.logger.Infof("%d", durationInSeconds)
	//duration := time.Second * time.Duration(20)
	duration := time.Second * time.Duration(durationInSeconds)
	app.logger.Infof("setupNudgesTimer -> first call after %s", duration)

	app.dailyNudgesTimer = time.NewTimer(duration)
	select {
	case <-app.dailyNudgesTimer.C:
		app.logger.Info("setupNudgesTimer -> nudges timer expired")
		app.dailyNudgesTimer = nil

		app.processNudges()
	case <-app.timerDone:
		// timer aborted
		app.logger.Info("setupNudgesTimer -> nudges timer aborted")
		app.dailyNudgesTimer = nil
	}
}

func (app *Application) processNudges() {
	app.logger.Info("processNudges")

	//process nudges
	app.processAllNudges()

	//generate new processing after 24 hours
	duration := time.Hour * 24
	app.logger.Infof("processNudges -> next call after %s", duration)
	app.dailyNudgesTimer = time.NewTimer(duration)
	select {
	case <-app.dailyNudgesTimer.C:
		app.logger.Info("processNudges -> nudges timer expired")
		app.dailyNudgesTimer = nil

		app.processNudges()
	case <-app.timerDone:
		// timer aborted
		app.logger.Info("processNudges -> nudges timer aborted")
		app.dailyNudgesTimer = nil
	}
}

func (app *Application) processAllNudges() {
	app.logger.Info("processAllNudges")

	//1. get all nudges
	nudges, err := app.storage.LoadAllNudges()
	if err != nil {
		app.logger.Errorf("error on processing all nudges - %s", err)
		return
	}

	//2. get all users
	users, err := app.groupsBB.GetUsers()
	if err != nil {
		app.logger.Errorf("error getting all users - %s", err)
		return
	}

	//process every nudge
	for _, nudge := range nudges {
		app.processNudge(nudge, users)
	}
}

func (app *Application) processNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	app.logger.Infof("processNudge - %s", nudge.ID)

	switch nudge.ID {
	case "last_login":
		app.processLastLoginNudge(nudge, allUsers)
	default:
		app.logger.Infof("Not supported nudge - %s", nudge.ID)
	}
}

func (app *Application) processLastLoginNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	app.logger.Infof("processLastLoginNudge - %s", nudge.ID)
}

// NewApplication creates new Application
func NewApplication(version string, build string, storage Storage, provider Provider,
	groupsBB GroupsBB, notificationsBB NotificationsBB,
	cacheadapter *cacheadapter.CacheAdapter, logger *logs.Logger) *Application {
	timerDone := make(chan bool)
	application := Application{
		version:         version,
		build:           build,
		provider:        provider,
		groupsBB:        groupsBB,
		notificationsBB: notificationsBB,
		storage:         storage,
		cacheAdapter:    cacheadapter,
		logger:          logger,
		timerDone:       timerDone}

	// add the drivers ports/interfaces
	application.Services = &servicesImpl{app: &application}
	application.Administration = &administrationImpl{app: &application}

	return &application
}
