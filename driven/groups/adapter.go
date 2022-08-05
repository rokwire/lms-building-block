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

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"lms/core"
	"log"
	"net/http"
)

//Adapter implements the groups BB interface
type Adapter struct {
	host   string
	apiKey string
}

//GetUsers get user from the groups BB
func (a *Adapter) GetUsers(groupName string, offset int, limit int) ([]core.GroupsBBUser, error) {

	url := fmt.Sprintf("%s/api/int/group/title/%s/members", a.host, groupName)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("error creating load groups members data request - %s", err)
		return nil, err
	}
	req.Header.Set("INTERNAL-API-KEY", a.apiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error loading groups members data - %s", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("error with response code - %d", resp.StatusCode)
		return nil, errors.New("error with response code != 200")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error reading the body data for the loading groups members data request - %s", err)
		return nil, err
	}

	var result []core.GroupsBBUser
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("error converting data for the loading groups members data request - %s", err)
		return nil, err
	}

	return result, nil
}

//NewGroupsAdapter creates a new groups BB adapter
func NewGroupsAdapter(host string, apiKey string) *Adapter {
	return &Adapter{host: host, apiKey: apiKey}
}
