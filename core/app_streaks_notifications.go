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
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
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
		userCourses, currentUserUnits, _, err := n.getUserDataForTimezone(config, config.StreaksNotificationsConfig.StreaksProcessTime, nowSeconds)
		if err != nil {
			n.logger.Errorf("%s -> error finding user courses and user units for course key %s: %v", funcName, config.CourseKey, err)
			continue
		}

		usePause := make([]string, 0)    // list of userIDs where pauses should be decremented
		resetStreak := make([]string, 0) // list of userIDs where streak should be reset
		for userID, userUnits := range currentUserUnits {
			var userCourse *model.UserCourse
			for i, uc := range userCourses {
				if uc.UserID == userID {
					userCourse = &userCourses[i]
					break
				}
			}
			if userCourse == nil {
				n.logger.Errorf("%s -> error matching user course for user ID %s: %v", funcName, userID, err)
				continue
			}

			incompleteTaskHandler := func(incompleteUserID string) error {
				// if task is incomplete, use a pause or reset the streak depending on the current number of pauses
				if userCourse.Pauses > 0 {
					usePause = append(usePause, incompleteUserID)
				} else {
					resetStreak = append(resetStreak, incompleteUserID)
				}
				return nil
			}
			completeTaskHandler := func(storage interfaces.Storage, item model.UserUnit, remainsCurrent bool) error {
				// the previous task was completed, so set the start time of the new task to now (beginning of the day)
				item.Completed++
				item.Current = remainsCurrent
				if remainsCurrent {
					userScheduleItem, _, _, _ := item.GetScheduleItem("", true)
					if userScheduleItem == nil {
						return errors.ErrorData(logutils.StatusMissing, model.TypeScheduleItem, &logutils.FieldArgs{"current": true})
					}
					userScheduleItem.DateStarted = &now
				}

				err := storage.UpdateUserUnit(item)
				if err != nil {
					return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
				}

				// if there are no more required schedule items to be done in the course, set date completed
				// allow the new current schedule item to be returned if current schedule item not required because userUnit.Completed has already been incremented
				if userCourse.Course.GetNextRequiredScheduleItem(item.ModuleKey, item.Unit.Key, item.Completed, true) == nil {
					if userCourse.CompletedModules == nil {
						userCourse.CompletedModules = make(map[string]time.Time)
					}
					userCourse.CompletedModules[item.ModuleKey] = now

					if userCourse.DateCompleted == nil && userCourse.IsComplete() {
						userCourse.DateCompleted = &now // prevents streak timer from operating on any data associated with this UserCourse
					}

					err := storage.UpdateUserCourse(*userCourse)
					if err != nil {
						return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserCourse, nil, err)
					}
				}

				return nil
			}
			completeUnitHandler := func(storage interfaces.Storage, item model.UserUnit) error {
				// insert the next user unit if the unit exists since the current one has been completed
				nextUnit := userCourse.Course.GetNextUnit(item.ModuleKey, item.Unit.Key)
				if nextUnit != nil {
					nextUserSchedule := nextUnit.CreateUserSchedule()
					nextUserSchedule[0].DateStarted = &now
					nextUserUnit := model.UserUnit{ID: uuid.NewString(), AppID: config.AppID, OrgID: config.OrgID, UserID: item.UserID, CourseKey: item.CourseKey, ModuleKey: item.ModuleKey,
						Unit: *nextUnit, Completed: 0, Current: true, UserSchedule: nextUserSchedule, DateCreated: time.Now().UTC()}

					err := storage.InsertUserUnit(nextUserUnit)
					if err != nil {
						return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
					}
				}

				return nil
			}

			err = n.checkScheduleTaskCompletion(userID, userUnits, now, incompleteTaskHandler, 0, completeTaskHandler, completeUnitHandler)
			if err != nil {
				n.logger.Errorf("%s -> error checking task completion for user userID %s: %v", funcName, userID, err)
				continue
			}
		}

		if len(usePause) > 0 {
			err = n.storage.DecrementUserCoursePauses(config.AppID, config.OrgID, usePause, config.CourseKey)
			if err != nil {
				n.logger.Errorf("%s -> error decrementing pauses for course config %s: %v", funcName, config.ID, err)
			}
		}
		if len(resetStreak) > 0 {
			err = n.storage.ResetUserCourseStreaks(config.AppID, config.OrgID, resetStreak, config.CourseKey)
			if err != nil {
				n.logger.Errorf("%s -> error reseting streaks for course config %s: %v", funcName, config.ID, err)
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
		completed := false
		userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, tzOffsets.GeneratePairs(config.StreaksNotificationsConfig.PreferEarly), &completed)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		configOffset := config.StreaksNotificationsConfig.TimezoneOffset
		offsetDiff := offset - configOffset
		offsetDiffPlusDay := offsetDiff + utils.SecondsInDay
		offsetDiffMinusDay := offsetDiff - utils.SecondsInDay
		if offsetDiff == 0 || offsetDiffPlusDay == 0 || offsetDiffMinusDay == 0 {
			// load all user courses for this course
			completed := false
			userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, nil, &completed)
			if err != nil {
				return nil, nil, nil, err
			}
		}
	}

	if len(userCourses) > 0 {
		// there should be one userCourse per userID (user cannot take the same course multiple times simultaneously)
		userIDs := make([]string, len(userCourses))
		for i, userCourse := range userCourses {
			userIDs[i] = userCourse.UserID
		}

		current := true
		userUnits, err := n.storage.FindUserUnits(config.AppID, config.OrgID, userIDs, config.CourseKey, nil, &current)
		if err != nil {
			return nil, nil, nil, err
		}
		currentUserUnits := make(map[string][]model.UserUnit)
		for _, userUnit := range userUnits {
			if currentUserUnits[userUnit.UserID] == nil {
				currentUserUnits[userUnit.UserID] = []model.UserUnit{}
			}
			currentUserUnits[userUnit.UserID] = append(currentUserUnits[userUnit.UserID], userUnit)
		}
		return userCourses, currentUserUnits, userIDs, nil
	}

	return nil, nil, nil, nil
}

