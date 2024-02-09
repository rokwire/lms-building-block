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

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	//TypeCourseConfig course config type
	TypeCourseConfig logutils.MessageDataType = "course config"
	//TypeStreaksNotificationsConfig streaks notifications config type
	TypeStreaksNotificationsConfig logutils.MessageDataType = "streaks notifications config"
	//TypeUserCourse user course type
	TypeUserCourse logutils.MessageDataType = "user course"
	//TypeUserUnit user unit type
	TypeUserUnit logutils.MessageDataType = "user unit"
	//TypeUserContent user content type
	TypeUserContent logutils.MessageDataType = "user content"
	//TypeCourse course type
	TypeCourse logutils.MessageDataType = "course"
	//TypeModule module type
	TypeModule logutils.MessageDataType = "module"
	//TypeUnit unit type
	TypeUnit logutils.MessageDataType = "unit"
	//TypeContent content type
	TypeContent logutils.MessageDataType = "content"
	//TypeScheduleItem schedule item type
	TypeScheduleItem logutils.MessageDataType = "schedule item"
	//TypeTimezone timezone type
	TypeTimezone logutils.MessageDataType = "timezone"

	//UserTimezone indicates the user's timezone should be used
	UserTimezone string = "user"
	//UserContentCompleteKey is the key into user data to check for task completion
	UserContentCompleteKey string = "complete"
)

// UserCourse represents a copy of a course that the user modifies as progress is made
type UserCourse struct {
	ID     string `json:"id"`
	AppID  string `json:"app_id"`
	OrgID  string `json:"org_id"`
	UserID string `json:"user_id"`

	Timezone // include user timezone info

	Streak         int         `json:"streak"`
	StreakResets   []time.Time `json:"streak_resets"`   // timestamps when the streak is reset for this course
	StreakRestarts []time.Time `json:"streak_restarts"` // timestamps when the streak is restarted for this course
	Pauses         int         `json:"pauses"`
	PauseProgress  int         `json:"pause_progress"`
	PauseUses      []time.Time `json:"pause_uses"` // timestamps when a pause is used for this course

	LastResponded *time.Time `json:"last_responded"`

	Course Course `json:"course"`

	DateCreated   time.Time  `json:"date_created"`
	DateUpdated   *time.Time `json:"date_updated"`
	DateCompleted *time.Time `json:"date_completed"`
	DateDropped   *time.Time `json:"date_dropped"`
}

// MostRecentStreakProcessTime gives the time when the most recent daily streak process ran for a user
func (u *UserCourse) MostRecentStreakProcessTime(now *time.Time, snConfig StreaksNotificationsConfig) *time.Time {
	if u == nil {
		return nil
	}
	if now == nil {
		newNow := time.Now()
		now = &newNow
	}

	var loc *time.Location
	var err error
	if snConfig.TimezoneName == UserTimezone {
		loc = time.FixedZone(u.Timezone.Name, u.Timezone.Offset)
	} else {
		loc, err = time.LoadLocation(snConfig.TimezoneName)
		if err != nil {
			loc = time.FixedZone(snConfig.TimezoneName, snConfig.TimezoneOffset)
		}
	}
	nowLocal := now.In(loc)
	nowLocalSeconds := utils.SecondsInHour*nowLocal.Hour() + 60*nowLocal.Minute() + nowLocal.Second()

	hour := snConfig.StreaksProcessTime / utils.SecondsInHour
	minute := (snConfig.StreaksProcessTime % utils.SecondsInHour) / 60
	second := (snConfig.StreaksProcessTime % utils.SecondsInHour) % 60
	mostRecent := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), hour, minute, second, 0, loc).UTC()
	if nowLocalSeconds < snConfig.StreaksProcessTime {
		// go back one day if the current moment is before the process time in the current day
		mostRecent = mostRecent.Add(time.Duration(-24) * time.Hour)
	}
	return &mostRecent
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
	if c == nil {
		return nil
	}

	returnNextUnit := false
	for _, module := range c.Modules {
		for _, unit := range module.Units {
			if returnNextUnit {
				nextUnit := unit
				return &nextUnit
			}
			if unit.Key == currentUnitKey {
				returnNextUnit = true
			}
		}
	}
	return nil
}

