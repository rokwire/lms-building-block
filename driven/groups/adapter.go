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

package groups

import "lms/core"

//Adapter implements the groups BB interface
type Adapter struct {
	testUserID string
	testNetID  string

	testUserID2 string
	testNetID2  string
}

//GetUsers get user from the groups BB
func (a *Adapter) GetUsers() ([]core.GroupsBBUser, error) {
	//TODO
	users := []core.GroupsBBUser{{UserID: a.testUserID, NetID: a.testNetID},
		{UserID: a.testUserID2, NetID: a.testNetID2}}
	return users, nil
}

//NewGroupsAdapter creates a new groups BB adapter
func NewGroupsAdapter(testUserID string, testNetID string, testUserID2 string, testNetID2 string) *Adapter {
	return &Adapter{testUserID: testUserID, testNetID: testNetID,
		testUserID2: testUserID2, testNetID2: testNetID2}
}