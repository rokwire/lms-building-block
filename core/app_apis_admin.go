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
	"errors"
	"lms/core/model"
	"strings"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
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
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) CreateCustomCourse(claims *tokenauth.Claims, item model.Course) (*model.Course, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) GetCustomCourse(claims *tokenauth.Claims, id string) (*model.Course, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) UpdateCustomCourse(claims *tokenauth.Claims, id string, item model.Course) (*model.Course, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) DeleteCustomCourse(claims *tokenauth.Claims, id string) error {
	//TODO: implement
	return errors.New(logutils.Unimplemented)
}

func (s *adminImpl) GetCustomModules(claims *tokenauth.Claims, id *string, name *string, key *string, unitID *string) ([]model.Module, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) CreateCustomModule(claims *tokenauth.Claims, item model.Module) (*model.Module, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) GetCustomModule(claims *tokenauth.Claims, id string) (*model.Module, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) UpdateCustomModule(claims *tokenauth.Claims, id string, item model.Module) (*model.Module, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) DeleteCustomModule(claims *tokenauth.Claims, id string) error {
	//TODO: implement
	return errors.New(logutils.Unimplemented)
}

func (s *adminImpl) GetCustomUnits(claims *tokenauth.Claims, id *string, name *string, key *string, contentID *string) ([]model.Unit, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) CreateCustomUnit(claims *tokenauth.Claims, item model.Unit) (*model.Unit, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) GetCustomUnit(claims *tokenauth.Claims, id string) (*model.Unit, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) UpdateCustomUnit(claims *tokenauth.Claims, id string, item model.Unit) (*model.Unit, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) DeleteCustomUnit(claims *tokenauth.Claims, id string) error {
	//TODO: implement
	return errors.New(logutils.Unimplemented)
}

func (s *adminImpl) GetCustomContents(claims *tokenauth.Claims, id *string, name *string, key *string) ([]model.Content, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) CreateCustomContent(claims *tokenauth.Claims, item model.Content) (*model.Content, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) GetCustomContent(claims *tokenauth.Claims, id string) (*model.Content, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) UpdateCustomContent(claims *tokenauth.Claims, id string, item model.Content) (*model.Content, error) {
	//TODO: implement
	return nil, errors.New(logutils.Unimplemented)
}

func (s *adminImpl) DeleteCustomContent(claims *tokenauth.Claims, id string) error {
	//TODO: implement
	return errors.New(logutils.Unimplemented)
}
