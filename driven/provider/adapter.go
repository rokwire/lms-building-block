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
	"fmt"
	"io"
	"lms/core/interfaces"
	"lms/core/model"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

// Adapter implements the Provider interface
type Adapter struct {
	host      string
	token     string
	tokenType string

	storage interfaces.Storage

	logger *logs.Logger
}

// GetCourses gets the user courses
func (a *Adapter) GetCourses(userID string, limit *int) ([]model.ProviderCourse, error) {
	return a.loadCourses(userID, limit)
}

// GetCourse gives the the course for the provided id
func (a *Adapter) GetCourse(userID string, courseID int) (*model.ProviderCourse, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
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
	var course *model.ProviderCourse
	err = json.Unmarshal(data, &course)
	if err != nil {
		log.Print("error converting course")
		return nil, err
	}
	return course, nil
}

// GetAssignmentGroups gives the the course assignment groups for the user
func (a *Adapter) GetAssignmentGroups(userID string, courseID int, includeAssignments bool, includeSubmission bool) ([]model.AssignmentGroup, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	includes := []string{}
	if includeAssignments {
		includes = append(includes, "assignments")
	}
	if includeSubmission {
		includes = append(includes, "submission")
	}

	if len(includes) > 0 {
		queryParamsItems["include[]"] = includes
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

// GetCourseUsers gives the all users for a course
func (a *Adapter) GetCourseUsers(courseID int) ([]model.User, error) {

	//path
	pathAndParams := fmt.Sprintf("/api/v1/courses/%d/users", courseID)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Print("error getting courses")
		return nil, err
	}

	//prepare the response and return it
	var users []model.User
	err = json.Unmarshal(data, &users)
	if err != nil {
		log.Printf("error converting users for course %d", courseID)
		return nil, err
	}
	return users, nil
}

// GetCourseUser gives the course user
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

// GetCurrentUser gives the current user
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

// CacheCommonData caches users and courses data
func (a *Adapter) CacheCommonData(usersIDs map[string]string) error {
	//1. cache users
	err := a.cacheUsers(usersIDs)
	if err != nil {
		return err
	}

	//2. cache users courses and courses assignments
	err = a.cacheUsersCoursesAndCoursesAssignments(usersIDs)
	if err != nil {
		return err
	}

	return nil
}

func (a *Adapter) cacheUsers(usersIDs map[string]string) error {
	a.logger.Info("start processing cacheUsers")

	for netID, userID := range usersIDs {
		a.cacheUser(netID, userID)
	}

	return nil
}

func (a *Adapter) cacheUser(netID string, userID string) error {
	a.logger.Infof("cache user - %s", netID)

	// check if the user exist
	exists, err := a.storage.UserExist(netID)
	if err != nil {
		a.logger.Errorf("error checking if the user exists - %s", netID)
		return err
	}

	if *exists {
		a.logger.Infof("%s exists, so not cache it", netID)
		return nil
	}

	a.logger.Infof("%s needs to be cached", netID)
	//load it from the provider
	loadedUser, err := a.loadUser(netID)
	if err != nil {
		a.logger.Errorf("error loading user - %s", netID)
		return err
	}
	//store it
	providerUser := model.ProviderUser{ID: userID, NetID: netID, User: *loadedUser, SyncDate: time.Now()}
	err = a.storage.InsertUser(providerUser)
	if err != nil {
		a.logger.Errorf("error inserting user - %s", netID)
		return err
	}

	return nil
}

func (a *Adapter) loadUser(netID string) (*model.User, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", netID)}
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

	return user, nil
}

func (a *Adapter) cacheUsersCoursesAndCoursesAssignments(usersIDs map[string]string) error {
	a.logger.Info("start processing cacheUsersCoursesAndCoursesAssignments")

	//for now process record by record..

	var err error

	//We do not ask the provider for every user. The courses and the assignemnts are the same as entities for the different users
	// and we use already what we have found
	allCourses := map[int]model.ProviderUserCourse{}

	for netID := range usersIDs {
		allCourses, err = a.cacheUserCoursesAndCoursesAssignments(netID, allCourses)
		if err != nil {
			a.logger.Errorf("error on caching user courses for - %s", netID)
			return err
		}
	}

	return nil
}

func (a *Adapter) cacheUserCoursesAndCoursesAssignments(netID string, allCourses map[int]model.ProviderUserCourse) (map[int]model.ProviderUserCourse, error) {
	a.logger.Infof("cache user courses and courses assignments - %s", netID)

	//get the user from the cache
	cachedUser, err := a.storage.FindUser(netID)
	if err != nil {
		a.logger.Errorf("error finding user for - %s", netID)
		return nil, err
	}

	//check if the user has courses data
	if cachedUser != nil {
		if cachedUser.Courses == nil {
			a.logger.Infof("there is no cached courses for %s, so loading them", netID)

			var userCourses *model.ProviderUserCourses
			userCourses, allCourses, err = a.loadCoursesAndAssignments(netID, allCourses)
			if err != nil {
				a.logger.Errorf("error loading user courses for - %s", netID)
				return nil, err
			}

			//add the courses data to the user
			cachedUser.Courses = userCourses

			//cache the user
			err = a.storage.SaveUser(*cachedUser)
			if err != nil {
				a.logger.Errorf("error saving user - %s", netID)
				return nil, err
			}
		} else {
			a.logger.Infof("there is cached courses for %s, so need to decide if we have to to refresh it", netID)

			currentUserCourses := cachedUser.Courses
			passedTimeInSecconds := time.Now().Unix() - currentUserCourses.SyncDate.Unix()

			//432000 seconds  = 5 days - to put it in the config
			if passedTimeInSecconds > 432000 {
				//if passedTimeInSecconds > 1 {
				a.logger.Infof("we need to refresh courses for - %s", netID)

				var loadedUserCourses *model.ProviderUserCourses
				loadedUserCourses, allCourses, err = a.loadCoursesAndAssignments(netID, allCourses)
				if err != nil {
					a.logger.Errorf("error loading user courses for - %s on refresh", netID)
					return nil, err
				}

				//do not loose the submissions when we refresh the courses data/submissions are not part of it/
				readyUserCourses := a.getSubmissionsFromCurrent(*currentUserCourses, *loadedUserCourses)

				//add the courses data to the user
				cachedUser.Courses = &readyUserCourses

				//cache the user
				err = a.storage.SaveUser(*cachedUser)
				if err != nil {
					a.logger.Errorf("error saving user - %s", netID)
					return nil, err
				}
			} else {
				a.logger.Infof("no need to refresh courses for - %s", netID)
			}
		}
	}

	return allCourses, nil
}

// puts the submissions data from the current to the new one. The new one does not have submissions in it, so we do not want to loose it.
func (a *Adapter) getSubmissionsFromCurrent(current model.ProviderUserCourses, new model.ProviderUserCourses) model.ProviderUserCourses {
	userCourses := new.Data
	if len(userCourses) == 0 {
		//no courses
		return new
	}

	resultUserCourses := make([]model.ProviderUserCourse, len(userCourses))
	for i, course := range userCourses {

		assignments := course.Assignments

		resultAssignments := make([]model.CourseAssignment, len(assignments))
		for j, assignment := range assignments {
			assignment.Submission = a.findSubmission(assignment.Data.ID, current)
			resultAssignments[j] = assignment
		}
		course.Assignments = resultAssignments
		resultUserCourses[i] = course
	}

	new.Data = resultUserCourses
	return new
}

func (a *Adapter) findSubmission(assignmentID int, current model.ProviderUserCourses) *model.ProviderSubmission {
	userCourses := current.Data
	if len(userCourses) == 0 {
		return nil
	}

	for _, course := range userCourses {
		assignments := course.Assignments

		if len(assignments) == 0 {
			continue
		}

		for _, assignment := range assignments {
			if assignment.Data.ID == assignmentID {
				return assignment.Submission
			}
		}
	}

	return nil
}

// check if the courses are available in allCourses otherwise load them
func (a *Adapter) loadCoursesAndAssignments(netID string, allCourses map[int]model.ProviderUserCourse) (*model.ProviderUserCourses, map[int]model.ProviderUserCourse, error) {
	//prepare the result variable
	now := time.Now()
	loadedUserCourses := model.ProviderUserCourses{SyncDate: now}
	data := []model.ProviderUserCourse{} //to be loaded in the function

	// first load the courses for the id
	courses, err := a.loadCourses(netID, nil)
	if err != nil {
		a.logger.Errorf("error loading user courses from the provider for - %s", netID)
		return nil, nil, err
	}

	//loop through all user courses and determine if they are already loaded or need to be loaded from the provider
	for _, course := range courses {
		//check if already exists
		value, ok := allCourses[course.ID]
		if ok {
			a.logger.Infof("we have course %d in the memory, so use it", course.ID)
			data = append(data, value)
		} else {
			a.logger.Infof("we do NOT have course %d in the memory, so need to load the data for it", course.ID)
			courseData, err := a.loadCourseData(netID, course, now)
			if err != nil {
				a.logger.Errorf("error loading course data for course and user - %d - %s", course.ID, netID)
				return nil, nil, err
			}
			if courseData == nil {
				return nil, nil, errors.Newf("there is no course data for - %d - %s", course.ID, netID)
			}
			data = append(data, *courseData)

			//keep the loaded data in the memory
			allCourses[course.ID] = *courseData
		}
	}

	//set the loaded user courses
	loadedUserCourses.Data = data

	return &loadedUserCourses, allCourses, nil
}

func (a *Adapter) loadCourseData(netID string, course model.ProviderCourse, syncDate time.Time) (*model.ProviderUserCourse, error) {
	now := time.Now()
	userCourse := model.ProviderUserCourse{Data: course, Assignments: nil, SyncDate: now}
	//to load the assignments

	loadedAssignments, err := a.getAssignments(course.ID, netID, false)
	if err != nil {
		a.logger.Errorf("error getting assignments for course and user - %d - %s", course.ID, netID)
		// some cources have restricted access, so we do not have to fail if we meet a such course
	}

	assignments := make([]model.CourseAssignment, len(loadedAssignments))
	for i, assignment := range loadedAssignments {
		assignments[i] = model.CourseAssignment{Data: assignment, Submission: nil, SyncDate: syncDate}
	}

	//add the loaded assignments
	userCourse.Assignments = assignments

	return &userCourse, nil
}

func (a *Adapter) loadCourses(userID string, limit *int) ([]model.ProviderCourse, error) {
	//params
	queryParamsItems := map[string][]string{}

	//as_user_id
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}

	//limit
	if limit != nil {
		queryParamsItems["per_page"] = []string{fmt.Sprintf("%d", *limit)}
	}

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
	var courses []model.ProviderCourse
	err = json.Unmarshal(data, &courses)
	if err != nil {
		log.Print("error converting courses")
		return nil, err
	}
	return courses, nil
}

