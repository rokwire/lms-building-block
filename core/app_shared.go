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
	nudgesBlocks, err := s.app.storage.LoadNudgesBlocksByUserID(claims.Subject)
	if err != nil {
		return nil, err
	}
	var nudgesBlocksResponse []model.NudgesBlocksResponse
	var nudgesProcessResponse []model.NudgesProcessesResponse
	var sendNudgesResponse []model.SentNudgeResponse
	var processIDs []string
	for _, nb := range nudgesBlocks {
		nbr := model.NudgesBlocksResponse{ID: nb.ProcessID, UserID: claims.Subject}
		nudgesBlocksResponse = append(nudgesBlocksResponse, nbr)
		processIDs = append(processIDs, nb.ProcessID)
	}

	nudgesProcess, err := s.app.storage.FindNudgesProcesses(0, 0)
	if err != nil {
		return nil, err
	}

	for _, np := range nudgesProcess {
		for _, n := range processIDs {
			if np.ID == n {
				npr := model.NudgesProcessesResponse{ID: np.ID, UserID: claims.Subject, Status: np.Status}
				nudgesProcessResponse = append(nudgesProcessResponse, npr)
			}
		}
	}

	sendNudges, err := s.app.storage.FindSendNudgesByUserID(claims.Subject)
	if err != nil {
		return nil, err
	}

	for _, sn := range sendNudges {
		snr := model.SentNudgeResponse{UserID: claims.Subject, ID: sn.ID, NudgeID: sn.NudgeID}
		sendNudgesResponse = append(sendNudgesResponse, snr)
	}

	userData := model.UserDataResponse{NudgesBlocksResponse: nudgesBlocksResponse, NudgesProcessResponse: nudgesProcessResponse, SentNudgeResponse: sendNudgesResponse}
	return &userData, nil
}

// newAppShared creates new appShared
func newAppShared(app *Application) appShared {
	return appShared{app: app}
}
