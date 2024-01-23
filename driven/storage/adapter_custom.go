package storage

import (
	"lms/core/model"
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetCustomCourses finds courses by a set of parameters
func (sa *Adapter) GetCustomCourses(appID string, orgID string, id []string, name []string, key []string, moduleKeys []string) ([]model.Course, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID}
	if len(id) != 0 {
		filter["_id"] = bson.M{"$in": id}
	}

	if len(name) != 0 {
		filter["name"] = bson.M{"$in": name}
	}

	if len(key) != 0 {
		filter["key"] = bson.M{"$in": key}
	}

	if len(moduleKeys) > 0 {
		filter["module_keys"] = bson.M{"$in": moduleKeys}
	}

	var result []course

	err := sa.db.customCourses.Find(sa.context, filter, &result, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourse, &errArgs, err)
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}

	var convertedResult []model.Course
	for _, retrievedCourse := range result {
		singleConverted, err := sa.customCourseConversionStorageToAPI(retrievedCourse)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// GetCustomCourse finds a course by id
func (sa *Adapter) GetCustomCourse(appID string, orgID string, key string) (*model.Course, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	var result course
	err := sa.db.customCourses.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.customCourseConversionStorageToAPI(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertCustomCourse inserts a course
func (sa *Adapter) InsertCustomCourse(item model.Course) error {
	item.DateCreated = time.Now()
	item.DateUpdated = nil
	course := sa.customCourseConversionAPIToStorage(item)

	_, err := sa.db.customCourses.InsertOne(sa.context, course)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeCourse, nil, err)
	}
	return nil
}

// UpdateCustomCourse updates a course
func (sa *Adapter) UpdateCustomCourse(key string, item model.Course) error {
	//parse into the storage format and pass parameters
	var moduleKeys []string
	for _, val := range item.Modules {
		moduleKeys = append(moduleKeys, val.Key)
	}

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "key": key}
	update := bson.M{
		"$set": bson.M{
			"date_updated": time.Now(),
			"name":         item.Name,
			"module_keys":  moduleKeys,
		},
	}
	result, err := sa.db.customCourses.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeCourse, &logutils.FieldArgs{"org_id": item.OrgID, "app_id": item.AppID, "key": key}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeCourse, &logutils.FieldArgs{"org_id": item.OrgID, "app_id": item.AppID, "key": key}, err)
	}
	return nil
}

// UpdateUserCourses updates all user_course that matches given courseKey
func (sa *Adapter) UpdateUserCourses(key string, item model.Course) error {
	//parse into the storage format and pass parameters
	var moduleKeys []string
	for _, val := range item.Modules {
		moduleKeys = append(moduleKeys, val.Key)
	}

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "course.key": key}
	update := bson.M{
		"$set": bson.M{
			"course.date_updated": time.Now(),
			"course.name":         item.Name,
			"course.module_keys":  moduleKeys,
		},
	}
	_, err := sa.db.userCourses.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &errArgs, err)
	}
	return nil
}

// DeleteCustomCourse deletes a course
func (sa *Adapter) DeleteCustomCourse(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	errArgs := logutils.FieldArgs(filter)
	result, err := sa.db.customCourses.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeCourse, &errArgs, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "delete course result", &errArgs, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeCourse, &errArgs, err)
	}
	return nil
}

// GetCustomModules finds courses by a set of parameters
func (sa *Adapter) GetCustomModules(appID string, orgID string, id []string, name []string, key []string, unitKeys []string) ([]model.Module, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID}
	if len(id) != 0 {
		filter["_id"] = bson.M{"$in": id}
	}

	if len(name) != 0 {
		filter["name"] = bson.M{"$in": name}
	}

	if len(key) != 0 {
		filter["key"] = bson.M{"$in": key}
	}

	if len(unitKeys) > 0 {
		filter["unit_keys"] = bson.M{"$in": unitKeys}
	}

	var result []module

	err := sa.db.customModules.Find(sa.context, filter, &result, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeModule, &errArgs, err)
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}

	var convertedResult []model.Module
	for _, retrievedModule := range result {
		singleConverted, err := sa.customModuleConversionStorageToAPI(retrievedModule)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// GetCustomModule finds a module by id
func (sa *Adapter) GetCustomModule(appID string, orgID string, key string) (*model.Module, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	var result module
	err := sa.db.customModules.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.customModuleConversionStorageToAPI(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertCustomModule inserts a module
func (sa *Adapter) InsertCustomModule(item model.Module) error {
	item.DateCreated = time.Now()
	item.DateUpdated = nil
	module := sa.customModuleConversionAPIToStorage(item)
	_, err := sa.db.customModules.InsertOne(sa.context, module)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeModule, nil, err)
	}
	return nil
}

// InsertCustomModules insert an array of modules
func (sa *Adapter) InsertCustomModules(items []model.Module) error {
	storeItems := make([]interface{}, len(items))
	for i, item := range items {
		item.DateCreated = time.Now()
		item.DateUpdated = nil
		module := sa.customModuleConversionAPIToStorage(item)
		storeItems[i] = module
	}

	_, err := sa.db.customModules.InsertMany(sa.context, storeItems, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeModule, nil, err)
	}
	return nil
}