// FindUsersByCanvasUserID finds cached users by canvas user ids
func (a *Adapter) FindUsersByCanvasUserID(canvasUserIds []int) ([]model.ProviderUser, error) {
	return a.storage.FindUsersByCanvasUserID(canvasUserIds)
}

// FindCachedData finds a cached data
func (a *Adapter) FindCachedData(usersIDs []string) ([]model.ProviderUser, error) {
	return a.storage.FindUsers(usersIDs)
}

// GetLastLogin gives the last login date for the user
func (a *Adapter) GetLastLogin(userID string) (*time.Time, error) {
	//TODO remove this function
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

// CacheUserData caches the user object
func (a *Adapter) CacheUserData(user model.ProviderUser) (*model.ProviderUser, error) {
	//1. load the user from the provider
	loadedUser, err := a.loadUser(user.NetID)
	if err != nil {
		return nil, err
	}

	//2 update the new user and store it
	user.User = *loadedUser
	user.SyncDate = time.Now()
	err = a.storage.SaveUser(user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// CacheUserCoursesData caches the user courses data
func (a *Adapter) CacheUserCoursesData(user model.ProviderUser, coursesIDs []int) (*model.ProviderUser, error) {
	if len(coursesIDs) == 0 {
		return &user, nil
	}

	// load the assignments(+submissions) data for all courses
	newData := map[int][]model.Assignment{}
	for _, courseID := range coursesIDs {
		assignments, err := a.getAssignments(courseID, user.NetID, true)
		if err != nil {
			return nil, err
		}
		newData[courseID] = assignments
	}

	//add the new data to the user object
	currentUserCourses := user.Courses.Data
	newUserCoursesData := []model.ProviderUserCourse{}
	for _, uc := range currentUserCourses {
		//get the data from the loaded ones
		loadedAssignments, has := newData[uc.Data.ID]
		if has {
			//use the new data

			now := time.Now()
			newCAs := make([]model.CourseAssignment, len(loadedAssignments))
			for j, assignment := range loadedAssignments {

				submission := model.ProviderSubmission{Data: assignment.Submission, SyncDate: now}

				newCA := model.CourseAssignment{Data: assignment, Submission: &submission, SyncDate: now}
				newCAs[j] = newCA
			}

			nuc := model.ProviderUserCourse{Data: uc.Data, Assignments: newCAs, SyncDate: now}
			newUserCoursesData = append(newUserCoursesData, nuc)
		} else {
			//use the old one
			newUserCoursesData = append(newUserCoursesData, uc)
		}

	}
	user.Courses.Data = newUserCoursesData

	//save the updated user data
	err := a.storage.SaveUser(user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetMissedAssignments gives the missed assignments of the user
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

// GetCompletedAssignments gives the completed assignments of the user
func (a *Adapter) GetCompletedAssignments(userID string) ([]model.Assignment, error) {
	//1. first we need to find all courses for the user
	userCourses, err := a.GetCourses(userID, nil)
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
		courseAssignments, err := a.getAssignments(course.ID, userID, true)
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

func (a *Adapter) getAssignments(courseID int, userID string, includeSubmission bool) ([]model.Assignment, error) {
	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", userID)}
	if includeSubmission {
		queryParamsItems["include[]"] = []string{"submission"}
	}
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

// GetCalendarEvents gives the events of the user
func (a *Adapter) GetCalendarEvents(netID string, providerUserID int, courseID int, startAt time.Time, endAt time.Time) ([]model.CalendarEvent, error) {
	// load the calendar events

	//params
	queryParamsItems := map[string][]string{}
	queryParamsItems["as_user_id"] = []string{fmt.Sprintf("sis_user_id:%s", netID)}
	queryParamsItems["per_page"] = []string{"50"}
	queryParamsItems["start_date"] = []string{startAt.Format(time.RFC3339)}
	queryParamsItems["end_date"] = []string{endAt.Format(time.RFC3339)}

	contextCodes := []string{}
	contextCodes = append(contextCodes, fmt.Sprintf("user_%d", providerUserID))
	contextCodes = append(contextCodes, fmt.Sprintf("course_%d", courseID))
	queryParamsItems["context_codes[]"] = contextCodes

	queryParams := a.constructQueryParams(queryParamsItems)

	//path + params
	pathAndParams := fmt.Sprintf("/api/v1/calendar_events%s", queryParams)

	//execute query
	data, err := a.executeQuery(http.NoBody, pathAndParams, "GET")
	if err != nil {
		log.Printf("error getting calendar events - %s", err)
		return nil, err
	}

	//prepare the response and return it
	var calendarEvents []model.CalendarEvent
	err = json.Unmarshal(data, &calendarEvents)
	if err != nil {
		log.Print("error converting missing calendar events")
		return nil, err
	}

	return calendarEvents, nil
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
	requestBody, err := io.ReadAll(body)
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
	data, err := io.ReadAll(resp.Body)
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

// NewProviderAdapter creates a new provider adapter
func NewProviderAdapter(host string, token string, tokenType string, storage interfaces.Storage, logger *logs.Logger) *Adapter {
	return &Adapter{host: host, token: token, tokenType: tokenType, storage: storage, logger: logger}
}
