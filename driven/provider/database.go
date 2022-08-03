// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"log"
	"time"

	"github.com/rokwire/logging-library-go/errors"
	"github.com/rokwire/logging-library-go/logs"
	"github.com/rokwire/logging-library-go/logutils"
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

	users *collectionWrapper
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

	users := &collectionWrapper{database: m, coll: db.Collection("adapter_pr_users")}
	err = m.applyUsersChecks(users)
	if err != nil {
		return err
	}

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.users = users

	return nil
}

func (m *database) applyUsersChecks(users *collectionWrapper) error {
	m.logger.Info("apply adapter users checks.....")

	//add net id index
	err := users.AddIndex(bson.D{primitive.E{Key: "net_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	//add user id index
	err = users.AddIndex(bson.D{primitive.E{Key: "user.id", Value: 1}}, false)
	if err != nil {
		return err
	}

	m.logger.Info("adapter users check passed")
	return nil
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

// Data managment

func (m *database) userExist(netID string) (*bool, error) {
	filter := bson.D{primitive.E{Key: "net_id", Value: netID}}

	count, err := m.users.CountDocuments(filter)
	if err != nil {
		return nil, errors.WrapErrorAction("error counting user for net id", "", &logutils.FieldArgs{"net_id": netID}, err)
	}

	var result bool
	if count == 1 {
		result = true
	} else {
		result = false
	}
	return &result, nil
}

func (m *database) insertUser(user providerUser) error {
	_, err := m.users.InsertOne(user)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "provider user", &logutils.FieldArgs{"net_id": user.NetID}, err)
	}
	return nil
}

func (m *database) findUser(netID string) (*providerUser, error) {
	filter := bson.D{primitive.E{Key: "net_id", Value: netID}}
	var result []providerUser
	err := m.users.Find(filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}

	user := result[0]
	return &user, nil
}
