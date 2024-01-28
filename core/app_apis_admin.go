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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

type adminImpl struct {
	app *Application
}

// model.NudgesProcess

func (s *adminImpl) GetNudgesConfig(claims *tokenauth.Claims) (*model.NudgesConfig, error) {
	// find the nudges config
	nudgesConfig, err := s.app.storage.FindNudgesConfig()
	if err != nil {
		return nil, err
	}
	return nudgesConfig, nil
}

// UpdateNudgesConfig(active bool, groupName string, testGroupName string, mode string, processTime *int, blockSize *int) error
func (s *adminImpl) UpdateNudgesConfig(claims *tokenauth.Claims, item model.NudgesConfig) (*model.NudgesConfig, error) {
	err := s.app.storage.SaveNudgesConfig(item)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *adminImpl) GetNudges(claims *tokenauth.Claims) ([]model.Nudge, error) {
	// find all active nudges
	nudges, err := s.app.storage.LoadAllNudges()
	if err != nil {
		return nil, err
	}
	return nudges, nil
}

// CreateNudge(ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSource) error
func (s *adminImpl) CreateNudge(claims *tokenauth.Claims, item model.Nudge) (*model.Nudge, error) {
	//create and insert nudge
	err := s.app.storage.InsertNudge(item)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// UpdateNudge(ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSources) error
func (s *adminImpl) UpdateNudge(claims *tokenauth.Claims, id string, item model.Nudge) (*model.Nudge, error) {
	item.ID = id
	err := s.app.storage.UpdateNudge(item)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *adminImpl) DeleteNudge(claims *tokenauth.Claims, id string) error {
	err := s.app.storage.DeleteNudge(id)
	if err != nil {
		return nil
	}
	return err
}

func (s *adminImpl) FindSentNudges(claims *tokenauth.Claims, nudgeID *string, userID *string, netID *string, mode *string) ([]model.SentNudge, error) {
	sentNudges, err := s.app.storage.FindSentNudges(nudgeID, userID, netID, nil, mode)
	if err != nil {
		return nil, err
	}
	return sentNudges, nil
}

func (s *adminImpl) DeleteSentNudges(claims *tokenauth.Claims, ids *string) error {
	idList := []string{}
	if ids != nil {
		idList = strings.Split(*ids, ",")
	}

	err := s.app.storage.DeleteSentNudges(idList, "")
	if err != nil {
		return err
	}
	return nil
}

func (s *adminImpl) ClearTestSentNudges(claims *tokenauth.Claims) error {
	err := s.app.storage.DeleteSentNudges(nil, "test")
	if err != nil {
		return err
	}
	return nil
}

func (s *adminImpl) FindNudgesProcesses(claims *tokenauth.Claims, limit *int, offset *int) ([]model.NudgesProcess, error) {
	limitVal := 5
	if limit != nil {
		limitVal = *limit
	}

	offsetVal := 0
	if offset != nil {
		offsetVal = *offset
	}

	nudgesProcess, err := s.app.storage.FindNudgesProcesses(limitVal, offsetVal)
	if err != nil {
		return nil, err
	}
	return nudgesProcess, nil
}

func (s *adminImpl) GetCustomCourses(claims *tokenauth.Claims, id *string, name *string, key *string, moduleKey *string) ([]model.Course, error) {
	var idArr, nameArr, keyArr, moduleKeys []string

	//parse moduleID comma seperated string into array
	if id != nil {
		idArr = strings.Split(*id, ",")
	}
	if name != nil {
		nameArr = strings.Split(*name, ",")
	}
	if key != nil {
		keyArr = strings.Split(*key, ",")
	}
	if moduleKey != nil {
		moduleKeys = strings.Split(*moduleKey, ",")
	}

	courses, err := s.app.storage.FindCustomCourses(claims.AppID, claims.OrgID, idArr, nameArr, keyArr, moduleKeys)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func (s *adminImpl) CreateCustomCourse(claims *tokenauth.Claims, item model.Course) (*model.Course, error) {
	transaction := func(storageTransaction interfaces.Storage) error {
		item.ID = uuid.NewString()
		item.AppID = claims.AppID
		item.OrgID = claims.OrgID

		//extract sublayer(s) key and struct, insert those not yet present in db according to key
		var modules, newModules []model.Module
		var units, newUnits []model.Unit
		var contents, newContents []model.Content
		var contentKeys []string

		for _, module := range item.Modules {
			for _, unit := range module.Units {
				if len(unit.Schedule) == 0 {
					return errors.ErrorData(logutils.StatusMissing, "unit schedule", &logutils.FieldArgs{"key": unit.Key})
				}
				if unit.ScheduleStart < 0 {
					return errors.ErrorData(logutils.StatusInvalid, "unit schedule start", &logutils.FieldArgs{"schedule_start": unit.ScheduleStart})
				}

				for _, content := range unit.Contents {
					if !utils.Exist[string](contentKeys, content.Key) {
						content.ID = uuid.NewString()
						content.AppID = claims.AppID
						content.OrgID = claims.OrgID
						contents = append(contents, content)
						contentKeys = append(contentKeys, content.Key)
					}
				}
				for _, scheduleItem := range unit.Schedule {
					for _, userContent := range scheduleItem.UserContent {
						if !utils.Exist[string](contentKeys, userContent.ContentKey) {
							return errors.ErrorData(logutils.StatusInvalid, "schedule content key", &logutils.FieldArgs{"content_key": userContent.ContentKey})
						}
					}
				}

				unit.ID = uuid.NewString()
				unit.AppID = claims.AppID
				unit.OrgID = claims.OrgID
				unit.Required = len(unit.Schedule)
				unit.DateCreated = time.Now().UTC()
				units = append(units, unit)
			}
			module.ID = uuid.NewString()
			module.AppID = claims.AppID
			module.OrgID = claims.OrgID
			modules = append(modules, module)
		}
		newModules, err := s.modulesNotInDB(storageTransaction, claims.AppID, claims.OrgID, modules)
		if err != nil {
			return err
		}
		newUnits, err = s.unitsNotInDB(storageTransaction, claims.AppID, claims.OrgID, units)
		if err != nil {
			return err
		}
		newContents, err = s.contentsNotInDB(storageTransaction, claims.AppID, claims.OrgID, contents)
		if err != nil {
			return err
		}

		err = storageTransaction.InsertCustomCourse(item)
		if err != nil {
			return err
		}

		if len(newModules) != 0 {
			err = storageTransaction.InsertCustomModules(newModules)
			if err != nil {
				return err
			}
		}

		if len(newUnits) != 0 {
			err = storageTransaction.InsertCustomUnits(newUnits)
			if err != nil {
				return err
			}
		}

		if len(newContents) != 0 {
			err = storageTransaction.InsertCustomContents(newContents)
			if err != nil {
				return err
			}
		}

		return nil
	}
	return nil, s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) GetCustomCourse(claims *tokenauth.Claims, key string) (*model.Course, error) {
	appID := claims.AppID
	orgID := claims.OrgID
	course, err := s.app.storage.FindCustomCourse(appID, orgID, key)
	if err != nil {
		return nil, err
	}
	return course, nil
}

func (s *adminImpl) UpdateCustomCourse(claims *tokenauth.Claims, key string, item model.Course) (*model.Course, error) {
	transaction := func(storageTransaction interfaces.Storage) error {
		item.AppID = claims.AppID
		item.OrgID = claims.OrgID

		// prevent empty key and key mismatch. current implementation disallow key update
		item.Key = key

		course, err := storageTransaction.FindCustomCourse(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		// checks if subcategory associated array keys are updated
		var curKeys, newKeys []string
		for _, val := range course.Modules {
			curKeys = append(curKeys, val.Key)
		}
		for _, val := range item.Modules {
			newKeys = append(newKeys, val.Key)
		}

		// checks if new associated array keys all present in database
		if !utils.Equal(curKeys, newKeys, false) {
			returnedStructs, err := storageTransaction.FindCustomModules(claims.AppID, claims.OrgID, nil, nil, newKeys, nil)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionFind, model.TypeModule, nil, err)
			}
			if len(newKeys) != len(returnedStructs) {
				return errors.WrapErrorData(logutils.StatusMissing, model.TypeModule, nil, err)
			}
		}

		err = storageTransaction.UpdateCustomCourse(key, item)
		if err != nil {
			return err
		}

		err = storageTransaction.UpdateUserCourses(key, item)
		if err != nil {
			return err
		}
		return nil
	}
	return nil, s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) DeleteCustomCourse(claims *tokenauth.Claims, key string) error {
	transaction := func(storageTransaction interfaces.Storage) error {
		appID := claims.AppID
		orgID := claims.OrgID
		err := storageTransaction.DeleteCustomCourse(appID, orgID, key)
		if err != nil {
			return err
		}

		// delete all derieved user course
		err = storageTransaction.DeleteUserCourses(appID, orgID, key)
		if err != nil {
			return err
		}

		return err
	}
	return s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) GetCustomModules(claims *tokenauth.Claims, id *string, name *string, key *string, unitKey *string) ([]model.Module, error) {
	var idArr, nameArr, keyArr, unitKeys []string

	//parse moduleID comma seperated string into array
	if id != nil {
		idArr = strings.Split(*id, ",")
	}
	if name != nil {
		nameArr = strings.Split(*name, ",")
	}
	if key != nil {
		keyArr = strings.Split(*key, ",")
	}
	if unitKey != nil {
		unitKeys = strings.Split(*unitKey, ",")
	}

	modules, err := s.app.storage.FindCustomModules(claims.AppID, claims.OrgID, idArr, nameArr, keyArr, unitKeys)
	if err != nil {
		return nil, err
	}
	return modules, nil
}

func (s *adminImpl) CreateCustomModule(claims *tokenauth.Claims, item model.Module) (*model.Module, error) {
	transaction := func(storageTransaction interfaces.Storage) error {
		item.ID = uuid.NewString()
		item.AppID = claims.AppID
		item.OrgID = claims.OrgID

		//extract sublayer(s) key and struct, insert those not yet present in db according to key
		var units, newUnits []model.Unit
		var contents, newContents []model.Content
		var contentKeys []string

		for _, unit := range item.Units {
			if len(unit.Schedule) == 0 {
				return errors.ErrorData(logutils.StatusMissing, "unit schedule", &logutils.FieldArgs{"key": unit.Key})
			}
			if unit.ScheduleStart < 0 {
				return errors.ErrorData(logutils.StatusInvalid, "unit schedule start", &logutils.FieldArgs{"schedule_start": unit.ScheduleStart})
			}

			for _, content := range unit.Contents {
				if !utils.Exist[string](contentKeys, content.Key) {
					content.ID = uuid.NewString()
					content.AppID = claims.AppID
					content.OrgID = claims.OrgID
					contents = append(contents, content)
					contentKeys = append(contentKeys, content.Key)
				}
			}
			for _, scheduleItem := range unit.Schedule {
				for _, userContent := range scheduleItem.UserContent {
					if !utils.Exist[string](contentKeys, userContent.ContentKey) {
						return errors.ErrorData(logutils.StatusInvalid, "schedule content key", &logutils.FieldArgs{"content_key": userContent.ContentKey})
					}
				}
			}

			unit.ID = uuid.NewString()
			unit.AppID = claims.AppID
			unit.OrgID = claims.OrgID
			unit.Required = len(unit.Schedule)
			unit.DateCreated = time.Now().UTC()
			units = append(units, unit)
		}

		newUnits, err := s.unitsNotInDB(storageTransaction, claims.AppID, claims.OrgID, units)
		if err != nil {
			return err
		}
		newContents, err = s.contentsNotInDB(storageTransaction, claims.AppID, claims.OrgID, contents)
		if err != nil {
			return err
		}

		err = storageTransaction.InsertCustomModule(item)
		if err != nil {
			return err
		}

		if len(newUnits) != 0 {
			err = storageTransaction.InsertCustomUnits(newUnits)
			if err != nil {
				return err
			}
		}

		if len(newContents) != 0 {
			err = storageTransaction.InsertCustomContents(newContents)
			if err != nil {
				return err
			}
		}

		return nil
	}
	return nil, s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) GetCustomModule(claims *tokenauth.Claims, key string) (*model.Module, error) {
	appID := claims.AppID
	orgID := claims.OrgID
	module, err := s.app.storage.FindCustomModule(appID, orgID, key)
	if err != nil {
		return nil, err
	}
	return module, nil
}

func (s *adminImpl) UpdateCustomModule(claims *tokenauth.Claims, key string, item model.Module) (*model.Module, error) {
	transaction := func(storageTransaction interfaces.Storage) error {

		item.AppID = claims.AppID
		item.OrgID = claims.OrgID

		// prevent empty key and key mismatch. current implementation disallow key update
		item.Key = key

		// checks if updated key correctly associate with existing struct in db
		module, err := storageTransaction.FindCustomModule(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		// checks if subcategory associated array keys are updated
		var curKeys, newKeys []string
		for _, val := range module.Units {
			curKeys = append(curKeys, val.Key)
		}
		for _, val := range item.Units {
			newKeys = append(newKeys, val.Key)
		}

		// checks if new associated array keys all present in database
		if !utils.Equal(curKeys, newKeys, false) {
			returnedStructs, err := storageTransaction.FindCustomUnits(claims.AppID, claims.OrgID, nil, nil, newKeys, nil)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionFind, model.TypeUnit, nil, err)
			}
			if len(newKeys) != len(returnedStructs) {
				return errors.WrapErrorData(logutils.StatusMissing, model.TypeUnit, nil, err)
			}
		}

		err = storageTransaction.UpdateCustomModule(key, item)
		if err != nil {
			return err
		}

		return nil
	}

	return nil, s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) DeleteCustomModule(claims *tokenauth.Claims, key string) error {
	transaction := func(storageTransaction interfaces.Storage) error {
		appID := claims.AppID
		orgID := claims.OrgID
		err := storageTransaction.DeleteCustomModule(appID, orgID, key)
		if err != nil {
			return err
		}

		err = storageTransaction.DeleteModuleKeyFromCourses(appID, orgID, key)
		if err != nil {
			return err
		}

		err = storageTransaction.DeleteModuleKeyFromUserCourses(appID, orgID, key)
		if err != nil {
			return err
		}

		return err
	}
	return s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) GetCustomUnits(claims *tokenauth.Claims, id *string, name *string, key *string, contentKey *string) ([]model.Unit, error) {
	var idArr, nameArr, keyArr, contentKeys []string
	if id != nil {
		idArr = strings.Split(*id, ",")
	}
	if name != nil {
		nameArr = strings.Split(*name, ",")
	}
	if key != nil {
		keyArr = strings.Split(*key, ",")
	}
	if contentKey != nil {
		contentKeys = strings.Split(*contentKey, ",")
	}

	result, err := s.app.storage.FindCustomUnits(claims.AppID, claims.OrgID, idArr, nameArr, keyArr, contentKeys)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *adminImpl) CreateCustomUnit(claims *tokenauth.Claims, item model.Unit) (*model.Unit, error) {
	transaction := func(storageTransaction interfaces.Storage) error {
		if len(item.Schedule) == 0 {
			return errors.ErrorData(logutils.StatusMissing, "unit schedule", &logutils.FieldArgs{"key": item.Key})
		}
		if item.ScheduleStart < 0 {
			return errors.ErrorData(logutils.StatusInvalid, "unit schedule start", &logutils.FieldArgs{"schedule_start": item.ScheduleStart})
		}

		item.ID = uuid.NewString()
		item.AppID = claims.AppID
		item.OrgID = claims.OrgID

		//extract sublayer(s) key and struct, insert those not yet present in db according to key
		var contents, newContents []model.Content
		var contentKeys []string

		//contents
		for _, content := range item.Contents {
			if !utils.Exist[string](contentKeys, content.Key) {
				content.ID = uuid.NewString()
				content.AppID = claims.AppID
				content.OrgID = claims.OrgID
				contents = append(contents, content)
				contentKeys = append(contentKeys, content.Key)
			}
		}
		for _, scheduleItem := range item.Schedule {
			for _, userContent := range scheduleItem.UserContent {
				if !utils.Exist[string](contentKeys, userContent.ContentKey) {
					return errors.ErrorData(logutils.StatusInvalid, "schedule content key", &logutils.FieldArgs{"content_key": userContent.ContentKey})
				}
			}
		}

		newContents, err := s.contentsNotInDB(storageTransaction, claims.AppID, claims.OrgID, contents)
		if err != nil {
			return err
		}

		if len(newContents) != 0 {
			err = storageTransaction.InsertCustomContents(newContents)
			if err != nil {
				return err
			}
		}

		item.Required = len(item.Schedule)
		item.DateCreated = time.Now().UTC()
		err = storageTransaction.InsertCustomUnit(item)
		if err != nil {
			return err
		}
		return nil
	}
	return nil, s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) GetCustomUnit(claims *tokenauth.Claims, key string) (*model.Unit, error) {
	result, err := s.app.storage.FindCustomUnit(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *adminImpl) UpdateCustomUnit(claims *tokenauth.Claims, key string, item model.Unit) (*model.Unit, error) {
	transaction := func(storageTransaction interfaces.Storage) error {
		if len(item.Schedule) == 0 {
			return errors.ErrorData(logutils.StatusMissing, "unit schedule", &logutils.FieldArgs{"key": item.Key})
		}
		if item.ScheduleStart < 0 {
			return errors.ErrorData(logutils.StatusInvalid, "unit schedule start", &logutils.FieldArgs{"schedule_start": item.ScheduleStart})
		}

		item.AppID = claims.AppID
		item.OrgID = claims.OrgID
		// prevent empty key and key mismatch. current implementation disallow key update
		item.Key = key

		unit, err := storageTransaction.FindCustomUnit(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		// checks if subcategory associated array keys are updated
		var curKeys, newKeys []string
		for _, val := range unit.Contents {
			curKeys = append(curKeys, val.Key)
		}
		for _, val := range item.Contents {
			newKeys = append(newKeys, val.Key)
		}

		// checks if new associated array keys all present in database
		if !utils.Equal(curKeys, newKeys, false) {
			returnedStructs, err := storageTransaction.FindCustomContents(claims.AppID, claims.OrgID, nil, nil, newKeys)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionFind, model.TypeContent, nil, err)
			}
			if len(newKeys) != len(returnedStructs) {
				return errors.WrapErrorData(logutils.StatusMissing, model.TypeContent, nil, err)
			}
		}

		err = storageTransaction.UpdateCustomUnit(key, item)
		if err != nil {
			return err
		}

		err = storageTransaction.UpdateUserUnits(key, item)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, nil, err)
		}

		return nil
	}
	return nil, s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) DeleteCustomUnit(claims *tokenauth.Claims, key string) error {
	transaction := func(storageTransaction interfaces.Storage) error {
		err := storageTransaction.DeleteCustomUnit(claims.AppID, claims.OrgID, key)
		if err != nil {
			return nil
		}
		// delete userUnit derived from customUnit
		err = storageTransaction.DeleteUserUnit(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		err = storageTransaction.DeleteUnitKeyFromModules(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		return nil
	}
	return s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) GetCustomContents(claims *tokenauth.Claims, id *string, name *string, key *string) ([]model.Content, error) {
	var idArr, nameArr, keyArr []string
	if id != nil {
		idArr = strings.Split(*id, ",")
	}
	if name != nil {
		nameArr = strings.Split(*name, ",")
	}
	if key != nil {
		keyArr = strings.Split(*key, ",")
	}

	result, err := s.app.storage.FindCustomContents(claims.AppID, claims.OrgID, idArr, nameArr, keyArr)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *adminImpl) CreateCustomContent(claims *tokenauth.Claims, item model.Content) (*model.Content, error) {
	item.ID = uuid.NewString()
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID
	err := s.app.storage.InsertCustomContent(item)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *adminImpl) GetCustomContent(claims *tokenauth.Claims, key string) (*model.Content, error) {
	result, err := s.app.storage.FindCustomContent(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *adminImpl) UpdateCustomContent(claims *tokenauth.Claims, key string, item model.Content) (*model.Content, error) {
	transaction := func(storageTransaction interfaces.Storage) error {
		item.AppID = claims.AppID
		item.OrgID = claims.OrgID

		// prevent empty key and key mismatch. current implementation disallow key update
		item.Key = key

		content, err := storageTransaction.FindCustomContent(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		// checks if subcategory associated array keys are updated
		var curKeys, newKeys []string
		curKeys = append(curKeys, content.LinkedContent...)
		newKeys = append(newKeys, item.LinkedContent...)

		// checks if new associated array keys all present in database
		if !utils.Equal(curKeys, newKeys, false) {
			returnedStructs, err := storageTransaction.FindCustomContents(claims.AppID, claims.OrgID, nil, nil, newKeys)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionFind, model.TypeContent, nil, err)
			}
			if len(newKeys) != len(returnedStructs) {
				return errors.WrapErrorData(logutils.StatusMissing, model.TypeContent, nil, err)
			}
		}

		err = storageTransaction.UpdateCustomContent(key, item)
		if err != nil {
			return err
		}
		return nil
	}
	return nil, s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) DeleteCustomContent(claims *tokenauth.Claims, key string) error {
	transaction := func(storageTransaction interfaces.Storage) error {
		err := storageTransaction.DeleteCustomContent(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		err = storageTransaction.DeleteContentKeyFromLinkedContents(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		err = storageTransaction.DeleteContentKeyFromUnits(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}

		err = storageTransaction.DeleteContentKeyFromUserUnits(claims.AppID, claims.OrgID, key)
		if err != nil {
			return err
		}
		return err
	}
	return s.app.storage.PerformTransaction(transaction)
}

func (s *adminImpl) GetCustomCourseConfigs(claims *tokenauth.Claims) ([]model.CourseConfig, error) {
	courseConfigs, err := s.app.storage.FindCourseConfigs(&claims.AppID, &claims.OrgID, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourseConfig, nil, err)
	}

	return courseConfigs, nil
}

func (s *adminImpl) CreateCustomCourseConfig(claims *tokenauth.Claims, item model.CourseConfig) (*model.CourseConfig, error) {
	err := item.StreaksNotificationsConfig.ValidateTimings()
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, model.TypeStreaksNotificationsConfig, nil, err)
	}

	item.ID = uuid.NewString()
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID
	item.DateCreated = time.Now().UTC()

	return nil, s.app.storage.InsertCourseConfig(item)
}

func (s *adminImpl) GetCustomCourseConfig(claims *tokenauth.Claims, key string) (*model.CourseConfig, error) {
	courseConfig, err := s.app.storage.FindCourseConfig(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourseConfig, nil, err)
	}

	return courseConfig, nil
}

