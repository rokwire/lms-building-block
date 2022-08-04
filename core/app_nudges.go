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
	"bytes"
	"fmt"
	"lms/core/model"
	"lms/utils"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/logs"
)

type nudgesLogic struct {
	logger *logs.Logger

	provider        Provider
	groupsBB        GroupsBB
	notificationsBB NotificationsBB

	storage Storage

	//nudges timer
	dailyNudgesTimer *time.Timer
	timerDone        chan bool

	config *model.NudgesConfig
}

func (n nudgesLogic) start() {
	//1. find the nudges config
	nudgesConfig, err := n.storage.FindNudgesConfig()
	if err != nil {
		n.logger.Errorf("error finding nudges config - %s", err)
		return
	}
	//2. check if we have nudges config
	if nudgesConfig == nil {
		n.logger.Error("nudges config is not set")
		return
	}

	//3. here we have a config, so set it
	n.config = nudgesConfig

	//4. setup nudges timer
	go n.setupNudgesTimer()
}

func (n nudgesLogic) setupNudgesTimer() {
	n.logger.Info("Setup nudges timer")

	//cancel if active
	if n.dailyNudgesTimer != nil {
		n.logger.Info("setupNudgesTimer -> there is active timer, so cancel it")

		n.timerDone <- true
		n.dailyNudgesTimer.Stop()
	}

	//wait until it is the correct moment from the day
	location, err := time.LoadLocation("America/Chicago")
	if err != nil {
		n.logger.Errorf("Error getting location:%s\n", err.Error())
	}
	now := time.Now().In(location)
	n.logger.Infof("setupNudgesTimer -> now - hours:%d minutes:%d seconds:%d\n", now.Hour(), now.Minute(), now.Second())

	nowSecondsInDay := 60*60*now.Hour() + 60*now.Minute() + now.Second()
	desiredMoment := 39600 //default desired moment in the day in seconds, i.e. 11:00 AM
	if n.config != nil && n.config.ProcessTime != nil {
		desiredMoment = *n.config.ProcessTime
	}

	var durationInSeconds int
	n.logger.Infof("setupNudgesTimer -> nowSecondsInDay:%d desiredMoment:%d\n", nowSecondsInDay, desiredMoment)
	if nowSecondsInDay <= desiredMoment {
		n.logger.Info("setupNudgesTimer -> not processed nudges today, so the first nudges process will be today")
		durationInSeconds = desiredMoment - nowSecondsInDay
	} else {
		n.logger.Info("setupNudgesTimer -> the nudges have already been processed today, so the first nudges process will be tomorrow")
		leftToday := 86400 - nowSecondsInDay
		durationInSeconds = leftToday + desiredMoment // the time which left today + desired moment from tomorrow
	}
	//app.logger.Infof("%d", durationInSeconds)
	//duration := time.Second * time.Duration(3)
	duration := time.Second * time.Duration(durationInSeconds)
	n.logger.Infof("setupNudgesTimer -> first call after %s", duration)

	n.dailyNudgesTimer = time.NewTimer(duration)
	select {
	case <-n.dailyNudgesTimer.C:
		n.logger.Info("setupNudgesTimer -> nudges timer expired")
		n.dailyNudgesTimer = nil

		n.processNudges()
	case <-n.timerDone:
		// timer aborted
		n.logger.Info("setupNudgesTimer -> nudges timer aborted")
		n.dailyNudgesTimer = nil
	}
}

func (n nudgesLogic) processNudges() {
	n.logger.Info("processNudges")

	//process nudges
	n.processAllNudges()

	//generate new processing after 24 hours
	duration := time.Hour * 24
	n.logger.Infof("processNudges -> next call after %s", duration)
	n.dailyNudgesTimer = time.NewTimer(duration)
	select {
	case <-n.dailyNudgesTimer.C:
		n.logger.Info("processNudges -> nudges timer expired")
		n.dailyNudgesTimer = nil

		n.processNudges()
	case <-n.timerDone:
		// timer aborted
		n.logger.Info("processNudges -> nudges timer aborted")
		n.dailyNudgesTimer = nil
	}
}

