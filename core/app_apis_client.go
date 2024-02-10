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
		return nil, errors.WrapErrorAction(logutils.ActionGet, model.TypeUser, nil, err)
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
		return nil, errors.WrapErrorAction(logutils.ActionGet, model.TypeUser, nil, err)
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

	userCourses, err := s.app.storage.FindUserCourses(idArr, claims.AppID, claims.OrgID, nameArr, keyArr, &userID, nil, nil)
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
		return nil, errors.WrapErrorAction(logutils.ActionValidate, model.TypeTimezone, nil, err)
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

// delete all user course derieved from a custom course
func (s *clientImpl) DeleteUserCourse(claims *tokenauth.Claims, courseKey string) error {
	transaction := func(storage interfaces.Storage) error {
		err := storage.DeleteUserCourse(claims.AppID, claims.OrgID, claims.Subject, courseKey)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionDelete, model.TypeUserCourse, nil, err)
		}

		err = storage.DeleteUserUnits(claims.AppID, claims.OrgID, claims.Subject, courseKey)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionDelete, model.TypeUserUnit, nil, err)
		}

		return nil
	}

	return s.app.storage.PerformTransaction(transaction)
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

// get all userUnits from a user's given course
func (s *clientImpl) GetUserCourseUnits(claims *tokenauth.Claims, courseKey string) ([]model.UserUnit, error) {
	userUnits, err := s.app.storage.FindUserUnits(claims.AppID, claims.OrgID, []string{claims.Subject}, courseKey, nil)
	if err != nil {
		return nil, err
	}
	return userUnits, nil
}

