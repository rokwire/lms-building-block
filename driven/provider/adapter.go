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

package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"lms/core/model"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//Adapter implements the Provider interface
type Adapter struct {
	host      string
	token     string
	tokenType string
}

//GetCourses gets the user courses
func (a *Adapter) GetCourses(userID string) ([]model.Course, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/courses%s", queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting courses")
		return nil, err
	}

	//prepare the response and return it
	var courses []model.Course
	err = json.Unmarshal(data, &courses)
	if err != nil {
		log.Print("error converting courses")
		return nil, err
	}
	return courses, nil
}

//GetCourse gives the the course for the provided id
func (a *Adapter) GetCourse(userID string, courseID int, include *string) (*model.Course, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	if include != nil {
		queryParamsItems["include[]"] = []string{*include}
	}
	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/courses/%d%s", courseID, queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting courses")
		return nil, err
	}

	//prepare the response and return it
	var course *model.Course
	err = json.Unmarshal(data, &course)
	if err != nil {
		log.Print("error converting course")
		return nil, err
	}
	return course, nil
}

//GetAssignmentGroups gives the the course assignment groups for the user
func (a *Adapter) GetAssignmentGroups(userID string, courseID int, include *string) ([]model.AssignmentGroup, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	if include != nil {
		queryParamsItems["include[]"] = []string{*include}
	}
	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/courses/%d/assignment_groups%s", courseID, queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting courses")
		return nil, err
	}

	//prepare the response and return it
	var assignmentGroups []model.AssignmentGroup
	err = json.Unmarshal(data, &assignmentGroups)
	if err != nil {
		log.Print("error converting course")
		return nil, err
	}
	return assignmentGroups, nil
}

//GetCourseUser gives the course user
func (a *Adapter) GetCourseUser(userID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	includes := []string{}
	if includeEnrolments {
		includes = append(includes, "enrollments")
	}
	if includeScores {
		includes = append(includes, "current_grading_period_scores")
	}
	if len(includes) > 0 {
		queryParamsItems["include[]"] = includes
	}

	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/courses/%d/users/self%s", courseID, queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting courses")
		return nil, err
	}

	//prepare the response and return it
	var user *model.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		log.Print("error converting users")
		return nil, err
	}
	return user, nil
}

//GetCurrentUser gives the current user
func (a *Adapter) GetCurrentUser(userID string) (*model.User, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/users/self%s", queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting courses")
		return nil, err
	}

	//prepare the response and return it
	var user *model.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		log.Print("error converting users")
		return nil, err
	}
	return user, nil
}

//GetLastLogin gives the last login date for the user
func (a *Adapter) GetLastLogin(userID string) (*time.Time, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	queryParamsItems["include[]"] = []string{"last_login"}
	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/users/self%s", queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting last login")
		return nil, err
	}

	//prepare the response and return it
	var user *model.User
	err = json.Unmarshal(data, &user)
	if err != nil {
		log.Print("error converting user")
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return user.LastLogin, nil
}

//GetMissedAssignments gives the missed assignments of the user
func (a *Adapter) GetMissedAssignments(userID string) ([]model.Assignment, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/users/self/missing_submissions%s", queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting last login")
		return nil, err
	}

	//prepare the response and return it
	var assignments []model.Assignment
	err = json.Unmarshal(data, &assignments)
	if err != nil {
		log.Print("error converting missing assignments")
		return nil, err
	}

	return assignments, nil
}

