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
	queryParamsItems := map[string]string{}
	queryParamsItems["as_user_id"] = fmt.Sprintf("sis_user_id:%s", userID)
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

func (a *Adapter) constructQueryParams(items map[string]string) string {
	if len(items) == 0 {
		return ""
	}

	values := url.Values{}

	for k, v := range items {
		values.Add(k, v)
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

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		//we have an error
		errorMessage := fmt.Sprintf("error with response code %d", resp.StatusCode)
		log.Print(errorMessage)
		return nil, errors.New(errorMessage)
	}

	//return the response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error converting response body - %s", pathAndParams)
		return nil, err
	}
	return data, nil
}

//NewProviderAdapter creates a new provider adapter
func NewProviderAdapter(host string, token string, tokenType string) *Adapter {
	return &Adapter{host: host, token: token, tokenType: tokenType}
}
