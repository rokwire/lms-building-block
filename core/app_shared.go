// Copyright 2022 Board of Trustees of the University of Illinois.
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

package core

import (
	"lms/core/model"
	"sync"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth/tokenauth"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

// appShared contains shared implementations
type appShared struct {
	app *Application
}

func (s *appShared) GetUserData(claims *tokenauth.Claims) (*model.UserDataResponse, error) {
	providerUserID := s.getProviderUserID(claims)
	if len(providerUserID) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, "net_id", nil)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errList error

	var providerCourses []model.ProviderCourse
	var user *model.User
	var assignments []model.Assignment
	var courses []model.UserCourse
	var units []model.UserUnit
	var contents []model.UserContent

	wg.Add(5) // Number of asynchronous operations

	// Fetch provider courses asynchronously
	go func() {
		defer wg.Done()
		c, err := s.app.provider.GetCourses(providerUserID, nil)
		mu.Lock()
		if err != nil {
			errList = errors.WrapErrorAction(logutils.ActionGet, "provider course", nil, err)
		} else {
			providerCourses = c
		}
		mu.Unlock()
	}()

	// Fetch completed assignments asynchronously
	go func() {
		defer wg.Done()
		a, err := s.app.provider.GetCompletedAssignments(providerUserID)
		mu.Lock()
		if err != nil {
			errList = errors.WrapErrorAction(logutils.ActionGet, "assignments", nil, err)
		} else {
			assignments = a
		}
		mu.Unlock()
	}()

	// Fetch user courses asynchronously
	go func() {
		defer wg.Done()
		c, err := s.app.storage.FindUserCourses(nil, claims.AppID, claims.OrgID, nil, nil, &claims.Subject, nil, nil)
		mu.Lock()
		if err != nil {
			errList = errors.WrapErrorAction(logutils.ActionGet, "courses", nil, err)
		} else {
			courses = c
		}
		mu.Unlock()
	}()

	// Fetch user units asynchronously
	go func() {
		defer wg.Done()
		u, err := s.app.storage.FindUserUnitsByUserID(claims.Subject)
		mu.Lock()
		if err != nil {
			errList = errors.WrapErrorAction(logutils.ActionGet, "user units", nil, err)
		} else {
			units = u
		}
		mu.Unlock()
	}()

	// Fetch user contents asynchronously
	go func() {
		defer wg.Done()
		co, err := s.app.storage.FindUserContents(nil, claims.AppID, claims.OrgID, claims.Subject)
		mu.Lock()
		if err != nil {
			errList = errors.WrapErrorAction(logutils.ActionGet, "user contents", nil, err)
		} else {
			contents = co
		}
		mu.Unlock()
	}()

	// Fetch current user synchronously as it's required before async operations
	user, err := s.app.provider.GetCurrentUser(providerUserID)
	if err != nil {
		return nil, err
	}

	// Wait for all async operations to complete
	wg.Wait()

	// Check if there were any errors
	if errList != nil {
		return nil, errList
	}

	// Construct the user data response
	userData := model.UserDataResponse{
		ProviderCourses:    providerCourses,
		ProviderAccount:    user,
		ProviderAssignment: assignments,
		Courses:            courses,
		Units:              units,
		Content:            contents,
	}

	return &userData, nil
}

func (s *appShared) getProviderUserID(claims *tokenauth.Claims) string {
	if claims == nil {
		return ""
	}
	return claims.ExternalIDs["net_id"]
}

// newAppShared creates new appShared
func newAppShared(app *Application) appShared {
	return appShared{app: app}
}
