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
	"lms/core/model"
)

//Adapter implements the Provider interface
type Adapter struct {
	host  string
	token string
}

//GetCourses gets the user courses
func (a *Adapter) GetCourses(userID string) ([]model.Course, error) {
	//TODO
	return nil, nil
}

//NewProviderAdapter creates a new provider adapter
func NewProviderAdapter(host string, token string) *Adapter {
	return &Adapter{host: host, token: token}
}
