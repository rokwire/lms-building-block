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

const (
	minTZOffset int = -43200
	maxTZOffset int = 50400
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
	// change to minute for testing.
	initialDuration = 60
	utils.StartTimer(n.notificationsTimer, n.notificationsTimerDone, &initialDuration, time.Minute, n.processNotifications, "processNotifications", n.logger)
	//utils.StartTimer(n.notificationsTimer, n.notificationsTimerDone, &initialDuration, time.Hour, n.processNotifications, "processNotifications", n.logger)
}

func (n streaksNotifications) processNotifications() {
	funcName := "processNotifications"
	now := time.Now().UTC()
	nowSeconds := 60*60*now.Hour() + 60*now.Minute() + now.Second()

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
				userCourses, err := n.getUserCoursesForTimezone(config, notification.ProcessTime, nowSeconds)
				if err != nil {
					n.logger.Errorf("%s -> error finding user courses for course key %s: %v", funcName, config.CourseKey, err)
					continue
				}

				recipients := make([]notifications.Recipient, 0)
				for _, userCourse := range userCourses {
					recipients = append(recipients, notifications.Recipient{UserID: userCourse.UserID})
				}
				if len(recipients) == 0 {
					n.logger.Infof("%s -> no recipients for notification %s for course key %s", funcName, notification.Subject, config.CourseKey)
					continue
				}

				err = n.notificationsBB.SendNotifications(recipients, notification.Subject, notification.Body, notification.Params)
				if err != nil {
					// why would usercourse be modified here?
					n.logger.Errorf("%s -> error sending notification %s for course key %s: %v", funcName, notification.Subject, config.CourseKey, err)
					continue
				}
				n.logger.Infof("%s -> sent notification %s for course key %s", funcName, notification.Subject, config.CourseKey)
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
	//change initialduration to 60 and time.hour to time.minute for testing purpose
	//initialDuration = 60
	//utils.StartTimer(n.streaksTimer, n.streaksTimerDone, &initialDuration, time.Minute, n.processStreaks, "processStreaks", n.logger)
	utils.StartTimer(n.streaksTimer, n.streaksTimerDone, &initialDuration, time.Hour, n.processStreaks, "processStreaks", n.logger)

}

func (n streaksNotifications) processStreaks() {
	funcName := "processStreaks"
	now := time.Now().UTC()
	nowSeconds := 60*60*now.Hour() + 60*now.Minute() + now.Second()

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
		userCourses, err := n.getUserCoursesForTimezone(config, config.StreaksNotificationsConfig.StreaksProcessTime, nowSeconds)
		if err != nil {
			n.logger.Errorf("%s -> error finding user courses for course key %s: %v", funcName, config.CourseKey, err)
			continue
		}

		current := true
		// there should be one userCourse per userID (user cannot take the same course multiple times simultaneously)
		userIDs := make([]string, len(userCourses))
		for i, userCourse := range userCourses {
			userIDs[i] = userCourse.UserID
		}

		userUnits, err := n.storage.FindUserUnits(config.AppID, config.OrgID, userIDs, config.CourseKey, &current)
		if err != nil {
			n.logger.Errorf("%s -> error finding user units for course key %s: %v", funcName, config.CourseKey, err)
			continue
		}

		usePause := make([]string, 0)    // list of userIDs where pauses should be decremented
		resetStreak := make([]string, 0) // list of userIDs where streak should be reset
		for _, userUnit := range userUnits {
			var userCourse *model.UserCourse
			for i, userID := range userIDs {
				if userID == userUnit.UserID {
					userCourse = &userCourses[i]
					break
				}
			}
			if userCourse == nil {
				//TODO: return error here?
				continue
			}

			if userUnit.Completed == 0 {
				if userCourse.Pauses > 0 {
					usePause = append(usePause, userUnit.UserID)
				} else {
					resetStreak = append(resetStreak, userCourse.UserID)
				}
			} else if userUnit.Completed < userUnit.Unit.Required {
				// check if the last completed schedule item was completed within days before now
				days := userUnit.Unit.Schedule[userUnit.Completed-1].Duration
				if userUnit.LastCompleted != nil && userUnit.LastCompleted.Add(time.Duration(days)*24*time.Hour).Before(now) {
					if userCourse.Pauses > 0 {
						usePause = append(usePause, userUnit.UserID)
					} else {
						resetStreak = append(resetStreak, userCourse.UserID)
					}
				}
			} else {
				// insert the next user unit since the current one has been completed
				transaction := func(storage interfaces.Storage) error {
					nextUnit := userCourse.Course.GetNextUnit(userUnit.Unit.Key)
					if nextUnit != nil {
						nextUserUnit := model.UserUnit{ID: uuid.NewString(), AppID: config.AppID, OrgID: config.OrgID, UserID: userUnit.UserID, CourseKey: userUnit.CourseKey,
							Unit: *nextUnit, Completed: 0, Current: true, DateCreated: time.Now().UTC()}
						err := storage.InsertUserUnit(nextUserUnit)
						if err != nil {
							return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
						}
					}

					// set current to false on the current user unit since there is a new curernt one or the course is finished
					userUnit.Current = false
					err := storage.UpdateUserUnit(config.AppID, config.OrgID, userUnit.UserID, userUnit.CourseKey, userUnit)
					if err != nil {
						return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
					}

					return nil
				}

				err = n.storage.PerformTransaction(transaction)
				if err != nil {
					n.logger.Errorf("%s -> error finding user units for course key %s: %v", funcName, config.CourseKey, err)
				}
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

func (n streaksNotifications) getUserCoursesForTimezone(config model.CourseConfig, processTime int, nowSeconds int) ([]model.UserCourse, error) {
	tzOffsets := make(model.TZOffsets, 0)
	var userCourses []model.UserCourse
	var err error

	offset := processTime - nowSeconds
	if config.StreaksNotificationsConfig.TimezoneName == "user" {
		if offset >= minTZOffset && offset <= maxTZOffset {
			tzOffsets = append(tzOffsets, offset)
		}

		if offset+utils.SecondsInDay <= maxTZOffset {
			tzOffsets = append(tzOffsets, offset+utils.SecondsInDay)
		}
		if offset-utils.SecondsInDay >= minTZOffset {
			tzOffsets = append(tzOffsets, offset-utils.SecondsInDay)
		}

		// load user courses for this course based on timezone offsets
		userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, tzOffsets.GeneratePairs(config.StreaksNotificationsConfig.PreferEarly))
	} else {
		configOffset := config.StreaksNotificationsConfig.TimezoneOffset
		if offset == configOffset || offset+utils.SecondsInDay == configOffset || offset-utils.SecondsInDay == configOffset {
			// load all user courses for this course
			userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, nil)
		}
	}

	return userCourses, err
}

// func (n streaksNotifications) processNotificationRequirements(requirements map[string]interface{}, timezoneName string, timezoneOffset int) (map[string]interface{}, error) {
// 	newRequirements := make(map[string]interface{})
// 	for reqKey, reqVal := range requirements {
// 		newRequirements[reqKey] = reqVal
// 		if reqKey == "completed_tasks" {
// 			userLoc := time.FixedZone(timezoneName, timezoneOffset)
// 			if userLoc == nil {
// 				return nil, errors.ErrorData(logutils.StatusInvalid, "user timezone", &logutils.FieldArgs{"name": timezoneName, "offset": timezoneOffset})
// 			}

// 			newRequirements["completed_tasks_time"] = time.Now().In(userLoc)
// 		}
// 	}
// 	return newRequirements, nil
// }