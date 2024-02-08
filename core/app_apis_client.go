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
	err := item.Validate()
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, "user timezone", nil, err)
	}

	userCourse := &model.UserCourse{ID: uuid.NewString(), AppID: claims.AppID, OrgID: claims.OrgID, UserID: claims.Subject, Timezone: item, DateCreated: time.Now()}
	transaction := func(storage interfaces.Storage) error {
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
		userCourse.PauseProgress = 0
		userCourse.DateCreated = time.Now().UTC()

		// unique index on user courses collection will ensure user cannot take a course multiple times simultaneously
		err = storage.InsertUserCourse(*userCourse)
		if err != nil {
			return err
		}

		return nil
	}

	err = s.app.storage.PerformTransaction(transaction)
	if err != nil {
		return nil, err
	}
	return userCourse, nil
}

// get all userUnits from a user's given course
func (s *clientImpl) GetUserCourseUnits(claims *tokenauth.Claims, courseKey string) ([]model.UserUnit, error) {
	userUnits, err := s.app.storage.FindUserUnits(claims.AppID, claims.OrgID, []string{claims.Subject}, courseKey, nil)
	if err != nil {
		return nil, err
	}
	return userUnits, nil
}

// delete all user course derieved from a custom course
func (s *clientImpl) DeleteUserCourse(claims *tokenauth.Claims, courseKey string) error {
	err := s.app.storage.DeleteUserCourse(claims.AppID, claims.OrgID, claims.Subject, courseKey)
	if err != nil {
		return err
	}
	return nil
}

