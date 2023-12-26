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
	"lms/utils"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"
)

const (
	secondsInDay int = 86400
	minTZOffset  int = -43200
	maxTZOffset  int = 50400
)

type streaksNotifications struct {
	logger *logs.Logger

	notificationsBB interfaces.NotificationsBB

	storage interfaces.Storage

	//notifications timer
	notificationsTimer     *time.Timer
	notificationsTimerDone chan bool
	//TODO: add streaks timer
}

func (n streaksNotifications) start() {
	//setup hourly notifications timer
	go n.setupNotificationsTimer()

	//TODO: setup daily streaks timer
}

func (n streaksNotifications) setupNotificationsTimer() {
	now := time.Now().UTC()
	nowSecondsInHour := now.Second() // 60*now.Minute() + now.Second()
	desiredMoment := 0               //default desired moment of the hour in seconds (beginning of the hour)

	var durationInSeconds int
	n.logger.Infof("setupNotificationsTimer -> nowSecondsInHour:%d", nowSecondsInHour)
	if nowSecondsInHour <= desiredMoment {
		n.logger.Info("setupNotificationsTimer -> notifications not yet processed this hour")
		durationInSeconds = desiredMoment - nowSecondsInHour
	} else {
		n.logger.Info("setupNotificationsTimer -> notifications have already been processed this hour")
		// durationInSeconds = (3600 - nowSecondsInHour) + desiredMoment // the time which left this hour + desired moment from next hour
		durationInSeconds = (60 - nowSecondsInHour) + desiredMoment // the time which left this hour + desired moment from next hour
	}

	initialDuration := time.Second * time.Duration(durationInSeconds)
	utils.StartTimer(n.notificationsTimer, n.notificationsTimerDone, &initialDuration, time.Second, n.processHourlyNotifications, "processHourlyNotifications", n.logger)
}

func (n streaksNotifications) processHourlyNotifications() {
	now := time.Now().UTC()
	nowSeconds := (now.Second() % 24) * 3600 // 60*60*now.Hour() + 60*now.Minute() + now.Second()

	// active := true
	// courseConfigs, err := n.storage.FindCourseConfigs(&active)
	// if err != nil {
	// 	n.logger.Errorf("processHourlyNotifications -> error finding active course configs: %v", err)
	// 	return
	// }
	// if len(courseConfigs) == 0 {
	// 	n.logger.Errorf("processHourlyNotifications -> no active course configs for notifications")
	// 	return
	// }

	courseConfigs := []model.CourseConfig{
		{CourseKey: "test1", NotificationsConfig: model.NotificationsConfig{
			TimezoneName:   "America/Chicago",
			TimezoneOffset: -21600, //CST (UTC-6)
			Active:         true,
			Notifications: []model.Notification{
				{Subject: "notifyMorning", ProcessTime: 25200, Active: true}, // 7AM
				{Subject: "notifyNoon", ProcessTime: 43200, Active: true},    // 12PM
				{Subject: "notifyEvening", ProcessTime: 64800, Active: true}, // 6PM (not notifying for CST)
			},
		}},
	}

	for _, config := range courseConfigs {
		offsets := make([]int, 0)
		for _, notification := range config.NotificationsConfig.Notifications {
			if config.NotificationsConfig.TimezoneName != "user" && config.NotificationsConfig.TimezoneOffset != 0 {
				//TODO: handle single timezones other than UTC
			}

			lowerOffset := notification.ProcessTime - nowSeconds
			if config.NotificationsConfig.TimezoneName == "user" {
				if lowerOffset >= minTZOffset && lowerOffset <= maxTZOffset {
					offsets = append(offsets, lowerOffset)
				}

				if lowerOffset+secondsInDay <= maxTZOffset {
					offsets = append(offsets, lowerOffset+secondsInDay)
				}
				if lowerOffset-secondsInDay >= minTZOffset {
					offsets = append(offsets, lowerOffset-secondsInDay)
				}
			} else {
				configOffset := config.NotificationsConfig.TimezoneOffset
				if lowerOffset == configOffset || lowerOffset+secondsInDay == configOffset || lowerOffset-secondsInDay == configOffset {
					//TODO: get all user courses
					n.logger.Infof("hour %d: notify %s %d", nowSeconds/3600, config.NotificationsConfig.TimezoneName, config.NotificationsConfig.TimezoneOffset)
				}
			}

			// userCourses
			// _, err := n.storage.FindUserCourses(nil, nil, []string{config.CourseKey}, nil, offsets)
			// if err != nil {
			// 	n.logger.Errorf("processHourlyNotifications -> error finding user courses for course key %s: %v", config.CourseKey, err)
			// 	continue
			// }

			if notification.Active {
				//TODO: check whether userCourse meets requirements for notification to be sent
			}
		}
		n.logger.Infof("hour %d: offsets: %v", nowSeconds/3600, offsets)
	}
}
