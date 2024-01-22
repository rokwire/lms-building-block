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
	userCourse, err := s.app.storage.FindUserCourse(claims.AppID, claims.OrgID, claims.Subject, courseKey)
	if err != nil {
		return nil, err
	}
	return userCourse, nil
}

// pass course key to create a new user course
func (s *clientImpl) CreateUserCourse(claims *tokenauth.Claims, courseKey string, item model.Timezone) (*model.UserCourse, error) {
	var userCourse *model.UserCourse
	transaction := func(storage interfaces.Storage) error {
		userCourse := model.UserCourse{ID: uuid.NewString(), AppID: claims.AppID, OrgID: claims.OrgID, UserID: claims.Subject, Timezone: item, DateCreated: time.Now()}

		//retrieve course with coursekey
		course, err := storage.FindCustomCourse(claims.AppID, claims.OrgID, courseKey)
		if err != nil {
			return err
		}
		userCourse.Course = *course

		courseConfig, err := storage.FindCourseConfig(course.AppID, course.OrgID, course.Key)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionFind, model.TypeCourseConfig, nil, err)
		}
		userCourse.Streak = 0
		userCourse.Pauses = courseConfig.InitialPauses

		// unique index on user courses collection will ensure user cannot take a course multiple times simultaneously
		err = storage.InsertUserCourse(userCourse)
		if err != nil {
			return err
		}

		return nil
	}

	err := s.app.storage.PerformTransaction(transaction)
	if err != nil {
		return nil, err
	}
	return userCourse, nil
}

// delete all user course derieved from a custom course
func (s *clientImpl) DeleteUserCourse(claims *tokenauth.Claims, courseKey string) error {

	//item.UserID = claims.Subject

	err := s.app.storage.DeleteUserCourse(claims.AppID, claims.OrgID, claims.Subject, courseKey)
	if err != nil {
		return err
	}
	return nil
}

func (s *clientImpl) UpdateUserCourseUnitProgress(claims *tokenauth.Claims, courseKey string, unitKey string, item model.UnitWithTimezone) (*model.UnitWithTimezone, error) {
	transaction := func(storageTransaction interfaces.Storage) error {
		// find the current user unit (this is managed by the streaks timer)
		userUnit, err := storageTransaction.FindUserUnit(claims.AppID, claims.OrgID, claims.Id, courseKey, &unitKey)
		if err != nil {
			return err
		}
		if userUnit == nil {
			// create a userUnit here if it doesn't already exist
			userUnit = &model.UserUnit{ID: uuid.NewString(), AppID: claims.AppID, OrgID: claims.OrgID, UserID: claims.Subject, CourseKey: courseKey,
				Completed: 0, Current: true, DateCreated: time.Now().UTC()}

			unit, err := storageTransaction.FindCustomUnit(claims.AppID, claims.OrgID, unitKey)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionFind, model.TypeUnit, nil, err)
			}
			item.Unit = *unit

			// user started the course and created the first user unit
			err = storageTransaction.InsertUserUnit(*userUnit)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
			}
		} else {
			// only updates to the current user unit are allowed
			if !userUnit.Current {
				return errors.ErrorData(logutils.StatusInvalid, model.TypeUserUnit, &logutils.FieldArgs{"current": false})
			}

			now := time.Now().UTC()
			userUnit.Unit.Schedule = item.Unit.Schedule
			userUnit.Completed++
			userUnit.LastCompleted = &now

			err = storageTransaction.UpdateUserUnit(claims.AppID, claims.OrgID, claims.Subject, courseKey, *userUnit)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
			}
		}

		userCourse, err := storageTransaction.FindUserCourse(claims.AppID, claims.OrgID, claims.Subject, courseKey)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionFind, model.TypeUserCourse, nil, err)
		}
		if userCourse == nil {
			return errors.ErrorData(logutils.StatusMissing, model.TypeUserCourse, &logutils.FieldArgs{"app_id": claims.AppID, "org_id": claims.OrgID, "user_id": claims.Subject, "key": courseKey})
		}

		courseConfig, err := storageTransaction.FindCourseConfig(userCourse.AppID, userCourse.OrgID, userCourse.Course.Key)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionFind, model.TypeCourseConfig, nil, err)
		}

		// update timezone name and offset for all user_course of a user
		err = storageTransaction.UpdateUserTimezone(userCourse.AppID, userCourse.OrgID, userCourse.UserID, item.Name, item.Offset)
		if err != nil {
			return err
		}

		// update streak and pauses immediately, pauses are used and streaks are reset if necessary in the streaks timer
		newStreak := userCourse.Streak + 1
		newPauses := userCourse.Pauses
		if newStreak%courseConfig.PauseRewardStreak == 0 && userCourse.Pauses < courseConfig.MaxPauses {
			newPauses++
		}
		err = storageTransaction.UpdateUserCourse(userCourse.AppID, userCourse.OrgID, userCourse.UserID, nil, userCourse.Course.Key, &newStreak, &newPauses)
		if err != nil {
			return err
		}
		return nil
	}
	return nil, s.app.storage.PerformTransaction(transaction)
}

func (s *clientImpl) GetCustomCourseConfig(claims *tokenauth.Claims, key string) (*model.CourseConfig, error) {
	courseConfig, err := s.app.storage.FindCourseConfig(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourseConfig, nil, err)
	}

	return courseConfig, nil
}

func (s *clientImpl) getProviderUserID(claims *tokenauth.Claims) string {
	if claims == nil {
		return ""
	}
	return claims.ExternalIDs["net_id"]
}

// marks a userCourse as deleted, as oppose to remove from database
func (s *clientImpl) DropUserCourse(claims *tokenauth.Claims, key string) (*model.UserCourse, error) {
	transaction := func(storageTransaction interfaces.Storage) error {
		appID := claims.AppID
		orgID := claims.OrgID

		err := storageTransaction.MarkUserCourseAsDelete(appID, orgID, key)
		if err != nil {
			return err
		}

		return err
	}
	return nil, s.app.storage.PerformTransaction(transaction)
}
