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

func (s *clientImpl) GetCourses(claims *tokenauth.Claims, courseType *string, limit *int) ([]model.ProviderCourse, error) {
	providerUserID := s.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, "net_id", nil)
	}

	courses, err := s.app.provider.GetCourses(providerUserID, limit)
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

		err = storage.DeleteUserContents(claims.AppID, claims.OrgID, claims.Subject, &courseKey)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionDelete, model.TypeUserContent, nil, err)
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
	userUnits, err := s.app.storage.FindUserUnits(claims.AppID, claims.OrgID, []string{claims.Subject}, courseKey, nil, nil)
	if err != nil {
		return nil, err
	}
	return userUnits, nil
}

func (s *clientImpl) UpdateUserCourseModuleProgress(claims *tokenauth.Claims, courseKey string, moduleKey string, item model.UserResponse) (*model.UserUnit, error) {
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
			return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeTimezone, nil, err)
		}
		userCourse.Timezone.Name = item.Name
		userCourse.Timezone.Offset = item.Offset

		// get all userUnits under this module and check for multiple current user units
		moduleUserUnits, err := storageTransaction.FindUserUnits(claims.AppID, claims.OrgID, []string{claims.Subject}, courseKey, &moduleKey, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionFind, model.TypeUserUnit, nil, err)
		}
		var currentModuleUserUnit *model.UserUnit
		for i, uUnit := range moduleUserUnits {
			if uUnit.Current {
				if currentModuleUserUnit != nil {
					return errors.ErrorData(logutils.StatusInvalid, model.TypeUserUnit, &logutils.FieldArgs{"id": uUnit.ID, "module_key": moduleKey, "current": true, "current_id": currentModuleUserUnit.ID})
				}
				currentModuleUserUnit = &moduleUserUnits[i]
			}
		}

		var userScheduleItem *model.UserScheduleItem
		isCurrent := false  // whether the schedule item being updated is current
		isRequired := false // whether the schedule item being updated is required
		updatedUserCourse := false
		now := time.Now().UTC()
		var lastStreakProcess *time.Time
		if len(moduleUserUnits) == 0 {
			// get the requested module and validate the first unit
			module, err := storageTransaction.FindCustomModule(claims.AppID, claims.OrgID, moduleKey)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionFind, model.TypeModule, nil, err)
			}
			if len(module.Units) == 0 {
				return errors.ErrorData(logutils.StatusInvalid, model.TypeModule, &logutils.FieldArgs{"units.length": 0})
			}

			unit := module.Units[0]
			if unit.Key != item.UnitKey {
				return errors.ErrorData(logutils.StatusInvalid, model.TypeUnit, &logutils.FieldArgs{"unit_key": item.UnitKey})
			}

			err = unit.Validate(nil)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionValidate, model.TypeUnit, &logutils.FieldArgs{"app_id": claims.AppID, "org_id": claims.OrgID, "key": unit.Key}, err)
			}

			userUnit = &model.UserUnit{ID: uuid.NewString(), AppID: claims.AppID, OrgID: claims.OrgID, UserID: claims.Subject, CourseKey: courseKey, ModuleKey: moduleKey, Unit: unit,
				Completed: 0, Current: true, UserSchedule: unit.CreateUserSchedule(), DateCreated: now}

			_, lastStreakProcess, err := s.updateUserContent(storageTransaction, userUnit, item, userCourse, &now, courseConfig.StreaksNotificationsConfig)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserContent, nil, err)
			}

			userScheduleItem, isCurrent, isRequired, updatedUserCourse, err = s.updateUserScheduleItem(storageTransaction, userUnit, item, userCourse, lastStreakProcess, &now, courseConfig.StreaksNotificationsConfig)
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
			for _, moduleUserUnit := range moduleUserUnits {
				if moduleUserUnit.Unit.Key == item.UnitKey {
					userUnit = &moduleUserUnit
					break
				}
			}
			if userUnit == nil {
				return errors.ErrorData(logutils.StatusMissing, model.TypeUserUnit, &logutils.FieldArgs{"app_id": claims.AppID, "org_id": claims.OrgID, "user_id": claims.Subject, "course_key": courseKey, "module_key": courseKey, "unit.key": item.UnitKey})
			}

			shouldUpdateUserUnit := true
			shouldUpdateUserUnit, lastStreakProcess, err = s.updateUserContent(storageTransaction, userUnit, item, userCourse, &now, courseConfig.StreaksNotificationsConfig)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserContent, nil, err)
			}
			if !shouldUpdateUserUnit {
				return nil
			}

			userScheduleItem, isCurrent, isRequired, updatedUserCourse, err = s.updateUserScheduleItem(storageTransaction, userUnit, item, userCourse, lastStreakProcess, &now, courseConfig.StreaksNotificationsConfig)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeScheduleItem, &logutils.FieldArgs{"current": true}, err)
			}

			err = storageTransaction.UpdateUserUnit(*userUnit)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
			}
		}

		if isCurrent && isRequired && userScheduleItem.IsComplete() {
			// update streak when user completes any required current task for the first time since last streak process
			if userCourse.CanIncrementStreak(lastStreakProcess, &now, courseConfig.StreaksNotificationsConfig) {
				userCourse.Streak++

				// if the user has no active streak and no remaining pauses, then add a streak restart (user has resumed progress after some extended time)
				if userCourse.Streak == 1 && userCourse.Pauses == 0 {
					if userCourse.StreakRestarts == nil {
						userCourse.StreakRestarts = make([]time.Time, 0)
					}
					userCourse.StreakRestarts = append(userCourse.StreakRestarts, now)
				}
			}

			userCourse.LastCompleted = &now // a current schedule item has been completed
			updatedUserCourse = true
		}

		if isRequired {
			// update pause progress and pauses when user responds to any required task for the first time since last streak process
			if userCourse.CanMakePauseProgress(lastStreakProcess, &now, courseConfig.StreaksNotificationsConfig) {
				if userCourse.Pauses < courseConfig.MaxPauses {
					userCourse.PauseProgress++

					// if the user has enough pause progress and has not reached the pause limit, add a pause
					if userCourse.PauseProgress%courseConfig.PauseProgressReward == 0 {
						userCourse.Pauses++
						userCourse.PauseProgress -= courseConfig.PauseProgressReward
					}
				}
			}

			userCourse.LastResponded = &now // the user has responded to a required task
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
	now *time.Time, snConfig model.StreaksNotificationsConfig) (bool, *time.Time, error) {
	userContentReference, isCurrent := userUnit.GetUserContentReferenceForKey(userResponse.ContentKey)
	if userContentReference == nil {
		return false, nil, errors.ErrorData(logutils.StatusMissing, model.TypeUserContentReference, &logutils.FieldArgs{"user_unit.id": userUnit.ID, "content_key": userResponse.ContentKey})
	}

	content, err := storage.FindCustomContent(userUnit.AppID, userUnit.OrgID, userResponse.ContentKey)
	if err != nil {
		return false, nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeContent, nil, err)
	}
	if content == nil {
		return false, nil, errors.ErrorData(logutils.StatusMissing, model.TypeUnit, &logutils.FieldArgs{"key": userResponse.ContentKey})
	}

	if now == nil {
		nowVal := time.Now().UTC()
		now = &nowVal
	}

	var lastStreakProcess *time.Time
	if len(userContentReference.IDs) == 0 {
		// the user may not respond to a task for the first time if it is not the current task
		if !isCurrent {
			return false, nil, errors.ErrorData(logutils.StatusInvalid, model.TypeUserContentReference, &logutils.FieldArgs{"unit_key": userUnit.Unit.Key, "content_key": content.Key, "current": false})
		}
		// the user has not saved any responses for this content yet, so create new user content
		err = s.createUserContent(storage, userUnit, userContentReference, *content, userResponse.Response, *now)
		if err != nil {
			return false, nil, err
		}
	} else {
		// the user has saved a response for this content before
		userContents, err := storage.FindUserContents(userContentReference.IDs, userUnit.AppID, userUnit.OrgID, userUnit.UserID)
		if err != nil {
			return false, nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUserContent, nil, err)
		}
		if len(userContents) == 0 {
			return false, nil, errors.ErrorData(logutils.StatusMissing, model.TypeUserContent, &logutils.FieldArgs{"ids": userContentReference.IDs})
		}

		lastUserContent := userContents[0] // FindUserContents returns UserContents sorted in reverse chronological order by DateCreated
		lastStreakProcess = userCourse.MostRecentStreakProcessTime(now, snConfig)
		if lastStreakProcess == nil {
			return false, nil, errors.ErrorData(logutils.StatusInvalid, "last streak process", &logutils.FieldArgs{"user_course.id": userCourse.ID, "process_time": snConfig.StreaksProcessTime, "timezone_name": snConfig.TimezoneName})
		}

		if lastUserContent.DateCreated.Before(*lastStreakProcess) {
			// last response was created before the most recent streak process, so create new user content
			err = s.createUserContent(storage, userUnit, userContentReference, *content, userResponse.Response, *now)
			if err != nil {
				return false, nil, err
			}
		} else {
			// the user has responded to this content since the last streak process, so update the saved response
			updatedResponse, completedNow := lastUserContent.UpdateResponse(userResponse.Response)
			updateContent := !content.Equals(&lastUserContent.Content)
			if updateContent {
				lastUserContent.Content = *content
			}
			userContentReference.Complete = userContentReference.Complete || lastUserContent.IsComplete() // reference is complete if completed now or already complete

			if updatedResponse || updateContent {
				err = storage.UpdateUserContent(lastUserContent, updateContent)
				if err != nil {
					return false, nil, errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserContent, nil, err)
				}
			}

			return completedNow, lastStreakProcess, nil // should update user unit, user course only if the task was just completed
		}
	}

	return true, lastStreakProcess, nil // user response was inserted, so update user unit, user course
}

