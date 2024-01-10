// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package interfaces

import (
	"lms/core/model"
	"lms/driven/groups"
	"lms/driven/notifications"
	"time"
)

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	SetListener(listener CollectionListener)

	PerformTransaction(transaction func(storage Storage) error) error

	UserExist(netID string) (*bool, error)
	InsertUser(user model.ProviderUser) error
	FindUser(netID string) (*model.ProviderUser, error)
	FindUsers(netIDs []string) ([]model.ProviderUser, error)
	FindUsersByCanvasUserID(canvasUserIds []int) ([]model.ProviderUser, error)
	SaveUser(providerUser model.ProviderUser) error

	CreateNudgesConfig(nudgesConfig model.NudgesConfig) error
	FindNudgesConfig() (*model.NudgesConfig, error)
	SaveNudgesConfig(nudgesConfig model.NudgesConfig) error

	LoadAllNudges() ([]model.Nudge, error)
	LoadActiveNudges() ([]model.Nudge, error)
	InsertNudge(item model.Nudge) error
	UpdateNudge(item model.Nudge) error
	DeleteNudge(ID string) error

	InsertSentNudge(sentNudge model.SentNudge) error
	InsertSentNudges(sentNudge []model.SentNudge) error
	FindSentNudge(nudgeID string, userID string, netID string, criteriaHash uint32, mode string) (*model.SentNudge, error)
	FindSentNudges(nudgeID *string, userID *string, netID *string, criteriaHash *[]uint32, mode *string) ([]model.SentNudge, error)
	DeleteSentNudges(ids []string, mode string) error

	InsertNudgesProcess(nudgesProcess model.NudgesProcess) error
	UpdateNudgesProcess(ID string, completedAt time.Time, status string, err *string) error
	CountNudgesProcesses(status string) (*int64, error)
	FindNudgesProcesses(limit int, offset int) ([]model.NudgesProcess, error)

	InsertBlock(block model.Block) error
	InsertBlocks(blocks []model.Block) error
	FindBlock(processID string, blockNumber int) (*model.Block, error)

	GetCustomCourses(appID string, orgID string, id []string, name []string, key []string, moduleKeys []string) ([]model.Course, error)
	GetCustomCourse(appID string, orgID string, key string) (*model.Course, error)
	InsertCustomCourse(item model.Course) error
	InsertCustomCourses(items []model.Course) error
	UpdateCustomCourse(key string, item model.Course) error
	DeleteCustomCourse(appID string, orgID string, key string) error

	GetCustomModules(appID string, orgID string, id []string, name []string, key []string, unitKeys []string) ([]model.Module, error)
	GetCustomModule(appID string, orgID string, key string) (*model.Module, error)
	InsertCustomModule(item model.Module) error
	InsertCustomModules(items []model.Module) error
	UpdateCustomModule(key string, item model.Module) error
	DeleteCustomModule(appID string, orgID string, key string) error

	GetCustomUnits(appID string, orgID string, id []string, name []string, key []string, contentKeys []string) ([]model.Unit, error)
	GetCustomUnit(appID string, orgID string, key string) (*model.Unit, error)
	InsertCustomUnit(item model.Unit) error
	InsertCustomUnits(items []model.Unit) error
	UpdateCustomUnit(key string, item model.Unit) error
	DeleteCustomUnit(appID string, orgID string, key string) error

	GetCustomContents(appID string, orgID string, id []string, name []string, key []string) ([]model.Content, error)
	GetCustomContent(appID string, orgID string, key string) (*model.Content, error)
	InsertCustomContent(item model.Content) error
	InsertCustomContents(items []model.Content) error
	UpdateCustomContent(key string, item model.Content) error
	DeleteCustomContent(appID string, orgID string, key string) error

	GetUserCourses(id []string, name []string, key []string, userID string) ([]model.UserCourse, error)
	GetUserCourse(appID string, orgID string, userID string, courseKey string) (*model.UserCourse, error)
	InsertUserCourse(item model.UserCourse) error
	UpdateUserCourses(key string, item model.Course) error
	DeleteUserCourse(appID string, orgID string, userID string, courseKey string) error

	InsertUserUnit(item model.UserUnit) error
	UpdateUserUnit(appID string, orgID string, userID string, item model.Unit) error
	UpdateUserUnits(key string, item model.Unit) error

	DeleteContentKeyFromLinkedContents(appID string, orgID string, key string) error
	DeleteContentKeyFromUnits(appID string, orgID string, key string) error
	DeleteContentKeyFromUserUnits(appID string, orgID string, key string) error
	DeleteUnitKeyFromModules(appID string, orgID string, key string) error
	DeleteModuleKeyFromCourses(appID string, orgID string, key string) error
	DeleteModuleKeyFromUserCourses(appID string, orgID string, key string) error
	MarkUserCourseAsDelete(appID string, orgID string, key string) error

	UpdateReferenceKeyToClientUnits(oldCourseKey string, newCourseKey string) error
}

// Provider interface for LMS provider
type Provider interface {
	GetCourses(userID string) ([]model.ProviderCourse, error)
	GetCourse(userID string, courseID int) (*model.ProviderCourse, error)
	GetCourseUsers(courseID int) ([]model.User, error)
	GetAssignmentGroups(userID string, courseID int, includeAssignments bool, includeSubmission bool) ([]model.AssignmentGroup, error)
	GetCourseUser(userID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error)
	GetCurrentUser(userID string) (*model.User, error)

	FindUsersByCanvasUserID(canvasUserIds []int) ([]model.ProviderUser, error)

	CacheCommonData(usersIDs map[string]string) error
	FindCachedData(usersIDs []string) ([]model.ProviderUser, error)
	CacheUserData(user model.ProviderUser) (*model.ProviderUser, error)
	CacheUserCoursesData(user model.ProviderUser, coursesIDs []int) (*model.ProviderUser, error)

	GetLastLogin(userID string) (*time.Time, error)
	GetMissedAssignments(userID string) ([]model.Assignment, error)
	GetCompletedAssignments(userID string) ([]model.Assignment, error)
	GetCalendarEvents(netID string, providerUserID int, courseID int, startAt time.Time, endAt time.Time) ([]model.CalendarEvent, error)
}

// GroupsBB interface for the Groups building block communication
type GroupsBB interface {
	GetUsers(groupName string, offset int, limit int) ([]groups.User, error)
}

// NotificationsBB interface for the Notifications building block communication
type NotificationsBB interface {
	SendNotifications(recipients []notifications.Recipient, text string, body string, data map[string]string) error
}

// CollectionListener listens for collection updates
type CollectionListener interface {
	OnConfigsUpdated()
}