//GetCompletedAssignments gives the completed assignments of the user
func (a *Adapter) GetCompletedAssignments(userID string) ([]model.Assignment, error) {
	//1. first we need to find all courses for the user
	userCourses, err := a.GetCourses(userID)
	if err != nil {
		log.Print("error getting user courses for early completed assignments")
		return nil, err
	}
	if len(userCourses) == 0 {
		//not courses for this user
		return nil, nil
	}

	//2. get the assignemnts for every course
	result := []model.Assignment{}
	for _, course := range userCourses {
		courseAssignments, err := a.getAssignments(course.ID, userID)
		if err != nil {
			log.Printf("error getting assignments for - %d - %s", course.ID, userID)
			continue
		}
		if len(courseAssignments) == 0 {
			continue
		}

		//check submission for every assignment
		for _, cAssignment := range courseAssignments {
			// get only the submitted ones
			submission := cAssignment.Submission
			if submission != nil && submission.SubmittedAt != nil {
				result = append(result, cAssignment)
			}
		}
	}

	return result, nil
}

func (a *Adapter) getAssignments(courseID int, userID string) ([]model.Assignment, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	queryParamsItems["include[]"] = []string{"submission"}
	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/courses/%d/assignments%s", courseID, queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting assignments")
		return nil, err
	}

	//prepare the response and return it
	var assignments []model.Assignment
	err = json.Unmarshal(data, &assignments)
	if err != nil {
		log.Print("error converting assignments")
		return nil, err
	}
	return assignments, nil
}

func (a *Adapter) constructQueryParams(items map[string][]string) string {
	if len(items) == 0 {
		return ""
	}

	values := url.Values{}

	for k, list := range items {
		for _, listItem := range list {
			values.Add(k, listItem)
		}
	}

	query := values.Encode()
	return fmt.Sprintf("?%s", query)
}

func (a *Adapter) executeQuery(body io.Reader, pathAndParams string, method string) ([]byte, error) {
	//body
	requestBody, err := ioutil.ReadAll(body)
	if err != nil {
		log.Printf("error getting body - %s", pathAndParams)
		return nil, err
	}

	//url
	url := fmt.Sprintf("%s%s", a.host, pathAndParams)

	//request
	req, err := http.NewRequest(method, url, strings.NewReader(string(requestBody)))
	if err != nil {
		log.Printf("error creating request - %s", pathAndParams)
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", a.tokenType, a.token))

	//execute
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error executing request - %s", pathAndParams)
		return nil, err
	}
	defer resp.Body.Close()

	//return the response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error converting response body - %s", pathAndParams)
		return nil, err
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		//we have an error
		errorMessage := fmt.Sprintf("error with response code %d: %s", resp.StatusCode, string(data))
		log.Print(errorMessage)
		return nil, errors.New(errorMessage)
	}

	return data, nil
}

//GetCalendarEvents gives the events of the user
func (a *Adapter) GetCalendarEvents(userID string, startAt time.Time, endAt time.Time) ([]model.CalendarEvent, error) {
	//TODO

	fmt.Println(startAt.Format(time.RFC3339))
	fmt.Println(endAt.Format(time.RFC3339))
	return nil, nil
	/*	//params
		queryParamsItems := map[string][]string{}
		queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
		queryParamsItems["course_id"] = []string{fmt.Sprintf("course_id:%s", courseID)}
		queryParamsItems["start_at"] = []string{fmt.Sprintf("start_at:%s", startAt)}
		queryParamsItems["end_at"] = []string{fmt.Sprintf("end_at:%s", endAt)}
		queryParamsItems["per_page"] = []string{fmt.Sprintf("per_page:%s", perPage)}

		queryParams := a.constructQueryParams(queryParamsItems)

		//path + params
		pathAndParams := fmt.Sprintf("api/v1/calendar_events%s", queryParams, courseID, startAt, endAt, perPage)

		//execute query
		data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
		if err != nil {
			log.Print("error getting last login")
			return nil, err
		}

		//prepare the response and return it
		var calendarEvents []model.CalendarEvent
		err = json.Unmarshal(data, &calendarEvents)
		if err != nil {
			log.Print("error converting missing calendar events")
			return nil, err
		}

		return calendarEvents, nil */
}

//NewProviderAdapter creates a new provider adapter
func NewProviderAdapter(host string, token string, tokenType string) *Adapter {
	return &Adapter{host: host, token: token, tokenType: tokenType}
}