// UpdateCustomModule updates a module
func (sa *Adapter) UpdateCustomModule(key string, item model.Module) error {
	//parse into the storage format and pass parameters
	var unitKeys []string
	for _, val := range item.Units {
		unitKeys = append(unitKeys, val.Key)
	}

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "key": key}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$set": bson.M{
			"date_updated": time.Now(),
			"name":         item.Name,
			"unit_keys":    unitKeys,
		},
	}
	result, err := sa.db.customModules.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeModule, &errArgs, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeModule, &errArgs, err)
	}
	return nil
}

// DeleteCustomModule deletes a module
func (sa *Adapter) DeleteCustomModule(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	errArgs := logutils.FieldArgs(filter)
	result, err := sa.db.customModules.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeModule, &errArgs, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "delete module result", &errArgs, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeModule, &errArgs, err)
	}
	return nil
}

// GetCustomUnits finds units by a set of parameters
func (sa *Adapter) GetCustomUnits(appID string, orgID string, id []string, name []string, key []string, contentKeys []string) ([]model.Unit, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID}
	if len(id) != 0 {
		filter["_id"] = bson.M{"$in": id}
	}

	if len(name) != 0 {
		filter["name"] = bson.M{"$in": name}
	}

	if len(key) != 0 {
		filter["key"] = bson.M{"$in": key}
	}
	if len(contentKeys) != 0 {
		filter["content_keys"] = bson.M{"$in": contentKeys}
	}
	var result []unit
	err := sa.db.customUnits.Find(sa.context, filter, &result, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUnit, &errArgs, err)
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}

	var convertedResult []model.Unit
	for _, retrievedUnit := range result {
		singleConverted, err := sa.customUnitConversionStorageToAPI(retrievedUnit)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// GetCustomUnit finds a unit by id
func (sa *Adapter) GetCustomUnit(appID string, orgID string, key string) (*model.Unit, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	var result unit
	err := sa.db.customUnits.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.customUnitConversionStorageToAPI(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertCustomUnit inserts a unit
func (sa *Adapter) InsertCustomUnit(item model.Unit) error {
	item.DateCreated = time.Now()
	item.DateUpdated = nil
	unit := sa.customUnitConversionAPIToStorage(item)
	_, err := sa.db.customUnits.InsertOne(sa.context, unit)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUnit, nil, err)
	}
	return nil
}

// InsertCustomUnits insert an array of units
func (sa *Adapter) InsertCustomUnits(items []model.Unit) error {
	storeItems := make([]interface{}, len(items))
	for i, item := range items {
		item.DateCreated = time.Now()
		item.DateUpdated = nil
		unit := sa.customUnitConversionAPIToStorage(item)
		storeItems[i] = unit
	}

	_, err := sa.db.customUnits.InsertMany(sa.context, storeItems, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUnit, nil, err)
	}
	return nil
}

// UpdateCustomUnit updates a unit
func (sa *Adapter) UpdateCustomUnit(key string, item model.Unit) error {
	//parse into the storage format and pass parameters
	var extractedKey []string
	for _, val := range item.Contents {
		extractedKey = append(extractedKey, val.Key)
	}

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "key": key}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$set": bson.M{
			"name":         item.Name,
			"content_keys": extractedKey,
			"schedule":     item.Schedule,
			"date_updated": time.Now(),
		},
	}
	result, err := sa.db.customUnits.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUnit, &errArgs, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeUnit, &errArgs, err)
	}
	return nil
}

// UpdateUserUnits updates all user_unit that matches given key
func (sa *Adapter) UpdateUserUnits(key string, item model.Unit) error {
	//parse into the storage format and pass parameters
	var contentKeys []string
	for _, val := range item.Contents {
		contentKeys = append(contentKeys, val.Key)
	}

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "unit.key": key}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$set": bson.M{
			"unit.name":         item.Name,
			"unit.content_keys": contentKeys,
			"unit.schedule":     item.Schedule,
			"unit.date_updated": time.Now(),
		},
	}
	_, err := sa.db.userUnits.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUnit, &errArgs, err)
	}
	return nil
}

