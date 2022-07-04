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
	courses, err := app.Provider.GetCourses(providerUserID)
	if err != nil {
		l.Debugf("error getting courses - %s", err)
		return nil, err
	}
	return courses, nil
}

func (app *Application) getCourse(l *logs.Log, providerUserID string, courseID int, include *string) (*model.Course, error) {
	course, err := app.Provider.GetCourse(providerUserID, courseID, include)
	if err != nil {
		l.Debugf("error getting course - %s", err)
		return nil, err
	}
	return course, nil
}

func (app *Application) getAssignmentGroups(l *logs.Log, providerUserID string, courseID int, include *string) ([]model.AssignmentGroup, error) {
	assignmentGroups, err := app.Provider.GetAssignmentGroups(providerUserID, courseID, include)
	if err != nil {
		l.Debugf("error getting assignment groups - %s", err)
		return nil, err
	}
	return assignmentGroups, nil
}

func (app *Application) getUsers(l *logs.Log, providerUserID string, courseID int, includeEnrolments bool, includeScores bool) ([]model.User, error) {
	return nil, nil
}

// OnCollectionUpdated callback that indicates the reward types collection is changed
func (app *Application) OnCollectionUpdated(name string) {

}
