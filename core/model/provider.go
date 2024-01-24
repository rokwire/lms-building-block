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

import (
	"time"
)

// ProviderUser cache entity
type ProviderUser struct {
	ID       string    `bson:"_id"`    //core BB account id
	NetID    string    `bson:"net_id"` //core BB external system id
	User     User      `bson:"user"`
	SyncDate time.Time `bson:"sync_date"`

	Courses *UserCourses `bson:"courses"`
}

// UserCourses cache entity
type UserCourses struct {
	Data     []UserCourse `bson:"data"`
	SyncDate time.Time    `bson:"sync_date"`
}

// UserCourse cache entity
type UserCourse struct {
	Data        ProviderCourse     `bson:"data"`
	Assignments []CourseAssignment `bson:"assignments"`
	SyncDate    time.Time          `bson:"sync_date"`
}

// CourseAssignment cache entity
type CourseAssignment struct {
	Data       Assignment          `bson:"data"`
	Submission *ProviderSubmission `bson:"submission"`
	SyncDate   time.Time           `bson:"sync_date"`
}

// ProviderSubmission cache entity
type ProviderSubmission struct {
	Data     *Submission `bson:"data"`
	SyncDate time.Time   `bson:"sync_date"`
}
