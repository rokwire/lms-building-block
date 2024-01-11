package storage

import (
	"lms/core/model"
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
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
		return nil, err
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

// customCourseConversionStorageToAPI formats storage struct to appropirate struct for API request
func (sa *Adapter) customCourseConversionStorageToAPI(item course) (model.Course, error) {
	var result model.Course
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.Key = item.Key
	result.Name = item.Name
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated
	if len(item.ModuleKeys) > 0 {
		var linked []module
		subFilter := bson.M{"org_id": item.OrgID, "app_id": item.AppID}
		subFilter["key"] = bson.M{"$in": item.ModuleKeys}
		err := sa.db.customModules.Find(sa.context, subFilter, &linked, nil)
		if err != nil {
			return result, err
		}

		for _, singleContent := range linked {
			convertedContent, err := sa.customModuleConversionStorageToAPI(singleContent)
			if err != nil {
				return result, err
			}
			result.Modules = append(result.Modules, convertedContent)
		}
	}
	return result, nil
}

// InsertCustomCourse inserts a course
func (sa *Adapter) InsertCustomCourse(item model.Course) error {
	item.DateCreated = time.Now()
	item.DateUpdated = nil
	course := sa.customCourseConversionAPIToStorage(item)

	_, err := sa.db.customCourses.InsertOne(sa.context, course)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "error inserting courses", nil, err)
	}
	return nil
}

// InsertCustomCourses insert an array of courses
func (sa *Adapter) InsertCustomCourses(items []model.Course) error {
	storeItems := make([]interface{}, len(items))
	for i, item := range items {
		item.DateCreated = time.Now()
		item.DateUpdated = nil
		course := sa.customCourseConversionAPIToStorage(item)
		storeItems[i] = course
	}

	_, err := sa.db.customCourses.InsertMany(sa.context, storeItems, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "error inserting courses", nil, err)
	}
	return nil
}

// customCourseConversionAPIToStorage formats API struct to stroage struct
func (sa *Adapter) customCourseConversionAPIToStorage(item model.Course) course {
	//parse into the storage format and pass parameters
	var moduleKeys []string
	for _, val := range item.Modules {
		moduleKeys = append(moduleKeys, val.Key)
	}

	var course course
	course.ID = item.ID
	course.AppID = item.AppID
	course.OrgID = item.OrgID
	course.Key = item.Key
	course.Name = item.Name
	course.ModuleKeys = moduleKeys
	course.DateCreated = item.DateCreated
	course.DateUpdated = item.DateUpdated

	return course
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
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
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
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

// DeleteCustomCourse deletes a course
func (sa *Adapter) DeleteCustomCourse(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	result, err := sa.db.customCourses.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"key": key}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
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
		return nil, err
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

// customModuleConversionStorageToAPI formats storage struct to appropirate struct for API request
func (sa *Adapter) customModuleConversionStorageToAPI(item module) (model.Module, error) {
	var result model.Module
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.Key = item.Key
	result.Name = item.Name
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated
	if len(item.UnitKeys) > 0 {
		var linked []unit
		subFilter := bson.M{"org_id": item.OrgID, "app_id": item.AppID}
		subFilter["key"] = bson.M{"$in": item.UnitKeys}
		err := sa.db.customUnits.Find(sa.context, subFilter, &linked, nil)
		if err != nil {
			return result, err
		}

		for _, singleContent := range linked {
			convertedContent, err := sa.customUnitConversionStorageToAPI(singleContent)
			if err != nil {
				return result, err
			}
			result.Units = append(result.Units, convertedContent)
		}
	}
	return result, nil
}

// InsertCustomModule inserts a module
func (sa *Adapter) InsertCustomModule(item model.Module) error {
	item.DateCreated = time.Now()
	item.DateUpdated = nil
	module := sa.customModuleConversionAPIToStorage(item)
	_, err := sa.db.customModules.InsertOne(sa.context, module)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "error inserting module", nil, err)
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
		return errors.WrapErrorAction(logutils.ActionInsert, "error inserting modules", nil, err)
	}
	return nil
}

