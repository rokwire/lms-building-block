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
	"lms/core/interfaces"
	"lms/core/model"
	"lms/driven/notifications"
	"lms/utils"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

type streaksNotifications struct {
	logger *logs.Logger

	notificationsBB interfaces.NotificationsBB

	storage interfaces.Storage

	//notifications timer
	notificationsTimer     *time.Timer
	notificationsTimerDone chan bool
	//streaks timer
	streaksTimer     *time.Timer
	streaksTimerDone chan bool
}

func (n streaksNotifications) start() {
	//setup hourly notifications timer
	go n.setupNotificationsTimer()
	//setup hourly streaks timer
	go n.setupStreaksTimer()
}

func (n streaksNotifications) setupNotificationsTimer() {
	now := time.Now().UTC()
	nowSecondsInHour := utils.SecondsInMinute*now.Minute() + now.Second()
	desiredMoment := 0 //default desired moment of the hour in seconds (beginning of the hour)

	var durationInSeconds int
	n.logger.Infof("setupNotificationsTimer -> nowSecondsInHour:%d", nowSecondsInHour)
	if nowSecondsInHour <= desiredMoment {
		n.logger.Info("setupNotificationsTimer -> notifications not yet processed this hour")
		durationInSeconds = desiredMoment - nowSecondsInHour
	} else {
		n.logger.Info("setupNotificationsTimer -> notifications have already been processed this hour")
		durationInSeconds = (utils.SecondsInHour - nowSecondsInHour) + desiredMoment // the time which left this hour + desired moment from next hour
	}

	initialDuration := time.Second * time.Duration(durationInSeconds)
	utils.StartTimer(n.notificationsTimer, n.notificationsTimerDone, &initialDuration, time.Hour, n.processNotifications, "processNotifications", n.logger)
}

func (n streaksNotifications) processNotifications() {
	funcName := "processNotifications"
	now := time.Now().UTC()
	nowSeconds := utils.SecondsInHour*now.Hour() + utils.SecondsInMinute*now.Minute() + now.Second()

	active := true
	courseConfigs, err := n.storage.FindCourseConfigs(nil, nil, &active)
	if err != nil {
		n.logger.Errorf("%s -> error finding active course configs: %v", funcName, err)
		return
	}
	if len(courseConfigs) == 0 {
		n.logger.Errorf("%s -> no active course configs for notifications", funcName)
		return
	}

	for _, config := range courseConfigs {
		for _, notification := range config.StreaksNotificationsConfig.Notifications {
			if notification.Active {
				_, userUnits, userIDs, err := n.getUserDataForTimezone(config, notification.ProcessTime, nowSeconds)
				if err != nil {
					n.logger.Errorf("%s -> error finding user courses and user units for course key %s: %v", funcName, config.CourseKey, err)
					continue
				}

				for reqKey, reqVal := range notification.Requirements {
					if reqKey == "completed" && reqVal == false {
						userIDs, err = n.filterUsersByIncomplete(userUnits, userIDs, now, config.StreaksNotificationsConfig.StreaksProcessTime, notification.ProcessTime)
						if err != nil {
							n.logger.Errorf("%s -> error filtering users by incomplete for notification %s in course config %s: %v", funcName, notification.Subject, config.ID, err)
							continue
						}
					}
					//TODO: add more requirement checks as needed here (userIDs, err = n.<function name>(userUnits, userIDs, ...))
				}

				if len(userIDs) == 0 {
					n.logger.Infof("%s -> no recipients for notification %s for course key %s", funcName, notification.Subject, config.CourseKey)
					continue
				}

				recipients := make([]notifications.Recipient, len(userIDs))
				for i, userID := range userIDs {
					recipients[i] = notifications.Recipient{UserID: userID}
				}

				switch config.StreaksNotificationsConfig.NotificationsMode {
				case "normal":
					err = n.notificationsBB.SendNotifications(recipients, notification.Subject, notification.Body, notification.Params)
					if err != nil {
						n.logger.Errorf("%s -> error sending notification %s for course key %s: %v", funcName, notification.Subject, config.CourseKey, err)
					} else {
						n.logger.Infof("%s -> sent notification %s for course key %s", funcName, notification.Subject, config.CourseKey)
					}
				case "test":
					n.logger.Infof("%s -> (test) notification %s would be sent to %d users for course key %s", funcName, notification.Subject, len(userIDs), config.CourseKey)
				}
			}
		}
	}
}