//TODO - decide if we need to loop through nudges or through all users(are the users the same for the nudges?)
func (n nudgesLogic) processAllNudges() {
	n.logger.Info("processAllNudges")

	//1. first check if we have a config and the config is set to active
	if n.config == nil {
		n.logger.Error("the config is not set and the nudges will not be processed")
		return
	}
	if !n.config.Active {
		n.logger.Info("the config active is set to false")
		return
	}
	n.logger.Info("the nudges processing is active")

	//2. get all active nudges
	nudges, err := n.storage.LoadActiveNudges()
	if err != nil {
		n.logger.Errorf("error on processing all nudges - %s", err)
		return
	}
	if len(nudges) == 0 {
		n.logger.Info("no active nudges for processing")
	}

	//3. get all users
	groupName := n.getGroupName()
	users, err := n.groupsBB.GetUsers(groupName)
	if err != nil {
		n.logger.Errorf("error getting all users - %s", err)
		return
	}

	//4. process every nudge
	for _, nudge := range nudges {
		n.processNudge(nudge, users)
	}
}

func (n nudgesLogic) getGroupName() string {
	if n.config.Mode == "normal" {
		return n.config.GroupName //normal mode
	}
	return n.config.TestGroupName //test mode
}

func (n nudgesLogic) processNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	n.logger.Infof("processNudge - %s", nudge.ID)

	switch nudge.ID {
	case "last_login":
		n.processLastLoginNudge(nudge, allUsers)
	case "missed_assignment":
		n.processMissedAssignmentNudge(nudge, allUsers)
	case "completed_assignment_early":
		n.processCompletedAssignmentEarlyNudge(nudge, allUsers)
	case "today_calendar_events":
		n.processTodayCalendarEventsNudge(nudge, allUsers)
	default:
		n.logger.Infof("Not supported nudge - %s", nudge.ID)
	}
}

func (n nudgesLogic) prepareNotificationData(deepLink string) map[string]string {
	data := map[string]string{}

	data["click_action"] = "FLUTTER_NOTIFICATION_CLICK"
	data["type"] = "canvas_app_deeplink"
	data["deep_link"] = deepLink

	return data
}

// last_login nudge

func (n nudgesLogic) processLastLoginNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	n.logger.Infof("processLastLoginNudge - %s", nudge.ID)

	for _, user := range allUsers {
		n.processLastLoginNudgePerUser(nudge, user)
	}
}

func (n nudgesLogic) processLastLoginNudgePerUser(nudge model.Nudge, user GroupsBBUser) {
	n.logger.Infof("processLastLoginNudgePerUser - %s", nudge.ID)

	//get last login date
	lastLogin, err := n.provider.GetLastLogin(user.NetID)
	if err != nil {
		n.logger.Errorf("error getting last login for - %s", user.NetID)
	}

	//if last login is not available we do nothing
	if lastLogin == nil {
		n.logger.Debugf("last login is not available for user - %s", user.NetID)
		return
	}

	//determine if needs to send notification
	hours := float64(nudge.Params["hours"].(int32))
	now := time.Now()
	difference := now.Sub(*lastLogin) //difference between now and the last login
	differenceInHours := difference.Hours()
	if differenceInHours <= hours {
		//not reached the max hours, so not send notification
		n.logger.Infof("not reached the max hours, so not send notification - %s", user.NetID)
		return
	}

	//need to send but first check if it has been send before

	//check if has been sent before
	criteriaHash := n.generateLastLoginHash(*lastLogin, hours)
	sentNudge, err := n.storage.FindSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash, n.config.Mode)
	if err != nil {
		//not reached the max hours, so not send notification
		n.logger.Errorf("error checking if sent nudge exists - %s - %s", nudge.ID, user.NetID)
		return
	}
	if sentNudge != nil {
		n.logger.Infof("this has been already sent - %s - %s", nudge.ID, user.NetID)
		return
	}

	//it has not been sent, so sent it
	n.sendLastLoginNudgeForUser(nudge, user, *lastLogin, hours)
}

