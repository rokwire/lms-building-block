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

// ProviderCourse entity
type ProviderCourse struct {
	ID                     int    `json:"id"  bson:"id"`
	Name                   string `json:"name" bson:"name"`
	AccessRestrictedByDate bool   `json:"access_restricted_by_date" bson:"access_restricted_by_date"`
	AccountID              int    `json:"account_id" bson:"account_id"`
}

// Assignment entity
type Assignment struct {
	ID         int         `json:"id" bson:"id"`
	Name       string      `json:"name" bson:"name"`
	CourseID   int         `json:"course_id" bson:"course_id"`
	HTMLUrl    string      `json:"html_url" bson:"html_url"`
	Position   *int        `json:"position" bson:"position"`
	CreatedAt  *time.Time  `json:"created_at" bson:"created_at"`
	DueAt      *time.Time  `json:"due_at" bson:"due_at"`
	Submission *Submission `json:"submission" bson:"submission"`
}

// Submission entity
type Submission struct {
	ID          int        `json:"id" bson:"id"`
	SubmittedAt *time.Time `json:"submitted_at" bson:"submitted_at"`
}

// AssignmentGroup entity
type AssignmentGroup struct {
	ID          int          `json:"id" bson:"id"`
	Assignments []Assignment `json:"assignments" bson:"assignments"`
}

// Grade entity
type Grade struct {
	CurrentScore *float64 `json:"current_score" bson:"current_score"`
}

// Enrollment entity
type Enrollment struct {
	ID    int    `json:"id" bson:"id"`
	Type  string `json:"type" bson:"type"`
	Grade *Grade `json:"grades" bson:"grades"`
}

// User entity
type User struct {
	ID          int          `json:"id" bson:"id"`
	Name        string       `json:"name" bson:"name"`
	LoginID     string       `json:"login_id" bson:"login_id"`
	LastLogin   *time.Time   `json:"last_login" bson:"last_login"`
	Enrollments []Enrollment `json:"enrollments" bson:"enrollments"`
}

// CalendarEvent entity
type CalendarEvent struct {
	ID    int    `json:"id" bson:"id"`
	Title string `json:"title" bson:"title"`
}
