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

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) getCourses(l *logs.Log, providerUserID string) ([]model.Course, error) {
	courses, err := app.provider.GetCourses(providerUserID)
	if err != nil {
		l.Debugf("error getting courses - %s", err)
		return nil, err
	}
	return courses, nil
}

func (app *Application) getCourse(l *logs.Log, providerUserID string, courseID int, include *string) (*model.Course, error) {
	course, err := app.provider.GetCourse(providerUserID, courseID, include)
	if err != nil {
		l.Debugf("error getting course - %s", err)
		return nil, err
	}
	return course, nil
}

func (app *Application) getAssignmentGroups(l *logs.Log, providerUserID string, courseID int, include *string) ([]model.AssignmentGroup, error) {
	assignmentGroups, err := app.provider.GetAssignmentGroups(providerUserID, courseID, include)
	if err != nil {
		l.Debugf("error getting assignment groups - %s", err)
		return nil, err
	}
	return assignmentGroups, nil
}

func (app *Application) getCourseUser(l *logs.Log, providerUserID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error) {
	user, err := app.provider.GetCourseUser(providerUserID, courseID, includeEnrolments, includeScores)
	if err != nil {
		l.Debugf("error getting user - %s", err)
		return nil, err
	}
	return user, nil
}

func (app *Application) getCurrentUser(l *logs.Log, providerUserID string) (*model.User, error) {
	user, err := app.provider.GetCurrentUser(providerUserID)
	if err != nil {
		l.Debugf("error getting user - %s", err)
		return nil, err
	}
	return user, nil
}

// OnConfigsUpdated is called when the config collection is updates
func (app *Application) OnConfigsUpdated() {
	config, err := app.storage.FindNudgesConfig()
	if err != nil {
		app.logger.Error("error finding nudge configs on configs changed")
	}

	oldConfig := app.nudgesLogic.config
	app.nudgesLogic.config = config
	if config.ProcessTime != nil {
		if oldConfig == nil || oldConfig.ProcessTime == nil || *oldConfig.ProcessTime != *config.ProcessTime {
			app.nudgesLogic.setupNudgesTimer()
		}
	}
}
