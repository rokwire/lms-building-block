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
	"lms/utils"
	"time"

	"github.com/google/uuid"
)

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
	//duration := time.Second * time.Duration(durationInSeconds)
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
	case "completed_assignment_early":
		app.processCompletedAssignmentEarlyNudge(nudge, allUsers)
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

	//get missed assignments
	missedAssignments, err := app.provider.GetMissedAssignments(user.NetID)
	if err != nil {
		app.logger.Errorf("error getting missed assignments for - %s", user.NetID)
	}
	if len(missedAssignments) == 0 {
		//no missed assignments
		app.logger.Infof("no missed assignments, so not send notifications - %s", user.NetID)
		return
	}

	//determine for which of the assignments we need to send notifications
	hours := float64(nudge.Params["hours"].(int32))
	now := time.Now()
	missedAssignments, err = app.findMissedAssignments(hours, now, missedAssignments)
	if err != nil {
		app.logger.Errorf("error finding missed assignments for - %s", user.NetID)
	}
	if len(missedAssignments) == 0 {
		//no missed assignments
		app.logger.Infof("no missed assignments after checking due date, so not send notifications - %s", user.NetID)
		return
	}

	//here we have the assignments we need to send notifications for

	//process the missed assignments
	for _, assignment := range missedAssignments {
		app.processMissedAssignment(nudge, user, assignment, hours)
	}
}

func (app *Application) processMissedAssignment(nudge model.Nudge, user GroupsBBUser, assignment model.Assignment, hours float64) {
	app.logger.Infof("processMissedAssignment - %s - %s - %s", nudge.ID, user.NetID, assignment.Name)

	//need to send but first check if it has been send before

	//check if has been sent before
	criteriaHash := app.generateMissedAssignmentHash(assignment.ID, hours)
	sentNudge, err := app.storage.FindSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash)
	if err != nil {
		//not reached the max hours, so not send notification
		app.logger.Errorf("error checking if sent nudge exists for missed assignment - %s - %s", nudge.ID, user.NetID)
		return
	}
	if sentNudge != nil {
		app.logger.Infof("this has been already sent - %s - %s", nudge.ID, user.NetID)
		return
	}

	//it has not been sent, so sent it
	app.sendMissedAssignmentNudgeForUser(nudge, user, assignment, hours)
}

func (app *Application) sendMissedAssignmentNudgeForUser(nudge model.Nudge, user GroupsBBUser,
	assignment model.Assignment, hours float64) {
	app.logger.Infof("sendMissedAssignmentNudgeForUser - %s - %s", nudge.ID, user.UserID)

	//send push notification
	recipient := Recipient{UserID: user.UserID, Name: ""}
	body := fmt.Sprintf(nudge.Body, assignment.Name)
	err := app.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, body)
	if err != nil {
		app.logger.Debugf("error sending notification for %s - %s", user.UserID, err)
		return
	}

	//insert sent nudge
	criteriaHash := app.generateMissedAssignmentHash(assignment.ID, hours)
	sentNudge := app.createSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash)
	err = app.storage.InsertSentNudge(sentNudge)
	if err != nil {
		app.logger.Errorf("error saving sent missed assignment nudge for %s - %s", user.UserID, err)
		return
	}
}

func (app *Application) generateMissedAssignmentHash(assignemntID int, hours float64) uint32 {
	assignmentIDComponent := fmt.Sprintf("%d", assignemntID)
	hoursComponent := fmt.Sprintf("%f", hours)
	component := fmt.Sprintf("%s+%s", assignmentIDComponent, hoursComponent)
	hash := utils.Hash(component)
	return hash
}

func (app *Application) findMissedAssignments(hours float64, now time.Time, assignments []model.Assignment) ([]model.Assignment, error) {
	app.logger.Info("findMissedAssignments")

	resultList := []model.Assignment{}
	for _, assignment := range assignments {
		if assignment.DueAt == nil {
			continue
		}

		difference := now.Sub(*assignment.DueAt) //difference between now and the due at date
		differenceInHours := difference.Hours()
		if differenceInHours > hours {
			resultList = append(resultList, assignment)
		}
	}

	return resultList, nil
}

// end missed_assignemnt nudge

// completed_assignment_early nudge

func (app *Application) processCompletedAssignmentEarlyNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	app.logger.Infof("processCompletedAssignmentEarlyNudge - %s", nudge.ID)

	for _, user := range allUsers {
		app.processCompletedAssignmentEarlyNudgePerUser(nudge, user)
	}
}

func (app *Application) processCompletedAssignmentEarlyNudgePerUser(nudge model.Nudge, user GroupsBBUser) {
	app.logger.Infof("processCompletedAssignmentEarlyNudgePerUser - %s", nudge.ID)

	//get completed assignments
	ecAssignments, err := app.provider.GetCompletedAssignments(user.NetID)
	if err != nil {
		app.logger.Errorf("error getting early completed assignments for - %s", user.NetID)
	}
	if len(ecAssignments) == 0 {
		//no early completed assignments
		app.logger.Infof("no early completed assignments, so not send notifications - %s", user.NetID)
		return
	}
	/*
		//determine for which of the assignments we need to send notifications
		hours := float64(nudge.Params["hours"].(int32))
		now := time.Now()
		missedAssignments, err = app.findMissedAssignments(hours, now, missedAssignments)
		if err != nil {
			app.logger.Errorf("error finding missed assignments for - %s", user.NetID)
		}
		if len(missedAssignments) == 0 {
			//no missed assignments
			app.logger.Infof("no missed assignments after checking due date, so not send notifications - %s", user.NetID)
			return
		}

		//here we have the assignments we need to send notifications for

		//process the missed assignments
		for _, assignment := range missedAssignments {
			app.processMissedAssignment(nudge, user, assignment, hours)
		} */
}

// end completed_assignment_early nudge