func (n streaksNotifications) setupStreaksTimer() {
	//TODO: setup hourly streaks timer (streaks must be updated according to user timezone)
	now := time.Now().UTC()
	nowSecondsInHour := utils.SecondsInMinute*now.Minute() + now.Second()
	desiredMoment := 0 //default desired moment of the hour in seconds (beginning of the hour)
	var durationInSeconds int
	n.logger.Infof("setupStreaksTimer -> nowSecondsInHour:%d", nowSecondsInHour)
	if nowSecondsInHour <= desiredMoment {
		n.logger.Info("setupStreaksTimer -> streaks not yet processed this hour")
		durationInSeconds = desiredMoment - nowSecondsInHour
	} else {
		n.logger.Info("setupStreaksTimer -> streaks have already been processed this hour")
		durationInSeconds = (utils.SecondsInHour - nowSecondsInHour) + desiredMoment // the time which left this hour + desired moment from next hour
	}
	initialDuration := time.Second * time.Duration(durationInSeconds)
	//change initialduration to 60 and time.hour to time.minute for testing purpose
	//initialDuration = 60
	//utils.StartTimer(n.streaksTimer, n.streaksTimerDone, &initialDuration, time.Minute, n.processStreaks, "processStreaks", n.logger)
	utils.StartTimer(n.streaksTimer, n.streaksTimerDone, &initialDuration, time.Hour, n.processStreaks, "processStreaks", n.logger)

}

func (n streaksNotifications) processStreaks() {
	funcName := "processStreaks"
	now := time.Now().UTC()
	nowSeconds := utils.SecondsInHour*now.Hour() + utils.SecondsInMinute*now.Minute() + now.Second()

	courseConfigs, err := n.storage.FindCourseConfigs(nil, nil, nil)
	if err != nil {
		n.logger.Errorf("%s -> error finding active course configs: %v", funcName, err)
		return
	}
	if len(courseConfigs) == 0 {
		n.logger.Errorf("%s -> no active course configs for streaks", funcName)
		return
	}

	// batch the following if number of user courses gets large
	// careful running this with multiple service instances - not all storage operations used here are idempotent
	for _, config := range courseConfigs {
		userCourses, userUnitsMap, _, err := n.getUserDataForTimezone(config, config.StreaksNotificationsConfig.StreaksProcessTime, nowSeconds)
		if err != nil {
			n.logger.Errorf("%s -> error finding user courses and user units for course key %s: %v", funcName, config.CourseKey, err)
			continue
		}

		usePause := make([]string, 0)    // list of userIDs where pauses should be decremented
		resetStreak := make([]string, 0) // list of userIDs where streak should be reset
		for userID, userUnits := range userUnitsMap {
			var userCourse *model.UserCourse
			for i, thisUserCourse := range userCourses {
				if thisUserCourse.UserID == userID {
					userCourse = &userCourses[i]
					break
				}
			}
			if userCourse == nil {
				n.logger.Errorf("%s -> error matching user course for user ID %s: %v", funcName, userID, err)
				continue
			}
			incompleteTaskHandler := func() error {
				// if task is incomplete, use a pause or reset the streak depending on the current number of pauses
				if userCourse.Pauses > 0 {
					usePause = append(usePause, userID)
				} else {
					resetStreak = append(resetStreak, userID)
				}
				return nil
			}
			completeTaskHandler := func(userUnit model.UserUnit) error {
				// the previous task was completed, so set the start time of the new task to now (beginning of the day)
				userUnit.Completed++
				userUnit.Unit.Schedule[userUnit.Completed].DateStarted = &now
				err := n.storage.UpdateUserUnit(userUnit)
				if err != nil {
					return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
				}
				return nil
			}
			completeUnitHandler := func(userUnit model.UserUnit) error {
				// insert the next user unit since the current one has been completed
				transaction := func(storage interfaces.Storage) error {
					nextUnit := userCourse.Course.GetNextUnit(userUnit)
					if nextUnit != nil {
						nextUnit.Schedule[nextUnit.ScheduleStart].DateStarted = &now
						nextUserUnit := model.UserUnit{ID: uuid.NewString(), AppID: config.AppID, OrgID: config.OrgID, UserID: userUnit.UserID, CourseKey: userUnit.CourseKey,
							ModuleKey: userUnit.ModuleKey, Unit: *nextUnit, Completed: nextUnit.ScheduleStart, Current: true, LastCompleted: userUnit.LastCompleted, DateCreated: time.Now().UTC()}
						err := storage.InsertUserUnit(nextUserUnit)
						if err != nil {
							return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
						}
					}

					// set current to false on the current user unit since there is a new curernt one or the course is finished
					userUnit.Current = false
					userUnit.Completed++
					err := storage.UpdateUserUnit(userUnit)
					if err != nil {
						return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
					}

					return nil
				}

				return n.storage.PerformTransaction(transaction)
			}

			err = n.checkScheduleTaskCompletion(userUnits, now, incompleteTaskHandler, 0, completeTaskHandler, completeUnitHandler)
			if err != nil {
				n.logger.Errorf("%s -> error checking task completion for user userID %s: %v", funcName, userID, err)
				continue
			}
		}

		if len(usePause) > 0 {
			err = n.storage.DecrementUserCoursePauses(config.AppID, config.OrgID, usePause, config.CourseKey)
			if err != nil {
				n.logger.Errorf("%s -> error decrementing pauses for course key %s: %v", funcName, config.CourseKey, err)
			}
		}
		if len(resetStreak) > 0 {
			err = n.storage.ResetUserCourseStreaks(config.AppID, config.OrgID, resetStreak, config.CourseKey)
			if err != nil {
				n.logger.Errorf("%s -> error reseting streaks for course key %s: %v", funcName, config.CourseKey, err)
			}
		}
	}
}

