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
	"go.mongodb.org/mongo-driver/bson/primitive"
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

//InsertNudge inserts a new Nudge
func (sa *Adapter) InsertNudge(item model.Nudge) error {
	nudge := model.Nudge{ID: item.ID, Name: item.Name, Body: item.Body, Params: item.Params}
	_, err := sa.db.nudges.InsertOne(nudge)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "", nil, err)
	}
	return nil
}

//UpdateNudge updates nudge
func (sa *Adapter) UpdateNudge(ID string, name string, body string, params *map[string]interface{}) error {

	nudgeFilter := bson.D{primitive.E{Key: "_id", Value: ID}}
	updateNudge := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "name", Value: name},
			primitive.E{Key: "body", Value: body},
			primitive.E{Key: "params", Value: params},
		}},
	}

	result, err := sa.db.nudges.UpdateOne(nudgeFilter, updateNudge, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"id": ID}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"id": ID}, err)
	}

	return nil
}

//DeleteNudge deletes nudge
func (sa *Adapter) DeleteNudge(ID string) error {
	filter := bson.M{"_id": ID}
	result, err := sa.db.nudges.DeleteOne(filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"_id": ID}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"_id": ID}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"_id": ID}, err)
	}
	return nil
}

//InsertSentNudge inserts sent nudge entity
func (sa *Adapter) InsertSentNudge(sentNudge model.SentNudge) error {
	_, err := sa.db.sentNudges.InsertOne(sentNudge)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "sent nudge", nil, err)
	}

	return nil
}

//FindSentNudge finds sent nudge entity
func (sa *Adapter) FindSentNudge(nudgeID string, userID string, netID string, criteriaHash uint32) (*model.SentNudge, error) {
	filter := bson.D{
		primitive.E{Key: "nudge_id", Value: nudgeID},
		primitive.E{Key: "user_id", Value: userID},
		primitive.E{Key: "net_id", Value: netID},
		primitive.E{Key: "criteria_hash", Value: criteriaHash}}

	var result []model.SentNudge
	err := sa.db.sentNudges.Find(filter, &result, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "sent nudge", nil, err)
	}
	if len(result) == 0 {
		//no record
		return nil, nil
	}
	sentNudge := result[0]
	return &sentNudge, nil
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
