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

//Course entity
type Course struct {
	ID                     int    `json:"id"`
	Name                   string `json:"name"`
	AccessRestrictedByDate bool   `json:"access_restricted_by_date"`
}

//Assignment entity
type Assignment struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	CourseID int        `json:"course_id"`
	HTMLUrl  string     `json:"html_url"`
	Position *int       `json:"position"`
	DueAt    *time.Time `json:"due_at"`
}

//AssignmentGroup entity
type AssignmentGroup struct {
	ID          int          `json:"id"`
	Assignments []Assignment `json:"assignments"`
}

//Grade entity
type Grade struct {
	CurrentScore *float64 `json:"current_score"`
}

//Enrollment entity
type Enrollment struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Grade *Grade `json:"grade"`
}

//User entity
type User struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	LastLogin   *time.Time   `json:"last_login"`
	Enrollments []Enrollment `json:"enrollments"`
}
