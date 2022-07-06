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

package storage

import (
	"lms/core/model"
	"log"
	"strconv"
	"time"

	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"
	"go.mongodb.org/mongo-driver/bson"
)

// Adapter implements the Storage interface
type Adapter struct {
	db *database
}

// Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
	return err
}

// NewStorageAdapter creates a new storage adapter instance
func NewStorageAdapter(mongoDBAuth string, mongoDBName string, mongoTimeout string, logger *logs.Logger) *Adapter {
	timeout, err := strconv.Atoi(mongoTimeout)
	if err != nil {
		log.Println("Set default timeout - 500")
		timeout = 500
	}
	timeoutMS := time.Millisecond * time.Duration(timeout)

	db := &database{mongoDBAuth: mongoDBAuth, mongoDBName: mongoDBName, mongoTimeout: timeoutMS, logger: logger}
	return &Adapter{db: db}
}

// SetListener sets the upper layer listener for sending collection changed callbacks
func (sa *Adapter) SetListener(listener CollectionListener) {
	sa.db.listener = listener
}

// LoadAllNudges loads all nudges
func (sa *Adapter) LoadAllNudges() ([]model.Nudge, error) {
	filter := bson.D{}
	var result []model.Nudge
	err := sa.db.nudges.Find(filter, &result, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "nudge", nil, err)
	}
	return result, nil
}

// Event

func (m *database) onDataChanged(changeDoc map[string]interface{}) {
	if changeDoc == nil {
		return
	}
	log.Printf("onDataChanged: %+v\n", changeDoc)
	ns := changeDoc["ns"]
	if ns == nil {
		return
	}
	nsMap := ns.(map[string]interface{})
	coll := nsMap["coll"]

	if m.listener != nil {
		m.listener.OnCollectionUpdated(coll.(string))
	}
}