// DeleteUserUnit deletes all userUnit derieved from a custom unit
func (sa *Adapter) DeleteUserUnit(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "unit.key": key}
	result, err := sa.db.userUnits.DeleteMany(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeUnit, &logutils.FieldArgs{"key": key}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"key": key}, err)
	}
	// deletedCount := result.DeletedCount
	// if deletedCount == 0 {
	// 	return errors.WrapErrorData(logutils.StatusMissing, model.TypeUnit, &logutils.FieldArgs{"key": key}, err)
	// }
	return nil
}

// DeleteCustomUnit deletes a unit
func (sa *Adapter) DeleteCustomUnit(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	errArgs := logutils.FieldArgs(filter)
	result, err := sa.db.customUnits.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeUnit, &errArgs, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "delete unit result", &errArgs, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeUnit, &errArgs, err)
	}
	return nil
}

// GetCustomContents finds contents by a set of parameters
func (sa *Adapter) GetCustomContents(appID string, orgID string, id []string, name []string, key []string) ([]model.Content, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID}
	if len(id) != 0 {
		filter["_id"] = bson.M{"$in": id}
	}

	if len(name) != 0 {
		filter["name"] = bson.M{"$in": name}
	}

	if len(key) != 0 {
		filter["key"] = bson.M{"$in": key}
	}
	var result []content
	err := sa.db.customContents.Find(sa.context, filter, &result, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeContent, &errArgs, err)
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}

	var convertedResult []model.Content
	for _, retrievedContent := range result {
		singleConverted, err := sa.customContentConversionStorageToAPI(retrievedContent)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// GetCustomContent finds a content by id
func (sa *Adapter) GetCustomContent(appID string, orgID string, key string) (*model.Content, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	var result content
	err := sa.db.customContents.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.customContentConversionStorageToAPI(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertCustomContent inserts a content
func (sa *Adapter) InsertCustomContent(item model.Content) error {
	item.DateCreated = time.Now()
	item.DateUpdated = nil
	content := sa.customContentConversionAPIToStorage(item)
	_, err := sa.db.customContents.InsertOne(sa.context, content)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeContent, nil, err)
	}
	return nil
}

// InsertCustomContents insert an array of contents
func (sa *Adapter) InsertCustomContents(items []model.Content) error {
	storeItems := make([]interface{}, len(items))
	for i, item := range items {
		item.DateCreated = time.Now()
		item.DateUpdated = nil
		content := sa.customContentConversionAPIToStorage(item)
		storeItems[i] = content
	}

	_, err := sa.db.customContents.InsertMany(sa.context, storeItems, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeContent, nil, err)
	}
	return nil
}

// UpdateCustomContent updates a content
func (sa *Adapter) UpdateCustomContent(key string, item model.Content) error {
	//parse into the storage format and pass parameters
	// var extractedKey []string
	// for _, val := range item.LinkedContent {
	// 	extractedKey = append(extractedKey, val.Key)
	// }

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "key": key}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$set": bson.M{
			"type":           item.Type,
			"details":        item.Details,
			"name":           item.Name,
			"reference":      item.ContentReference,
			"linked_content": item.LinkedContent,
			"date_updated":   time.Now(),
		},
	}
	result, err := sa.db.customContents.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeContent, &errArgs, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeContent, &errArgs, err)
	}
	return nil
}

// DeleteCustomContent deletes a content
func (sa *Adapter) DeleteCustomContent(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	errArgs := logutils.FieldArgs(filter)
	result, err := sa.db.customContents.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeContent, &errArgs, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "delete content result", &errArgs, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeContent, &errArgs, err)
	}
	return nil
}

// DeleteContentKeyFromLinkedContents deletes a content key from linkedContent field within customContents collection
func (sa *Adapter) DeleteContentKeyFromLinkedContents(appID string, orgID string, key string) error {
	var keyArr []string
	keyArr = append(keyArr, key)
	filter := bson.M{"org_id": orgID, "app_id": appID}
	//delete key from linked_content
	update := bson.M{
		"$pull": bson.M{
			"linked_content": bson.M{
				"$in": keyArr,
			},
		},
		// "$pull": bson.M{
		// 	"linked_content": key,
		// },
	}

	_, err := sa.db.customContents.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeContent, &errArgs, err)
	}
	return nil
}