func (s *adminImpl) UpdateCustomCourseConfig(claims *tokenauth.Claims, key string, item model.CourseConfig) (*model.CourseConfig, error) {
	err := item.StreaksNotificationsConfig.ValidateTimings()
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionValidate, model.TypeStreaksNotificationsConfig, nil, err)
	}

	item.AppID = claims.AppID
	item.OrgID = claims.OrgID
	item.CourseKey = key

	return nil, s.app.storage.UpdateCourseConfig(item)
}

func (s *adminImpl) DeleteCustomCourseConfig(claims *tokenauth.Claims, key string) error {
	return s.app.storage.DeleteCourseConfig(claims.AppID, claims.OrgID, key)
}

// return those inside the array that are not present in database determined by key
func (s *adminImpl) modulesNotInDB(storage interfaces.Storage, appID string, orgID string, modules []model.Module) ([]model.Module, error) {
	var keys, returnedKeys []string
	var resultStructs []model.Module
	for _, val := range modules {
		keys = append(keys, val.Key)
	}
	returnedStructs, err := storage.FindCustomModules(appID, orgID, nil, nil, keys, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeModule, nil, err)
	}

	for _, val := range returnedStructs {
		returnedKeys = append(returnedKeys, val.Key)
	}

	for _, dataStruct := range modules {
		if !utils.Exist(returnedKeys, dataStruct.Key) {
			resultStructs = append(resultStructs, dataStruct)
		}
	}
	return resultStructs, nil
}