func (n streaksNotifications) getUserDataForTimezone(config model.CourseConfig, processTime int, nowSeconds int) ([]model.UserCourse, map[string][]model.UserUnit, []string, error) {
	tzOffsets := make(model.TZOffsets, 0)
	var userCourses []model.UserCourse
	var err error

	offset := processTime - nowSeconds
	if config.StreaksNotificationsConfig.TimezoneName == model.UserTimezone {
		if offset >= utils.MinTZOffset && offset <= utils.MaxTZOffset {
			tzOffsets = append(tzOffsets, offset)
		}

		if offset+utils.SecondsInDay <= utils.MaxTZOffset {
			tzOffsets = append(tzOffsets, offset+utils.SecondsInDay)
		}
		if offset-utils.SecondsInDay >= utils.MinTZOffset {
			tzOffsets = append(tzOffsets, offset-utils.SecondsInDay)
		}

		// load user courses for this course based on timezone offsets
		userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, tzOffsets.GeneratePairs(config.StreaksNotificationsConfig.PreferEarly))
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		configOffset := config.StreaksNotificationsConfig.TimezoneOffset
		tolerance := config.StreaksNotificationsConfig.TimerDelayTolerance
		offsetDiff := offset - configOffset
		offsetDiffPlusDay := offset + utils.SecondsInDay - configOffset
		offsetDiffMinusDay := offset - utils.SecondsInDay - configOffset
		if offsetDiff <= tolerance || offsetDiff >= -tolerance || offsetDiffPlusDay <= tolerance || offsetDiffPlusDay >= -tolerance || offsetDiffMinusDay <= tolerance || offsetDiffMinusDay >= -tolerance {
			// load all user courses for this course
			userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, nil)
			if err != nil {
				return nil, nil, nil, err
			}
		}
	}

	if len(userCourses) > 0 {
		current := true
		// there should be one userCourse per userID (user cannot take the same course multiple times simultaneously)
		userIDs := make([]string, len(userCourses))
		for i, userCourse := range userCourses {
			userIDs[i] = userCourse.UserID
		}

		userUnits, err := n.storage.FindUserUnits(config.AppID, config.OrgID, userIDs, config.CourseKey, nil, &current)
		if err != nil {
			return nil, nil, nil, err
		}
		userUnitsMap := make(map[string][]model.UserUnit)
		for _, userUnit := range userUnits {
			if userUnitsMap[userUnit.UserID] == nil {
				userUnitsMap[userUnit.UserID] = []model.UserUnit{}
			}
			userUnitsMap[userUnit.UserID] = append(userUnitsMap[userUnit.UserID], userUnit)
		}
		return userCourses, userUnitsMap, userIDs, nil
	}

	return nil, nil, nil, nil
}

