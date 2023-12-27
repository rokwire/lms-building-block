// Copyright 2023 Board of Trustees of the University of Illinois.
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

package storage

import (
	"lms/core/model"
	"time"
)

type userCourse struct {
	ID     string `bson:"_id"`
	AppID  string `bson:"app_id"`
	OrgID  string `bson:"org_id"`
	UserID string `bson:"user_id"`

	// Notification Requirements fields
	Streak         int  `bson:"streak"`
	Pauses         int  `bson:"pauses"`
	CompletedTasks bool `bson:"completed_tasks"`

	Course course `bson:"course"`

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}

type course struct {
	ID    string `bson:"_id"`
	AppID string `bson:"app_id"`
	OrgID string `bson:"org_id"`

	Key        string   `bson:"key"`
	Name       string   `bson:"name"`
	ModuleKeys []string `bson:"module_keys"`

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}

type userModule struct {
	ID          string     `bson:"_id"`
	AppID       string     `bson:"app_id"`
	OrgID       string     `bson:"org_id"`
	UserID      string     `bson:"user_id"`
	CourseKey   string     `bson:"course_key"`
	Module      module     `bson:"module"`
	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}

type module struct {
	ID    string `bson:"_id"`
	AppID string `bson:"app_id"`
	OrgID string `bson:"org_id"`

	CourseKey string   `bson:"course_key"`
	Key       string   `bson:"key"`
	Name      string   `bson:"name"`
	UnitKeys  []string `bson:"unit_keys"`

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}

type userUnit struct {
	ID          string     `bson:"_id"`
	AppID       string     `bson:"app_id"`
	OrgID       string     `bson:"org_id"`
	UserID      string     `bson:"user_id"`
	CourseKey   string     `bson:"course_key"`
	ModuleKey   string     `bson:"module_key"`
	Unit        unit       `bson:"unit"`
	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}

type unit struct {
	ID    string `bson:"_id"`
	AppID string `bson:"app_id"`
	OrgID string `bson:"org_id"`

	CourseKey   string               `bson:"course_key"`
	ModuleKey   string               `bson:"module_key"`
	Key         string               `bson:"key"`
	Name        string               `bson:"name"`
	ContentKeys []string             `bson:"content_keys"`
	Schedule    []model.ScheduleItem `bson:"schedule"`

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}

type content struct {
	ID    string `bson:"_id"`
	AppID string `bson:"app_id"`
	OrgID string `bson:"org_id"`

	CourseKey        string          `bson:"course_key"`
	ModuleKey        string          `bson:"module_key"`
	UnitKey          string          `bson:"unit_key"`
	Key              string          `bson:"key"`
	Type             string          `bson:"type"` // assignment, resource, reward, evaluation
	Name             string          `bson:"name"`
	Details          string          `bson:"details"`
	ContentReference model.Reference `bson:"reference"`
	LinkedContent    []string        `bson:"linked_content"` // Content Keys

	DateCreated time.Time  `bson:"date_created"`
	DateUpdated *time.Time `bson:"date_updated"`
}
