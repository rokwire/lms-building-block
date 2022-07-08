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
	"lms/core/model"
	"lms/driven/storage"
	"time"

	"github.com/rokwire/logging-library-go/logs"
)

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	GetCourses(l *logs.Log, providerUserID string) ([]model.Course, error)
	GetCourse(l *logs.Log, providerUserID string, courseID int, include *string) (*model.Course, error)
	GetAssignmentGroups(l *logs.Log, providerUserID string, courseID int, include *string) ([]model.AssignmentGroup, error)
	GetCourseUser(l *logs.Log, providerUserID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error)
	GetCurrentUser(l *logs.Log, providerUserID string) (*model.User, error)
}

type Administration interface {
	GetNudges() ([]model.Nudge, error)
	CreateNudge(l *logs.Log, ID string, name string, body string, params *map[string]interface{}) error
	UpdateNudge(l *logs.Log, ID string, name string, body string, params *map[string]interface{}) error
	DeleteNudge(l *logs.Log, ID string) error
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

func (s *servicesImpl) GetCourse(l *logs.Log, providerUserID string, courseID int, include *string) (*model.Course, error) {
	return s.app.getCourse(l, providerUserID, courseID, include)
}

func (s *servicesImpl) GetAssignmentGroups(l *logs.Log, providerUserID string, courseID int, include *string) ([]model.AssignmentGroup, error) {
	return s.app.getAssignmentGroups(l, providerUserID, courseID, include)
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

func (s *administrationImpl) GetNudges() ([]model.Nudge, error) {
	return s.app.getNudges()
}

func (s *administrationImpl) CreateNudge(l *logs.Log, ID string, name string, body string, params *map[string]interface{}) error {
	return s.app.createNudge(l, ID, name, body, params)
}

func (s *administrationImpl) UpdateNudge(l *logs.Log, ID string, name string, body string, params *map[string]interface{}) error {
	return s.app.updateNudge(l, ID, name, body, params)
}

func (s *administrationImpl) DeleteNudge(l *logs.Log, ID string) error {
	return s.app.deleteNudge(l, ID)
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	SetListener(listener storage.CollectionListener)

	LoadAllNudges() ([]model.Nudge, error)
	InsertNudge(item model.Nudge) error
	UpdateNudge(ID string, name string, body string, params *map[string]interface{}) error
	DeleteNudge(ID string) error

	InsertSentNudge(sentNudge model.SentNudge) error
}

//Provider interface for LMS provider
type Provider interface {
	GetCourses(userID string) ([]model.Course, error)
	GetCourse(userID string, courseID int, include *string) (*model.Course, error)
	GetAssignmentGroups(userID string, courseID int, include *string) ([]model.AssignmentGroup, error)
	GetCourseUser(userID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error)
	GetCurrentUser(userID string) (*model.User, error)
	GetLastLogin(userID string) (*time.Time, error)
}

//GroupsBB interface for the Groups building block communication
type GroupsBB interface {
	GetUsers() ([]GroupsBBUser, error)
}

//GroupsBBUser entity
type GroupsBBUser struct {
	UserID string
	NetID  string
}

//NotificationsBB interface for the Notifications building block communication
type NotificationsBB interface {
	SendNotifications() error
}
