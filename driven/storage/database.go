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

package storage

import (
	"context"
	"lms/core/interfaces"
	"log"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	listener interfaces.CollectionListener

	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	logger *logs.Logger

	db       *mongo.Database
	dbClient *mongo.Client

	configs         *collectionWrapper
	users           *collectionWrapper
	nudges          *collectionWrapper
	sentNudges      *collectionWrapper
	nudgesProcesses *collectionWrapper
	nudgesBlocks    *collectionWrapper
	customCourses   *collectionWrapper
	customModules   *collectionWrapper
	customUnits     *collectionWrapper
	customContents  *collectionWrapper
	userCourses     *collectionWrapper
	userUnits       *collectionWrapper
}

func (m *database) start() error {

	m.logger.Info("database -> start")

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

	configs := &collectionWrapper{database: m, coll: db.Collection("configs")}
	err = m.applyConfigsChecks(configs)
	if err != nil {
		return err
	}

	users := &collectionWrapper{database: m, coll: db.Collection("adapter_pr_users")}
	err = m.applyUsersChecks(users)
	if err != nil {
		return err
	}

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

	nudgesProcesses := &collectionWrapper{database: m, coll: db.Collection("nudges_processes")}
	err = m.applyNudgesProcessesChecks(nudgesProcesses)
	if err != nil {
		return err
	}

	nudgesBlocks := &collectionWrapper{database: m, coll: db.Collection("nudges_blocks")}
	err = m.applyNudgesBlocksChecks(nudgesBlocks)
	if err != nil {
		return err
	}

	customCourses := &collectionWrapper{database: m, coll: db.Collection("custom_courses")}
	err = m.applyCustomCoursesChecks(customCourses)
	if err != nil {
		return err
	}

	customModules := &collectionWrapper{database: m, coll: db.Collection("custom_modules")}
	err = m.applyCustomModulesChecks(customModules)
	if err != nil {
		return err
	}

	customUnits := &collectionWrapper{database: m, coll: db.Collection("custom_units")}
	err = m.applyCustomUnitsChecks(customUnits)
	if err != nil {
		return err
	}

	customContents := &collectionWrapper{database: m, coll: db.Collection("custom_contents")}
	err = m.applyCustomContentChecks(customContents)
	if err != nil {
		return err
	}

	userCourses := &collectionWrapper{database: m, coll: db.Collection("user_courses")}
	err = m.applyUserCoursesChecks(userCourses)
	if err != nil {
		return err
	}

	userUnits := &collectionWrapper{database: m, coll: db.Collection("user_units")}
	err = m.applyUserUnitsChecks(userUnits)
	if err != nil {
		return err
	}

	//asign the db, db client and the collections
	m.db = db
	m.dbClient = client

	m.configs = configs
	m.users = users
	m.nudges = nudges
	m.sentNudges = sentNudges
	m.nudgesProcesses = nudgesProcesses
	m.nudgesBlocks = nudgesBlocks
	m.customCourses = customCourses
	m.customModules = customModules
	m.customUnits = customUnits
	m.customContents = customContents
	m.userCourses = userCourses
	m.userUnits = userUnits

	go m.configs.Watch(nil, m.logger)

	return nil
}

func (m *database) applyConfigsChecks(configs *collectionWrapper) error {
	m.logger.Info("apply configs checks.....")

	m.logger.Info("configs check passed")
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

func (m *database) applyNudgesProcessesChecks(nudgesProcesses *collectionWrapper) error {
	m.logger.Info("apply nudges processes checks.....")

	//add blocks number index
	err := nudgesProcesses.AddIndex(bson.D{primitive.E{Key: "blocks.number", Value: 1}}, false)
	if err != nil {
		return err
	}

	m.logger.Info("nudges processes check passed")
	return nil
}

func (m *database) applyNudgesBlocksChecks(nudgesProcesses *collectionWrapper) error {
	m.logger.Info("apply nudges blocks checks.....")

	//add process id index
	err := nudgesProcesses.AddIndex(bson.D{primitive.E{Key: "process_id", Value: 1}}, false)
	if err != nil {
		return err
	}

	//add blocks number index
	err = nudgesProcesses.AddIndex(bson.D{primitive.E{Key: "number", Value: 1}}, false)
	if err != nil {
		return err
	}

	m.logger.Info("nudges blocks check passed")
	return nil
}

// Custom Course
func (m *database) applyCustomCoursesChecks(customCourses *collectionWrapper) error {
	m.logger.Info("apply custom course check.....")
	err := customCourses.AddIndex(
		bson.D{
			primitive.E{Key: "app_id", Value: 1},
			primitive.E{Key: "org_id", Value: 1},
			primitive.E{Key: "key", Value: 1},
		}, true)
	if err != nil {
		return err
	}
	m.logger.Info("custom course check passed")
	return nil
}

// Custom Module
func (m *database) applyCustomModulesChecks(customModules *collectionWrapper) error {
	m.logger.Info("apply custom module check.....")
	err := customModules.AddIndex(
		bson.D{
			primitive.E{Key: "app_id", Value: 1},
			primitive.E{Key: "org_id", Value: 1},
			primitive.E{Key: "key", Value: 1},
		}, true)
	if err != nil {
		return err
	}
	m.logger.Info("custom module check passed")
	return nil
}

// Custom Unit
func (m *database) applyCustomUnitsChecks(customUnits *collectionWrapper) error {
	m.logger.Info("apply custom unit check.....")
	err := customUnits.AddIndex(
		bson.D{
			primitive.E{Key: "app_id", Value: 1},
			primitive.E{Key: "org_id", Value: 1},
			primitive.E{Key: "key", Value: 1},
		}, true)
	if err != nil {
		return err
	}
	m.logger.Info("custom unit check passed")
	return nil
}

// Custom Content
func (m *database) applyCustomContentChecks(customContents *collectionWrapper) error {
	m.logger.Info("apply custom content check.....")
	err := customContents.AddIndex(
		bson.D{
			primitive.E{Key: "app_id", Value: 1},
			primitive.E{Key: "org_id", Value: 1},
			primitive.E{Key: "key", Value: 1},
		}, true)
	if err != nil {
		return err
	}
	m.logger.Info("custom content check passed")
	return nil
}

// User Course
func (m *database) applyUserCoursesChecks(userCourses *collectionWrapper) error {
	m.logger.Info("apply user course check.....")
	err := userCourses.AddIndex(
		bson.D{
			primitive.E{Key: "app_id", Value: 1},
			primitive.E{Key: "org_id", Value: 1},
			primitive.E{Key: "user_id", Value: 1},
			primitive.E{Key: "course.key", Value: 1},
		}, true)
	if err != nil {
		return err
	}
	m.logger.Info("user course check passed")
	return nil
}

// User Unit
func (m *database) applyUserUnitsChecks(userUnits *collectionWrapper) error {
	m.logger.Info("apply user unit check.....")
	err := userUnits.AddIndex(
		bson.D{
			primitive.E{Key: "app_id", Value: 1},
			primitive.E{Key: "org_id", Value: 1},
			primitive.E{Key: "user_id", Value: 1},
		}, false)
	if err != nil {
		return err
	}
	m.logger.Info("user unit check passed")
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

	switch coll {
	case "configs":
		log.Println("configs collection changed")

		if m.listener != nil {
			m.listener.OnConfigsUpdated()
		}
	}
}
