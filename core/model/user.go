/*
 *   Copyright (c) 2023 Board of Trustees of the University of Illinois.
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

package model

// DeletedUserData represents a user-deleted
type DeletedUserData struct {
	AppID       string              `json:"app_id"`
	Memberships []DeletedMembership `json:"memberships"`
	OrgID       string              `json:"org_id"`
}

// DeletedMembership defines model for DeletedMembership.
type DeletedMembership struct {
	AccountID   string                  `json:"account_id"`
	Context     *map[string]interface{} `json:"context,omitempty"`
	ExternalIDs *map[string]string      `json:"external_ids"`
}

// UserDataResponse represents a user data response
type UserDataResponse struct {
	ProviderCourses    []ProviderCourse `json:"my_provider_courses"`
	ProviderAssignment []Assignment     `json:"my_provider_assignments"`
	ProviderAccount    *ProviderUser    `json:"provider_account"`
	Courses            []UserCourse     `json:"my_courses"`
	Units              []UserUnit       `json:"my_unit"`
	Content            []UserContent    `json:"my_contents"`
}
