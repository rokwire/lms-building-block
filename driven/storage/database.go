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
	"context"
	"log"
	"time"

	"github.com/rokwire/logging-library-go/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionListener listens for collection updates
type CollectionListener interface {
	OnCollectionUpdated(name string)
}

type database struct {
	listener CollectionListener

	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	logger *logs.Logger

	db       *mongo.Database
	dbClient *mongo.Client

	nudges     *collectionWrapper
	sentNudges *collectionWrapper
}

func (m *database) start() error {

	log.Println("database -> start")

	//connect to the database
	clientOptions := options.Client().ApplyURI(m.mongoDBAuth)
	connectContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	client, err := mongo.Connect(connectContext, clientOptions)
	cancel()
	if err != nil {
		return err
	}

	//ping the database
	pingContext, cancel := context.WithTimeout(context.Background(), m.mongoTimeout)
	err = client.Ping(pingContext, nil)
	cancel()
	if err != nil {
		return err
	}

	//apply checks
	db := client.Database(m.mongoDBName)

	nudges := &collectionWrapper{database: m, coll: db.Collection("nudges")}
	err = m.applyNudgesChecks(nudges)
	if err != nil {
		return err
	}

	sentNudges := &collectionWrapper{database: m, coll: db.Collection("sent_nudges")}
	err = m.applySentNudgesChecks(sentNudges)
	if err != nil {
		return err
	}

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.nudges = nudges
	m.sentNudges = sentNudges

	return nil
}

func (m *database) applyNudgesChecks(nudges *collectionWrapper) error {
	m.logger.Info("apply nudges checks.....")

	m.logger.Info("nudges check passed")
	return nil
}

func (m *database) applySentNudgesChecks(sentNudges *collectionWrapper) error {
	m.logger.Info("apply sent nudges checks.....")

	//add nudge_id index
	err := sentNudges.AddIndex(bson.D{primitive.E{Key: "nudge_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	//add user_id index
	err = sentNudges.AddIndex(bson.D{primitive.E{Key: "user_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	//add net_id index
	err = sentNudges.AddIndex(bson.D{primitive.E{Key: "net_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	//add criteria_hash index
	err = sentNudges.AddIndex(bson.D{primitive.E{Key: "criteria_hash", Value: 1}}, false)
	if err != nil {
		return err
	}

	m.logger.Info("sent nudges check passed")
	return nil
}
