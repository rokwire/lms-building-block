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

import (
	"time"

	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	//TypeUserCourse user course type
	TypeUserCourse logutils.MessageDataType = "user course"
	//TypeUserUnit user unit type
	TypeUserUnit logutils.MessageDataType = "user unit"
	//TypeCourse course type
	TypeCourse logutils.MessageDataType = "course"
	//TypeModule module type
	TypeModule logutils.MessageDataType = "module"
	//TypeUnit unit type
	TypeUnit logutils.MessageDataType = "unit"
	//TypeContent content type
	TypeContent logutils.MessageDataType = "content"
)

// UserCourse represents a copy of a course that the user modifies as progress is made
type UserCourse struct {
	ID     string `json:"id"`
	AppID  string `json:"app_id"`
	OrgID  string `json:"org_id"`
	UserID string `json:"user_id"`

	Course Course `json:"course"`

	DateCreated time.Time  `json:"date_created"`
	DateUpdated *time.Time `json:"date_updated"`
	DateDropped *time.Time `json:"dete_dropped"`

	//TODO: add a timestamp for a user dropping a course?
}

// Course represents a custom-defined course (e.g. Essential Skills Coaching)
type Course struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`
	OrgID string `json:"org_id"`

	Key     string   `json:"key"`
	Name    string   `json:"name"`
	Modules []Module `json:"modules"`

	DateCreated time.Time
	DateUpdated *time.Time
}

// Module represents an individual module of a Course (e.g. Conversational Skills)
type Module struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`
	OrgID string `json:"org_id"`

	//CourseKey string `json:"course_key"`
	Key   string `json:"key"`
	Name  string `json:"name"`
	Units []Unit `json:"units"`

	DateCreated time.Time
	DateUpdated *time.Time
}

// UserUnit represents a copy of a unit that the user modifies as progress is made
type UserUnit struct {
	ID        string `json:"id"`
	AppID     string `json:"app_id"`
	OrgID     string `json:"org_id"`
	UserID    string `json:"user_id"`
	CourseKey string `json:"course_key"`
	//ModuleKey   string     `json:"module_key"`
	Unit        Unit       `json:"unit"`
	DateCreated time.Time  `json:"date_created"`
	DateUpdated *time.Time `json:"date_updated"`
}

// Unit represents an individual unit of a Module (e.g. The Physical Side of Communication)
type Unit struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`
	OrgID string `json:"org_id"`

	//CourseKey string         `json:"course_key"`
	//ModuleKey string         `json:"module_key"`
	Key      string         `json:"key"`
	Name     string         `json:"name"`
	Contents []Content      `json:"content"`
	Schedule []ScheduleItem `json:"schedule"`

	DateCreated time.Time
	DateUpdated *time.Time
}

// ScheduleItem represents a set of Content items to be completed in a certain amount of time
type ScheduleItem struct {
	Name        string          `bson:"name" json:"name"`
	UserContent []UserReference `bson:"user_content" json:"user_content"`
	Duration    int             `bson:"duration" json:"duration"`
}

// Content represents some Unit content
type Content struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`
	OrgID string `json:"org_id"`

	//CourseKey        string    `json:"course_key"`
	//ModuleKey        string    `json:"module_key"`
	//UnitKey          string    `json:"unit_key"`
	Key              string    `json:"key"`
	Type             string    `json:"type"` // assignment, resource, reward, evaluation
	Name             string    `json:"name"`
	Details          string    `json:"details"`
	ContentReference Reference `json:"reference"`
	LinkedContent    []string  `json:"linked_content"`

	DateCreated time.Time
	DateUpdated *time.Time
}

// Reference represents a reference to another entity
type Reference struct {
	Name         string `bson:"name" json:"name"`
	Type         string `bson:"type" json:"type"` // content item, video, PDF, survey, web URL
	ReferenceKey string `bson:"reference_key" json:"reference_key"`
}

// UserReference represents a reference with some additional data about user interactions
type UserReference struct {
	Reference

	// user fields (populated as user takes a course)
	UserData      map[string]interface{} `bson:"user_data,omitempty" json:"user_data,omitempty"`
	DateStarted   *time.Time             `bson:"date_started,omitempty" json:"date_started,omitempty"`
	DateCompleted *time.Time             `bson:"date_completed,omitempty" json:"date_completed,omitempty"`
}