// return those inside the array that are not present in database determined by key
func (s *adminImpl) unitsNotInDB(storage interfaces.Storage, appID string, orgID string, units []model.Unit) ([]model.Unit, error) {
	var keys, returnedKeys []string
	var resultStructs []model.Unit
	for _, val := range units {
		keys = append(keys, val.Key)
	}
	returnedStructs, err := storage.FindCustomUnits(appID, orgID, nil, nil, keys, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUnit, nil, err)
	}

	for _, val := range returnedStructs {
		returnedKeys = append(returnedKeys, val.Key)
	}

	for _, dataStruct := range units {
		if !utils.Exist(returnedKeys, dataStruct.Key) {
			resultStructs = append(resultStructs, dataStruct)
		}
	}
	return resultStructs, nil
}

// return those inside the array that are not present in database determined by key
func (s *adminImpl) contentsNotInDB(storage interfaces.Storage, appID string, orgID string, contents []model.Content) ([]model.Content, error) {
	var keys, returnedKeys []string
	var resultStructs []model.Content
	for _, val := range contents {
		keys = append(keys, val.Key)
	}
	returnedStructs, err := storage.FindCustomContents(appID, orgID, nil, nil, keys)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeContent, nil, err)
	}

	for _, val := range returnedStructs {
		returnedKeys = append(returnedKeys, val.Key)
	}

	for _, dataStruct := range contents {
		if !utils.Exist(returnedKeys, dataStruct.Key) {
			resultStructs = append(resultStructs, dataStruct)
		}
	}
	return resultStructs, nil
}
