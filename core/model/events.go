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
	ID                   int               `json:"id"`
	Title                string            `json:"title"`
	StartAt              *time.Time        `json:"start_at"`
	EndAt                *time.Time        `json:"end_at"`
	Description          string            `json:"desctiption"`
	LocationName         string            `json:"location_name"`
	LocationAddress      string            `json:"location_address"`
	ContextCode          string            `json:"context_code"`
	EffectiveContextCode *string           `json:"effective_context_code"`
	ContextName          string            `json:"context_name"`
	AllContextCodes      []string          `json:"all_context_codes"`
	WorkflowState        string            `json:"workflow_state"`
	Hidden               bool              `json:"hidden"`
	ParentEventID        *string           `json:"parent_event_id"`
	ChildEventsCount     int               `json:"child_events_count"`
	ChildEvents          *string           `json:"child_events"`
	Url                  string            `json:"url"`      // URL for this calendar event (to update, delete, etc.)
	HTMLUrl              string            `json:"html_url"` // URL for a user to view this event
	AllDayDate           string            `json:"all_day_date"`
	AllDay               bool              `json:"all_day"`
	CreatedAt            string            `json:"created_at"`
	UpdatedAt            string            `json:"updated_at"`
	AppointmentGroup     *AppointmentGroup `json:"appointments"`
}

type AppointmentGroup struct {
	AppointmentGroupID        *string `json:"appointment_group_id"`
	AppointmentGroupURL       *string `json:"appointment_group_url"`
	OwnReservation            bool    `json:"own_reservation"`
	ReserveURL                *string `json:"reserve_url"`
	Reserve                   bool    `json:"reserved"`
	PerticipantType           *string `json:"participant_type"`
	PerticipantPerAppointment *int    `json:"participants_per_appointment"`
	AvailableSlots            *int    `json:"available_slots"`
	User                      *string `json:"user"`
	Group                     *string `json:"group"`
	ImportantDates            bool    `json:"important_dates"`
	SeriesUUID                *string `json:"series_uuid"`
	RRule                     *int    `json:"rrule"`
}