func (n nudgesLogic) sendLastLoginNudgeForUser(nudge model.Nudge, user GroupsBBUser,
	lastLogin time.Time, hours float64) {
	n.logger.Infof("sendLastLoginNudgeForUser - %s - %s", nudge.ID, user.UserID)

	//send push notification
	recipient := Recipient{UserID: user.UserID, Name: ""}
	data := n.prepareNotificationData(nudge.DeepLink)
	err := n.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, nudge.Body, data)
	if err != nil {
		n.logger.Debugf("error sending notification for %s - %s", user.UserID, err)
		return
	}

	//insert sent nudge
	criteriaHash := n.generateLastLoginHash(lastLogin, hours)
	sentNudge := n.createSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash, n.config.Mode)
	err = n.storage.InsertSentNudge(sentNudge)
	if err != nil {
		n.logger.Errorf("error saving sent nudge for %s - %s", user.UserID, err)
		return
	}
}

func (n nudgesLogic) generateLastLoginHash(lastLogin time.Time, hours float64) uint32 {
	lastLoginComponent := fmt.Sprintf("%d", lastLogin.Unix())
	hoursComponent := fmt.Sprintf("%f", hours)
	component := fmt.Sprintf("%s+%s", lastLoginComponent, hoursComponent)
	hash := utils.Hash(component)
	return hash
}

func (n nudgesLogic) createSentNudge(nudgeID string, userID string, netID string, criteriaHash uint32, mode string) model.SentNudge {
	id, _ := uuid.NewUUID()
	return model.SentNudge{ID: id.String(), NudgeID: nudgeID, UserID: userID,
		NetID: netID, CriteriaHash: criteriaHash, DateSent: time.Now(), Mode: mode}
}

// end last_login nudge

// missed_assignemnt nudge

func (n nudgesLogic) processMissedAssignmentNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	n.logger.Infof("processMissedAssignmentNudge - %s", nudge.ID)

	for _, user := range allUsers {
		n.processMissedAssignmentNudgePerUser(nudge, user)
	}
}

func (n nudgesLogic) processMissedAssignmentNudgePerUser(nudge model.Nudge, user GroupsBBUser) {
	n.logger.Infof("processMissedAssignmentNudgePerUser - %s", nudge.ID)

	//get missed assignments
	missedAssignments, err := n.provider.GetMissedAssignments(user.NetID)
	if err != nil {
		n.logger.Errorf("error getting missed assignments for - %s", user.NetID)
	}
	if len(missedAssignments) == 0 {
		//no missed assignments
		n.logger.Infof("no missed assignments, so not send notifications - %s", user.NetID)
		return
	}

	//determine for which of the assignments we need to send notifications
	hours := float64(nudge.Params["hours"].(int32))
	now := time.Now()
	missedAssignments, err = n.findMissedAssignments(hours, now, missedAssignments)
	if err != nil {
		n.logger.Errorf("error finding missed assignments for - %s", user.NetID)
	}
	if len(missedAssignments) == 0 {
		//no missed assignments
		n.logger.Infof("no missed assignments after checking due date, so not send notifications - %s", user.NetID)
		return
	}

	//here we have the assignments we need to send notifications for

	//process the missed assignments
	for _, assignment := range missedAssignments {
		n.processMissedAssignment(nudge, user, assignment, hours)
	}
}

func (n nudgesLogic) processMissedAssignment(nudge model.Nudge, user GroupsBBUser, assignment model.Assignment, hours float64) {
	n.logger.Infof("processMissedAssignment - %s - %s - %s", nudge.ID, user.NetID, assignment.Name)

	//need to send but first check if it has been send before

	//check if has been sent before
	criteriaHash := n.generateMissedAssignmentHash(assignment.ID, hours)
	sentNudge, err := n.storage.FindSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash, n.config.Mode)
	if err != nil {
		//not reached the max hours, so not send notification
		n.logger.Errorf("error checking if sent nudge exists for missed assignment - %s - %s", nudge.ID, user.NetID)
		return
	}
	if sentNudge != nil {
		n.logger.Infof("this has been already sent - %s - %s", nudge.ID, user.NetID)
		return
	}

	//it has not been sent, so sent it
	n.sendMissedAssignmentNudgeForUser(nudge, user, assignment, hours)
}