func (s *clientImpl) UpdateUserCourseUnitProgress(claims *tokenauth.Claims, courseKey string, unitKey string, item model.UserResponse) (*model.UserUnit, error) {
	err := item.Validate()
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, model.TypeTimezone, nil, err)
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
		userCourseUnits, err := storageTransaction.FindUserUnits(claims.AppID, claims.OrgID, []string{claims.Subject}, courseKey, nil)
		if err != nil {
			return err
		}

		var userScheduleItem *model.UserScheduleItem
		isCurrent := false  // whether the schedule item being updated is current
		isRequired := false // whether the schedule item being updated is required
		updatedUserCourse := false
		now := time.Now().UTC()
		if len(userCourseUnits) == 0 {
			unit, err := storageTransaction.FindCustomUnit(claims.AppID, claims.OrgID, unitKey)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionFind, model.TypeUnit, nil, err)
			}
			if unit == nil {
				return errors.ErrorData(logutils.StatusMissing, model.TypeUnit, &logutils.FieldArgs{"key": unitKey})
			}
			err = unit.Validate(nil)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionValidate, model.TypeUnit, &logutils.FieldArgs{"app_id": claims.AppID, "org_id": claims.OrgID, "key": unit.Key}, err)
			}

			userUnit = &model.UserUnit{ID: uuid.NewString(), AppID: claims.AppID, OrgID: claims.OrgID, UserID: claims.Subject, CourseKey: courseKey, Unit: *unit,
				Completed: 0, Current: true, UserSchedule: unit.CreateUserSchedule(), DateCreated: now}

			lastStreakProcess, err := s.updateUserContent(storageTransaction, userUnit, item, userCourse, &now, courseConfig.StreaksNotificationsConfig)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserContent, nil, err)
			}

			userScheduleItem, isCurrent, isRequired, updatedUserCourse, err = s.updateUserScheduleItem(userUnit, item, userCourse, lastStreakProcess, &now, courseConfig.StreaksNotificationsConfig)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeScheduleItem, &logutils.FieldArgs{"current": true}, err)
			}

			// user started the course so create the first user unit
			err = storageTransaction.InsertUserUnit(*userUnit)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
			}
		} else {
			// find the user unit the request wants to update
			for _, courseUnit := range userCourseUnits {
				if courseUnit.Unit.Key == unitKey {
					userUnit = &courseUnit
					break
				}
			}
			if userUnit == nil {
				return errors.ErrorData(logutils.StatusMissing, model.TypeUserUnit, &logutils.FieldArgs{"app_id": claims.AppID, "org_id": claims.OrgID, "user_id": claims.Subject, "course_key": courseKey, "unit.key": unitKey})
			}

			lastStreakProcess, err := s.updateUserContent(storageTransaction, userUnit, item, userCourse, &now, courseConfig.StreaksNotificationsConfig)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserContent, nil, err)
			}

			userScheduleItem, isCurrent, isRequired, updatedUserCourse, err = s.updateUserScheduleItem(userUnit, item, userCourse, lastStreakProcess, &now, courseConfig.StreaksNotificationsConfig)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeScheduleItem, &logutils.FieldArgs{"current": true}, err)
			}

			err = storageTransaction.UpdateUserUnit(*userUnit)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
			}
		}

		// update streak if the following are true:
		// 1. the current schedule item is required
		// 2. the current schedule item is completed
		// 3. the previous schedule item was completed before current schedule item was started or there is no previous schedule item
		var previousScheduleItem *model.UserScheduleItem
		var previousScheduleItemCompleted *time.Time
		if isCurrent {
			previousScheduleItem = userUnit.PreviousScheduleItem()
		}
		if previousScheduleItem != nil {
			previousScheduleItemCompleted = previousScheduleItem.DateCompleted
		} else if isCurrent && userUnit.LastCompleted != nil {
			previousScheduleItemCompleted = userUnit.LastCompleted
		}

		userScheduleItemStart := userScheduleItem.DateStarted
		if isCurrent && isRequired && userScheduleItem.IsComplete() && (previousScheduleItemCompleted == nil || (userScheduleItemStart != nil && previousScheduleItemCompleted.Before(*userScheduleItemStart))) {
			newStreak := userCourse.Streak + 1

			// if the user has no active streak and no remaining pauses, then add a streak restart (user has resumed progress after some extended time)
			if userCourse.Streak == 0 && userCourse.Pauses == 0 {
				if userCourse.StreakRestarts == nil {
					userCourse.StreakRestarts = make([]time.Time, 0)
				}
				userCourse.StreakRestarts = append(userCourse.StreakRestarts, now)
			}
			userCourse.Streak = newStreak
			updatedUserCourse = true
		}

		// update pause progress and pauses when user responds to a required task, regardless of completion, for the first time since last streak process ("start of day")
		if isRequired && (userCourse.LastResponded == nil || (userScheduleItemStart != nil && userCourse.LastResponded.Before(*userScheduleItemStart))) {
			userCourse.PauseProgress++
			userCourse.LastResponded = &now
			if userCourse.PauseProgress == courseConfig.PauseProgressReward && userCourse.Pauses < courseConfig.MaxPauses {
				userCourse.Pauses++
				userCourse.PauseProgress = 0
			}
			updatedUserCourse = true
		}

		if updatedUserCourse {
			err = storageTransaction.UpdateUserCourse(*userCourse)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = s.app.storage.PerformTransaction(transaction)
	if err != nil {
		return nil, err
	}
	return userUnit, nil
}

func (s *clientImpl) updateUserContent(storage interfaces.Storage, userUnit *model.UserUnit, userResponse model.UserResponse, userCourse *model.UserCourse,
	now *time.Time, snConfig model.StreaksNotificationsConfig) (*time.Time, error) {
	userContentReference := userUnit.GetUserContentReferenceForKey(userResponse.ContentKey)
	if userContentReference == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeUserContentReference, &logutils.FieldArgs{"user_unit.id": userUnit.ID, "content_key": userResponse.ContentKey})
	}

	content, err := storage.FindCustomContent(userUnit.AppID, userUnit.OrgID, userResponse.ContentKey)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeContent, nil, err)
	}
	if content == nil {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeUnit, &logutils.FieldArgs{"key": userResponse.ContentKey})
	}

	if now == nil {
		nowVal := time.Now().UTC()
		now = &nowVal
	}

	var lastStreakProcess *time.Time
	if len(userContentReference.IDs) == 0 {
		// the user has not saved any responses for this content yet, so create new user content
		err = s.createUserContent(storage, userUnit, userContentReference, *content, userResponse.Response, *now)
		if err != nil {
			return nil, err
		}
	} else {
		// the user has saved a response for this content before
		userContents, err := storage.FindUserContents(userContentReference.IDs, userUnit.AppID, userUnit.OrgID, userUnit.UserID)
		if err != nil {
			return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUserContent, nil, err)
		}
		if len(userContents) == 0 {
			return nil, errors.ErrorData(logutils.StatusMissing, model.TypeUserContent, &logutils.FieldArgs{"ids": userContentReference.IDs})
		}

		lastUserContent := userContents[0] // FindUserContents returns UserContents sorted in reverse chronological order by DateCreated
		lastStreakProcess = userCourse.MostRecentStreakProcessTime(now, snConfig)
		if lastStreakProcess == nil {
			return nil, errors.ErrorData(logutils.StatusInvalid, "last streak process", &logutils.FieldArgs{"user_course.id": userCourse.ID, "process_time": snConfig.StreaksProcessTime, "timezone_name": snConfig.TimezoneName})
		}

		if lastUserContent.DateCreated.Before(*lastStreakProcess) {
			// last response was created before the most recent streak process, so create new user content
			err = s.createUserContent(storage, userUnit, userContentReference, *content, userResponse.Response, *now)
			if err != nil {
				return nil, err
			}
		} else {
			// the user has responded to this content since the last streak process, so update the saved response
			lastUserContent.Response = userResponse.Response
			updateContent := !content.Equals(&lastUserContent.Content)
			if updateContent {
				lastUserContent.Content = *content
			}
			userContentReference.Complete = userContentReference.Complete || lastUserContent.IsComplete()

			err = storage.UpdateUserContent(lastUserContent, updateContent)
			if err != nil {
				return nil, errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserContent, nil, err)
			}
		}
	}

	return lastStreakProcess, nil
}

