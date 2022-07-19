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

//CalendarEvent entity
type CalendarEvent struct {
	ID              int        `json:"id"`
	Title           string     `json:"title"`
	StartAt         *time.Time `json:"start_at"`
	EndAt           *time.Time `json:"end_at"`
	Description     string     `json:"desctiption"`
	LocationName    string     `json:"location_name"`
	LocationAddress string     `json:"location_address"`
	Url             string     `json:"url"`      // URL for this calendar event (to update, delete, etc.)
	HTMLUrl         string     `json:"html_url"` // URL for a user to view this event
	AllDayDate      string     `json:"all_day_date"`
	AllDay          bool       `json:"all_day"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}
