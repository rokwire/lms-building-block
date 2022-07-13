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

	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logs"
)

func (app *Application) getNudges() ([]model.Nudge, error) {
	// find all the nudges
	nudges, err := app.storage.LoadAllNudges()
	if err != nil {
		return nil, nil
	}
	if nudges == nil {
		return nil, errors.New("can't find the nudges")
	}
	return nudges, nil
}

func (app *Application) createNudge(l *logs.Log, ID string, name string, body string, params *map[string]interface{}) error {
	//create and insert nudge
	nudge := model.Nudge{ID: ID, Name: name, Body: body, Params: *params}
	err := app.storage.InsertNudge(nudge)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) updateNudge(l *logs.Log, ID string, name string, body string, params *map[string]interface{}) error {
	err := app.storage.UpdateNudge(ID, name, body, params)
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
