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
	"lms/core/model"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type configEntity struct {
	Name   string      `bson:"_id"`
	Config interface{} `bson:"config"`
}

// Adapter implements the Storage interface
type Adapter struct {
	db *database

	context mongo.SessionContext
}

// Start starts the storage
func (sa *Adapter) Start() error {
	err := sa.db.start()
	return err
}

// SetListener sets the upper layer listener for sending collection changed callbacks
func (sa *Adapter) SetListener(listener interfaces.CollectionListener) {
	sa.db.listener = listener
}

// PerformTransaction performs a transaction
func (sa *Adapter) PerformTransaction(transaction func(storage interfaces.Storage) error) error {
	// transaction
	callback := func(sessionContext mongo.SessionContext) (interface{}, error) {
		adapter := sa.withContext(sessionContext)

		err := transaction(adapter)
		if err != nil {
			if wrappedErr, ok := err.(interface {
				Internal() error
			}); ok && wrappedErr.Internal() != nil {
				return nil, wrappedErr.Internal()
			}
			return nil, err
		}

		return nil, nil
	}

	session, err := sa.db.dbClient.StartSession()
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionStart, "mongo session", nil, err)
	}
	context := context.Background()
	defer session.EndSession(context)

	_, err = session.WithTransaction(context, callback)
	if err != nil {
		return errors.WrapErrorAction("performing", logutils.TypeTransaction, nil, err)
	}
	return nil
}

// UserExist returns whether a user exists with the given netID
func (sa *Adapter) UserExist(netID string) (*bool, error) {
	filter := bson.D{primitive.E{Key: "net_id", Value: netID}}

	count, err := sa.db.users.CountDocuments(sa.context, filter)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCount, model.TypeUser, &logutils.FieldArgs{"net_id": netID}, err)
	}

	var result bool
	if count == 1 {
		result = true
	} else {
		result = false
	}
	return &result, nil
}

// InsertUser inserts a provider user
func (sa *Adapter) InsertUser(user model.ProviderUser) error {
	_, err := sa.db.users.InsertOne(sa.context, user)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "provider user", &logutils.FieldArgs{"net_id": user.NetID}, err)
	}
	return nil
}

// FindUser finds a provider user by netID
func (sa *Adapter) FindUser(netID string) (*model.ProviderUser, error) {
	filter := bson.D{primitive.E{Key: "net_id", Value: netID}}
	var result []model.ProviderUser
	err := sa.db.users.Find(sa.context, filter, &result, nil)
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

// FindUsers finds provider users by netID
func (sa *Adapter) FindUsers(netIDs []string) ([]model.ProviderUser, error) {
	filter := bson.D{primitive.E{Key: "net_id", Value: bson.M{"$in": netIDs}}}
	var result []model.ProviderUser
	err := sa.db.users.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}
	return result, nil
}

// FindUsersByCanvasUserID finds provider users by canvas user ID
func (sa *Adapter) FindUsersByCanvasUserID(canvasUserIds []int) ([]model.ProviderUser, error) {
	filter := bson.D{primitive.E{Key: "user.id", Value: bson.M{"$in": canvasUserIds}}}
	var result []model.ProviderUser
	err := sa.db.users.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}
	return result, nil
}

// SaveUser saves a provider user
func (sa *Adapter) SaveUser(providerUser model.ProviderUser) error {
	filter := bson.M{"_id": providerUser.ID}
	err := sa.db.users.ReplaceOne(sa.context, filter, providerUser, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionSave, "provider user", &logutils.FieldArgs{"_id": providerUser.ID}, err)
	}

	return nil
}

// DeleteUsersByNetIDs deletes users by netIDs
func (sa *Adapter) DeleteUsersByNetIDs(log *logs.Log, netIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "net_id", Value: primitive.M{"$in": netIDs}},
	}
	_, err := sa.db.users.DeleteMany(nil, filter, nil)
	return err
}

// CreateNudgesConfig creates nudges config
func (sa *Adapter) CreateNudgesConfig(nudgesConfig model.NudgesConfig) error {
	storageConfig := configEntity{Name: "nudges", Config: nudgesConfig}
	_, err := sa.db.configs.InsertOne(sa.context, storageConfig)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "config", &logutils.FieldArgs{"name": "nudges"}, err)
	}
	return nil
}

