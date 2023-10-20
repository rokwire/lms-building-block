// Copyright 2023 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 the "License";
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

import "time"

// UserCourse represents a copy of a course that the user modifies as progress is made
type UserCourse struct {
	ID     string
	AppID  string
	OrgID  string
	UserID string

	Course Course

	DateCreated time.Time
	DateUpdated *time.Time
	//TODO: add a timestamp for a user dropping a course?
}

// Course represents a custom-defined course (e.g. Essential Skills Coaching)
type Course struct {
	ID    string
	AppID string
	OrgID string

	Key     string
	Name    string
	Modules []Module

	DateCreated time.Time
	DateUpdated *time.Time
}

// Module represents an individual module of a Course (e.g. Conversational Skills)
type Module struct {
	ID    string
	AppID string
	OrgID string

	CourseKey string
	Key       string
	Name      string
	Units     []Unit

	Completed *bool
	UserData  map[string]interface{}

	DateCreated time.Time
	DateUpdated *time.Time
}

// Unit represents an individual unit of a Module (e.g. The Physical Side of Communication)
type Unit struct {
	ID    string
	AppID string
	OrgID string

	CourseKey string
	ModuleKey string
	Key       string
	Name      string
	Contents  []Content
	Schedule  []ScheduleItem

	Completed *bool
	UserData  map[string]interface{}

	DateCreated time.Time
	DateUpdated *time.Time
}

// ScheduleItem represents a set of Content items to be completed in a certain amount of time
type ScheduleItem struct {
	Name        string   `bson:"name"`
	ContentKeys []string `bson:"content_keys"`
	Duration    int      `bson:"duration"`
}

// Content represents some Unit content
type Content struct {
	ID    string
	AppID string
	OrgID string

	CourseKey        string
	ModuleKey        string
	UnitKey          string
	Key              string
	Type             string // assignment, resource, reward, evaluation
	Name             string
	Details          string
	ContentReference Reference
	LinkedContent    []Content

	Completed *bool
	UserData  map[string]interface{}

	DateCreated time.Time
	DateUpdated *time.Time
}

// Reference represents a reference to another entity
type Reference struct {
	Name         string `bson:"name"`
	Type         string `bson:"type"` // content item, video, PDF, survey, web URL
	ReferenceKey string `bson:"reference_key"`
}