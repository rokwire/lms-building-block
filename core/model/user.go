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

// // UserDataResponse represents a user data response
type UserDataResponse struct {
	SentNudgeResponse     []SentNudgeResponse       `json:"sent_nudges"`
	NudgesProcessResponse []NudgesProcessesResponse `json:"nudges_processes"`
	NudgesBlocksResponse  []NudgesBlocksResponse    `json:"nudges_blocks"`
	UserContentResponse   []UserContentResponse     `json:"user_contents"`
	UserCoursesResponse   []UserCoursesResponse     `json:"user_courses"`
	UserUnitsResponse     []UserUnitsResponse       `json:"user_units"`
}

// SentNudgeResponse entity
type SentNudgeResponse struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

// NudgesProcessesResponse entity
type NudgesProcessesResponse struct {
	ID     string `json:"id" bson:"_id"`
	UserID string `json:"user_id"`
}

// NudgesBlocksResponse entity
type NudgesBlocksResponse struct {
	ID     string `json:"id" bson:"_id"`
	UserID string `json:"user_id"`
}

// UserContentResponse entity
type UserContentResponse struct {
	ID     string `json:"id" bson:"_id"`
	UserID string `json:"user_id"`
}

// UserCoursesResponse entity
type UserCoursesResponse struct {
	ID     string `json:"id" bson:"_id"`
	UserID string `json:"user_id"`
}

// UserUnitsResponse entity
type UserUnitsResponse struct {
	ID     string `json:"id" bson:"_id"`
	UserID string `json:"user_id"`
}
