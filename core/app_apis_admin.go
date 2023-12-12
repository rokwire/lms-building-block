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
	"strings"

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

func (s *adminImpl) GetCustomCourses(claims *tokenauth.Claims, id *string, name *string, key *string, moduleID *string) ([]model.Course, error) {
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
	if moduleID != nil {
		moduleKeys = strings.Split(*moduleID, ",")
	}

	courses, err := s.app.storage.GetCustomCourses(claims.AppID, claims.OrgID, idArr, nameArr, keyArr, moduleKeys)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

func (s *adminImpl) CreateCustomCourse(claims *tokenauth.Claims, item model.Course) (*model.Course, error) {
	item.ID = uuid.NewString()
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID
	err := s.app.storage.InsertCustomCourse(item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *adminImpl) GetCustomCourse(claims *tokenauth.Claims, key string) (*model.Course, error) {
	appID := claims.AppID
	orgID := claims.OrgID
	course, err := s.app.storage.GetCustomCourse(appID, orgID, key)
	if err != nil {
		return nil, err
	}
	return course, nil
}

func (s *adminImpl) UpdateCustomCourse(claims *tokenauth.Claims, key string, item model.Course) (*model.Course, error) {
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID

	if item.Key == "" {
		return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"item key empty": item.Key}, nil)
	}

	err := s.app.storage.UpdateCustomCourse(key, item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *adminImpl) DeleteCustomCourse(claims *tokenauth.Claims, key string) error {
	appID := claims.AppID
	orgID := claims.OrgID
	err := s.app.storage.DeleteCustomCourse(appID, orgID, key)
	if err != nil {
		return nil
	}
	return err
}

func (s *adminImpl) GetCustomModules(claims *tokenauth.Claims, id *string, name *string, key *string, unitID *string) ([]model.Module, error) {
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
	if unitID != nil {
		unitKeys = strings.Split(*unitID, ",")
	}

	modules, err := s.app.storage.GetCustomModules(claims.AppID, claims.OrgID, idArr, nameArr, keyArr, unitKeys)
	if err != nil {
		return nil, err
	}
	return modules, nil
}

func (s *adminImpl) CreateCustomModule(claims *tokenauth.Claims, item model.Module) (*model.Module, error) {
	item.ID = uuid.NewString()
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID
	err := s.app.storage.InsertCustomModule(item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *adminImpl) GetCustomModule(claims *tokenauth.Claims, key string) (*model.Module, error) {
	appID := claims.AppID
	orgID := claims.OrgID
	module, err := s.app.storage.GetCustomModule(appID, orgID, key)
	if err != nil {
		return nil, err
	}
	return module, nil
}

func (s *adminImpl) UpdateCustomModule(claims *tokenauth.Claims, key string, item model.Module) (*model.Module, error) {
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID

	if item.Key == "" {
		return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"item key empty": item.Key}, nil)
	}
	// checks if updated key correctly associate with existing struct in db
	module, err := s.app.storage.GetCustomModule(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, err
	}
	if module.CourseKey != item.CourseKey {
		_, err = s.app.storage.GetCustomCourse(claims.AppID, claims.OrgID, item.CourseKey)
		if err != nil {
			return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"course key non-exist": item.CourseKey}, err)
		}
	}

	err = s.app.storage.UpdateCustomModule(key, item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *adminImpl) DeleteCustomModule(claims *tokenauth.Claims, key string) error {
	appID := claims.AppID
	orgID := claims.OrgID
	err := s.app.storage.DeleteCustomModule(appID, orgID, key)
	if err != nil {
		return nil
	}
	return err
}

func (s *adminImpl) GetCustomUnits(claims *tokenauth.Claims, id *string, name *string, key *string, contentID *string) ([]model.Unit, error) {
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
	result, err := s.app.storage.GetCustomUnits(claims.AppID, claims.OrgID, idArr, nameArr, keyArr)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *adminImpl) CreateCustomUnit(claims *tokenauth.Claims, item model.Unit) (*model.Unit, error) {
	item.ID = uuid.NewString()
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID

	err := s.app.storage.InsertCustomUnit(item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *adminImpl) GetCustomUnit(claims *tokenauth.Claims, key string) (*model.Unit, error) {
	result, err := s.app.storage.GetCustomUnit(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *adminImpl) UpdateCustomUnit(claims *tokenauth.Claims, key string, item model.Unit) (*model.Unit, error) {
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID

	if item.Key == "" {
		return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"item key empty": item.Key}, nil)
	}
	// checks if updated key correctly associate with existing struct in db
	unit, err := s.app.storage.GetCustomUnit(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, err
	}
	if unit.CourseKey != item.CourseKey {
		_, err = s.app.storage.GetCustomCourse(claims.AppID, claims.OrgID, item.CourseKey)
		if err != nil {
			return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"course key non-exist": item.CourseKey}, err)
		}
	}

	if unit.ModuleKey != item.ModuleKey {
		_, err = s.app.storage.GetCustomModule(claims.AppID, claims.OrgID, item.ModuleKey)
		if err != nil {
			return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"module key non-exist": item.ModuleKey}, err)
		}
	}

	err = s.app.storage.UpdateCustomUnit(key, item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *adminImpl) DeleteCustomUnit(claims *tokenauth.Claims, key string) error {
	err := s.app.storage.DeleteCustomUnit(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil
	}
	return err
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

	result, err := s.app.storage.GetCustomContents(claims.AppID, claims.OrgID, idArr, nameArr, keyArr)
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
	return &item, nil
}

func (s *adminImpl) GetCustomContent(claims *tokenauth.Claims, key string) (*model.Content, error) {
	result, err := s.app.storage.GetCustomContent(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *adminImpl) UpdateCustomContent(claims *tokenauth.Claims, key string, item model.Content) (*model.Content, error) {
	item.AppID = claims.AppID
	item.OrgID = claims.OrgID

	if item.Key == "" {
		return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"item key empty": item.Key}, nil)
	}
	// checks if updated key correctly associate with existing struct in db
	content, err := s.app.storage.GetCustomContent(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil, err
	}
	if content.CourseKey != item.CourseKey {
		_, err = s.app.storage.GetCustomCourse(claims.AppID, claims.OrgID, item.CourseKey)
		if err != nil {
			return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"course key non-exist": item.CourseKey}, err)
		}
	}

	if content.ModuleKey != item.ModuleKey {
		_, err = s.app.storage.GetCustomModule(claims.AppID, claims.OrgID, item.ModuleKey)
		if err != nil {
			return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"module key non-exist": item.ModuleKey}, err)
		}
	}

	if content.UnitKey != item.UnitKey {
		_, err = s.app.storage.GetCustomUnit(claims.AppID, claims.OrgID, item.UnitKey)
		if err != nil {
			return nil, errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"unit key non-exist": item.UnitKey}, err)
		}
	}

	err = s.app.storage.UpdateCustomContent(key, item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *adminImpl) DeleteCustomContent(claims *tokenauth.Claims, key string) error {
	err := s.app.storage.DeleteCustomContent(claims.AppID, claims.OrgID, key)
	if err != nil {
		return nil
	}
	return err
}
