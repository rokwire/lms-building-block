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

package model

// CoreService wrapper record for the corresponding service record response
type CoreService struct {
	Host      string `json:"host"`
	ServiceID string `json:"service_id"`
}

// CoreAccount wraps the account structure from the Core BB
// @name CoreAccount
type CoreAccount struct {
	AuthTypes []struct {
		Active     bool   `json:"active"`
		Code       string `json:"code"`
		ID         string `json:"id"`
		Identifier string `json:"identifier"`
		Params     struct {
			User struct {
				Email          string        `json:"email"`
				FirstName      string        `json:"first_name"`
				Groups         []interface{} `json:"groups"`
				Identifier     string        `json:"identifier"`
				LastName       string        `json:"last_name"`
				MiddleName     string        `json:"middle_name"`
				Roles          []string      `json:"roles"`
				SystemSpecific struct {
					PreferredUsername string `json:"preferred_username"`
				} `json:"system_specific"`
			} `json:"user"`
		} `json:"params"`
	} `json:"auth_types"`
	Groups      []interface{} `json:"groups"`
	ID          string        `json:"id"`
	Permissions []interface{} `json:"permissions"`
	Preferences struct {
		Favorites interface{} `json:"favorites"`
		Interests struct {
		} `json:"interests"`
		PrivacyLevel int      `json:"privacy_level"`
		Roles        []string `json:"roles"`
		Settings     struct {
		} `json:"settings"`
		Tags struct {
		} `json:"tags"`
		Voter struct {
			RegisteredVoter bool        `json:"registered_voter"`
			VotePlace       string      `json:"vote_place"`
			Voted           bool        `json:"voted"`
			VoterByMail     interface{} `json:"voter_by_mail"`
		} `json:"voter"`
	} `json:"preferences"`
	Profile struct {
		Address   string `json:"address"`
		BirthYear int    `json:"birth_year"`
		Country   string `json:"country"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		ID        string `json:"id"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		PhotoURL  string `json:"photo_url"`
		State     string `json:"state"`
		ZipCode   string `json:"zip_code"`
	} `json:"profile"`
	Roles []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"roles"`
}

// GetUIN Gets the uin
func (a *CoreAccount) GetUIN() *string {
	for _, auth := range a.AuthTypes {
		return &auth.Params.User.Identifier
	}
	return nil
}

// GetNetID Gets the NetID
func (a *CoreAccount) GetNetID() *string {
	for _, auth := range a.AuthTypes {
		return &auth.Params.User.ExternalIDs.NetID
	}
	return nil
}
