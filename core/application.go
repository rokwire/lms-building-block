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
	"fmt"
	"lms/core/model"
	cacheadapter "lms/driven/cache"
	"lms/utils"
	"time"

	"github.com/google/uuid"
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
	/*
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
		} */
	//app.logger.Infof("%d", durationInSeconds)
	duration := time.Second * time.Duration(3)
	//åduration := time.Second * time.Duration(durationInSeconds)
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

//TODO - decide if we need to loop through nudges or through all users(are the users the same for the nudges?)
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
	case "missed_assignment":
		app.processMissedAssignmentNudge(nudge, allUsers)
	default:
		app.logger.Infof("Not supported nudge - %s", nudge.ID)
	}
}

// last_login nudge

func (app *Application) processLastLoginNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	app.logger.Infof("processLastLoginNudge - %s", nudge.ID)

	for _, user := range allUsers {
		app.processLastLoginNudgePerUser(nudge, user)
	}
}

func (app *Application) processLastLoginNudgePerUser(nudge model.Nudge, user GroupsBBUser) {
	app.logger.Infof("processLastLoginNudgePerUser - %s", nudge.ID)

	//get last login date
	lastLogin, err := app.provider.GetLastLogin(user.NetID)
	if err != nil {
		app.logger.Errorf("error getting last login for - %s", user.NetID)
	}

	//if last login is not available we do nothing
	if lastLogin == nil {
		app.logger.Debugf("last login is not available for user - %s", user.NetID)
		return
	}

	//determine if needs to send notification
	hours := float64(nudge.Params["hours"].(int32))
	now := time.Now()
	difference := now.Sub(*lastLogin) //difference between now and the last login
	differenceInHours := difference.Hours()
	if differenceInHours <= hours {
		//not reached the max hours, so not send notification
		app.logger.Infof("not reached the max hours, so not send notification - %s", user.NetID)
		return
	}

	//need to send but first check if it has been send before

	//check if has been sent before
	criteriaHash := app.generateLastLoginHash(*lastLogin, hours)
	sentNudge, err := app.storage.FindSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash)
	if err != nil {
		//not reached the max hours, so not send notification
		app.logger.Errorf("error checking if sent nudge exists - %s - %s", nudge.ID, user.NetID)
		return
	}
	if sentNudge != nil {
		app.logger.Infof("this has been already sent - %s - %s", nudge.ID, user.NetID)
		return
	}

	//it has not been sent, so sent it
	app.sendLastLoginNudgeForUser(nudge, user, *lastLogin, hours)
}

func (app *Application) sendLastLoginNudgeForUser(nudge model.Nudge, user GroupsBBUser,
	lastLogin time.Time, hours float64) {
	app.logger.Infof("sendLastLoginNudgeForUser - %s - %s", nudge.ID, user.UserID)

	//send push notification
	recipient := Recipient{UserID: user.UserID, Name: ""}
	err := app.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, nudge.Body)
	if err != nil {
		app.logger.Debugf("error sending notification for %s - %s", user.UserID, err)
		return
	}

	//insert sent nudge
	criteriaHash := app.generateLastLoginHash(lastLogin, hours)
	sentNudge := app.createSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash)
	err = app.storage.InsertSentNudge(sentNudge)
	if err != nil {
		app.logger.Errorf("error saving sent nudge for %s - %s", user.UserID, err)
		return
	}
}

func (app *Application) generateLastLoginHash(lastLogin time.Time, hours float64) uint32 {
	lastLoginComponent := fmt.Sprintf("%d", lastLogin.Unix())
	hoursComponent := fmt.Sprintf("%f", hours)
	component := fmt.Sprintf("%s+%s", lastLoginComponent, hoursComponent)
	hash := utils.Hash(component)
	return hash
}

func (app *Application) createSentNudge(nudgeID string, userID string, netID string, criteriaHash uint32) model.SentNudge {
	id, _ := uuid.NewUUID()
	return model.SentNudge{ID: id.String(), NudgeID: nudgeID, UserID: userID,
		NetID: netID, CriteriaHash: criteriaHash, DateSent: time.Now()}
}

// end last_login nudge

// missed_assignemnt nudge

func (app *Application) processMissedAssignmentNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	app.logger.Infof("processMissedAssignmentNudge - %s", nudge.ID)

	for _, user := range allUsers {
		app.processMissedAssignmentNudgePerUser(nudge, user)
	}
}

func (app *Application) processMissedAssignmentNudgePerUser(nudge model.Nudge, user GroupsBBUser) {
	app.logger.Infof("processMissedAssignmentNudgePerUser - %s", nudge.ID)

	//TODO
}

// end missed_assignemnt nudge

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
