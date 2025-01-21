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

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
)

// appShared contains shared implementations
type appShared struct {
	app *Application
}

func (s *appShared) GetUserData(claims *tokenauth.Claims) (*model.UserDataResponse, error) {
	var (
		nudgesBlocksResponse  []model.NudgesBlocksResponse
		nudgesProcessResponse []model.NudgesProcessesResponse
		sendNudgesResponse    []model.SentNudgeResponse
		userContentsResponse  []model.UserContentResponse
		userCoursesResponse   []model.UserCoursesResponse
		userUnitsResponse     []model.UserUnitsResponse
		processIDs            []string
	)

	errChan := make(chan error, 6) // To collect errors from goroutines
	done := make(chan struct{})    // To signal completion of all goroutines
	defer close(errChan)

	// Load Nudges Blocks
	go func() {
		defer func() { done <- struct{}{} }()
		nudgesBlocks, err := s.app.storage.LoadNudgesBlocksByUserID(claims.Subject)
		if err != nil {
			errChan <- err
			return
		}

		for _, nb := range nudgesBlocks {
			nbr := model.NudgesBlocksResponse{ID: nb.ProcessID, UserID: claims.Subject}
			nudgesBlocksResponse = append(nudgesBlocksResponse, nbr)
			processIDs = append(processIDs, nb.ProcessID)
		}
	}()

	// Load Nudges Process
	go func() {
		defer func() { done <- struct{}{} }()
		nudgesProcess, err := s.app.storage.FindNudgesProcesses(0, 0)
		if err != nil {
			errChan <- err
			return
		}

		for _, np := range nudgesProcess {
			for _, n := range processIDs {
				if np.ID == n {
					npr := model.NudgesProcessesResponse{ID: np.ID, UserID: claims.Subject, Status: np.Status}
					nudgesProcessResponse = append(nudgesProcessResponse, npr)
				}
			}
		}
	}()

	// Load Sent Nudges
	go func() {
		defer func() { done <- struct{}{} }()
		sendNudges, err := s.app.storage.FindSendNudgesByUserID(claims.Subject)
		if err != nil {
			errChan <- err
			return
		}

		for _, sn := range sendNudges {
			snr := model.SentNudgeResponse{UserID: claims.Subject, ID: sn.ID, NudgeID: sn.NudgeID}
			sendNudgesResponse = append(sendNudgesResponse, snr)
		}
	}()

	// Load User Contents
	go func() {
		defer func() { done <- struct{}{} }()
		userContents, err := s.app.storage.FindUserContents(nil, claims.AppID, claims.OrgID, claims.Subject)
		if err != nil {
			errChan <- err
			return
		}

		for _, uc := range userContents {
			ucr := model.UserContentResponse{ID: uc.ID, UserID: uc.UserID}
			userContentsResponse = append(userContentsResponse, ucr)
		}
	}()

	// Load User Courses
	go func() {
		defer func() { done <- struct{}{} }()
		userCourses, err := s.app.storage.FindUserCourses(nil, claims.AppID, claims.OrgID, nil, nil, &claims.Subject, nil, nil)
		if err != nil {
			errChan <- err
			return
		}

		for _, ucours := range userCourses {
			ucoursR := model.UserCoursesResponse{ID: ucours.ID, UserID: ucours.UserID}
			userCoursesResponse = append(userCoursesResponse, ucoursR)
		}
	}()

	// Load User Units
	go func() {
		defer func() { done <- struct{}{} }()
		userUnits, err := s.app.storage.FindUserUnitsByUserID(claims.Subject)
		if err != nil {
			errChan <- err
			return
		}

		for _, uu := range userUnits {
			uur := model.UserUnitsResponse{ID: uu.ID, UserID: uu.UserID}
			userUnitsResponse = append(userUnitsResponse, uur)
		}
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 6; i++ {
		<-done
	}

	// Check if there are any errors
	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	// Combine all data into the response
	userData := model.UserDataResponse{
		NudgesBlocksResponse:  nudgesBlocksResponse,
		NudgesProcessResponse: nudgesProcessResponse,
		SentNudgeResponse:     sendNudgesResponse,
		UserContentResponse:   userContentsResponse,
		UserCoursesResponse:   userCoursesResponse,
		UserUnitsResponse:     userUnitsResponse,
	}

	return &userData, nil
}

// newAppShared creates new appShared
func newAppShared(app *Application) appShared {
	return appShared{app: app}
}