func (n nudgesLogic) sendMissedAssignmentNudgeForUser(nudge model.Nudge, user GroupsBBUser,
	assignment model.Assignment, hours float64) {
	n.logger.Infof("sendMissedAssignmentNudgeForUser - %s - %s", nudge.ID, user.UserID)

	//send push notification
	recipient := Recipient{UserID: user.UserID, Name: ""}
	body := fmt.Sprintf(nudge.Body, assignment.Name)
	deepLink := fmt.Sprintf(nudge.DeepLink, assignment.CourseID, assignment.ID)
	data := n.prepareNotificationData(deepLink)
	err := n.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, body, data)
	if err != nil {
		n.logger.Debugf("error sending notification for %s - %s", user.UserID, err)
		return
	}

	//insert sent nudge
	criteriaHash := n.generateMissedAssignmentHash(assignment.ID, hours)
	sentNudge := n.createSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash, n.config.Mode)
	err = n.storage.InsertSentNudge(sentNudge)
	if err != nil {
		n.logger.Errorf("error saving sent missed assignment nudge for %s - %s", user.UserID, err)
		return
	}
}

func (n nudgesLogic) generateMissedAssignmentHash(assignemntID int, hours float64) uint32 {
	assignmentIDComponent := fmt.Sprintf("%d", assignemntID)
	hoursComponent := fmt.Sprintf("%f", hours)
	component := fmt.Sprintf("%s+%s", assignmentIDComponent, hoursComponent)
	hash := utils.Hash(component)
	return hash
}

func (n nudgesLogic) findMissedAssignments(hours float64, now time.Time, assignments []model.Assignment) ([]model.Assignment, error) {
	n.logger.Info("findMissedAssignments")

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

func (n nudgesLogic) processCompletedAssignmentEarlyNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	n.logger.Infof("processCompletedAssignmentEarlyNudge - %s", nudge.ID)

	for _, user := range allUsers {
		n.processCompletedAssignmentEarlyNudgePerUser(nudge, user)
	}
}

func (n nudgesLogic) processCompletedAssignmentEarlyNudgePerUser(nudge model.Nudge, user GroupsBBUser) {
	n.logger.Infof("processCompletedAssignmentEarlyNudgePerUser - %s", nudge.ID)

	//get completed assignments
	ecAssignments, err := n.provider.GetCompletedAssignments(user.NetID)
	if err != nil {
		n.logger.Errorf("error getting early completed assignments for - %s", user.NetID)
	}
	if len(ecAssignments) == 0 {
		//no early completed assignments
		n.logger.Infof("no early completed assignments, so not send notifications - %s", user.NetID)
		return
	}

	//determine for which of the submissions we need to send notifications
	hours := float64(nudge.Params["hours"].(int32))
	now := time.Now()
	ecAssignments, err = n.findCompletedEarlyAssignments(hours, now, ecAssignments)
	if err != nil {
		n.logger.Errorf("error finding early completed assignments for - %s", user.NetID)
	}
	if len(ecAssignments) == 0 {
		//no early completed assignments
		n.logger.Infof("no early completed assignments after checking submitted date, so not send notifications - %s", user.NetID)
		return
	}

	//here we have the assignments we need to send notifications for

	//process the early completed assignments
	for _, assignment := range ecAssignments {
		n.processCompletedAssignmentEarly(nudge, user, assignment, hours)
	}
}

func (n nudgesLogic) findCompletedEarlyAssignments(hours float64, now time.Time, assignments []model.Assignment) ([]model.Assignment, error) {
	n.logger.Info("findCompletedEarlyAssignments")

	hoursInSecs := hours * 60 * 60
	resultList := []model.Assignment{}
	for _, assignment := range assignments {

		if assignment.DueAt == nil || assignment.Submission == nil || assignment.Submission.SubmittedAt == nil {
			continue
		}
		dueAt := assignment.DueAt.Unix()
		submittedAt := assignment.Submission.SubmittedAt.Unix()
		//check if submitted is x hours before due
		difference := dueAt - submittedAt
		if difference > int64(hoursInSecs) {
			resultList = append(resultList, assignment)
		}
	}

	return resultList, nil
}