// FindNudgesConfig finds the nudges config
func (sa *Adapter) FindNudgesConfig() (*model.NudgesConfig, error) {
	filter := bson.D{primitive.E{Key: "_id", Value: "nudges"}}
	var result []configEntity
	err := sa.db.configs.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "configs", &logutils.FieldArgs{"name": "nudges"}, err)
	}
	if len(result) == 0 {
		return nil, nil
	}
	data := result[0].Config

	var nudgesConfig model.NudgesConfig
	bsonBytes, err := bson.Marshal(data)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionUnmarshal, "configs", &logutils.FieldArgs{"name": "nudges"}, err)
	}

	err = bson.Unmarshal(bsonBytes, &nudgesConfig)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionUnmarshal, "configs", &logutils.FieldArgs{"name": "nudges"}, err)
	}
	return &nudgesConfig, nil
}

// SaveNudgesConfig updates the nudges config
func (sa *Adapter) SaveNudgesConfig(nudgesConfig model.NudgesConfig) error {
	filter := bson.D{primitive.E{Key: "_id", Value: "nudges"}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "config", Value: nudgesConfig},
		}},
	}

	upsert := true
	opts := options.UpdateOptions{Upsert: &upsert}
	_, err := sa.db.configs.UpdateOne(sa.context, filter, update, &opts)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeNudgesConfig, &logutils.FieldArgs{"id": "nudges"}, err)
	}

	return nil
}

// LoadAllNudges loads all nudges
func (sa *Adapter) LoadAllNudges() ([]model.Nudge, error) {
	filter := bson.D{}
	var result []model.Nudge
	err := sa.db.nudges.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "nudge", nil, err)
	}
	if len(result) == 0 {
		return make([]model.Nudge, 0), nil
	}
	return result, nil
}

// LoadActiveNudges loads all active nudges
func (sa *Adapter) LoadActiveNudges() ([]model.Nudge, error) {
	filter := bson.D{primitive.E{Key: "active", Value: true}}
	var result []model.Nudge
	err := sa.db.nudges.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "nudge", nil, err)
	}
	if len(result) == 0 {
		return make([]model.Nudge, 0), nil
	}
	return result, nil
}

// InsertNudge inserts a new Nudge
func (sa *Adapter) InsertNudge(item model.Nudge) error {
	_, err := sa.db.nudges.InsertOne(sa.context, item)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeNudge, nil, err)
	}
	return nil
}

// UpdateNudge updates nudge
func (sa *Adapter) UpdateNudge(item model.Nudge) error {

	nudgeFilter := bson.D{primitive.E{Key: "_id", Value: item.ID}}
	updateNudge := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "name", Value: item.Name},
			primitive.E{Key: "body", Value: item.Body},
			primitive.E{Key: "deep_link", Value: item.DeepLink},
			primitive.E{Key: "params", Value: item.Params},
			primitive.E{Key: "active", Value: item.Active},
			primitive.E{Key: "users_sources", Value: item.UsersSources},
		}},
	}

	result, err := sa.db.nudges.UpdateOne(sa.context, nudgeFilter, updateNudge, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeNudge, &logutils.FieldArgs{"id": item.ID}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeNudge, &logutils.FieldArgs{"id": item.ID}, err)
	}

	return nil
}

// DeleteNudge deletes nudge
func (sa *Adapter) DeleteNudge(ID string) error {
	filter := bson.M{"_id": ID}
	result, err := sa.db.nudges.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeNudge, &logutils.FieldArgs{"_id": ID}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "delete nudge result", &logutils.FieldArgs{"_id": ID}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeNudge, &logutils.FieldArgs{"_id": ID}, err)
	}
	return nil
}

// InsertSentNudge inserts sent nudge entity
func (sa *Adapter) InsertSentNudge(sentNudge model.SentNudge) error {
	_, err := sa.db.sentNudges.InsertOne(sa.context, sentNudge)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "sent nudge", nil, err)
	}

	return nil
}

