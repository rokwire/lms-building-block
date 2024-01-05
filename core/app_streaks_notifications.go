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

	"github.com/rokwire/logging-library-go/v2/logs"
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
	utils.StartTimer(n.notificationsTimer, n.notificationsTimerDone, &initialDuration, time.Hour, n.processNotifications, "processNotifications", n.logger)
}

func (n streaksNotifications) processNotifications() {
	now := time.Now().UTC()
	nowSeconds := 60*60*now.Hour() + 60*now.Minute() + now.Second()

	active := true
	courseConfigs, err := n.storage.FindCourseConfigs(&active)
	if err != nil {
		n.logger.Errorf("processNotifications -> error finding active course configs: %v", err)
		return
	}
	if len(courseConfigs) == 0 {
		n.logger.Errorf("processNotifications -> no active course configs for notifications")
		return
	}

	for _, config := range courseConfigs {
		for _, notification := range config.StreaksNotificationsConfig.Notifications {
			if notification.Active {
				tzOffsets := make(model.TZOffsets, 0)
				var userCourses []model.UserCourse

				offset := notification.ProcessTime - nowSeconds
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
					userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, tzOffsets.GeneratePairs(config.StreaksNotificationsConfig.PreferEarly), notification.Requirements)
					if err != nil {
						n.logger.Errorf("processNotifications -> error finding user courses for course key %s: %v", config.CourseKey, err)
						continue
					}
				} else {
					configOffset := config.StreaksNotificationsConfig.TimezoneOffset
					if offset == configOffset || offset+utils.SecondsInDay == configOffset || offset-utils.SecondsInDay == configOffset {
						// load all user courses for this course
						userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, nil, notification.Requirements)
						if err != nil {
							n.logger.Errorf("processNotifications -> error finding user courses for course key %s: %v", config.CourseKey, err)
							continue
						}
					}
				}

				recipients := make([]notifications.Recipient, 0)
				for _, userCourse := range userCourses {
					recipients = append(recipients, notifications.Recipient{UserID: userCourse.UserID})
				}
				if len(recipients) == 0 {
					n.logger.Infof("processNotifications -> no recipients for notification %s for course key %s", notification.Subject, config.CourseKey)
					continue
				}

				err = n.notificationsBB.SendNotifications(recipients, notification.Subject, notification.Body, notification.Params)
				if err != nil {
					n.logger.Errorf("processNotifications -> error sending notification %s for course key %s: %v", notification.Subject, config.CourseKey, err)
					continue
				}
				n.logger.Infof("processNotifications -> sent notification %s for course key %s", notification.Subject, config.CourseKey)
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
	initialDuration = 60
	utils.StartTimer(n.streaksTimer, n.streaksTimerDone, &initialDuration, time.Minute, n.processStreaks, "processStreaks", n.logger)
	//utils.StartTimer(n.streaksTimer, n.streaksTimerDone, &initialDuration, time.Hour, n.processStreaks, "processStreaks", n.logger)

}

func (n streaksNotifications) processStreaks() {
	now := time.Now().UTC()
	nowSeconds := 60*60*now.Hour() + 60*now.Minute() + now.Second()

	courseConfigs, err := n.storage.FindCourseConfigs(nil)
	if err != nil {
		n.logger.Errorf("processStreaks -> error finding active course configs: %v", err)
		return
	}
	if len(courseConfigs) == 0 {
		n.logger.Errorf("processStreaks -> no active course configs for streaks")
		return
	}

	// completedtask
	// when complete usercourseunitprogress, get utc convert to local using timezone name and offset, store date
	// when we revist in process streaks, if same date, we ignore. if yesterday then increment

	// for every config in course_configs:
	// 		retrieve usercourse that passes 12am local time (how?) (also what if china to usa, experiencing same day mightnight twice)
	//		if CompletedTasks? (refer to comments above):
	//			streaks + 1; pause + 1
	//			if usercourse's new pause > max_pause:
	//				set to max_pause
	//		else (incomplete task):
	//			pause - 1
	//			if pause < 0:
	//				streak = 0; pause = 0
	//		updateUserCourses([usercourses struct]) (why db usercourse doesn't have timezone?)
	//		reset CompletedTasks?
	// remember to assign initial pause when creating usercourse struct

	for _, config := range courseConfigs {
		tzOffsets := make(model.TZOffsets, 0)
		var userCourses []model.UserCourse
		// 0 for streak processtime
		offset := 0 - nowSeconds
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
			userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, tzOffsets.GeneratePairs(config.StreaksNotificationsConfig.PreferEarly), nil)
			if err != nil {
				n.logger.Errorf("processStreaks -> error finding user courses for course key %s: %v", config.CourseKey, err)
				continue
			}
		} else {
			configOffset := config.StreaksNotificationsConfig.TimezoneOffset
			if offset == configOffset || offset+utils.SecondsInDay == configOffset || offset-utils.SecondsInDay == configOffset {
				// load all user courses for this course
				userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, nil, nil)
				if err != nil {
					n.logger.Errorf("processStreaks -> error finding user courses for course key %s: %v", config.CourseKey, err)
					continue
				}
			}
		}

		for i, userCourse := range userCourses {
			if userCourse.CompletedTasks != nil {
				uY, uM, uD := userCourse.CompletedTasks.Date()
				y, m, d := now.Date()
				pastY, pastM, pastD := now.AddDate(0, 0, -1).Date()
				// edge case between time zone traveling
				if uY == y && uM == m && uD == d {
					continue
					// yesterday
				} else if uY == pastY && uM == pastM && uD == pastD {
					userCourse.Streaks += 1
					userCourse.Pauses += 1
					if config.MaxPauses <= userCourse.Pauses {
						userCourse.Pauses = config.MaxPauses
					}
					// date earlier than yesterday, streak is broken
				} else {
					userCourse.Pauses -= 1
					if userCourse.Pauses < 0 {
						userCourse.Streaks = 0
						userCourse.Pauses = 0
					}
				}
				// User started a course and not never made any progress
			} else {
				userCourse.Pauses -= 1
				if userCourse.Pauses < 0 {
					userCourse.Streaks = 0
					userCourse.Pauses = 0
				}
			}
			userCourses[i] = userCourse
		}
		if len(userCourses) != 0 {
			err = n.UpdateManyUserCoursesStreaks(config.AppID, config.OrgID, userCourses)
			if err != nil {
				n.logger.Errorf("processStreaks -> error updating user courses for course key %s: %v", config.CourseKey, err)
			}
		}
	}

}

func (n streaksNotifications) UpdateManyUserCoursesStreaks(appID string, orgID string, UserCourses []model.UserCourse) error {
	for _, userCourse := range UserCourses {
		// do we use userCourse id or userID+coursekey
		err := n.storage.UpdateUserCourseStreaks(appID, orgID, nil, &userCourse.ID, userCourse.Course.Key, &userCourse.Streaks, &userCourse.Pauses, nil)
		if err != nil {
			return err
		}
		if err != nil {
			n.logger.Errorf("processStreaks -> error updating user courses for course id %s: %v", userCourse.ID, err)
			continue
		}
	}
	return nil
}
