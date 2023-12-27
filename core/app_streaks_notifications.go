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
		for _, notification := range config.NotificationsConfig.Notifications {
			if notification.Active {
				tzOffsets := make(model.TZOffsets, 0)
				var userCourses []model.UserCourse

				offset := notification.ProcessTime - nowSeconds
				if config.NotificationsConfig.TimezoneName == "user" {
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
					userCourses, err = n.storage.FindUserCourses(nil, config.AppID, config.OrgID, nil, []string{config.CourseKey}, nil, tzOffsets.GeneratePairs(notification.PreferEarly), notification.Requirements)
					if err != nil {
						n.logger.Errorf("processNotifications -> error finding user courses for course key %s: %v", config.CourseKey, err)
						continue
					}
				} else {
					configOffset := config.NotificationsConfig.TimezoneOffset
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
}