// InsertSentNudges inserts sent nudges entities
func (sa *Adapter) InsertSentNudges(sentNudges []model.SentNudge) error {
	data := make([]interface{}, len(sentNudges))
	for i, sn := range sentNudges {
		data[i] = sn
	}

	_, err := sa.db.sentNudges.InsertMany(sa.context, data, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "sent nudge", nil, err)
	}

	return nil
}

// FindSentNudge finds sent nudge entity
func (sa *Adapter) FindSentNudge(nudgeID string, userID string, netID string, criteriaHash uint32, mode string) (*model.SentNudge, error) {
	filter := bson.D{
		primitive.E{Key: "nudge_id", Value: nudgeID},
		primitive.E{Key: "user_id", Value: userID},
		primitive.E{Key: "net_id", Value: netID},
		primitive.E{Key: "criteria_hash", Value: criteriaHash},
		primitive.E{Key: "mode", Value: mode}}

	var result []model.SentNudge
	err := sa.db.sentNudges.Find(sa.context, filter, &result, nil)
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

// FindSentNudges finds sent nudges entities
func (sa *Adapter) FindSentNudges(nudgeID *string, userID *string, netID *string, criteriaHashes *[]uint32, mode *string) ([]model.SentNudge, error) {

	filter := bson.D{}

	if nudgeID != nil {
		filter = append(filter, primitive.E{Key: "nudge_id", Value: *nudgeID})
	}

	if userID != nil {
		filter = append(filter, primitive.E{Key: "user_id", Value: *userID})
	}

	if netID != nil {
		filter = append(filter, primitive.E{Key: "net_id", Value: *netID})
	}

	if criteriaHashes != nil {
		filter = append(filter, primitive.E{Key: "criteria_hash", Value: bson.M{"$in": *criteriaHashes}})
	}

	if mode != nil {
		filter = append(filter, primitive.E{Key: "mode", Value: *mode})
	}

	var result []model.SentNudge
	err := sa.db.sentNudges.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "sent nudge", nil, err)
	}
	return result, nil
}

// DeleteSentNudges deletes sent nudge
func (sa *Adapter) DeleteSentNudges(ids []string, mode string) error {
	filter := bson.M{}
	if ids != nil {
		filter["_id"] = bson.M{"$in": ids}
	}
	if mode != "" {
		filter["mode"] = mode
	}

	result, err := sa.db.sentNudges.DeleteMany(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeSentNudge, &logutils.FieldArgs{"_id": ids}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "delete sent nudges result", &logutils.FieldArgs{"_id": ids}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeSentNudge, &logutils.FieldArgs{"_id": ids}, err)
	}
	return nil
}

// FindNudgesProcesses finds all nudges-process
func (sa *Adapter) FindNudgesProcesses(limit int, offset int) ([]model.NudgesProcess, error) {
	filter := bson.D{}
	var result []model.NudgesProcess
	options := options.Find()
	options.SetLimit(int64(limit))
	options.SetSkip(int64(offset))
	err := sa.db.nudgesProcesses.Find(sa.context, filter, &result, options)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "nudges_process", nil, err)
	}
	if len(result) == 0 {
		return make([]model.NudgesProcess, 0), nil
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

// InsertNudgesProcess inserts nudges process
func (sa *Adapter) InsertNudgesProcess(nudgesProcess model.NudgesProcess) error {
	_, err := sa.db.nudgesProcesses.InsertOne(sa.context, nudgesProcess)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "nudges process", nil, err)
	}
	return nil
}

// UpdateNudgesProcess updates a nudges process
func (sa *Adapter) UpdateNudgesProcess(ID string, completedAt time.Time, status string, errStr *string) error {
	filter := bson.D{primitive.E{Key: "_id", Value: ID}}
	update := bson.D{
		primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "completed_at", Value: completedAt},
			primitive.E{Key: "status", Value: status},
			primitive.E{Key: "error", Value: errStr},
		}},
	}

	result, err := sa.db.nudgesProcesses.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "nudges process", &logutils.FieldArgs{"id": ID}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "nudges process", &logutils.FieldArgs{"id": ID}, err)
	}

	return nil
}

