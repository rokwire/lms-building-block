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
	"log"
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
	/*
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
		//app.logger.Infof("%d", durationInSeconds)*/
	duration := time.Second * time.Duration(3)
	//duration := time.Second * time.Duration(durationInSeconds)
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

func (n nudgesLogic) processAllNudges() {
	n.logger.Info("START nudges processing")

	// first check if we have a config and the config is set to active
	if n.config == nil {
		n.logger.Error("the config is not set and the nudges will not be processed")
		return
	}
	if !n.config.Active {
		n.logger.Info("the config active is set to false")
		return
	}

	// check if we already have a running nudges process
	hasProcess, err := n.hasRunningProcess()
	if err != nil {
		n.logger.Errorf("error on checking if has a running process - %s", err)
		return
	}
	if *hasProcess {
		n.logger.Info("cannot start as already has a running process")
		return
	}

	// check if we have active nudges
	nudges, err := n.storage.LoadActiveNudges()
	if err != nil {
		n.logger.Errorf("error on processing all nudges - %s", err)
		return
	}
	if len(nudges) == 0 {
		n.logger.Info("no active nudges for processing")
	}

	n.logger.Info("we are ready to start a process")

	//start process
	processID, err := n.startProcess()
	if err != nil {
		n.logger.Errorf("error on starting a process - %s", err)
		return
	}

	// process phase 1
	blocksSize, err := n.processPhase1(*processID)
	if err != nil {
		n.logger.Errorf("error on processing phase 1, so stopping the process and mark it as failed - %s", err)
		n.completeProcessFailed(*processID, err.Error())
		return
	}

	// process phase 2
	err = n.processPhase2(*processID, *blocksSize, nudges)
	if err != nil {
		n.logger.Errorf("error on processing phase 2, so stopping the process and mark it as failed - %s", err)
		n.completeProcessFailed(*processID, err.Error())
		return
	}

	//end process
	err = n.completeProcessSuccess(*processID)
	if err != nil {
		n.logger.Errorf("error on completing a process - %s", err)
		return
	}
}

func (n nudgesLogic) hasRunningProcess() (*bool, error) {
	//check count
	count, err := n.storage.CountNudgesProcesses("processing")
	if err != nil {
		return nil, err
	}

	has := false
	if *count > 0 {
		has = true
	}
	return &has, nil
}