// customModuleConversionAPIToStorage formats API struct to stroage struct
func (sa *Adapter) customModuleConversionAPIToStorage(item model.Module) module {
	//parse into the storage format and pass parameters
	var unitKeys []string
	for _, val := range item.Units {
		unitKeys = append(unitKeys, val.Key)
	}

	var module module
	module.ID = item.ID
	module.AppID = item.AppID
	module.OrgID = item.OrgID
	module.Key = item.Key
	module.Name = item.Name
	module.UnitKeys = unitKeys
	module.DateCreated = item.DateCreated
	module.DateUpdated = item.DateUpdated

	return module
}

// UpdateCustomModule updates a module
func (sa *Adapter) UpdateCustomModule(key string, item model.Module) error {
	//parse into the storage format and pass parameters
	var unitKeys []string
	for _, val := range item.Units {
		unitKeys = append(unitKeys, val.Key)
	}

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "key": key}
	update := bson.M{
		"$set": bson.M{
			"date_updated": time.Now(),
			"name":         item.Name,
			"unit_keys":    unitKeys,
		},
	}
	result, err := sa.db.customModules.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

// DeleteCustomModule deletes a module
func (sa *Adapter) DeleteCustomModule(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	result, err := sa.db.customModules.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"key": key}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
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
		return nil, err
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

// customUnitConversionStorageToAPI formats storage struct to appropirate struct for API request
func (sa *Adapter) customUnitConversionStorageToAPI(item unit) (model.Unit, error) {
	var result model.Unit
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.Key = item.Key
	result.Name = item.Name
	result.Schedule = item.Schedule
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated

	if len(item.ContentKeys) > 0 {
		var linked []content
		subFilter := bson.M{"org_id": item.OrgID, "app_id": item.AppID}
		subFilter["key"] = bson.M{"$in": item.ContentKeys}
		err := sa.db.customContent.Find(sa.context, subFilter, &linked, nil)
		if err != nil {
			return result, err
		}
		for _, singleContent := range linked {
			convertedContent, err := sa.customContentConversionStorageToAPI(singleContent)
			if err != nil {
				return result, err
			}
			result.Contents = append(result.Contents, convertedContent)
		}
	}
	return result, nil
}

// InsertCustomUnit inserts a unit
func (sa *Adapter) InsertCustomUnit(item model.Unit) error {
	item.DateCreated = time.Now()
	item.DateUpdated = nil
	unit := sa.customUnitConversionAPIToStorage(item)
	_, err := sa.db.customUnits.InsertOne(sa.context, unit)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "error inserting unit", nil, err)
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
		return errors.WrapErrorAction(logutils.ActionInsert, "error inserting units", nil, err)
	}
	return nil
}

// customUnitConversionAPIToStorage formats API struct to stroage struct
func (sa *Adapter) customUnitConversionAPIToStorage(item model.Unit) unit {
	//parse into the storage format and pass parameters
	var extractedKey []string
	for _, val := range item.Contents {
		extractedKey = append(extractedKey, val.Key)
	}

	var result unit
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.Key = item.Key
	result.Name = item.Name
	result.ContentKeys = extractedKey
	result.Schedule = item.Schedule
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated

	return result
}

// UpdateCustomUnit updates a unit
func (sa *Adapter) UpdateCustomUnit(key string, item model.Unit) error {
	//parse into the storage format and pass parameters
	var extractedKey []string
	for _, val := range item.Contents {
		extractedKey = append(extractedKey, val.Key)
	}

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "key": key}
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
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
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
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

// UpdateReferenceKeyToClientUnits doesn't work
func (sa *Adapter) UpdateReferenceKeyToClientUnits(oldCourseKey string, newCourseKey string) error {

	filter := bson.M{}
	update := bson.M{
		"$set": bson.M{
			"unit.course_key": bson.M{
				"$replaceAll": bson.M{
					"input":       "$course_key",
					"find":        oldCourseKey,
					"replacement": newCourseKey,
				},
			},
		},
	}

	_, err := sa.db.userUnits.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "", nil, err)
	}
	return nil
}