// CountNudgesProcesses counts the nudges process by status
func (sa *Adapter) CountNudgesProcesses(status string) (*int64, error) {
	filter := bson.D{primitive.E{Key: "status", Value: status}}

	count, err := sa.db.nudgesProcesses.CountDocuments(sa.context, filter)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCount, model.TypeNudgesProcess, &logutils.FieldArgs{"status": status}, err)
	}
	return &count, nil
}

// InsertBlock adds a block to a nudges process
func (sa *Adapter) InsertBlock(block model.Block) error {
	_, err := sa.db.nudgesBlocks.InsertOne(sa.context, block)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "nudge block", nil, err)
	}
	return nil
}

// InsertBlocks inserts blocks
func (sa *Adapter) InsertBlocks(blocks []model.Block) error {
	data := make([]interface{}, len(blocks))
	for i, sn := range blocks {
		data[i] = sn
	}

	_, err := sa.db.nudgesBlocks.InsertMany(sa.context, data, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "blocks", nil, err)
	}

	return nil
}

// FindBlock finds for a nudges process
func (sa *Adapter) FindBlock(processID string, blockNumber int) (*model.Block, error) {
	filter := bson.D{primitive.E{Key: "process_id", Value: processID},
		primitive.E{Key: "number", Value: blockNumber}}
	var result []model.Block
	err := sa.db.nudgesBlocks.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, "nudge block", nil, err)
	}
	if len(result) == 0 {
		return nil, nil
	}
	block := result[0]
	return &block, nil
}

// Creates a new Adapter with provided context
func (sa *Adapter) withContext(context mongo.SessionContext) *Adapter {
	return &Adapter{db: sa.db, context: context}
}

// DeleteNudgesBlocksByAccountsIDs deletes specific items from nudges blocks based on accountsIDs
func (sa *Adapter) DeleteNudgesBlocksByAccountsIDs(log *logs.Logger, accountsIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "items.user_id", Value: primitive.M{"$in": accountsIDs}},
	}
	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{
				"user_id": bson.M{"$in": accountsIDs},
			},
		},
	}

	_, err := sa.db.nudgesBlocks.UpdateMany(nil, filter, update, nil)
	return err
}

// DeleteSentNudgesByAccountsIDs deletes sent nudges by accountsIDs
func (sa *Adapter) DeleteSentNudgesByAccountsIDs(log *logs.Logger, accountsIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "user_id", Value: primitive.M{"$in": accountsIDs}},
	}
	_, err := sa.db.sentNudges.DeleteMany(nil, filter, nil)
	return err
}

// DeleteUserContentsByAccountsIDs deletes an user contents by accountsIDs
func (sa *Adapter) DeleteUserContentsByAccountsIDs(log *logs.Log, appID string, orgID string, accountsIDs []string) error {
	filter := bson.D{
		{Key: "app_id", Value: appID},
		{Key: "org_id", Value: orgID},
		primitive.E{Key: "user_id", Value: primitive.M{"$in": accountsIDs}},
	}
	_, err := sa.db.userContents.DeleteMany(nil, filter, nil)
	return err
}

// DeleteUserCoursesByAccountsIDs deletes an user courses by accountsIDs
func (sa *Adapter) DeleteUserCoursesByAccountsIDs(log *logs.Log, appID string, orgID string, accountsIDs []string) error {
	filter := bson.D{
		{Key: "app_id", Value: appID},
		{Key: "org_id", Value: orgID},
		primitive.E{Key: "user_id", Value: primitive.M{"$in": accountsIDs}},
	}
	_, err := sa.db.userCourses.DeleteMany(nil, filter, nil)
	return err
}

// DeleteUserUnitsByAccountsIDs deletes an user units by accountsIDs
func (sa *Adapter) DeleteUserUnitsByAccountsIDs(log *logs.Log, appID string, orgID string, accountsIDs []string) error {
	filter := bson.D{
		{Key: "app_id", Value: appID},
		{Key: "org_id", Value: orgID},
		primitive.E{Key: "user_id", Value: primitive.M{"$in": accountsIDs}},
	}
	_, err := sa.db.userUnits.DeleteMany(nil, filter, nil)
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
