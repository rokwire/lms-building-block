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

import "time"

//NudgesConfig entity
type NudgesConfig struct {
	Active        bool   `json:"active" bson:"active"` //if the nudges processing is "on" or "off"
	GroupName     string `json:"group_name" bson:"group_name"`
	TestGroupName string `json:"test_group_name" bson:"test_group_name"`
	Mode          string `json:"mode" bson:"mode"` // "normal" or "test"
}

//Nudge entity
type Nudge struct {
	ID       string                 `json:"id" bson:"_id"`              //last_login
	Name     string                 `json:"name" bson:"name"`           //"Last Canvas use was over 2 weeks"
	Body     string                 `json:"body" bson:"body"`           //"You have not used the Canvas Application in over 2 weeks."
	DeepLink string                 `json:"deep_link" bson:"deep_link"` //deep link
	Params   map[string]interface{} `json:"params" bson:"params"`       //Nudge specific settings
	Active   bool                   `json:"active" bson:"active"`       //true or false
}

//SentNudge entity
type SentNudge struct {
	ID           string    `json:"id" bson:"_id"`
	NudgeID      string    `json:"nudge_id" bson:"nudge_id"`
	UserID       string    `json:"user_id" bson:"user_id"`
	NetID        string    `json:"net_id" bson:"net_id"`
	CriteriaHash uint32    `json:"criteria_hash" bson:"criteria_hash"`
	DateSent     time.Time `json:"date_sent" bson:"date_sent"`
	Mode         string    `json:"mode" bson:"mode"`
}
