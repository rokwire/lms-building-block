// Copyright 2023 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 the "License";
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"lms/utils"
	"time"

	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	//TypeCourseConfig course config type
	TypeCourseConfig logutils.MessageDataType = "course config"
	//TypeUserCourse user course type
	TypeUserCourse logutils.MessageDataType = "user course"
	//TypeUserUnit user unit type
	TypeUserUnit logutils.MessageDataType = "user unit"
	//TypeCourse course type
	TypeCourse logutils.MessageDataType = "course"
	//TypeModule module type
	TypeModule logutils.MessageDataType = "module"
	//TypeUnit unit type
	TypeUnit logutils.MessageDataType = "unit"
	//TypeUnitWithTimezone unit with timezone type
	TypeUnitWithTimezone logutils.MessageDataType = "unit with timezone"
	//TypeContent content type
	TypeContent logutils.MessageDataType = "content"
	//TypeTimezone timezone type
	TypeTimezone logutils.MessageDataType = "timezone"
)

// UserCourse represents a copy of a course that the user modifies as progress is made
type UserCourse struct {
	ID     string `json:"id"`
	AppID  string `json:"app_id"`
	OrgID  string `json:"org_id"`
	UserID string `json:"user_id"`

	Timezone // include user timezone info

	Streak       int         `json:"streak"`
	StreakResets []time.Time `json:"streak_resets"` // timestamps when the streak is reset for this course
	Pauses       int         `json:"pauses"`
	PauseUses    []time.Time `json:"pause_uses"` // timestamps when a pause is used for this course

	Course Course `json:"course"`

	DateCreated time.Time  `json:"date_created"`
	DateUpdated *time.Time `json:"date_updated"`
	DateDropped *time.Time `json:"date_dropped"`
}

// Course represents a custom-defined course (e.g. Essential Skills Coaching)
type Course struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`
	OrgID string `json:"org_id"`

	Key     string   `json:"key"`
	Name    string   `json:"name"`
	Modules []Module `json:"modules"`

	DateCreated time.Time  `json:"-"`
	DateUpdated *time.Time `json:"-"`
}

// GetNextUnit returns the next unit in a course given the current unit key
func (c *Course) GetNextUnit(currentUnitKey string) *Unit {
	returnNextModuleUnit := false
	for _, module := range c.Modules {
		for i, unit := range module.Units {
			if returnNextModuleUnit {
				nextUnit := unit
				return &nextUnit
			}
			if unit.Key == currentUnitKey {
				if i+1 < len(module.Units) {
					nextUnit := module.Units[i+1]
					return &nextUnit
				}
				returnNextModuleUnit = true
			}
		}
	}
	return nil
}

// CourseConfig represents streak and notification settings for a course
type CourseConfig struct {
	ID        string `json:"id" bson:"_id"`
	AppID     string `json:"app_id" bson:"app_id"`
	OrgID     string `json:"org_id" bson:"org_id"`
	CourseKey string `json:"course_key" bson:"course_key"`

	InitialPauses     int `json:"initial_pauses" bson:"initial_pauses"`
	MaxPauses         int `json:"max_pauses" bson:"max_pauses"`
	PauseRewardStreak int `json:"pause_reward_streak" bson:"pause_reward_streak"`

	StreaksNotificationsConfig StreaksNotificationsConfig `json:"streaks_notifications_config" bson:"streaks_notifications_config"`

	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// StreaksNotificationsConfig entity
type StreaksNotificationsConfig struct {
	TimezoneName   string `json:"timezone_name" bson:"timezone_name"`     // either an IANA timezone database identifier or "user" to for users' most recent known timezone
	TimezoneOffset int    `json:"timezone_offset" bson:"timezone_offset"` // in seconds east of UTC (only valid if TimezoneName is not "user")

	StreaksProcessTime  int  `json:"streaks_process_time" bson:"streaks_process_time"` // seconds since midnight in selected timezone at which to process streaks
	PreferEarly         bool `json:"prefer_early" bson:"prefer_early"`                 // whether notification should be sent early or late if it cannot be sent at exactly ProcessTime
	NotificationsActive bool `json:"notifications_active" bson:"notifications_active"` // if the notifications processing is "on" or "off"
	// BlockSize int    `json:"block_size" bson:"block_size"` // TODO: needed?
	NotificationsMode string `json:"notifications_mode" bson:"notifications_mode"` // "normal" or "test"

	Notifications []Notification `json:"notifications" bson:"notifications"`
}

// Notification entity
type Notification struct {
	Subject string             `json:"subject" bson:"subject"` // e.g., "Daily task reminder" (a.k.a. "text")
	Body    string             `json:"body" bson:"body"`       // e.g., "Remember to complete your daily task."
	Params  NotificationParams `json:"params" bson:"params"`   // Notification specific settings (include deep link string if needed)

	ProcessTime int `json:"process_time" bson:"process_time"` // seconds since midnight in selected timezone at which to process notifications
	//PreferEarly bool `json:"prefer_early" bson:"prefer_early"` // whether notification should be sent early or late if it cannot be sent at exactly ProcessTime

	Active bool `json:"active" bson:"active"`

	Requirements map[string]interface{} `json:"requirements" bson:"requirements"`
}

// NotificationParams entity
type NotificationParams map[string]string

// Module represents an individual module of a Course (e.g. Conversational Skills)
type Module struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`
	OrgID string `json:"org_id"`

	Key   string `json:"key"`
	Name  string `json:"name"`
	Units []Unit `json:"units"`

	DateCreated time.Time  `json:"-"`
	DateUpdated *time.Time `json:"-"`
}

