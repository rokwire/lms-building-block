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
	//TypeUserContentReference user content reference type
	TypeUserContentReference logutils.MessageDataType = "user content reference"
	//TypeUserResponse user response type
	TypeUserResponse logutils.MessageDataType = "user response"
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

// CanMakePauseProgress returns whether the last user response was before the most recent streak process
func (u *UserCourse) CanMakePauseProgress(now *time.Time, snConfig StreaksNotificationsConfig) bool {
	if u == nil {
		return false
	}

	lastStreakProcess := u.MostRecentStreakProcessTime(now, snConfig)
	if lastStreakProcess == nil {
		return false
	}
	return u.LastResponded == nil || u.LastResponded.Before(*lastStreakProcess)
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
	returnNextRequired := false
	for _, module := range c.Modules {
		for _, unit := range module.Units {
			if returnNextRequired || unit.Key == currentUnitKey {
				for j, scheduleItem := range unit.Schedule {
					if (returnNextRequired || j > scheduleIndex || (allowCurrent && j == scheduleIndex)) && scheduleItem.IsRequired() {
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
			return errors.WrapErrorAction(logutils.ActionValidate, TypeTimezone, nil, err)
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

	Completed    int                `json:"completed"` // number of schedule items the user has completed
	Current      bool               `json:"current"`
	UserSchedule []UserScheduleItem `json:"user_schedule"`

	LastCompleted *time.Time `json:"last_completed"` // when the last required task of the previous unit was completed
	DateCreated   time.Time  `json:"date_created"`
	DateUpdated   *time.Time `json:"date_updated"`
}

// GetScheduleItem returns pointers, current status, and required status for the UserScheduleItem and ScheduleItem requested by contentKey or forceCurrent
func (u *UserUnit) GetScheduleItem(contentKey string, forceCurrent bool) (*UserScheduleItem, *ScheduleItem, bool, bool) {
	if u == nil {
		return nil, nil, false, false
	}
	if u.Completed < 0 || u.Completed >= len(u.Unit.Schedule) || u.Completed >= len(u.UserSchedule) {
		return nil, nil, false, false
	}

	if forceCurrent {
		unitScheduleItem := u.Unit.Schedule[u.Completed]
		return &u.UserSchedule[u.Completed], &unitScheduleItem, true, unitScheduleItem.IsRequired()
	}

	for i, item := range u.Unit.Schedule {
		for _, key := range item.ContentKeys {
			if key == contentKey {
				isCurrent := u.Current && (i == u.Completed)
				return &u.UserSchedule[i], &item, isCurrent, item.IsRequired()
			}
		}
	}

	return nil, nil, false, false
}

// GetPreviousScheduleItem returns a pointer to a UserScheduleItem in the unit prior the current position based on forceRequired
func (u *UserUnit) GetPreviousScheduleItem(forceRequired bool) *UserScheduleItem {
	if u == nil {
		return nil
	}
	if u.Completed <= 0 || u.Completed > len(u.Unit.Schedule) || u.Completed > len(u.UserSchedule) {
		return nil
	}

	for i := 1; i <= u.Completed; i++ {
		if forceRequired && !u.Unit.Schedule[u.Completed-i].IsRequired() {
			continue
		}
		return &u.UserSchedule[u.Completed-i]
	}
	return nil
}

// GetNextScheduleItem returns a pointer to a UserScheduleItem in the unit after the current position based on forceRequired
func (u *UserUnit) GetNextScheduleItem(forceRequired bool) *UserScheduleItem {
	if u == nil {
		return nil
	}
	if u.Completed < -1 || u.Completed+1 > len(u.Unit.Schedule) || u.Completed+1 > len(u.UserSchedule) {
		return nil
	}

	for i := u.Completed + 1; i < u.Unit.Required; i++ {
		if forceRequired && !u.Unit.Schedule[i].IsRequired() {
			continue
		}
		return &u.UserSchedule[i]
	}
	return nil
}

// GetUserContentReferenceForKey returns the user content reference in the user schedule containing contentKey and whether the reference is in the current UserScheduleItem
func (u *UserUnit) GetUserContentReferenceForKey(contentKey string) (*UserContentReference, bool) {
	if u == nil {
		return nil, false
	}

	for i, scheduleItem := range u.UserSchedule {
		for j, reference := range scheduleItem.UserContent {
			if reference.ContentKey == contentKey {
				return &u.UserSchedule[i].UserContent[j], u.Current && i == u.Completed
			}
		}
	}
	return nil, false
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

	Required int `json:"required"` // number of schedule items to complete = length of Schedule (may add required flags to each schedule item in future)

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

	if len(contentKeys) == 0 {
		contentKeys = make([]string, 0)
		for _, content := range u.Contents {
			if !utils.Exist[string](contentKeys, content.Key) {
				contentKeys = append(contentKeys, content.Key)
			}
		}
	}

	for _, item := range u.Schedule {
		for _, key := range item.ContentKeys {
			if !utils.Exist[string](contentKeys, key) {
				return errors.ErrorData(logutils.StatusInvalid, "schedule content key", &logutils.FieldArgs{"content_key": key})
			}
		}
	}

	u.Required = len(u.Schedule)
	return nil
}

// CreateUserSchedule creates the list of UserScheduleItems for Schedule with no associated UserContent ids
func (u *Unit) CreateUserSchedule() []UserScheduleItem {
	if u == nil {
		return nil
	}

	userSchedule := make([]UserScheduleItem, len(u.Schedule))
	for i, item := range u.Schedule {
		userContentRefs := make([]UserContentReference, len(item.ContentKeys))
		for j, key := range item.ContentKeys {
			userContentRefs[j] = UserContentReference{ContentKey: key, Complete: false}
		}
		userSchedule[i] = UserScheduleItem{UserContent: userContentRefs}
	}
	return userSchedule
}

// UserScheduleItem represents a set of UserContent references and when the corresponding ScheduleItem was started and first completed
type UserScheduleItem struct {
	UserContent []UserContentReference `json:"user_content" bson:"user_content"`

	DateStarted   *time.Time `json:"date_started,omitempty" bson:"date_started,omitempty"`
	DateCompleted *time.Time `json:"date_completed,omitempty" bson:"date_completed,omitempty"`
}

// IsComplete gives whether all UserContent items are complete
func (u UserScheduleItem) IsComplete() bool {
	for _, reference := range u.UserContent {
		if !reference.Complete {
			return false
		}
	}
	return true
}

// UserContentReference represents a set of UserContent references
type UserContentReference struct {
	ContentKey string   `json:"content_key" bson:"content_key"`
	IDs        []string `json:"ids,omitempty" bson:"ids,omitempty"` // UserContent IDs
	Complete   bool     `json:"complete" bson:"complete"`
}

// ScheduleItem represents a set of Content items to be completed in a certain amount of time
type ScheduleItem struct {
	Name        string   `json:"name" bson:"name"`
	ContentKeys []string `json:"content_keys" bson:"content_keys"`
	Duration    *int     `json:"duration" bson:"duration,omitempty"` // in days (if nil, this ScheduleItem is considered optional to complete)
}

// IsRequired returns whether the schedule item is required
func (s ScheduleItem) IsRequired() bool {
	return (s.Duration != nil)
}

// UserContent represents
type UserContent struct {
	ID        string `json:"id" bson:"_id"`
	AppID     string `json:"app_id" bson:"app_id"`
	OrgID     string `json:"org_id" bson:"org_id"`
	UserID    string `json:"user_id" bson:"user_id"`
	CourseKey string `json:"course_key" bson:"course_key"`
	ModuleKey string `json:"module_key" bson:"module_key"`
	UnitKey   string `json:"unit_key" bson:"unit_key"`

	Content  Content                `json:"content" bson:"content"`
	Response map[string]interface{} `json:"response" bson:"response"`

	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// IsComplete gives whether Response contains the entry {UserContentCompleteKey: true}
func (u *UserContent) IsComplete() bool {
	if u == nil {
		return false
	}

	completeVal, exists := u.Response[UserContentCompleteKey]
	if !exists {
		return false
	}
	complete, ok := completeVal.(bool)
	return complete && ok
}

// UpdateResponse updates the existing Response data with the incoming data, except UserContentCompleteKey if u.IsComplete()
func (u *UserContent) UpdateResponse(response map[string]interface{}) (bool, bool) {
	if u == nil {
		return false, false
	}

	completedNow := !u.IsComplete() && (response[UserContentCompleteKey] == true)
	if !completedNow {
		response[UserContentCompleteKey] = u.IsComplete()
	}
	updatedResponse := !utils.DeepEqual(response, u.Response)
	u.Response = response

	return updatedResponse, completedNow
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

// Equals returns whether two Content objects have the same underlying data (except styles and timestamps)
func (c *Content) Equals(other *Content) bool {
	if (c == nil) != (other == nil) {
		return false
	}
	if c == nil && other == nil {
		return true
	}

	if c.ID != other.ID || c.AppID != other.AppID || c.OrgID != other.OrgID || c.Key != other.Key || c.Type != other.Type || c.Name != other.Name || c.Details != other.Details {
		return false
	}
	if c.Reference.Name != other.Reference.Name || c.Reference.Type != other.Reference.Type || c.Reference.ReferenceKey != other.Reference.ReferenceKey {
		return false
	}
	if !utils.Equal(c.LinkedContent, other.LinkedContent, false) {
		return false
	}
	if !utils.DeepEqual(c.Styles, other.Styles) {
		return false
	}
	return true
}

// UserResponse includes a user response to a task with timezone info
type UserResponse struct {
	Timezone                          // include user timezone info
	ContentKey string                 `json:"content_key"`
	Response   map[string]interface{} `json:"response"`
}

// Reference represents a reference to another entity
type Reference struct {
	Name         string `json:"name" bson:"name"`
	Type         string `json:"type" bson:"type"` // content item, video, PDF, survey, web URL
	ReferenceKey string `json:"reference_key" bson:"reference_key"`
}

// Timezone represents user timezone information received from the client
type Timezone struct {
	Name   string `json:"timezone_name"`
	Offset int    `json:"timezone_offset"` // in seconds east of UTC
}

// Validate checks whether name and offset refer to a valid timezone (sets offset if name is valid but offset is not)
func (t *Timezone) Validate() error {
	if t == nil {
		return errors.ErrorData(logutils.StatusMissing, TypeTimezone, nil)
	}

	if t.Offset < utils.MinTZOffset || t.Offset > utils.MaxTZOffset {
		tzLoc, err := time.LoadLocation(t.Name)
		if err != nil {
			return errors.WrapErrorData(logutils.StatusInvalid, TypeTimezone, &logutils.FieldArgs{"name": t.Name, "offset": t.Offset}, err)
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
