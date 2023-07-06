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

package core

import (
	"lms/core/model"
	"lms/driven/storage"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	GetCourses(l *logs.Log, providerUserID string) ([]model.Course, error)
	GetCourse(l *logs.Log, providerUserID string, courseID int) (*model.Course, error)
	GetAssignmentGroups(l *logs.Log, providerUserID string, courseID int, includeAssignments bool, includeSubmission bool) ([]model.AssignmentGroup, error)
	GetCourseUser(l *logs.Log, providerUserID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error)
	GetCurrentUser(l *logs.Log, providerUserID string) (*model.User, error)
}

// Administration exposes APIs for the driver adapters
type Administration interface {
	GetNudgesConfig(l *logs.Log) (*model.NudgesConfig, error)
	UpdateNudgesConfig(l *logs.Log, active bool, groupName string, testGroupName string, mode string, processTime *int, blockSize *int) error

	GetNudges() ([]model.Nudge, error)
	CreateNudge(l *logs.Log, ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSource) error
	UpdateNudge(l *logs.Log, ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSources) error
	DeleteNudge(l *logs.Log, ID string) error

	FindSentNudges(l *logs.Log, nudgeID *string, userID *string, netID *string, mode *string) ([]model.SentNudge, error)
	DeleteSentNudges(l *logs.Log, ids []string) error
	ClearTestSentNudges(l *logs.Log) error

	FindNudgesProcesses(l *logs.Log, limit int, offset int) ([]model.NudgesProcess, error)
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) GetCourses(l *logs.Log, providerUserID string) ([]model.Course, error) {
	return s.app.getCourses(l, providerUserID)
}

func (s *servicesImpl) GetCourse(l *logs.Log, providerUserID string, courseID int) (*model.Course, error) {
	return s.app.getCourse(l, providerUserID, courseID)
}

func (s *servicesImpl) GetAssignmentGroups(l *logs.Log, providerUserID string, courseID int, includeAssignments bool, includeSubmission bool) ([]model.AssignmentGroup, error) {
	return s.app.getAssignmentGroups(l, providerUserID, courseID, includeAssignments, includeSubmission)
}

func (s *servicesImpl) GetCourseUser(l *logs.Log, providerUserID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error) {
	return s.app.getCourseUser(l, providerUserID, courseID, includeEnrolments, includeScores)
}

func (s *servicesImpl) GetCurrentUser(l *logs.Log, providerUserID string) (*model.User, error) {
	return s.app.getCurrentUser(l, providerUserID)
}

//admin

type administrationImpl struct {
	app *Application
}

func (s *administrationImpl) GetNudgesConfig(l *logs.Log) (*model.NudgesConfig, error) {
	return s.app.getNudgesConfig(l)
}

func (s *administrationImpl) UpdateNudgesConfig(l *logs.Log, active bool, groupName string, testGroupName string, mode string, processTime *int, blockSize *int) error {
	return s.app.updateNudgesConfig(l, active, groupName, testGroupName, mode, processTime, blockSize)
}

func (s *administrationImpl) GetNudges() ([]model.Nudge, error) {
	return s.app.getNudges()
}

func (s *administrationImpl) CreateNudge(l *logs.Log, ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSource) error {
	return s.app.createNudge(l, ID, name, body, deepLink, params, active, usersSourse)
}

func (s *administrationImpl) UpdateNudge(l *logs.Log, ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSources) error {
	return s.app.updateNudge(l, ID, name, body, deepLink, params, active, usersSourse)
}

func (s *administrationImpl) DeleteNudge(l *logs.Log, ID string) error {
	return s.app.deleteNudge(l, ID)
}

func (s *administrationImpl) FindSentNudges(l *logs.Log, nudgeID *string, userID *string, netID *string, mode *string) ([]model.SentNudge, error) {
	return s.app.findSentNudges(l, nudgeID, userID, netID, nil, mode)
}

