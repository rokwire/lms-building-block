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
	nowSecondsInHour := 60*now.Minute() + now.Second()
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
	// omit minutes and seconds so that we only need to handle integer multiples of seconds per hour
	now := time.Now().UTC().Truncate(time.Hour)
	nowSeconds := utils.SecondsInHour * now.Hour()

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
	nowSecondsInHour := 60*now.Minute() + now.Second()
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
	utils.StartTimer(n.streaksTimer, n.streaksTimerDone, &initialDuration, time.Hour, n.processStreaks, "processStreaks", n.logger)

}

func (n streaksNotifications) processStreaks() {
	funcName := "processStreaks"
	// omit minutes and seconds so that we only need to handle integer multiples of seconds per hour
	now := time.Now().UTC().Truncate(time.Hour)
	nowSeconds := utils.SecondsInHour * now.Hour()

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
		userCourses, userUnits, userIDs, err := n.getUserDataForTimezone(config, config.StreaksNotificationsConfig.StreaksProcessTime, nowSeconds)
		if err != nil {
			n.logger.Errorf("%s -> error finding user courses and user units for course key %s: %v", funcName, config.CourseKey, err)
			continue
		}

		usePause := make([]string, 0)    // list of userIDs where pauses should be decremented
		resetStreak := make([]string, 0) // list of userIDs where streak should be reset
		for _, userUnit := range userUnits {
			var userCourse *model.UserCourse
			incompleteTaskHandler := func(item model.UserUnit) error {
				// if task is incomplete, use a pause or reset the streak depending on the current number of pauses
				if userCourse.Pauses > 0 {
					usePause = append(usePause, item.UserID)
				} else {
					resetStreak = append(resetStreak, item.UserID)
				}
				return nil
			}
			completeTaskHandler := func(storage interfaces.Storage, item model.UserUnit, remainsCurrent bool) error {
				// the previous task was completed, so set the start time of the new task to now (beginning of the day)
				item.Completed++
				item.Current = remainsCurrent
				item.Unit.Schedule[item.Completed].DateStarted = &now
				err := storage.UpdateUserUnit(item)
				if err != nil {
					return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
				}
				return nil
			}
			completeUnitHandler := func(storage interfaces.Storage, item model.UserUnit) error {
				// insert the next user unit since the current one has been completed
				nextUnit := userCourse.Course.GetNextUnit(item.Unit.Key)
				if nextUnit != nil {
					nextUnit.Schedule[0].DateStarted = &now
					nextUserUnit := model.UserUnit{ID: uuid.NewString(), AppID: config.AppID, OrgID: config.OrgID, UserID: item.UserID, CourseKey: item.CourseKey,
						Unit: *nextUnit, Completed: 0, Current: true, LastCompleted: item.LastCompleted, DateCreated: time.Now().UTC()}
					err := storage.InsertUserUnit(nextUserUnit)
					if err != nil {
						return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
					}
				}

				return nil
			}

			for i, userID := range userIDs {
				if userID == userUnit.UserID {
					userCourse = &userCourses[i]
					break
				}
			}
			if userCourse == nil {
				n.logger.Errorf("%s -> error matching user course for user unit %s: %v", funcName, userUnit.ID, err)
				continue
			}

			err = n.checkScheduleTaskCompletion(userUnit, now, incompleteTaskHandler, 0, completeTaskHandler, completeUnitHandler)
			if err != nil {
				n.logger.Errorf("%s -> error checking task completion for user unit %s: %v", funcName, userUnit.ID, err)
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

func (n streaksNotifications) getUserDataForTimezone(config model.CourseConfig, processTime int, nowSeconds int) ([]model.UserCourse, []model.UserUnit, []string, error) {
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

		userUnits, err := n.storage.FindUserUnits(config.AppID, config.OrgID, userIDs, config.CourseKey, &current)
		if err != nil {
			return nil, nil, nil, err
		}
		return userCourses, userUnits, userIDs, nil
	}

	return nil, nil, nil, nil
}

func (n streaksNotifications) filterUsersByIncomplete(userUnits []model.UserUnit, userIDs []string, now time.Time, streaksProcessTime int, notificationProcessTime int) ([]string, error) {
	filtered := make([]string, 0)
	for _, userUnit := range userUnits {
		if !utils.Exist[string](userIDs, userUnit.UserID) {
			continue
		}

		incompleteTaskHandler := func(item model.UserUnit) error {
			// if task is incomplete, use a pause or reset the streak depending on the current number of pauses
			filtered = append(filtered, item.UserID)
			return nil
		}
		// equal to difference between notification process time and streaks process time (start of "day") converted to hours
		incompleteTaskPeriodOffset := (notificationProcessTime - streaksProcessTime) / utils.SecondsInHour

		err := n.checkScheduleTaskCompletion(userUnit, now, incompleteTaskHandler, -incompleteTaskPeriodOffset, nil, nil)
		if err != nil {
			n.logger.Errorf("processNotifications -> error checking task completion for user unit %s: %v", userUnit.ID, err)
			continue
		}
	}

	return filtered, nil
}

func (n streaksNotifications) checkScheduleTaskCompletion(userUnit model.UserUnit, now time.Time, incompleteTaskHandler func(model.UserUnit) error, incompleteTaskPeriodOffset int,
	completeTaskHandler func(interfaces.Storage, model.UserUnit, bool) error, completeUnitHandler func(interfaces.Storage, model.UserUnit) error) error {
	if incompleteTaskHandler == nil {
		return errors.ErrorData(logutils.StatusInvalid, "incomplete task handler", nil)
	}

	//TODO: make sure there are no situations where no streak is earned and no pause is used
	// check if the last completed schedule item was completed within (24*days+offset) hours before now
	days := userUnit.Unit.Schedule[userUnit.Completed].Duration
	if userUnit.LastCompleted != nil && userUnit.LastCompleted.Add((24*time.Duration(days)+time.Duration(incompleteTaskPeriodOffset))*time.Hour).Before(now) { //TODO: may need to change this to handle user travelling, DST
		// not completed within specified period, so handle incomplete
		return incompleteTaskHandler(userUnit)
	} else if completeTaskHandler != nil {
		// completed within specified period, so handle complete if desired
		remainsCurrent := (userUnit.Completed+1 < userUnit.Unit.Required)
		if remainsCurrent {
			return completeTaskHandler(n.storage, userUnit, true)
		}

		// user completed the current unit because the end of the schedule has been reached, so complete the task then complete the unit
		transaction := func(storage interfaces.Storage) error {
			err := completeTaskHandler(storage, userUnit, false)
			if err != nil {
				return errors.WrapErrorAction("completing", "user unit schedule item", &logutils.FieldArgs{"id": userUnit.ID, "completed": userUnit.Completed}, err)
			}

			if completeUnitHandler != nil {
				return completeUnitHandler(storage, userUnit)
			}

			return nil
		}

		return n.storage.PerformTransaction(transaction)
	}

	return nil
}
