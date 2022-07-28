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

	"github.com/rokwire/logging-library-go/logs"
)

func (app *Application) getNudges() ([]model.Nudge, error) {
	// find all active nudges
	nudges, err := app.storage.LoadAllNudges()
	if err != nil {
		return nil, err
	}
	return nudges, nil
}

func (app *Application) createNudge(l *logs.Log, ID string, name string, body string, deepLink string, params *map[string]interface{}, active bool) error {
	//create and insert nudge
	nudge := model.Nudge{ID: ID, Name: name, Body: body, DeepLink: deepLink, Params: *params, Active: active}
	err := app.storage.InsertNudge(nudge)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) updateNudge(l *logs.Log, ID string, name string, body string, deepLink string, params *map[string]interface{}, active bool) error {
	err := app.storage.UpdateNudge(ID, name, body, deepLink, params, active)
	if err != nil {
		return nil
	}
	return err
}

func (app *Application) deleteNudge(l *logs.Log, ID string) error {
	err := app.storage.DeleteNudge(ID)
	if err != nil {
		return nil
	}
	return err
}

func (app *Application) findSentNudges(l *logs.Log, nudgeID *string, userID *string, netID *string, criteriaHash *[]uint32, mode *string) ([]model.SentNudge, error) {
	sentNudges, _ := app.storage.FindSentNudges(nudgeID, userID, netID, criteriaHash, mode)
	if sentNudges == nil {
		return nil, nil
	}
	return sentNudges, nil
}

func (app *Application) deleteSentNudges(l *logs.Log, ids []string) error {
	err := app.storage.DeleteSentNudges(ids)
	if err != nil {
		return nil
	}
	return err
}
