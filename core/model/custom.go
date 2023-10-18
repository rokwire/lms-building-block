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

// Course represents a custom-defined course (e.g. Essential Skills Coaching)
type Course struct {
	ID    string `json:"id" bson:"_id"`
	AppID string `json:"app_id" bson:"app_id"`
	OrgID string `json:"org_id" bson:"org_id"`

	Key        string   `json:"key" bson:"key"`
	Name       string   `json:"name" bson:"name"`
	ModuleKeys []string `json:"module_keys" bson:"module_keys"`

	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// Module represents an individual module of a Course (e.g. Conversational Skills)
type Module struct {
	ID    string `json:"id" bson:"_id"`
	AppID string `json:"app_id" bson:"app_id"`
	OrgID string `json:"org_id" bson:"org_id"`

	CourseKey string   `json:"course_key" bson:"course_key"`
	Key       string   `json:"key" bson:"key"`
	Name      string   `json:"name" bson:"name"`
	UnitKeys  []string `json:"unit_keys" bson:"unit_keys"`

	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// Unit represents an individual unit of a Module (e.g. The Physical Side of Communication)
type Unit struct {
	ID    string `json:"id" bson:"_id"`
	AppID string `json:"app_id" bson:"app_id"`
	OrgID string `json:"org_id" bson:"org_id"`

	CourseKey   string         `json:"course_key" bson:"course_key"`
	ModuleKey   string         `json:"module_key" bson:"module_key"`
	Key         string         `json:"key" bson:"key"`
	Name        string         `json:"name" bson:"name"`
	ContentKeys []string       `json:"content_keys" bson:"content_keys"`
	Schedule    []ScheduleItem `json:"schedule" bson:"schedule"`

	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// ScheduleItem represents a set of Content items to be completed in a certain amount of time
type ScheduleItem struct {
	Name        string   `json:"name" bson:"name"`
	ContentKeys []string `json:"content_keys" bson:"content_keys"`
	Duration    int      `json:"duration" bson:"duration"`
}

// Content represents some Unit content
type Content struct {
	ID    string `json:"id" bson:"_id"`
	AppID string `json:"app_id" bson:"app_id"`
	OrgID string `json:"org_id" bson:"org_id"`

	CourseKey        string    `json:"course_key" bson:"course_key"`
	ModuleKey        string    `json:"module_key" bson:"module_key"`
	UnitKey          string    `json:"unit_key" bson:"unit_key"`
	Key              string    `json:"key" bson:"key"`
	Type             string    `json:"type" bson:"type"` // assignment, resource, reward, evaluation
	Name             string    `json:"name" bson:"name"`
	Details          string    `json:"details" bson:"details"`
	ContentReference Reference `json:"reference" bson:"reference"`
	LinkedContent    []string  `json:"linked_content" bson:"linked_content"` // Content Keys

	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// Reference represents a reference to another entity
type Reference struct {
	Name         string `json:"name" bson:"name"`
	Type         string `json:"type" bson:"type"` // content item, video, PDF, survey, web URL
	ReferenceKey string `json:"reference_key" bson:"reference_key"`
}