func (n nudgesLogic) startProcess() (*string, error) {
	//create object
	uuidID, _ := uuid.NewUUID()
	id := uuidID.String()
	mode := n.config.Mode
	createdAt := time.Now()
	status := "processing"
	blocks := []model.Block{} //empty
	process := model.NudgesProcess{ID: id, Mode: mode, CreatedAt: createdAt, Status: status, Blocks: blocks}

	//store it
	err := n.storage.InsertNudgesProcess(process)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (n nudgesLogic) completeProcessSuccess(processID string) error {
	completedAt := time.Now()
	status := "success"
	err := n.storage.UpdateNudgesProcess(processID, completedAt, status, nil)
	if err != nil {
		return err
	}
	return nil
}

func (n nudgesLogic) completeProcessFailed(processID string, errStr string) error {
	completedAt := time.Now()
	status := "failed"
	err := n.storage.UpdateNudgesProcess(processID, completedAt, status, &errStr)
	if err != nil {
		return err
	}
	return nil
}

// as a result of phase 1 we have into our service a cached provider data for:
// all users
// users courses
// courses assignments
// - with acceptable sync date
func (n nudgesLogic) processPhase1(processID string) (*int, error) {
	n.logger.Info("START Phase1")

	/// get the users from the groups bb adapter on blocks
	groupName := n.getGroupName()
	offset := 0
	limit := n.config.BlockSize
	currentBlock := 0
	for {
		//get the block users from the groups bb adapter
		users, err := n.groupsBB.GetUsers(groupName, offset, limit)
		if err != nil {
			n.logger.Errorf("error getting all users - %s", err)
			return nil, err
		}

		if len(users) == 0 {
			//finish if there is no more users
			break
		}

		//process the block
		err = n.processPhase1Block(processID, currentBlock, users)
		if err != nil {
			n.logger.Errorf("error processing block - %s", err)
			return nil, err
		}

		//move offset
		offset += limit
		currentBlock++
	}

	n.logger.Info("END Phase1")
	return &currentBlock, nil
}

func (n nudgesLogic) processPhase1Block(processID string, curentBlock int, users []GroupsBBUser) error {
	//add the block to the process
	block := n.createBlock(curentBlock, users)
	err := n.storage.AddBlockToNudgesProcess(processID, block)
	if err != nil {
		n.logger.Errorf("error on adding block %d to process %s - %s", block.Number, processID, err)
		return err
	}

	//prepare the provider data for the block
	err = n.prepareProviderData(users)
	if err != nil {
		n.logger.Errorf("error on preparing the provider data - %s", err)
		return err
	}

	return nil
}

func (n nudgesLogic) createBlock(curentBlock int, users []GroupsBBUser) model.Block {
	items := make([]model.BlockItem, len(users))
	for i, user := range users {
		blockItem := model.BlockItem{NetID: user.NetID, UserID: user.UserID}
		items[i] = blockItem
	}
	return model.Block{Number: curentBlock, Items: items}
}

// phase2 operates over the data prepared in phase1 and apply the nudges for every user
func (n nudgesLogic) processPhase2(processID string, blocksSize int, nudges []model.Nudge) error {
	n.logger.Info("START Phase2")

	for blockNumber := 0; blockNumber < blocksSize; blockNumber++ {
		n.logger.Infof("block:%d", blockNumber)

		err := n.processPhase2Block(processID, blockNumber, nudges)
		if err != nil {
			n.logger.Errorf("error on process block %d - %s", blockNumber, err)
			return err
		}
	}

	n.logger.Info("END Phase2")
	return nil
}

func (n nudgesLogic) processPhase2Block(processID string, blockNumber int, nudges []model.Nudge) error {
	// load block data
	cachedData, err := n.getBlockData(processID, blockNumber)
	if err != nil {
		n.logger.Errorf("error on getting block data %s - %d - %s", processID, blockNumber, err)
		return err
	}

	// process every user
	for _, providerUser := range cachedData {
		err := n.processUser(providerUser, nudges)
		if err != nil {
			n.logger.Errorf("process provider user %s - %s", providerUser.NetID, err)
			return err
		}
	}
	return nil
}

func (n nudgesLogic) getBlockData(processID string, blockNumber int) ([]ProviderUser, error) {
	//get data
	block, err := n.storage.GetBlockFromNudgesProcess(processID, blockNumber)
	if err != nil {
		n.logger.Errorf("error on getting block data from the storage %s - %d - %s", processID, blockNumber, err)
		return nil, err
	}
	items := block.Items
	if len(items) == 0 {
		return []ProviderUser{}, nil
	}

	//get the cached data from the block
	usersIDs := make([]string, len(items))
	for i, item := range items {
		usersIDs[i] = item.NetID
	}
	cachedData, err := n.provider.FindCachedData(usersIDs)
	if err != nil {
		n.logger.Errorf("error on getting cached data %s - %d - %s", processID, blockNumber, err)
		return nil, err
	}

	return cachedData, nil
}

func (n nudgesLogic) prepareProviderData(users []GroupsBBUser) error {
	n.logger.Info("\tprepareProviderData")

	//get the net ids from the users
	usersIDs := n.prepareUsers(users)
	if len(usersIDs) == 0 {
		n.logger.Info("\t\tno users for processing")
		return nil
	}

	//process caching
	err := n.provider.CacheCommonData(usersIDs)
	if err != nil {
		n.logger.Errorf("error caching common data- %s", err)
		return err
	}

	return nil
}

// returns the net ids for all user who have it
func (n nudgesLogic) prepareUsers(users []GroupsBBUser) map[string]string {
	result := map[string]string{}
	for _, user := range users {
		if len(user.NetID) > 0 {
			result[user.NetID] = user.UserID
		}
	}
	return result
}

func (n nudgesLogic) getGroupName() string {
	if n.config.Mode == "normal" {
		return n.config.GroupName //normal mode
	}
	return n.config.TestGroupName //test mode
}

func (n nudgesLogic) processUser(user ProviderUser, nudges []model.Nudge) error {
	n.logger.Infof("\tprocess %s", user.NetID)

	for _, nudge := range nudges {
		processedUser, err := n.processNudge(nudge, user)
		if err != nil {
			return err
		}

		//in some nudges processment we could load a new data in the user, so pass all this object to the next nudge
		user = *processedUser
	}
	return nil
}

func (n nudgesLogic) processNudge(nudge model.Nudge, user ProviderUser) (*ProviderUser, error) {
	n.logger.Infof("\t\tprocessNudge - %s - %s", user.NetID, nudge.ID)

	switch nudge.ID {
	case "last_login":
		err := n.processLastLoginNudgePerUser(nudge, user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	case "missed_assignment":
		processedUser, err := n.processMissedAssignmentNudgePerUser(nudge, user)
		if err != nil {
			return nil, err
		}
		return processedUser, nil
	case "completed_assignment_early":
		processedUser, err := n.processCompletedAssignmentEarlyNudgePerUser(nudge, user)
		if err != nil {
			return nil, err
		}
		return processedUser, nil
	/*case "today_calendar_events":
	n.processTodayCalendarEventsNudgePerUser(nudge, user) */
	default:
		n.logger.Infof("\t\tnot supported nudge - %s", nudge.ID)
		return &user, nil
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

func (n nudgesLogic) processLastLoginNudgePerUser(nudge model.Nudge, user ProviderUser) error {
	n.logger.Infof("\t\t\tprocessLastLoginNudgePerUser - %s", nudge.ID)

	var err error

	//get last login date from the cache data
	lastLogin := user.User.LastLogin
	//if last login is not available we do nothing
	if lastLogin == nil {
		n.logger.Debugf("\t\t\t\tlast login is not available for user - %s", user.NetID)
		return nil
	}

	//prepare another needed data
	hours := nudge.Params["hours"].(float64)
	now := time.Now()

	//determine if needs to send notification - using the cached data
	needsSend := n.lastLoginNeedsToSend(hours, now, *lastLogin)
	if !needsSend {
		//not reached the max hours, so not send notification
		n.logger.Infof("\t\t\t\tnot reached the max hours, so not send notification - %s (cache)", user.NetID)
		return nil
	}

	//based on the cached data we need to send it
	//in this case we must refresh the login time with up to date data to determine if we really need to send it.
	lastLogin, err = n.lastLoginRefreshCache(user)
	if err != nil {
		n.logger.Errorf("\t\t\t\terror refreshing cache last login for - %s", user.NetID)
		return err
	}
	if lastLogin == nil {
		n.logger.Debugf("\t\t\t\tlast login is not available for user after refresh - %s", user.NetID)
		return nil
	}

	//determine if needs to send notification - using up to date login time
	needsSend = n.lastLoginNeedsToSend(hours, now, *lastLogin)
	if !needsSend {
		//not reached the max hours, so not send notification
		n.logger.Infof("\t\t\t\tnot reached the max hours, so not send notification - %s (up to date)", user.NetID)
		return nil
	}

	//need to send but first check if it has been send before

	//check if has been sent before
	criteriaHash := n.generateLastLoginHash(*lastLogin, hours)
	sentNudge, err := n.storage.FindSentNudge(nudge.ID, user.ID, user.NetID, criteriaHash, n.config.Mode)
	if err != nil {
		//not reached the max hours, so not send notification
		n.logger.Errorf("\t\t\t\terror checking if sent nudge exists - %s - %s", nudge.ID, user.NetID)
		return err
	}
	if sentNudge != nil {
		n.logger.Infof("\t\t\t\tthis has been already sent - %s - %s", nudge.ID, user.NetID)
		return err
	}

	//it has not been sent, so sent it
	err = n.sendLastLoginNudgeForUser(nudge, user, *lastLogin, hours)
	if err != nil {
		n.logger.Errorf("\t\t\t\terror send last login nudge - %s - %s", nudge.ID, user.NetID)
		return err
	}

	return nil
}

func (n nudgesLogic) lastLoginNeedsToSend(hours float64, now time.Time, lastLogin time.Time) bool {
	difference := now.Sub(lastLogin) //difference between now and the last login
	differenceInHours := difference.Hours()
	return differenceInHours > hours
}

func (n nudgesLogic) lastLoginRefreshCache(user ProviderUser) (*time.Time, error) {
	//cache the user data
	updatedUser, err := n.provider.CacheUserData(user)
	if err != nil {
		n.logger.Debugf("error caching user data %s - %s", user.NetID, err)
		return nil, err
	}

	//return the loaded value
	return updatedUser.User.LastLogin, nil
}

func (n nudgesLogic) sendLastLoginNudgeForUser(nudge model.Nudge, user ProviderUser,
	lastLogin time.Time, hours float64) error {
	n.logger.Infof("\t\t\t\tsendLastLoginNudgeForUser - %s - %s", nudge.ID, user.NetID)

	//insert sent nudge
	criteriaHash := n.generateLastLoginHash(lastLogin, hours)
	sentNudge := n.createSentNudge(nudge.ID, user.ID, user.NetID, criteriaHash, n.config.Mode)
	err := n.storage.InsertSentNudge(sentNudge)
	if err != nil {
		n.logger.Errorf("\t\t\t\terror saving sent nudge for %s - %s", user.ID, err)
		return err
	}

	//sending in another thread as it happen very slowly
	go func(user ProviderUser) {
		//send push notification
		recipient := Recipient{UserID: user.ID, Name: ""}
		data := n.prepareNotificationData(nudge.DeepLink)
		err := n.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, nudge.Body, data)
		if err != nil {
			log.Printf("\t\t\t\terror sending notification for %s - %s", user.NetID, err)
			return
		}
		//the logger doe snot work in another thread
		log.Printf("\t\t\t\tsuccess sending notification for %s", user.NetID)
	}(user)

	return nil
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

func (n nudgesLogic) processMissedAssignmentNudgePerUser(nudge model.Nudge, user ProviderUser) (*ProviderUser, error) {
	n.logger.Infof("\t\t\tprocessMissedAssignmentNudgePerUser - %s", nudge.ID)

	//fill the cache if empty
	userData, err := n.maFillCacheIfEmpty(user)
	if err != nil {
		n.logger.Debugf("\t\t\terror filling cache if empty [ma] %s - %s", user.NetID, err)
		return nil, err
	}
	user = *userData

	//in this moment we have ensured that we have cached data for the submissions

	//get the missed assignments based on the cache data
	missedAssignments := n.getMissedAssignmentsData(user, nil)

	if len(missedAssignments) == 0 {
		n.logger.Infof("\t\t\tno missed assignments, so not send notifications - %s", user.NetID)
		return &user, nil
	}

	//at this moment we have identified missed assignments but based on the cached data
	//now we have to determine for which courses we have to update the data and for which
	//we can use the cache data
	notValid, valid := n.maCheckDataValidity(missedAssignments)

	refreshedData := []CourseAssignment{}
	if len(notValid) > 0 {
		n.logger.Infof("\t\t\twe have old data, so need to refresh it - %s", user.NetID)
		updatedData, err := n.provider.CacheUserCoursesData(user, notValid)
		if err != nil {
			n.logger.Debugf("\t\t\terror getting not valid data [ma] %s - %s", user.NetID, err)
			return nil, err
		}
		user = *updatedData

		//once we have loaded the not valid data then we have to cehck if it is really missed assignments
		refreshedData = n.getMissedAssignmentsData(user, notValid)
	}

	//merge valid and unvalid
	readyData := n.maMergeData(refreshedData, valid)

	//determine for which of the assignments we need to send notifications
	hours := nudge.Params["hours"].(float64)
	now := time.Now()
	readyData, err = n.findMissedAssignments(hours, now, readyData)
	if err != nil {
		n.logger.Errorf("\t\t\terror finding missed assignments for - %s", user.NetID)
		return nil, err
	}
	if len(readyData) == 0 {
		//no missed assignments
		n.logger.Infof("\t\t\tno missed assignments after checking due date, so not send notifications - %s", user.NetID)
		return &user, nil
	}

	//here we have the assignments we need to send notifications for

	//process the missed assignments
	for _, assignment := range readyData {
		err = n.processMissedAssignment(nudge, user, assignment, hours)
		if err != nil {
			n.logger.Errorf("\t\t\terror process missed assignment for - %s - %s", user.NetID, assignment.Name)
			return nil, err
		}
	}

	return &user, nil
}

func (n nudgesLogic) maMergeData(part1 []CourseAssignment, part2 []CourseAssignment) []model.Assignment {
	result := []model.Assignment{}

	for _, c := range part1 {
		result = append(result, c.Data)
	}

	for _, c := range part2 {
		result = append(result, c.Data)
	}

	return result
}

func (n nudgesLogic) maCheckDataValidity(missedAssignments []CourseAssignment) ([]int, []CourseAssignment) {
	notValidMap := map[int]bool{}
	valid := []CourseAssignment{}

	now := time.Now()
	for _, ma := range missedAssignments {
		syncDate := ma.SyncDate
		if utils.DateEqual(now, syncDate) {
			//valid when in the same calendar day
			valid = append(valid, ma)
		} else {
			//not valid
			notValidMap[ma.Data.CourseID] = true
		}
	}

	//prepare not valid
	notValid := make([]int, len(notValidMap))
	i := 0
	for k := range notValidMap {
		notValid[i] = k
		i++
	}

	return notValid, valid
}

//if coursesIDs is empty then check for all courses
func (n nudgesLogic) getMissedAssignmentsData(user ProviderUser, coursesIDs []int) []CourseAssignment {
	userCourses := user.Courses
	if userCourses == nil || len(userCourses.Data) == 0 {
		return []CourseAssignment{}
	}

	result := []CourseAssignment{}
	for _, uc := range userCourses.Data {

		if len(coursesIDs) == 0 || utils.ExistInt(coursesIDs, uc.Data.ID) {

			assignments := uc.Assignments
			if len(assignments) > 0 {
				for _, assignment := range assignments {
					if n.isAssignmentMissed(assignment) {
						result = append(result, assignment)
					}
				}
			}
		}
	}

	return result
}

func (n nudgesLogic) isAssignmentMissed(ca CourseAssignment) bool {
	assignmentData := ca.Data
	assignmentDueAt := assignmentData.DueAt
	if assignmentDueAt == nil {
		n.logger.Debugf("\t\t\tthere is no due_at for assignment %s", assignmentData.Name)
		return false
	}

	submissionData := ca.Submission.Data
	now := time.Now()

	return now.After(*assignmentDueAt) &&
		((submissionData == nil) ||
			(submissionData.SubmittedAt == nil) ||
			(submissionData != nil && submissionData.SubmittedAt != nil && submissionData.SubmittedAt.After(*assignmentData.DueAt)))
}

func (n nudgesLogic) maFillCacheIfEmpty(user ProviderUser) (*ProviderUser, error) {
	//get the courses we need to load assignments data
	coursesIDs := []int{}

	userCourses := user.Courses
	if userCourses == nil || len(userCourses.Data) == 0 {
		n.logger.Debugf("\t\t\tno courses for %s", user.NetID)
		return &user, nil
	}

	for _, userCourse := range userCourses.Data {
		assignments := userCourse.Assignments
		if len(assignments) > 0 {
			for _, assignment := range assignments {
				if assignment.Submission == nil {
					coursesIDs = append(coursesIDs, userCourse.Data.ID)
					break
				}
			}
		}
	}

	if len(coursesIDs) == 0 {
		n.logger.Debugf("\t\t\tthere is no empty submissions for %s", user.NetID)
		return &user, nil
	}

	//we need to load the data for the empty ones
	updatedData, err := n.provider.CacheUserCoursesData(user, coursesIDs)
	if err != nil {
		n.logger.Debugf("\t\t\terror caching user courses data [ma] %s - %s", user.NetID, err)
		return nil, err
	}

	return updatedData, nil
}

func (n nudgesLogic) processMissedAssignment(nudge model.Nudge, user ProviderUser, assignment model.Assignment, hours float64) error {
	n.logger.Infof("\t\t\tprocessMissedAssignment - %s - %s - %s", nudge.ID, user.NetID, assignment.Name)

	//need to send but first check if it has been send before

	//check if has been sent before
	criteriaHash := n.generateMissedAssignmentHash(assignment.ID, hours)
	sentNudge, err := n.storage.FindSentNudge(nudge.ID, user.ID, user.NetID, criteriaHash, n.config.Mode)
	if err != nil {
		//not reached the max hours, so not send notification
		n.logger.Errorf("\t\t\terror checking if sent nudge exists for missed assignment - %s - %s", nudge.ID, user.NetID)
		return err
	}
	if sentNudge != nil {
		n.logger.Infof("\t\t\tthis has been already sent - %s - %s", nudge.ID, user.NetID)
		return err
	}

	//it has not been sent, so sent it
	err = n.sendMissedAssignmentNudgeForUser(nudge, user, assignment, hours)
	if err != nil {
		n.logger.Errorf("\t\t\terror sending missed assignment nudge - %s - %s", nudge.ID, user.NetID)
		return err
	}

	return nil
}

func (n nudgesLogic) sendMissedAssignmentNudgeForUser(nudge model.Nudge, user ProviderUser,
	assignment model.Assignment, hours float64) error {
	n.logger.Infof("\t\t\tsendMissedAssignmentNudgeForUser - %s - %s", nudge.ID, user.NetID)

	//insert sent nudge
	criteriaHash := n.generateMissedAssignmentHash(assignment.ID, hours)
	sentNudge := n.createSentNudge(nudge.ID, user.ID, user.NetID, criteriaHash, n.config.Mode)
	err := n.storage.InsertSentNudge(sentNudge)
	if err != nil {
		n.logger.Errorf("\t\t\terror saving sent missed assignment nudge for %s - %s", user.NetID, err)
		return err
	}

	//sending in another thread as it happen very slowly
	go func(user ProviderUser) {

		//send push notification
		recipient := Recipient{UserID: user.ID, Name: ""}
		body := fmt.Sprintf(nudge.Body, assignment.Name)
		deepLink := fmt.Sprintf(nudge.DeepLink, assignment.CourseID, assignment.ID)
		data := n.prepareNotificationData(deepLink)
		err := n.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, body, data)
		if err != nil {
			n.logger.Debugf("\t\t\terror sending notification for %s - %s", user.ID, err)
			return
		}

		//the logger doe snot work in another thread
		log.Printf("\t\t\t\tsuccess sending notification for %s", user.NetID)
	}(user)

	return nil
}

func (n nudgesLogic) generateMissedAssignmentHash(assignemntID int, hours float64) uint32 {
	assignmentIDComponent := fmt.Sprintf("%d", assignemntID)
	hoursComponent := fmt.Sprintf("%f", hours)
	component := fmt.Sprintf("%s+%s", assignmentIDComponent, hoursComponent)
	hash := utils.Hash(component)
	return hash
}

func (n nudgesLogic) findMissedAssignments(hours float64, now time.Time, assignments []model.Assignment) ([]model.Assignment, error) {
	n.logger.Info("\t\t\tfindMissedAssignments")

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

func (n nudgesLogic) processCompletedAssignmentEarlyNudgePerUser(nudge model.Nudge, user ProviderUser) (*ProviderUser, error) {
	n.logger.Infof("\t\t\tprocessCompletedAssignmentEarlyNudgePerUser - %s", nudge.ID)

	// find the early completion candidate assignments
	candidateAssignments := n.ecFindCandidateAssignments(user)
	if len(candidateAssignments) == 0 {
		n.logger.Infof("\t\t\tthere is no candidate assignments - %s", user.NetID)
		return &user, nil
	}

	// load data if necessary
	updatedUser, updatedCandidateAssignments, err := n.ecLoadDataIfNecessary(user, candidateAssignments)
	if err != nil {
		n.logger.Debugf("\t\t\terror loading data if necessary [ec] %s - %s", user.ID, err)
		return nil, err
	}
	user = *updatedUser

	// determine which of the assignments are early completed
	ecAssignments := []model.Assignment{}
	hours := nudge.Params["hours"].(float64)
	for _, assignment := range updatedCandidateAssignments {
		if n.ecIsEarlyCompleted(assignment, hours) {
			ecAssignments = append(ecAssignments, assignment.Data)
		}
	}

	//here we have the assignments we need to send notifications for

	//process the early completed assignments
	for _, assignment := range ecAssignments {
		err = n.processCompletedAssignmentEarly(nudge, user, assignment, hours)
		if err != nil {
			n.logger.Errorf("\t\t\terror process early complete assignment for - %s - %s", user.NetID, assignment.Name)
			return nil, err
		}
	}

	return &user, nil
}

func (n nudgesLogic) ecIsEarlyCompleted(assignment CourseAssignment, hours float64) bool {
	submission := assignment.Submission
	if submission == nil {
		return false
	}
	submissionData := submission.Data
	if submissionData == nil {
		return false
	}
	submittedAt := submissionData.SubmittedAt
	if submittedAt == nil {
		return false
	}

	dueAtSeconds := assignment.Data.DueAt.Unix()
	submittedAtSeconds := submittedAt.Unix()
	hoursInSecs := hours * 60 * 60
	//check if submitted is x hours before due
	difference := dueAtSeconds - submittedAtSeconds

	return difference > int64(hoursInSecs)
}

func (n nudgesLogic) ecFindCandidateAssignments(user ProviderUser) []CourseAssignment {
	userCourses := user.Courses
	if userCourses == nil || userCourses.Data == nil || len(userCourses.Data) == 0 {
		return []CourseAssignment{}
	}
	userCoursesData := userCourses.Data

	result := []CourseAssignment{}
	//now := time.Now()
	for _, uc := range userCoursesData {
		assignments := uc.Assignments
		if len(assignments) > 0 {
			for _, assignment := range assignments {
				dueAt := assignment.Data.DueAt
				if dueAt == nil {
					//we rely on due at date
					continue
				}
				//TODO revert
				//if now.Before(*dueAt) {
				result = append(result, assignment)
				//}
			}
		}
	}

	return result
}

func (n nudgesLogic) ecLoadDataIfNecessary(user ProviderUser, assignments []CourseAssignment) (*ProviderUser, []CourseAssignment, error) {
	result := []CourseAssignment{}
	forLoading := map[int][]int{}

	now := time.Now()
	for _, assignment := range assignments {
		submission := assignment.Submission

		if submission != nil {
			n.logger.Infof("\t\t\tsubmission has been loaded before - %s", assignment.Data.Name)
			syncDate := submission.SyncDate

			if utils.DateEqual(now, syncDate) {
				n.logger.Info("\t\t\t\tit is from today, no need to load")
				result = append(result, assignment)
			} else {
				n.logger.Info("\t\t\t\tit has been loaded before")

				submissionData := submission.Data
				if submissionData != nil {
					n.logger.Info("\t\t\t\t\tthere is a submission, no need to load")
					result = append(result, assignment)
				} else {
					n.logger.Info("\t\t\t\t\tthere is no submission yet")

					//add for loading
					value := forLoading[assignment.Data.CourseID]
					value = append(value, assignment.Data.ID)
					forLoading[assignment.Data.CourseID] = value
				}
			}
		} else {
			n.logger.Infof("\t\t\tsubmission has NOT been loaded before - %s", assignment.Data.Name)

			//add for loading
			value := forLoading[assignment.Data.CourseID]
			value = append(value, assignment.Data.ID)
			forLoading[assignment.Data.CourseID] = value
		}
	}

	if len(forLoading) > 0 {
		n.logger.Infof("\t\t\twe need to load assignments for %d courses", len(forLoading))

		coursesIDs := make([]int, len(forLoading))
		assignmentsIDs := []int{}
		i := 0
		for k, v := range forLoading {
			coursesIDs[i] = k
			assignmentsIDs = append(assignmentsIDs, v...)
			i++
		}
		updatedUser, err := n.provider.CacheUserCoursesData(user, coursesIDs)
		if err != nil {
			n.logger.Debugf("\t\t\terror caching user courses data [ma] %s - %s", user.NetID, err)
			return nil, nil, err
		}
		user = *updatedUser

		loadedAssignments := n.ecFindAssignments(user, assignmentsIDs)

		//add the loaded assignemnts to the rest
		result = append(result, loadedAssignments...)
	} else {
		n.logger.Info("\t\t\tno need to load assingments")
	}

	return &user, result, nil
}

func (n nudgesLogic) ecFindAssignments(user ProviderUser, assignmentsIDs []int) []CourseAssignment {
	userCourses := user.Courses
	if userCourses == nil || userCourses.Data == nil || len(userCourses.Data) == 0 {
		return []CourseAssignment{}
	}
	userCoursesData := userCourses.Data

	result := []CourseAssignment{}
	for _, uc := range userCoursesData {
		assignments := uc.Assignments
		if len(assignments) > 0 {
			for _, assignment := range assignments {
				if utils.ExistInt(assignmentsIDs, assignment.Data.ID) {
					result = append(result, assignment)
				}
			}
		}
	}

	return result
}

func (n nudgesLogic) processCompletedAssignmentEarly(nudge model.Nudge, user ProviderUser, assignment model.Assignment, hours float64) error {
	n.logger.Infof("\t\t\tprocessCompletedAssignmentEarly - %s - %s - %s", nudge.ID, user.NetID, assignment.Name)

	//need to send but first check if it has been send before

	//check if has been sent before
	criteriaHash := n.generateEarlyCompletedAssignmentHash(assignment.ID, assignment.Submission.ID, *assignment.Submission.SubmittedAt)
	sentNudge, err := n.storage.FindSentNudge(nudge.ID, user.ID, user.NetID, criteriaHash, n.config.Mode)
	if err != nil {
		//not reached the max hours, so not send notification
		n.logger.Errorf("\t\t\terror checking if sent nudge exists for missed assignment - %s - %s", nudge.ID, user.NetID)
		return err
	}
	if sentNudge != nil {
		n.logger.Infof("\t\t\tthis has been already sent - %s - %s", nudge.ID, user.NetID)
		return err
	}

	//it has not been sent, so sent it
	err = n.sendEarlyCompletedAssignmentNudgeForUser(nudge, user, assignment, hours)
	if err != nil {
		n.logger.Errorf("\t\t\terror send early completed assignment - %s - %s", nudge.ID, user.NetID)
		return err
	}
	return nil
}

func (n nudgesLogic) generateEarlyCompletedAssignmentHash(assignmentID int, submissionID int, submittedAt time.Time) uint32 {
	assignmentIDComponent := fmt.Sprintf("%d", assignmentID)
	submissionIDComponent := fmt.Sprintf("%d", submissionID)
	submittedAtComponent := fmt.Sprintf("%d", submittedAt.Unix())
	component := fmt.Sprintf("%s+%s+%s", assignmentIDComponent, submissionIDComponent, submittedAtComponent)
	hash := utils.Hash(component)
	return hash
}

func (n nudgesLogic) sendEarlyCompletedAssignmentNudgeForUser(nudge model.Nudge, user ProviderUser,
	assignment model.Assignment, hours float64) error {
	n.logger.Infof("\t\t\tsendEarlyCompletedAssignmentNudgeForUser - %s - %s", nudge.ID, user.NetID)

	//insert sent nudge
	criteriaHash := n.generateEarlyCompletedAssignmentHash(assignment.ID, assignment.Submission.ID, *assignment.Submission.SubmittedAt)
	sentNudge := n.createSentNudge(nudge.ID, user.ID, user.NetID, criteriaHash, n.config.Mode)
	err := n.storage.InsertSentNudge(sentNudge)
	if err != nil {
		n.logger.Errorf("\t\t\terror saving sent early completed assignment nudge for %s - %s", user.NetID, err)
		return err
	}

	//sending in another thread as it happen very slowly
	go func(user ProviderUser) {
		//send push notification
		recipient := Recipient{UserID: user.ID, Name: ""}
		deepLink := fmt.Sprintf(nudge.DeepLink, assignment.CourseID, assignment.ID)
		data := n.prepareNotificationData(deepLink)
		err := n.notificationsBB.SendNotifications([]Recipient{recipient}, nudge.Name, nudge.Body, data)
		if err != nil {
			n.logger.Debugf("\t\t\terror sending notification for %s - %s", user.ID, err)
			return
		}

		//the logger doe snot work in another thread
		log.Printf("\t\t\t\tsuccess sending notification for %s", user.NetID)
	}(user)

	return nil
}

// end completed_assignment_early nudge

// calendar_event nudge

func (n nudgesLogic) processTodayCalendarEventsNudgePerUser(nudge model.Nudge, user GroupsBBUser) {
	n.logger.Infof("\tprocessTodayCalendarEventsNudgePerUser - %s", nudge.ID)

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