// DeleteCustomUnit deletes a unit
func (sa *Adapter) DeleteCustomUnit(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	result, err := sa.db.customUnits.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"key": key}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
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
	err := sa.db.customContent.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
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
	err := sa.db.customContent.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.customContentConversionStorageToAPI(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// customContentConversionStorageToAPI formats storage struct to appropirate struct for API request
func (sa *Adapter) customContentConversionStorageToAPI(item content) (model.Content, error) {
	var result model.Content
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.Key = item.Key
	result.Type = item.Type
	result.Name = item.Name
	result.Details = item.Details
	result.ContentReference = item.ContentReference
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated

	if len(item.LinkedContent) > 0 {
		var linkedContents []content
		subFilter := bson.M{"org_id": item.OrgID, "app_id": item.AppID}
		subFilter["key"] = bson.M{"$in": item.LinkedContent}
		err := sa.db.customContent.Find(sa.context, subFilter, &linkedContents, nil)
		if err != nil {
			return result, err
		}
		for _, singleContent := range linkedContents {
			convertedContent, err := sa.customContentConversionStorageToAPI(singleContent)
			if err != nil {
				return result, err
			}
			result.LinkedContent = append(result.LinkedContent, convertedContent)
		}
	}
	return result, nil
}

// InsertCustomContent inserts a content
func (sa *Adapter) InsertCustomContent(item model.Content) error {
	item.DateCreated = time.Now()
	item.DateUpdated = nil
	content := sa.customContentConversionAPIToStorage(item)
	_, err := sa.db.customContent.InsertOne(sa.context, content)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "error inserting content", nil, err)
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

	_, err := sa.db.customContent.InsertMany(sa.context, storeItems, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "error inserting contents", nil, err)
	}
	return nil
}

// customContentConversionAPIToStorage formats API struct to stroage struct
func (sa *Adapter) customContentConversionAPIToStorage(item model.Content) content {
	//parse into the storage format and pass parameters
	var extractedKey []string
	for _, val := range item.LinkedContent {
		extractedKey = append(extractedKey, val.Key)
	}

	var content content
	content.ID = item.ID
	content.AppID = item.AppID
	content.OrgID = item.OrgID
	content.Key = item.Key
	content.Type = item.Type
	content.Name = item.Name
	content.Details = item.Details
	content.ContentReference = item.ContentReference
	content.LinkedContent = extractedKey
	content.DateCreated = item.DateCreated
	content.DateUpdated = item.DateUpdated

	return content
}

// UpdateCustomContent updates a content
func (sa *Adapter) UpdateCustomContent(key string, item model.Content) error {
	//parse into the storage format and pass parameters
	var extractedKey []string
	for _, val := range item.LinkedContent {
		extractedKey = append(extractedKey, val.Key)
	}

	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "key": key}
	update := bson.M{
		"$set": bson.M{
			"type":           item.Type,
			"details":        item.Details,
			"name":           item.Name,
			"reference":      item.ContentReference,
			"linked_content": extractedKey,
			"date_updated":   time.Now(),
		},
	}
	result, err := sa.db.customContent.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

// DeleteCustomContent deletes a content
func (sa *Adapter) DeleteCustomContent(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	result, err := sa.db.customContent.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"key": key}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

// FindCourseConfigs finds course configs by the given search parameters
func (sa *Adapter) FindCourseConfigs(notificationsActive *bool) ([]model.CourseConfig, error) {
	filter := bson.M{}

	if notificationsActive != nil {
		filter["streaks_notifications_config.notifications_active"] = *notificationsActive
	}

	var configs []model.CourseConfig
	err := sa.db.courseConfigs.Find(sa.context, filter, &configs, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourseConfig, nil, err)
	}
	return configs, nil
}

// DeleteContentKeyFromLinkedContents deletes a content key from linkedContent field within customContent collection
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

	_, err := sa.db.customContent.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
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
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
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
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
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
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
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
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
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
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

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
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}