// DeleteContentKeyFromUnits deletes a content key from contentKey field within customUnit collection
func (sa *Adapter) DeleteContentKeyFromUnits(appID string, orgID string, key string) error {
	var keyArr []string
	keyArr = append(keyArr, key)
	filter := bson.M{"org_id": orgID, "app_id": appID}
	update := bson.M{
		"$pull": bson.M{
			"content_keys": bson.M{
				"$in": keyArr,
			},
		},
	}
	_, err := sa.db.customUnits.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUnit, &errArgs, err)
	}
	return err
}

// DeleteContentKeyFromUserUnits deletes given content key from all content.contentKey field within user_unit collection
func (sa *Adapter) DeleteContentKeyFromUserUnits(appID string, orgID string, key string) error {
	var keyArr []string
	keyArr = append(keyArr, key)
	filter := bson.M{"org_id": orgID, "app_id": appID}
	update := bson.M{
		"$pull": bson.M{
			"unit.content_keys": bson.M{
				"$in": keyArr,
			},
		},
	}

	_, err := sa.db.userUnits.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, &errArgs, err)
	}
	return nil
}

// DeleteUnitKeyFromModules deletes a unit key from unitKey field within customModule collection
func (sa *Adapter) DeleteUnitKeyFromModules(appID string, orgID string, key string) error {
	var keyArr []string
	keyArr = append(keyArr, key)
	filter := bson.M{"org_id": orgID, "app_id": appID}
	update := bson.M{
		"$pull": bson.M{
			"unit_keys": bson.M{
				"$in": keyArr,
			},
		},
	}

	_, err := sa.db.customModules.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeModule, &errArgs, err)
	}
	return nil
}

// DeleteModuleKeyFromCourses deletes a module key from moduleKey field within customCourse collection
func (sa *Adapter) DeleteModuleKeyFromCourses(appID string, orgID string, key string) error {
	var keyArr []string
	keyArr = append(keyArr, key)
	filter := bson.M{"org_id": orgID, "app_id": appID}
	update := bson.M{
		"$pull": bson.M{
			"module_keys": bson.M{
				"$in": keyArr,
			},
		},
	}

	_, err := sa.db.customCourses.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeCourse, &errArgs, err)
	}
	return nil
}

// DeleteModuleKeyFromUserCourses deletes given module key from all module.moduleKey fields within user_course collection
func (sa *Adapter) DeleteModuleKeyFromUserCourses(appID string, orgID string, key string) error {
	var keyArr []string
	keyArr = append(keyArr, key)
	filter := bson.M{"org_id": orgID, "app_id": appID}
	update := bson.M{
		"$pull": bson.M{
			"course.module_keys": bson.M{
				"$in": keyArr,
			},
		},
	}

	_, err := sa.db.userCourses.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &errArgs, err)
	}
	return nil
}

// moved from adapter_client.go

// MarkUserCourseAsDelete mark given course as deleted in user_course collection
func (sa *Adapter) MarkUserCourseAsDelete(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "course.key": key}
	update := bson.M{
		"$set": bson.M{
			"date_dropped": time.Now(),
		},
	}

	_, err := sa.db.userCourses.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &errArgs, err)
	}
	return nil
}

// DeleteCustomCourse deletes a course
func (sa *Adapter) DeleteUserCourse(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "course.key": key}
	result, err := sa.db.userCourses.DeleteMany(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeCourse, &logutils.FieldArgs{"key": key}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"key": key}, err)
	}
	// deletedCount := result.DeletedCount
	// if deletedCount == 0 {
	// 	return errors.WrapErrorData(logutils.StatusMissing, model.TypeCourse, &logutils.FieldArgs{"key": key}, err)
	// }
	return nil
}