func (s *clientImpl) updateUserScheduleItem(userUnit *model.UserUnit, userResponse model.UserResponse, userCourse *model.UserCourse, lastStreakProcess *time.Time,
	now *time.Time, snConfig model.StreaksNotificationsConfig) (*model.UserScheduleItem, bool, bool, bool, error) {
	userScheduleItem, _, isCurrent, isRequired := userUnit.GetScheduleItem(userResponse.ContentKey, false)
	if userScheduleItem == nil {
		return nil, false, false, false, errors.ErrorData(logutils.StatusMissing, model.TypeScheduleItem, &logutils.FieldArgs{"current": true})
	}

	updatedUserCourse := false
	if isCurrent && userScheduleItem.DateStarted == nil {
		if lastStreakProcess != nil {
			userScheduleItem.DateStarted = lastStreakProcess
		} else {
			userScheduleItem.DateStarted = userCourse.MostRecentStreakProcessTime(now, snConfig)
		}
	}
	if isCurrent && userScheduleItem.IsComplete() {
		userScheduleItem.DateCompleted = now
		if !isRequired {
			userUnit.Completed++
		}

		// if there are no more required schedule items to be done in the course, set date completed
		// allow the new current schedule item to be returned if current schedule item not required because userUnit.Completed has already been incremented
		if userCourse.Course.NextRequiredScheduleItem(userUnit.Unit.Key, userUnit.Completed, !isRequired) == nil {
			// if the schedule item is required, increment completed because it will not be done by the streaks timer since userCourse.DateCompleted will be set
			if isRequired {
				userUnit.Completed++
			}
			userUnit.Current = false
			userCourse.DateCompleted = now
			updatedUserCourse = true
		}
	}

	return userScheduleItem, isCurrent, isRequired, updatedUserCourse, nil
}

func (s *clientImpl) createUserContent(storage interfaces.Storage, userUnit *model.UserUnit, userContentReference *model.UserContentReference, content model.Content, response map[string]interface{}, now time.Time) error {
	id := uuid.NewString()
	userContent := model.UserContent{ID: id, AppID: userUnit.AppID, OrgID: userUnit.OrgID, UserID: userUnit.UserID,
		CourseKey: userUnit.CourseKey, UnitKey: userUnit.Unit.Key, Content: content, Response: response, DateCreated: now}

	err := storage.InsertUserContent(userContent)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
	}

	if userContentReference.IDs == nil {
		userContentReference.IDs = make([]string, 0)
	}
	userContentReference.IDs = append(userContentReference.IDs, id)
	userContentReference.Complete = userContentReference.Complete || userContent.IsComplete()
	return nil
}

func (s *clientImpl) GetUserContents(claims *tokenauth.Claims, ids string) ([]model.UserContent, error) {
	idsList := strings.Split(ids, ",")
	userContents, err := s.app.storage.FindUserContents(idsList, claims.AppID, claims.OrgID, claims.Subject)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUserContent, nil, err)
	}

	return userContents, nil
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
