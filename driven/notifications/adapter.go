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

//Adapter implements the notifications BB interface
type Adapter struct {
}

//SendNotifications sends notifications via the Notifications BB
func (a *Adapter) SendNotifications() error {
	//TODO
	return nil
}

//NewNotificationsAdapter creates a new notifications BB adapter
func NewNotificationsAdapter() *Adapter {
	return &Adapter{}
}