func (s *administrationImpl) DeleteSentNudges(l *logs.Log, ids []string) error {
	return s.app.deleteSentNudges(l, ids)
}

func (s *administrationImpl) ClearTestSentNudges(l *logs.Log) error {
	return s.app.clearTestSentNudges(l)
}

func (s *administrationImpl) FindNudgesProcesses(l *logs.Log, limit int, offset int) ([]model.NudgesProcess, error) {
	return s.app.findNudgesProcesses(l, limit, offset)
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	SetListener(listener storage.CollectionListener)

	CreateNudgesConfig(nudgesConfig model.NudgesConfig) error
	FindNudgesConfig() (*model.NudgesConfig, error)
	SaveNudgesConfig(nudgesConfig model.NudgesConfig) error

	LoadAllNudges() ([]model.Nudge, error)
	LoadActiveNudges() ([]model.Nudge, error)
	InsertNudge(item model.Nudge) error
	UpdateNudge(ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSources) error
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
}

// Provider interface for LMS provider
type Provider interface {
	GetCourses(userID string) ([]model.Course, error)
	GetCourse(userID string, courseID int) (*model.Course, error)
	GetCourseUsers(courseID int) ([]model.User, error)
	GetAssignmentGroups(userID string, courseID int, includeAssignments bool, includeSubmission bool) ([]model.AssignmentGroup, error)
	GetCourseUser(userID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error)
	GetCurrentUser(userID string) (*model.User, error)

	FindUsersByCanvasUserID(canvasUserIds []int) ([]ProviderUser, error)

	CacheCommonData(usersIDs map[string]string) error
	FindCachedData(usersIDs []string) ([]ProviderUser, error)
	CacheUserData(user ProviderUser) (*ProviderUser, error)
	CacheUserCoursesData(user ProviderUser, coursesIDs []int) (*ProviderUser, error)

	GetLastLogin(userID string) (*time.Time, error)
	GetMissedAssignments(userID string) ([]model.Assignment, error)
	GetCompletedAssignments(userID string) ([]model.Assignment, error)
	GetCalendarEvents(netID string, providerUserID int, courseID int, startAt time.Time, endAt time.Time) ([]model.CalendarEvent, error)
}

//Cache entities

// ProviderUser cache entity
type ProviderUser struct {
	ID       string     `bson:"_id"`    //core BB account id
	NetID    string     `bson:"net_id"` //core BB external system id
	User     model.User `bson:"user"`
	SyncDate time.Time  `bson:"sync_date"`

	Courses *UserCourses `bson:"courses"`
}

// UserCourses cache entity
type UserCourses struct {
	Data     []UserCourse `bson:"data"`
	SyncDate time.Time    `bson:"sync_date"`
}

// UserCourse cache entity
type UserCourse struct {
	Data        model.Course       `bson:"data"`
	Assignments []CourseAssignment `bson:"assignments"`
	SyncDate    time.Time          `bson:"sync_date"`
}

// CourseAssignment cache entity
type CourseAssignment struct {
	Data       model.Assignment `bson:"data"`
	Submission *Submission      `bson:"submission"`
	SyncDate   time.Time        `bson:"sync_date"`
}

// Submission cache entity
type Submission struct {
	Data     *model.Submission `bson:"data"`
	SyncDate time.Time         `bson:"sync_date"`
}

// GroupsBB interface for the Groups building block communication
type GroupsBB interface {
	GetUsers(groupName string, offset int, limit int) ([]GroupsBBUser, error)
}

// GroupsBBUser entity
type GroupsBBUser struct {
	UserID string `json:"user_id"`
	NetID  string `json:"net_id"`
	Name   string `json:"name"`
}

// NotificationsBB interface for the Notifications building block communication
type NotificationsBB interface {
	SendNotifications(recipients []Recipient, text string, body string, data map[string]string) error
}

// Recipient entity
type Recipient struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}
