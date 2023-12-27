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
	"lms/utils"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

type clientImpl struct {
	app *Application
}

func (s *clientImpl) GetCourses(claims *tokenauth.Claims, courseType *string) ([]model.ProviderCourse, error) {
	providerUserID := s.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, "net_id", nil)
	}

	courses, err := s.app.provider.GetCourses(providerUserID)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, "course", nil, err)
	}
	return courses, nil
}

func (s *clientImpl) GetCourse(claims *tokenauth.Claims, id string) (*model.ProviderCourse, error) {
	providerUserID := s.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, "net_id", nil)
	}

	courseID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.WrapErrorData(logutils.StatusInvalid, "course id", nil, err)
	}

	course, err := s.app.provider.GetCourse(providerUserID, courseID)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, "course", nil, err)
	}
	return course, nil
}

func (s *clientImpl) GetAssignmentGroups(claims *tokenauth.Claims, id string, include *string) ([]model.AssignmentGroup, error) {
	providerUserID := s.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, "net_id", nil)
	}

	courseID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.WrapErrorData(logutils.StatusInvalid, "course id", nil, err)
	}

	includeParts := []string{}
	if include != nil {
		includeParts = strings.Split(*include, ",")
	}
	includeAssignments := utils.Exist(includeParts, "assignments")
	includeSubmission := utils.Exist(includeParts, "submission")

	assignmentGroups, err := s.app.provider.GetAssignmentGroups(providerUserID, courseID, includeAssignments, includeSubmission)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, "assignment groups", nil, err)
	}
	return assignmentGroups, nil
}

func (s *clientImpl) GetCourseUser(claims *tokenauth.Claims, id string, include *string) (*model.User, error) {
	providerUserID := s.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, "net_id", nil)
	}

	courseID, err := strconv.Atoi(id)
	if err != nil {
		return nil, errors.WrapErrorData(logutils.StatusInvalid, "course id", nil, err)
	}

	includeParts := []string{}
	if include != nil {
		includeParts = strings.Split(*include, ",")
	}
	includeEnrollments := utils.Exist(includeParts, "enrollments")
	includeScores := utils.Exist(includeParts, "scores")

	user, err := s.app.provider.GetCourseUser(providerUserID, courseID, includeEnrollments, includeScores)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, "user", nil, err)
	}
	return user, nil
}

func (s *clientImpl) GetCurrentUser(claims *tokenauth.Claims) (*model.User, error) {
	providerUserID := s.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, "net_id", nil)
	}

	user, err := s.app.provider.GetCurrentUser(providerUserID)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, "user", nil, err)
	}
	return user, nil
}

func (s *clientImpl) GetUserCourses(claims *tokenauth.Claims, id *string, name *string, courseKey *string) ([]model.UserCourse, error) {
	var idArr, nameArr, keyArr []string
	userID := claims.Subject
	//parse moduleID comma seperated string into array
	if id != nil {
		idArr = strings.Split(*id, ",")
	}
	if name != nil {
		nameArr = strings.Split(*name, ",")
	}
	if courseKey != nil {
		keyArr = strings.Split(*courseKey, ",")
	}

	userCourses, err := s.app.storage.FindUserCourses(idArr, claims.AppID, claims.OrgID, nameArr, keyArr, &userID, nil)
	if err != nil {
		return nil, err
	}
	return userCourses, nil
}

// pass usercourse id to retrieve usercourse struct
func (s *clientImpl) GetUserCourse(claims *tokenauth.Claims, courseKey string) (*model.UserCourse, error) {
	userCourse, err := s.app.storage.GetUserCourse(claims.AppID, claims.OrgID, claims.Subject, courseKey)
	if err != nil {
		return nil, err
	}
	return userCourse, nil
}

// pass course key to create a new user course
func (s *clientImpl) CreateUserCourse(claims *tokenauth.Claims, courseKey string) (*model.UserCourse, error) {
	var item model.UserCourse
	item.ID = uuid.NewString()
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID
	item.UserID = claims.Subject
	item.DateCreated = time.Now()

	//retrieve course with coursekey
	course, err := s.app.storage.GetCustomCourse(claims.AppID, claims.OrgID, courseKey)
	if err != nil {
		return nil, err
	}
	item.Course = *course
	err = s.app.storage.InsertUserCourse(item)
	if err != nil {
		return nil, err
	}

	// create user module and user unit
	for _, singleModule := range course.Modules {
		for _, singleUnit := range singleModule.Units {
			s.CreateUserUnit(claims, courseKey, singleModule.Key, singleUnit.Key)
		}
		s.CreateUserModule(claims, courseKey, singleModule.Key)
	}

	return &item, nil
}

// pass course key to create a new user course
func (s *clientImpl) CreateUserModule(claims *tokenauth.Claims, courseKey string, moduleKey string) (*model.UserModule, error) {
	var item model.UserModule
	item.ID = uuid.NewString()
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID
	item.UserID = claims.Subject
	item.CourseKey = courseKey
	item.DateCreated = time.Now()

	//retrieve moudle with moduleKey
	module, err := s.app.storage.GetCustomModule(claims.AppID, claims.OrgID, moduleKey)
	if err != nil {
		return nil, err
	}
	item.Module = *module
	err = s.app.storage.InsertUserModule(item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// pass unit key to create a new user unit
func (s *clientImpl) CreateUserUnit(claims *tokenauth.Claims, courseKey string, moduleKey string, unitKey string) (*model.UserUnit, error) {
	var item model.UserUnit
	item.ID = uuid.NewString()
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID
	item.UserID = claims.Subject
	item.CourseKey = courseKey
	item.ModuleKey = moduleKey
	item.DateCreated = time.Now()

	//retrieve moudle with unitKey
	unit, err := s.app.storage.GetCustomUnit(claims.AppID, claims.OrgID, unitKey)
	if err != nil {
		return nil, err
	}
	item.Unit = *unit
	err = s.app.storage.InsertUserUnit(item)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *clientImpl) DeleteUserCourse(claims *tokenauth.Claims, courseKey string) error {

	//item.UserID = claims.Subject

	err := s.app.storage.DeleteUserCourse(claims.AppID, claims.OrgID, claims.Subject, courseKey)
	if err != nil {
		return nil
	}
	return err
}

func (s *clientImpl) UpdateUserCourseUnitProgress(claims *tokenauth.Claims, courseKey string, moduleKey string, item model.Unit) (*model.Unit, error) {
	err := s.app.storage.UpdateUserUnit(claims.AppID, claims.OrgID, claims.Subject, courseKey, moduleKey, item)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *clientImpl) getProviderUserID(claims *tokenauth.Claims) string {
	if claims == nil {
		return ""
	}
	return claims.ExternalIDs["net_id"]
}