func (s *clientImpl) UpdateUserCourseUnitProgress(claims *tokenauth.Claims, courseKey string, unitKey string, item model.UserContentWithTimezone) (*model.UserUnit, error) {
	err := item.Validate()
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, "user timezone", nil, err)
	}

	var userUnit *model.UserUnit
	transaction := func(storageTransaction interfaces.Storage) error {
		userCourse, err := storageTransaction.FindUserCourse(claims.AppID, claims.OrgID, claims.Subject, courseKey)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionFind, model.TypeUserCourse, nil, err)
		}
		if userCourse == nil {
			return errors.ErrorData(logutils.StatusMissing, model.TypeUserCourse, &logutils.FieldArgs{"app_id": claims.AppID, "org_id": claims.OrgID, "user_id": claims.Subject, "key": courseKey})
		}
		if userCourse.DateDropped != nil {
			return errors.ErrorData(logutils.StatusInvalid, model.TypeUserCourse, &logutils.FieldArgs{"id": userCourse.ID, "date_dropped": userCourse.DateDropped})
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
		userCourse.Timezone.Name = item.Name
		userCourse.Timezone.Offset = item.Offset

		// find the current user unit (this is managed by the streaks timer)
		userUnit, err = storageTransaction.FindUserUnit(claims.AppID, claims.OrgID, claims.Subject, courseKey, &unitKey)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		var lastCompleted *time.Time
		var scheduleItem *model.ScheduleItem
		if userUnit == nil {
			// create a userUnit here if it doesn't already exist
			unit, err := storageTransaction.FindCustomUnit(claims.AppID, claims.OrgID, unitKey)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionFind, model.TypeUnit, nil, err)
			}
			if unit == nil {
				return errors.ErrorData(logutils.StatusMissing, model.TypeUnit, &logutils.FieldArgs{"key": unitKey})
			}
			if len(unit.Schedule) == 0 {
				return errors.ErrorData(logutils.StatusMissing, "unit schedule", &logutils.FieldArgs{"key": unit.Key})
			}

			userUnit = &model.UserUnit{ID: uuid.NewString(), AppID: claims.AppID, OrgID: claims.OrgID, UserID: claims.Subject, CourseKey: courseKey,
				Completed: 0, Current: true, DateCreated: time.Now().UTC()}
			userUnit.Unit = *unit

			scheduleItem = &userUnit.Unit.Schedule[0]
			scheduleItem.UpdateUserData(item.UserContent)
			scheduleItem.DateStarted = userCourse.MostRecentStreakProcessTime(&now, courseConfig.StreaksNotificationsConfig)
			if scheduleItem.IsComplete() {
				scheduleItem.DateCompleted = &now
				// userUnit.Completed++
				userUnit.LastCompleted = &now
				if userUnit.Completed < unit.ScheduleStart {
					userUnit.Completed++
				}
			}

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
			if userUnit.LastCompleted != nil && userUnit.Completed > userUnit.Unit.ScheduleStart {
				// make copy of LastCompleted time before possibly updating it
				lastCompletedVal := *userUnit.LastCompleted
				lastCompleted = &lastCompletedVal
			}

			scheduleItem = &userUnit.Unit.Schedule[userUnit.Completed]
			scheduleItem.UpdateUserData(item.UserContent)
			if scheduleItem.DateStarted == nil {
				scheduleItem.DateStarted = userCourse.MostRecentStreakProcessTime(&now, courseConfig.StreaksNotificationsConfig)
			}
			if scheduleItem.IsComplete() {
				scheduleItem.DateCompleted = &now
				// userUnit.Completed++
				userUnit.LastCompleted = &now
				if userUnit.Completed < userUnit.Unit.ScheduleStart {
					userUnit.Completed++
				}
			}

			err = storageTransaction.UpdateUserUnit(*userUnit)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
			}
		}

		// update streak and pauses immediately, pauses are used and streaks are reset if necessary in the streaks timer
		// only if the following are true:
		// 1. the current schedule item is required
		// 2. the current schedule item is completed
		// 3. the previous schedule item was completed before most recent streak timer run or there is no previous schedule item
		isRequired := userUnit.Completed >= userUnit.Unit.ScheduleStart
		lastStreakProcess := userUnit.Unit.Schedule[userUnit.Completed].DateStarted
		if isRequired && scheduleItem.IsComplete() && (lastCompleted == nil || (lastStreakProcess != nil && lastCompleted.Before(*lastStreakProcess))) {
			newStreak := userCourse.Streak + 1

			// if the user has no active streak and no remaining pauses, then mark now as a streak restart (user has resumed progress after some extended time)
			if userCourse.Streak == 0 && userCourse.Pauses == 0 {
				if userCourse.StreakRestarts == nil {
					userCourse.StreakRestarts = make([]time.Time, 0)
				}
				userCourse.StreakRestarts = append(userCourse.StreakRestarts, now)
			}
			userCourse.Streak = newStreak
		}

		// update pause progress and pauses when user responds to a required task, regardless of completion, for the first time since last streak process ("start of day")
		if isRequired && (userCourse.LastResponded == nil || (lastStreakProcess != nil && userCourse.LastResponded.Before(*lastStreakProcess))) {
			userCourse.PauseProgress++
			if userCourse.PauseProgress == courseConfig.PauseProgressReward && userCourse.Pauses < courseConfig.MaxPauses {
				userCourse.Pauses++
				userCourse.PauseProgress = 0
			}
		}
		userCourse.LastResponded = &now
		err = storageTransaction.UpdateUserCourse(*userCourse)
		if err != nil {
			return err
		}

		return nil
	}

	err = s.app.storage.PerformTransaction(transaction)
	if err != nil {
		return nil, err
	}
	return userUnit, nil
}

func (s *clientImpl) UpdateUserCourse(claims *tokenauth.Claims, key string, drop *bool) (*model.UserCourse, error) {
	userCourse, err := s.app.storage.FindUserCourse(claims.AppID, claims.OrgID, claims.Subject, key)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUserCourse, nil, err)
	}
	if userCourse == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeUserCourse, &logutils.FieldArgs{"course.key": key})
	}

	if drop != nil && *drop {
		if userCourse.DateDropped != nil {
			return nil, errors.ErrorData(logutils.StatusInvalid, model.TypeUserCourse, &logutils.FieldArgs{"course.key": key, "date_dropped": *userCourse.DateDropped})
		}

		now := time.Now().UTC()
		userCourse.DateDropped = &now
		err := s.app.storage.UpdateUserCourse(*userCourse)
		if err != nil {
			return nil, errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &logutils.FieldArgs{"drop": true}, err)
		}
	}

	return nil, nil
}

func (s *clientImpl) GetCustomCourses(claims *tokenauth.Claims) ([]model.Course, error) {
	courses, err := s.app.storage.FindCustomCourses(claims.AppID, claims.OrgID, nil, nil, nil, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourse, nil, err)
	}
	return courses, nil
}

func (s *clientImpl) GetCustomCourse(claims *tokenauth.Claims, key string) (*model.Course, error) {
	course, err := s.app.storage.FindCustomCourse(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourse, nil, err)
	}
	return course, nil
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