func (n nudgesLogic) processCompletedAssignmentEarly(nudge model.Nudge, user GroupsBBUser, assignment model.Assignment, hours float64) {
	n.logger.Infof("processCompletedAssignmentEarly - %s - %s - %s", nudge.ID, user.NetID, assignment.Name)

	//need to send but first check if it has been send before

	//check if has been sent before
	criteriaHash := n.generateEarlyCompletedAssignmentHash(assignment.ID, assignment.Submission.ID, *assignment.Submission.SubmittedAt)
	sentNudge, err := n.storage.FindSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash, n.config.Mode)
	if err != nil {
		//not reached the max hours, so not send notification
		n.logger.Errorf("error checking if sent nudge exists for missed assignment - %s - %s", nudge.ID, user.NetID)
		return
	}
	if sentNudge != nil {
		n.logger.Infof("this has been already sent - %s - %s", nudge.ID, user.NetID)
		return
	}

	//it has not been sent, so sent it
	n.sendEarlyCompletedAssignmentNudgeForUser(nudge, user, assignment, hours)
}

func (n nudgesLogic) generateEarlyCompletedAssignmentHash(assignmentID int, submissionID int, submittedAt time.Time) uint32 {
	assignmentIDComponent := fmt.Sprintf("%d", assignmentID)
	submissionIDComponent := fmt.Sprintf("%d", submissionID)
	submittedAtComponent := fmt.Sprintf("%d", submittedAt.Unix())
	component := fmt.Sprintf("%s+%s+%s", assignmentIDComponent, submissionIDComponent, submittedAtComponent)
	hash := utils.Hash(component)
	return hash
}

func (n nudgesLogic) sendEarlyCompletedAssignmentNudgeForUser(nudge model.Nudge, user GroupsBBUser,
	assignment model.Assignment, hours float64) {
	n.logger.Infof("sendEarlyCompletedAssignmentNudgeForUser - %s - %s", nudge.ID, user.UserID)

	//send push notification
	recipient := Recipient{UserID: user.UserID, Name: ""}
	deepLink := fmt.Sprintf(nudge.DeepLink, assignment.CourseID, assignment.ID)
	data := n.prepareNotificationData(deepLink)
	err := n.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, nudge.Body, data)
	if err != nil {
		n.logger.Debugf("error sending notification for %s - %s", user.UserID, err)
		return
	}

	//insert sent nudge
	criteriaHash := n.generateEarlyCompletedAssignmentHash(assignment.ID, assignment.Submission.ID, *assignment.Submission.SubmittedAt)
	sentNudge := n.createSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash, n.config.Mode)
	err = n.storage.InsertSentNudge(sentNudge)
	if err != nil {
		n.logger.Errorf("error saving sent early completed assignment nudge for %s - %s", user.UserID, err)
		return
	}
}

// end completed_assignment_early nudge

// calendar_event nudge
func (n nudgesLogic) processTodayCalendarEventsNudge(nudge model.Nudge, allUsers []GroupsBBUser) {
	n.logger.Infof("processTodayCalendarEventsNudge - %s", nudge.ID)

	for _, user := range allUsers {
		n.processTodayCalendarEventsNudgePerUser(nudge, user)
	}
}

func (n nudgesLogic) processTodayCalendarEventsNudgePerUser(nudge model.Nudge, user GroupsBBUser) {
	n.logger.Infof("processTodayCalendarEventsNudgePerUser - %s", nudge.ID)

	//get calendar events
	startDate, endDate := n.prepareTodayCalendarEventsDates()
	calendarEvents, err := n.provider.GetCalendarEvents(user.NetID, startDate, endDate)
	if err != nil {
		n.logger.Errorf("error getting calendar events for - %s", user.NetID)
	}
	if len(calendarEvents) == 0 {
		//no calendar events
		n.logger.Infof("no calendar events, so not send notifications - %s", user.NetID)
		return
	}

	//we have calendar events, so process them
	n.processCalendarEvents(nudge, user, calendarEvents)
}

