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

	"github.com/rokwire/logging-library-go/v2/logs"
)

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
