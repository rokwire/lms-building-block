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

package model

import (
	"lms/utils"
	"time"
)

// NudgesConfig entity
type NudgesConfig struct {
	Active        bool   `json:"active" bson:"active"` //if the nudges processing is "on" or "off"
	GroupName     string `json:"group_name" bson:"group_name"`
	TestGroupName string `json:"test_group_name" bson:"test_group_name"`
	ProcessTime   *int   `json:"process_time" bson:"process_time"` //seconds since midnight CT at which to process nudges
	BlockSize     int    `json:"block_size" bson:"block_size"`
	Mode          string `json:"mode" bson:"mode"` // "normal" or "test"
}

// Nudge entity
type Nudge struct {
	ID           string        `json:"id" bson:"_id"`                      //last_login
	Name         string        `json:"name" bson:"name"`                   //"Last Canvas use was over 2 weeks"
	Body         string        `json:"body" bson:"body"`                   //"You have not used the Canvas Application in over 2 weeks."
	DeepLink     string        `json:"deep_link" bson:"deep_link"`         //deep link
	Params       NudgeParams   `json:"params" bson:"params"`               //Nudge specific settings
	Active       bool          `json:"active" bson:"active"`               //true or false
	UsersSources []UsersSource `json:"users_sources" bson:"users_sources"` //it says where to take the users from for this nudge - groups-bb-group, canvas courses
}

// GetUsersSourcesCanvasCoursesIDs gives the uniques canvas courses ids
func (p Nudge) GetUsersSourcesCanvasCoursesIDs() []int {
	if len(p.UsersSources) == 0 {
		return []int{}
	}
	for _, source := range p.UsersSources {
		if source.Type == "canvas-courses" {
			currentCoursesIDs := source.Params["courses_ids"]
			return utils.AnyToArrayOfInt(currentCoursesIDs)
		}
	}
	return []int{}
}

// UsersSource entity
type UsersSource struct {
	Type   string         `json:"type" bson:"type"`     //groups-bb-group or canvas-course
	Params map[string]any `json:"params" bson:"params"` //nil for groups-bb-group and a list with canvas courses for canvas-courses
}
type UsersSources struct {
	Type   *string         `json:"type" bson:"type"`     //groups-bb-group or canvas-course
	Params *map[string]any `json:"params" bson:"params"` //nil for groups-bb-group and a list with canvas courses for canvas-courses
}

// NudgeParams entity
type NudgeParams map[string]any

// Hours Retrieves hours param
func (p NudgeParams) Hours() *float64 {
	if val, ok := p["hours"]; ok {
		rValue := utils.AnyToFloat64(val)
		return &rValue
	}
	return p.DefaultHours()
}

// CourseIDs Retrieves course_ids param
func (p NudgeParams) CourseIDs() []int {
	if val, ok := p["course_ids"]; ok {
		rValue := utils.AnyToArrayOfInt(val)
		return rValue
	}
	return nil
}

// AccountIDs Retrieves account_ids param
func (p NudgeParams) AccountIDs() []int {
	if val, ok := p["account_ids"]; ok {
		rValue := utils.AnyToArrayOfInt(val)
		return rValue
	}
	return nil
}

// DefaultHours Returns the default hours param
func (p *NudgeParams) DefaultHours() *float64 {
	val := float64(0)
	return &val
}

// SentNudge entity
type SentNudge struct {
	ID           string    `json:"id" bson:"_id"`
	NudgeID      string    `json:"nudge_id" bson:"nudge_id"`
	UserID       string    `json:"user_id" bson:"user_id"`
	NetID        string    `json:"net_id" bson:"net_id"`
	CriteriaHash uint32    `json:"criteria_hash" bson:"criteria_hash"`
	DateSent     time.Time `json:"date_sent" bson:"date_sent"`
	Mode         string    `json:"mode" bson:"mode"`
}

// NudgesProcess entity
type NudgesProcess struct {
	ID          string     `json:"id" bson:"_id"`
	Mode        string     `json:"mode" bson:"mode"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	CompletedAt *time.Time `json:"completed_at" bson:"completed_at"`
	Status      string     `json:"status" bson:"status"` //processing, success, failed
	Error       *string    `json:"error" bson:"error"`
}

// Block entity
type Block struct {
	ProcessID string      `json:"process_id" bson:"process_id"`
	Number    int         `json:"number" bson:"number"`
	Items     []BlockItem `json:"items" bson:"items"`
}

// BlockItem entity
type BlockItem struct {
	NetID     string   `json:"net_id" bson:"net_id"`
	UserID    string   `json:"user_id" bson:"user_id"`
	NudgesIDs []string `json:"nudges_ids" bson:"nudges_ids"`
}