func (n nudgesLogic) prepareTodayCalendarEventsDates() (time.Time, time.Time) {
	now := time.Now()

	start := time.Date(now.Year(), time.Month(now.Month()), now.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(now.Year(), time.Month(now.Month()), now.Day(), 23, 59, 59, 999, time.UTC)
	return start, end
}

func (n nudgesLogic) processCalendarEvents(nudge model.Nudge, user GroupsBBUser, events []model.CalendarEvent) {
	n.logger.Infof("processCalendarEvents - %s - %s - %d", nudge.ID, user.NetID, len(events))

	//find unset events
	unsentEvents, err := n.findUnsentEvents(nudge, user, events)
	if err != nil {
		n.logger.Errorf("error finding unset events - %s - %s", nudge.ID, user.NetID)
		return
	}
	if len(unsentEvents) == 0 {
		n.logger.Infof("unsent events is 0 - %s - %s", nudge.ID, user.NetID)
		return
	}

	//we have unsent events, so process them for sending
	n.sendCalendareEventNudgeForUsers(nudge, user, unsentEvents)
}

func (n nudgesLogic) findUnsentEvents(nudge model.Nudge, user GroupsBBUser, events []model.CalendarEvent) ([]model.CalendarEvent, error) {
	//get hashes for all events
	hashes := []uint32{}
	for _, event := range events {
		hash := n.generateCalendarEventHash(event.ID)
		hashes = append(hashes, hash)
	}

	//find the sent nudges
	sentNudges, err := n.storage.FindSentNudges(&nudge.ID, &user.UserID, &user.NetID, &hashes, &n.config.Mode)
	if err != nil {
		n.logger.Errorf("error checking if sent nudge exists for calendar events - %s - %s", nudge.ID, user.NetID)
		return nil, err
	}

	//prepare the result
	result := []model.CalendarEvent{}
	for _, event := range events {
		isSent := n.isEventSent(event, sentNudges)
		if !isSent {
			result = append(result, event)
		}
	}

	return result, nil
}

func (n nudgesLogic) isEventSent(event model.CalendarEvent, sentEvents []model.SentNudge) bool {
	cHash := n.generateCalendarEventHash(event.ID)
	for _, sent := range sentEvents {
		if cHash == sent.CriteriaHash {
			return true
		}
	}
	return false
}

func (n nudgesLogic) sendCalendareEventNudgeForUsers(nudge model.Nudge, user GroupsBBUser,
	events []model.CalendarEvent) {
	n.logger.Infof("sendCalendareEventNudgeForUsers - %s - %s", nudge.ID, user.UserID)

	//send push notification
	recipient := Recipient{UserID: user.UserID, Name: ""}
	body := n.prepareCalendarEventNudgeBody(nudge, events)
	data := n.prepareNotificationData(nudge.DeepLink)
	err := n.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, body, data)
	if err != nil {
		n.logger.Debugf("error sending notification for %s - %s", user.UserID, err)
		return
	}

	//insert sent nudge
	sentNudges := make([]model.SentNudge, len(events))
	for i, event := range events {
		criteriaHash := n.generateCalendarEventHash(event.ID)
		sentNudge := n.createSentNudge(nudge.ID, user.UserID, user.NetID, criteriaHash, n.config.Mode)
		sentNudges[i] = sentNudge
	}
	err = n.storage.InsertSentNudges(sentNudges)
	if err != nil {
		n.logger.Errorf("error saving sent calendar events nudge for %s - %s", user.UserID, err)
		return
	}
}

func (n nudgesLogic) prepareCalendarEventNudgeBody(nudge model.Nudge, events []model.CalendarEvent) string {
	var eventsNames bytes.Buffer
	for _, event := range events {
		eventsNames.WriteString(event.Title)
		eventsNames.WriteString("\n")
	}
	return fmt.Sprintf(nudge.Body, eventsNames.String())
}

func (n nudgesLogic) generateCalendarEventHash(eventID int) uint32 {
	eventIDComponent := fmt.Sprintf("%d", eventID)
	hash := utils.Hash(eventIDComponent)
	return hash
}

// end calendar_event nudge