func (s *clientImpl) createUserContent(storage interfaces.Storage, userUnit *model.UserUnit, userContentReference *model.UserContentReference, content model.Content, response map[string]interface{}, now time.Time) error {
	id := uuid.NewString()
	userContent := model.UserContent{ID: id, AppID: userUnit.AppID, OrgID: userUnit.OrgID, UserID: userUnit.UserID, CourseKey: userUnit.CourseKey,
		ModuleKey: userUnit.ModuleKey, UnitKey: userUnit.Unit.Key, Content: content, Response: response, DateCreated: now}

	err := storage.InsertUserContent(userContent)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
	}

	if userContentReference.IDs == nil {
		userContentReference.IDs = make([]string, 0)
	}
	userContentReference.IDs = append(userContentReference.IDs, id)
	userContentReference.Complete = userContentReference.Complete || userContent.IsComplete() // reference is complete if completed now or already complete
	return nil
}

func (s *clientImpl) updateUserScheduleItem(storage interfaces.Storage, userUnit *model.UserUnit, userResponse model.UserResponse, userCourse *model.UserCourse, lastStreakProcess *time.Time,
	now *time.Time, snConfig model.StreaksNotificationsConfig) (*model.UserScheduleItem, bool, bool, bool, error) {
	userScheduleItem, _, isCurrent, isRequired := userUnit.GetScheduleItem(userResponse.ContentKey, false)
	if userScheduleItem == nil {
		return nil, false, false, false, errors.ErrorData(logutils.StatusMissing, model.TypeScheduleItem, &logutils.FieldArgs{"user_unit.id": userUnit.ID, "content_key": userResponse.ContentKey})
	}

	updatedUserCourse := false
	if isCurrent && userScheduleItem.DateStarted == nil {
		userScheduleItem.DateStarted = lastStreakProcess
		if lastStreakProcess == nil {
			lastStreakProcess = userCourse.MostRecentStreakProcessTime(now, snConfig)
			userScheduleItem.DateStarted = lastStreakProcess
		}
	}
	if isCurrent && userScheduleItem.IsComplete() {
		userScheduleItem.DateCompleted = now
		if !isRequired {
			// set start time of next schedule item if immediately completing the current schedule item
			nextScheduleItem := userUnit.GetNextScheduleItem(false)
			if nextScheduleItem != nil {
				nextScheduleItem.DateStarted = lastStreakProcess
				if lastStreakProcess == nil {
					lastStreakProcess = userCourse.MostRecentStreakProcessTime(now, snConfig)
					nextScheduleItem.DateStarted = lastStreakProcess
				}
			}

			userUnit.Completed++

			// handle course, module, and unit completion if this optional schedule item is last in the
			// if there are no more required schedule items to be done in the course, set date completed
			// allow the new current schedule item to be returned if current schedule item not required because userUnit.Completed has already been incremented
			if userCourse.Course.GetNextRequiredScheduleItem(userUnit.ModuleKey, userUnit.Unit.Key, userUnit.Completed, true) == nil {
				userUnit.Current = (userUnit.Completed < userUnit.Unit.Required) // allow user to submit responses to any optional ScheduleItems remaining in the course

				if userCourse.CompletedModules == nil {
					userCourse.CompletedModules = make(map[string]time.Time)
				}
				// do not update module completion time if completing an optional schedule item
				if _, exists := userCourse.CompletedModules[userUnit.ModuleKey]; !exists {
					userCourse.CompletedModules[userUnit.ModuleKey] = *now
					if userCourse.IsComplete() {
						userCourse.DateCompleted = now // prevents streak timer from operating on any data associated with this UserCourse
					}
					updatedUserCourse = true
				}
			}

			if userUnit.Completed == userUnit.Unit.Required {
				// user has just completed the current unit by completing an optional schedule item so insert the next user unit if necessary
				nextUnit := userCourse.Course.GetNextUnit(userUnit.ModuleKey, userUnit.Unit.Key)
				if nextUnit != nil {
					nextUserSchedule := nextUnit.CreateUserSchedule()
					nextUserSchedule[0].DateStarted = lastStreakProcess
					if lastStreakProcess == nil {
						lastStreakProcess = userCourse.MostRecentStreakProcessTime(now, snConfig)
						nextUserSchedule[0].DateStarted = lastStreakProcess
					}
					nextUserUnit := model.UserUnit{ID: uuid.NewString(), AppID: userUnit.AppID, OrgID: userUnit.OrgID, UserID: userUnit.UserID, CourseKey: userUnit.CourseKey,
						ModuleKey: userUnit.ModuleKey, Unit: *nextUnit, Completed: 0, Current: true, UserSchedule: nextUserSchedule, DateCreated: time.Now().UTC()}

					err := storage.InsertUserUnit(nextUserUnit)
					if err != nil {
						return nil, false, false, false, errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
					}
				}
			}
		}
	}

	return userScheduleItem, isCurrent, isRequired, updatedUserCourse, nil
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
