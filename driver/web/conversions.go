// Copyright 2023 Board of Trustees of the University of Illinois.
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

package web

import (
	"errors"
	"lms/core/model"
	Def "lms/driver/web/docs/gen"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

func nudgesConfigFromDef(claims *tokenauth.Claims, item *Def.NudgesConfig) (*model.NudgesConfig, error) {
	if item == nil {
		return nil, nil
	}

	blockSizeVal := 50
	if item.BlockSize != nil {
		blockSizeVal = *item.BlockSize
	}

	nudgesConfig := model.NudgesConfig{Active: item.Active, GroupName: item.GroupName, TestGroupName: item.TestGroupName, Mode: string(item.Mode),
		ProcessTime: item.ProcessTime, BlockSize: blockSizeVal}
	return &nudgesConfig, errors.New(logutils.Unimplemented)
}

func nudgeFromDefAdminReqCreate(claims *tokenauth.Claims, item *Def.AdminReqCreateNudge) (*model.Nudge, error) {
	if item == nil {
		return nil, nil
	}

	var usersSources []model.UsersSource
	if item.UsersSources != nil {
		usersSources := make([]model.UsersSource, len(*item.UsersSources))
		for i, u := range *item.UsersSources {
			usersSources[i] = model.UsersSource{Type: u.Type}
			if u.Params != nil {
				usersSources[i].Params = *u.Params
			}
		}
	}

	nudge := model.Nudge{ID: item.Id, Name: item.Name, Body: item.Body, DeepLink: item.DeepLink, Params: item.Params, Active: item.Active, UsersSources: usersSources}
	return &nudge, nil
}

func nudgeFromDefAdminReqUpdate(claims *tokenauth.Claims, item *Def.AdminReqUpdateNudge) (*model.Nudge, error) {
	if item == nil {
		return nil, nil
	}

	var usersSources []model.UsersSource
	if item.UsersSources != nil {
		usersSources := make([]model.UsersSource, len(*item.UsersSources))
		for i, u := range *item.UsersSources {
			usersSources[i] = model.UsersSource{Type: u.Type}
			if u.Params != nil {
				usersSources[i].Params = *u.Params
			}
		}
	}

	nudge := model.Nudge{Name: item.Name, Body: item.Body, DeepLink: item.DeepLink, Params: item.Params, Active: item.Active, UsersSources: usersSources}
	return &nudge, nil
}