func (n streaksNotifications) filterUsersByIncomplete(currentUserUnits map[string][]model.UserUnit, userIDs []string, now time.Time, streaksProcessTime int, notificationProcessTime int) ([]string, error) {
	filtered := make([]string, 0)
	for userID, userUnits := range currentUserUnits {
		incompleteTaskHandler := func(incompleteUserID string) error {
			filtered = append(filtered, incompleteUserID)
			return nil
		}

		// offsetSeconds is difference between next streak process time (start of "day") and notification process time
		offsetSeconds := 0
		if streaksProcessTime <= notificationProcessTime {
			offsetSeconds = utils.SecondsInDay - notificationProcessTime + streaksProcessTime
		} else {
			offsetSeconds = streaksProcessTime - notificationProcessTime
		}

		err := n.checkScheduleTaskCompletion(userID, userUnits, now, incompleteTaskHandler, -offsetSeconds/utils.SecondsInHour, nil, nil)
		if err != nil {
			n.logger.Errorf("processNotifications -> error checking task completion for userID %s: %v", userID, err)
			continue
		}
	}

	return filtered, nil
}

// iterate through array of current userUnits
// run completeTaskHandler on all qualifying userUnits, run incompleteTaskHandler if none qualify
func (n streaksNotifications) checkScheduleTaskCompletion(userID string, userUnits []model.UserUnit, now time.Time, incompleteTaskHandler func(string) error, incompleteTaskPeriodOffset int,
	completeTaskHandler func(interfaces.Storage, model.UserUnit, bool) error, completeUnitHandler func(interfaces.Storage, model.UserUnit) error) error {
	if incompleteTaskHandler == nil {
		return errors.ErrorData(logutils.StatusMissing, "incomplete task handler", nil)
	}

	allIncomplete := true
	// all storage operations for completed schedule items done here must be atomic
	transaction := func(storage interfaces.Storage) error {
		for _, userUnit := range userUnits {
			userScheduleItem, scheduleItem, _, isRequired := userUnit.GetScheduleItem("", true)
			if userScheduleItem == nil || scheduleItem == nil {
				return errors.ErrorData(logutils.StatusMissing, model.TypeScheduleItem, &logutils.FieldArgs{"current": true})
			}

			// if the current schedule item is not required, it means the user has not completed the first required schedule item after it
			if !isRequired || scheduleItem.Duration == nil {
				// return incompleteTaskHandler(userUnit)
				continue
			}

			//TODO: may need to change this check to handle user travelling, DST
			// check if the current schedule item is incomplete and current schedule item start date is missing or at least (24*days+offset) hours before now
			startDateOffset := (24*time.Duration(*scheduleItem.Duration) + time.Duration(incompleteTaskPeriodOffset)) * time.Hour
			if !userScheduleItem.IsComplete() && (userScheduleItem.DateStarted == nil || !userScheduleItem.DateStarted.Add(startDateOffset).After(now)) {
				// not completed within specified period
				// return incompleteTaskHandler(userUnit)
				continue
			} else {
				allIncomplete = false // user has completed the current schedule item in at least one current user unit
				if completeTaskHandler != nil {
					// completed within specified period
					remainsCurrent := (userUnit.Completed+1 < userUnit.Unit.Required)
					err := completeTaskHandler(storage, userUnit, remainsCurrent)
					if err != nil {
						return errors.WrapErrorAction("completing", model.TypeScheduleItem, &logutils.FieldArgs{"user_unit.id": userUnit.ID, "completed": userUnit.Completed}, err)
					}

					// user completed the current unit because the end of the schedule has been reached, so complete the unit
					if !remainsCurrent && completeUnitHandler != nil {
						err = completeUnitHandler(storage, userUnit)
						if err != nil {
							return errors.WrapErrorAction("completing", model.TypeUserUnit, &logutils.FieldArgs{"user_unit.id": userUnit.ID, "completed": userUnit.Completed}, err)
						}
					}
				}
			}
		}

		if allIncomplete {
			return incompleteTaskHandler(userID)
		}

		return nil
	}

	return n.storage.PerformTransaction(transaction)
}
