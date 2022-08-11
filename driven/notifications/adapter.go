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

package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lms/core"
	"log"
	"net/http"
)

//Adapter implements the notifications BB interface
type Adapter struct {
	host           string
	internalAPIKey string
}

//SendNotifications sends notifications via the Notifications BB
func (a *Adapter) SendNotifications(recipients []core.Recipient, text string, body string, data map[string]string) error {
	if len(recipients) > 0 {
		url := fmt.Sprintf("%s/api/int/message", a.host)

		bodyData := map[string]interface{}{
			"priority":   10,
			"recipients": recipients,
			"topic":      nil,
			"subject":    text,
			"body":       body,
			"data":       data,
		}
		bodyBytes, err := json.Marshal(bodyData)
		if err != nil {
			log.Printf("error creating notification request - %s", err)
			return err
		}

		client := &http.Client{}
		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("error creating load user data request - %s", err)
			return err
		}
		req.Header.Set("INTERNAL-API-KEY", a.internalAPIKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("error loading user data - %s", err)
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			log.Printf("notifications error with response code - %d", resp.StatusCode)
			return fmt.Errorf("error with response code != 200")
		}
	}
	return nil
}

//NewNotificationsAdapter creates a new notifications BB adapter
func NewNotificationsAdapter(notificationHost string, internalAPIKey string) *Adapter {
	return &Adapter{host: notificationHost, internalAPIKey: internalAPIKey}
}
