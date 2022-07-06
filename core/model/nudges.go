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

//Nudge entity
type Nudge struct {
	ID     int                    `json:"id" bson:"_id"`        //last_login
	Name   string                 `json:"name" bson:"name"`     //"Last Canvas use was over 2 weeks"
	Body   string                 `json:"body" bson:"body"`     //"You have not used the Canvas Application in over 2 weeks."
	Params map[string]interface{} `json:"params" bson:"params"` //Nudge specific settings
}