// NextRequiredScheduleItem returns the next required schedule item in a course given the current unit key and schedule index (set allowCurrent to return the schedule item corresponding to scheduleIndex if required)
func (c *Course) NextRequiredScheduleItem(currentUnitKey string, scheduleIndex int, allowCurrent bool) *ScheduleItem {
	//TODO: not working
	returnNextRequired := false
	for _, module := range c.Modules {
		for _, unit := range module.Units {
			if returnNextRequired || unit.Key == currentUnitKey {
				for j, scheduleItem := range unit.Schedule {
					if (returnNextRequired || j > scheduleIndex || (allowCurrent && j == scheduleIndex)) && j >= unit.ScheduleStart {
						nextScheduleItem := scheduleItem
						return &nextScheduleItem
					}
				}
				returnNextRequired = true
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

	InitialPauses       int `json:"initial_pauses" bson:"initial_pauses"`
	MaxPauses           int `json:"max_pauses" bson:"max_pauses"`
	PauseProgressReward int `json:"pause_progress_reward" bson:"pause_progress_reward"`

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

// ValidateTimings checks timezone information is valid and streaks and notifications process times are valid
func (c *StreaksNotificationsConfig) ValidateTimings() error {
	if c == nil {
		return errors.ErrorData(logutils.StatusMissing, TypeStreaksNotificationsConfig, nil)
	}

	if c.TimezoneName != UserTimezone {
		timezone := Timezone{Name: c.TimezoneName, Offset: c.TimezoneOffset}
		err := timezone.Validate()
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionValidate, "streaks and notifications timezone", nil, err)
		}
		c.TimezoneOffset = timezone.Offset
	}
	if c.StreaksProcessTime < 0 || c.StreaksProcessTime >= utils.SecondsInDay || c.StreaksProcessTime%utils.SecondsInHour != 0 {
		return errors.ErrorData(logutils.StatusInvalid, "streaks process time", &logutils.FieldArgs{"streaks_process_time": c.StreaksProcessTime})
	}
	for _, notification := range c.Notifications {
		if notification.ProcessTime < 0 || notification.ProcessTime >= utils.SecondsInDay || c.StreaksProcessTime%utils.SecondsInHour != 0 {
			return errors.ErrorData(logutils.StatusInvalid, "notification process time", &logutils.FieldArgs{"process_time": notification.ProcessTime})
		}
	}

	return nil
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

	Styles Styles `json:"styles"`

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

	LastCompleted *time.Time `json:"last_completed"` // when the last required task of the previous unit was completed
	DateCreated   time.Time  `json:"date_created"`
	DateUpdated   *time.Time `json:"date_updated"`
}

// CurrentScheduleItem returns a pointer to the schedule item in the unit the user is working on
func (u *UserUnit) CurrentScheduleItem() *ScheduleItem {
	if u == nil {
		return nil
	}
	if u.Completed < 0 || u.Completed >= u.Unit.Required {
		return nil
	}

	return &u.Unit.Schedule[u.Completed]
}

// PreviousScheduleItem returns a pointer to the schedule item in the unit immediately before the current on
func (u *UserUnit) PreviousScheduleItem() *ScheduleItem {
	if u == nil {
		return nil
	}
	if u.Completed <= 0 || u.Completed > u.Unit.Required {
		return nil
	}

	return &u.Unit.Schedule[u.Completed-1]
}

// IsCurrentScheduleItemRequired returns whether the current schedule item is required
func (u *UserUnit) IsCurrentScheduleItemRequired() *bool {
	if u == nil {
		return nil
	}
	if u.Completed < 0 || u.Completed >= u.Unit.Required {
		return nil
	}

	isRequired := u.Completed >= u.Unit.ScheduleStart
	return &isRequired
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

	ScheduleStart int `json:"schedule_start"` // index of the first schedule item the user should submit data for
	Required      int `json:"required"`       // number of schedule items required to be completed (may add required flags to each schedule item in future)

	DateCreated time.Time  `json:"-"`
	DateUpdated *time.Time `json:"-"`
}

// Validate checks the unit schedule to make sure it is valid
func (u *Unit) Validate(contentKeys []string) error {
	if u == nil {
		return errors.ErrorData(logutils.StatusMissing, TypeUnit, nil)
	}
	if len(u.Schedule) == 0 {
		return errors.ErrorData(logutils.StatusMissing, "unit schedule", &logutils.FieldArgs{"key": u.Key})
	}
	if u.ScheduleStart < 0 || u.ScheduleStart >= len(u.Schedule) {
		return errors.ErrorData(logutils.StatusInvalid, "unit schedule start", &logutils.FieldArgs{"schedule_start": u.ScheduleStart})
	}
	if len(contentKeys) == 0 {
		contentKeys = make([]string, 0)
		for _, content := range u.Contents {
			if !utils.Exist[string](contentKeys, content.Key) {
				contentKeys = append(contentKeys, content.Key)
			}
		}
	}

	for i, item := range u.Schedule {
		if i >= u.ScheduleStart && item.Duration == nil {
			return errors.ErrorData(logutils.StatusInvalid, TypeScheduleItem, &logutils.FieldArgs{"name": item.Name, "duration": item.Duration})
		}

		for _, userContent := range item.UserContent {
			if !utils.Exist[string](contentKeys, userContent.ContentKey) {
				return errors.ErrorData(logutils.StatusInvalid, "schedule content key", &logutils.FieldArgs{"content_key": userContent.ContentKey})
			}
		}
	}

	u.Required = len(u.Schedule)
	return nil
}

// UserContentWithTimezone wraps unit with time information
type UserContentWithTimezone struct {
	UserContent UserContent `json:"user_content"`
	Timezone                // include user timezone info
}

// ScheduleItem represents a set of Content items to be completed in a certain amount of time
type ScheduleItem struct {
	Name        string        `json:"name" bson:"name"`
	UserContent []UserContent `json:"user_content" bson:"user_content"`
	Duration    *int          `json:"duration" bson:"duration,omitempty"` // in days

	DateStarted   *time.Time `json:"date_started,omitempty" bson:"date_started,omitempty"`
	DateCompleted *time.Time `json:"date_completed,omitempty" bson:"date_completed,omitempty"`
}

// UpdateUserData updates the stored data for the user content matching item.ContentKey in the schedule item
func (s *ScheduleItem) UpdateUserData(item UserContent) error {
	if s == nil {
		return errors.ErrorData(logutils.StatusMissing, TypeScheduleItem, nil)
	}
	for i, userContent := range s.UserContent {
		if userContent.ContentKey == item.ContentKey {
			s.UserContent[i].UserData = item.UserData
			return nil
		}
	}

	return errors.ErrorData(logutils.StatusMissing, TypeUserContent, &logutils.FieldArgs{"content_key": item.ContentKey})
}

// IsComplete gives whether every user content item in the schedule item has user data
func (s *ScheduleItem) IsComplete() bool {
	for _, userContent := range s.UserContent {
		if !userContent.IsComplete() {
			return false
		}
	}
	return true
}

// Content represents some Unit content
type Content struct {
	ID    string `json:"id" bson:"_id"`
	AppID string `json:"app_id" bson:"app_id"`
	OrgID string `json:"org_id" bson:"org_id"`

	Key           string    `json:"key" bson:"key"`
	Type          string    `json:"type" bson:"type"` // assignment, resource, reward, evaluation
	Name          string    `json:"name" bson:"name"`
	Details       string    `json:"details" bson:"details"`
	Reference     Reference `json:"reference" bson:"reference"`
	LinkedContent []string  `json:"linked_content" bson:"linked_content"`

	Styles Styles `json:"styles" bson:"styles"`

	DateCreated time.Time  `json:"-" bson:"date_created"`
	DateUpdated *time.Time `json:"-" bson:"date_updated"`
}

// Reference represents a reference to another entity
type Reference struct {
	Name         string `json:"name" bson:"name"`
	Type         string `json:"type" bson:"type"` // content item, video, PDF, survey, web URL
	ReferenceKey string `json:"reference_key" bson:"reference_key"`
}

// UserContent represents a Content reference with some additional user data
type UserContent struct {
	ContentKey string `json:"content_key" bson:"content_key"`

	// user fields (populated as user takes a course)
	UserData map[string]interface{} `json:"user_data,omitempty" bson:"user_data,omitempty"`
}

// IsComplete gives whether the user has completed the content corresponding to ContentKey
func (uc *UserContent) IsComplete() bool {
	if uc == nil {
		return false
	}

	completeVal, ok := uc.UserData[UserContentCompleteKey]
	if !ok {
		return false
	}
	complete, ok := completeVal.(bool)
	if !ok {
		return false
	}
	return complete
}

// Timezone represents user timezone information received from the client
type Timezone struct {
	Name   string `json:"timezone_name"`
	Offset int    `json:"timezone_offset"` // in seconds east of UTC
}

// Validate checks whether name and offset refer to a valid timezone (sets offset if name is valid but offset is not)
func (t *Timezone) Validate() error {
	if t == nil {
		return errors.ErrorData(logutils.StatusMissing, "timezone", nil)
	}

	if t.Offset < utils.MinTZOffset || t.Offset > utils.MaxTZOffset {
		tzLoc, err := time.LoadLocation(t.Name)
		if err != nil {
			return errors.WrapErrorData(logutils.StatusInvalid, "user timezone", &logutils.FieldArgs{"name": t.Name, "offset": t.Offset}, err)
		}

		// set the offset if it was invalid, but could load location from name
		_, t.Offset = time.Now().In(tzLoc).Zone()
	}

	return nil
}

// TZOffsets entity represents a set of single timezone offsets
type TZOffsets []int

// GeneratePairs gives the set of offset pairs to use for search according to preferEarly
func (tz TZOffsets) GeneratePairs(preferEarly bool) []TZOffsetPair {
	pairs := make([]TZOffsetPair, len(tz))
	for i, offset := range tz {
		if preferEarly {
			pairs[i] = TZOffsetPair{Lower: offset - utils.SecondsInHour + 1, Upper: offset}
		} else {
			pairs[i] = TZOffsetPair{Lower: offset, Upper: offset + utils.SecondsInHour - 1}
		}
	}
	return pairs
}

// TZOffsetPair represents a set of timezone offset ranges used to find users when sending notifications
type TZOffsetPair struct {
	Lower int
	Upper int
}

// Styles represents data used to determine how to display course data in the client
type Styles struct {
	Colors  map[string]interface{} `json:"colors,omitempty" bson:"colors,omitempty"`
	Images  map[string]interface{} `json:"images,omitempty" bson:"images,omitempty"`
	Strings map[string]interface{} `json:"strings,omitempty" bson:"strings,omitempty"`
}