// GetUserCourses finds user course by a set of parameters
func (sa *Adapter) GetUserCourses(id []string, name []string, key []string, userID string) ([]model.UserCourse, error) {
	filter := bson.M{}
	if len(id) != 0 {
		filter["_id"] = bson.M{"$in": id}
	}

	if len(name) != 0 {
		filter["course.name"] = bson.M{"$in": name}
	}

	if len(key) != 0 {
		filter["course.key"] = bson.M{"$in": key}
	}

	if len(userID) != 0 {
		filter["user_id"] = userID
	}

	var result []userCourse
	err := sa.db.userCourses.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}

	var convertedResult []model.UserCourse
	for _, retrievedResult := range result {
		singleConverted, err := sa.userCourseConversionStorageToAPI(retrievedResult)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// GetUserCourse finds a user course by id
func (sa *Adapter) GetUserCourse(appID string, orgID string, userID string, courseKey string) (*model.UserCourse, error) {
	filter := bson.M{"app_id": appID, "org_id": orgID, "user_id": userID, "course.key": courseKey}
	var result userCourse
	err := sa.db.userCourses.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.userCourseConversionStorageToAPI(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertUserCourse inserts a user course
func (sa *Adapter) InsertUserCourse(item model.UserCourse) error {
	var userCourse userCourse
	userCourse.ID = item.ID
	userCourse.AppID = item.AppID
	userCourse.OrgID = item.OrgID
	userCourse.UserID = item.UserID
	userCourse.DateCreated = time.Now()
	userCourse.DateUpdated = nil
	userCourse.Course = sa.customCourseConversionAPIToStorage(item.Course)

	_, err := sa.db.userCourses.InsertOne(sa.context, userCourse)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserCourse, nil, err)
	}
	return nil
}

// get UserCourseUnits gets all userUnits within a user's course
func (sa *Adapter) GetUserCourseUnits(appID string, orgID string, userID string, courseKey string) ([]model.UserUnit, error) {
	filter := bson.M{"app_id": appID, "org_id": orgID, "user_id": userID, "course_key": courseKey}
	var result []userUnit
	err := sa.db.userUnits.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		//no data
		return nil, nil
	}

	var convertedResult []model.UserUnit
	for _, retrievedResult := range result {
		singleConverted, err := sa.userUnitConversionStorageToAPI(retrievedResult)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// GetUserUnit finds a user unit
func (sa *Adapter) GetUserUnitExist(appID string, orgID string, userID string, courseKey string, unitKey string) (bool, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "user_id": userID, "course_key": courseKey, "unit.key": unitKey}
	var result userUnit
	// user find instead of findOne to avoid document none-exist error
	err := sa.db.userUnits.FindOne(sa.context, filter, &result, nil)
	if err != nil && err != mongo.ErrNoDocuments {
		return false, err
	}
	// if err != nil {

	// }

	// no function needs to return UserUnit so not implementating this function yet
	//convertedResult, err := sa.userUnitConversionStorageToAPI(result)
	// if err != nil {
	// 	return nil, err
	// }
	ifExist := result.ID != ""
	return ifExist, nil
}

// InsertUserUnit inserts a user unit
func (sa *Adapter) InsertUserUnit(item model.UserUnit) error {
	var userUnit userUnit
	userUnit.ID = item.ID
	userUnit.AppID = item.AppID
	userUnit.OrgID = item.OrgID
	userUnit.UserID = item.UserID
	userUnit.DateCreated = time.Now()
	userUnit.DateUpdated = nil
	userUnit.CourseKey = item.CourseKey
	userUnit.Unit = sa.customUnitConversionAPIToStorage(item.Unit)

	_, err := sa.db.userUnits.InsertOne(sa.context, userUnit)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
	}
	return nil
}

// UpdateUserUnit updates shcedules in a user unit
func (sa *Adapter) UpdateUserUnit(appID string, orgID string, userID string, courseKey string, item model.Unit) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "user_id": userID, "course_key": courseKey}
	update := bson.M{
		"$set": bson.M{
			"unit.schedule": item.Schedule,
			"date_updated":  time.Now(),
		},
	}
	result, err := sa.db.userUnits.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, &logutils.FieldArgs{"app_id": appID, "org_id": orgID, "user_id": userID, "course_key": courseKey}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeUserUnit, &logutils.FieldArgs{"app_id": appID, "org_id": orgID, "user_id": userID, "course_key": courseKey}, err)
	}
	return nil
}

// DeleteUserCourse deletes a user course
// func (sa *Adapter) DeleteUserCourse(appID string, orgID string, userID string, courseKey string) error {
// 	filter := bson.M{"app_id": appID, "org_id": orgID, "user_id": userID, "course.key": courseKey}
// 	result, err := sa.db.userCourses.DeleteOne(sa.context, filter, nil)
// 	if err != nil {
// 		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeUserCourse, &logutils.FieldArgs{"courseKey": courseKey}, err)
// 	}
// 	if result == nil {
// 		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"courseKey": courseKey}, err)
// 	}
// 	deletedCount := result.DeletedCount
// 	if deletedCount == 0 {
// 		return errors.WrapErrorData(logutils.StatusMissing, model.TypeUserCourse, &logutils.FieldArgs{"courseKey": courseKey}, err)
// 	}
// 	return nil
// }