// UserUnit represents a copy of a unit that the user modifies as progress is made
type UserUnit struct {
	ID        string `json:"id"`
	AppID     string `json:"app_id"`
	OrgID     string `json:"org_id"`
	UserID    string `json:"user_id"`
	CourseKey string `json:"course_key"`
	Unit      Unit   `json:"unit"`

	Completed int  `json:"completed"` // number of schedule items the user has completed
	Current   bool `json:"current"`

	LastCompleted *time.Time `json:"last_completed"`
	DateCreated   time.Time  `json:"date_created"`
	DateUpdated   *time.Time `json:"date_updated"`
}

// Unit represents an individual unit of a Module (e.g. The Physical Side of Communication)
type Unit struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`
	OrgID string `json:"org_id"`

	Key      string         `json:"key"`
	Name     string         `json:"name"`
	Contents []Content      `json:"content"`
	Schedule []ScheduleItem `json:"schedule"`

	Required int `json:"required"` // number of schedule items required to be completed (may add required flags to each schedule item in future)

	DateCreated time.Time  `json:"-"`
	DateUpdated *time.Time `json:"-"`
}

// UnitWithTimezone wraps unit with time information
type UnitWithTimezone struct {
	Unit     Unit `json:"unit"`
	Timezone      // include user timezone info
}

// ScheduleItem represents a set of Content items to be completed in a certain amount of time
type ScheduleItem struct {
	Name        string          `bson:"name" json:"name"`
	UserContent []UserReference `bson:"user_content" json:"user_content"`
	Duration    int             `bson:"duration" json:"duration"` // in days
}

// Content represents some Unit content
type Content struct {
	ID    string `json:"id" bson:"_id"`
	AppID string `json:"app_id" bson:"app_id"`
	OrgID string `json:"org_id" bson:"org_id"`

	Key              string    `json:"key" bson:"key"`
	Type             string    `json:"type" bson:"type"` // assignment, resource, reward, evaluation
	Name             string    `json:"name" bson:"name"`
	Details          string    `json:"details" bson:"details"`
	ContentReference Reference `json:"reference" bson:"reference"`
	LinkedContent    []string  `json:"linked_content" bson:"linked_content"`

	DateCreated time.Time  `json:"-" bson:"date_created"`
	DateUpdated *time.Time `json:"-" bson:"date_updated"`
}

// Reference represents a reference to another entity
type Reference struct {
	Name         string `bson:"name" json:"name"`
	Type         string `bson:"type" json:"type"` // content item, video, PDF, survey, web URL
	ReferenceKey string `bson:"reference_key" json:"reference_key"`
}

// UserReference represents a reference with some additional data about user interactions
type UserReference struct {
	Reference

	// user fields (populated as user takes a course)
	UserData      map[string]interface{} `bson:"user_data,omitempty" json:"user_data,omitempty"`
	DateStarted   *time.Time             `bson:"date_started,omitempty" json:"date_started,omitempty"`
	DateCompleted *time.Time             `bson:"date_completed,omitempty" json:"date_completed,omitempty"`
}

// Timezone represents user timezone information received from the client
type Timezone struct {
	Name   string `json:"timezone_name"`
	Offset int    `json:"timezone_offset"` // in seconds east of UTC
}

// TZOffsets entity represents a set of single timezone offsets
type TZOffsets []int

// GeneratePairs gives the set of offset pairs to use for search according to preferEarly
func (tz TZOffsets) GeneratePairs(preferEarly bool) []TZOffsetPair {
	pairs := make([]TZOffsetPair, len(tz))
	for i, offset := range tz {
		if preferEarly {
			pairs[i] = TZOffsetPair{Lower: offset, Upper: offset + utils.SecondsInHour - 1}
		} else {
			pairs[i] = TZOffsetPair{Lower: offset - utils.SecondsInHour + 1, Upper: offset}
		}
	}
	return pairs
}

// TZOffsetPair represents a set of timezone offset ranges used to find users when sending notifications
type TZOffsetPair struct {
	Lower int
	Upper int
}