func (n streaksNotifications) filterUsersByIncomplete(userUnitsMap map[string][]model.UserUnit, userIDs []string, now time.Time, streaksProcessTime int, notificationProcessTime int) ([]string, error) {
	filtered := make([]string, 0)
	for userID, userUnits := range userUnitsMap {
		incompleteTaskHandler := func() error {
			// if task is incomplete, use a pause or reset the streak depending on the current number of pauses
			filtered = append(filtered, userID)
			return nil
		}
		// equal to difference between notification process time and streaks process time (start of "day") converted to hours
		incompleteTaskPeriodOffset := (notificationProcessTime - streaksProcessTime) / utils.SecondsInHour

		err := n.checkScheduleTaskCompletion(userUnits, now, incompleteTaskHandler, -incompleteTaskPeriodOffset, nil, nil)
		if err != nil {
			n.logger.Errorf("processNotifications -> error checking task completion for userID %s: %v", userID, err)
			continue
		}
	}

	return filtered, nil
}

// iterate through array of active userUnit
// run completeTaskHandler like normal if any qualifies, halt for incompleteTaskHandler
// only run incompleteTaskHandler if none of active userUnit qualifies as completed task
func (n streaksNotifications) checkScheduleTaskCompletion(userUnits []model.UserUnit, now time.Time, incompleteTaskHandler func() error, incompleteTaskPeriodOffset int,
	completeTaskHandler func(model.UserUnit) error, completeUnitHandler func(model.UserUnit) error) error {
	if incompleteTaskHandler == nil {
		return errors.ErrorData(logutils.StatusInvalid, "incomplete task handler", nil)
	}

	allIncomplete := true
	for _, userUnit := range userUnits {
		// if userUnit.Completed == userUnit.Unit.ScheduleStart {
		// 	// user has not completed the current task
		// } else
		if userUnit.Completed+1 < userUnit.Unit.Required {
			// check if the last completed schedule item was completed within (24*days+offset) hours before now
			// Note: is completed suppose to be num of completed or index of latest completed? how to handle 0
			days := userUnit.Unit.Schedule[userUnit.Completed].Duration
			//days := userUnit.Unit.Schedule[userUnit.Completed-1].Duration
			if userUnit.LastCompleted != nil && userUnit.LastCompleted.Add((time.Duration(utils.HoursInDay*days)+time.Duration(incompleteTaskPeriodOffset))*time.Hour).Before(now) {
				// not completed within specified period, so handle incomplete
				//return incompleteTaskHandler()
			} else if completeTaskHandler != nil {
				// completed within specified period, so handle complete if desired
				err := completeTaskHandler(userUnit)
				if err != nil {
					n.logger.Error("checkScheduleTaskCompletion completeTaskHandler error: " + err.Error())
				}
				allIncomplete = false
			}
		} else if userUnit.Completed+1 == userUnit.Unit.Required && completeUnitHandler != nil {
			// user completed the current unit
			err := completeUnitHandler(userUnit)
			if err != nil {
				n.logger.Error("checkScheduleTaskCompletion completeUnitHandler error: " + err.Error())
			}
			allIncomplete = false
		}
	}
	if allIncomplete {
		return incompleteTaskHandler()
	}
	return nil
}
