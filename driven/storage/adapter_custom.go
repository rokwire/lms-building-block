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

	err := sa.db.customCourse.Find(sa.context, filter, &result, nil)
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
	err := sa.db.customCourse.FindOne(sa.context, filter, &result, nil)
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
		err := sa.db.customModule.Find(sa.context, subFilter, &linked, nil)
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

	_, err := sa.db.customCourse.InsertOne(sa.context, course)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "", nil, err)
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
			"key":          item.Key,
			"name":         item.Name,
			"module_keys":  moduleKeys,
		},
	}
	result, err := sa.db.customCourse.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

// DeleteCustomCourse deletes a course
func (sa *Adapter) DeleteCustomCourse(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	result, err := sa.db.customCourse.DeleteOne(sa.context, filter, nil)
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

	err := sa.db.customModule.Find(sa.context, filter, &result, nil)
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
	err := sa.db.customModule.FindOne(sa.context, filter, &result, nil)
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
	result.CourseKey = item.CourseKey
	result.Key = item.Key
	result.Name = item.Name
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated
	if len(item.UnitKeys) > 0 {
		var linked []unit
		subFilter := bson.M{"org_id": item.OrgID, "app_id": item.AppID}
		subFilter["key"] = bson.M{"$in": item.UnitKeys}
		err := sa.db.customUnit.Find(sa.context, subFilter, &linked, nil)
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
	_, err := sa.db.customModule.InsertOne(sa.context, module)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "", nil, err)
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
	module.CourseKey = item.CourseKey
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
			"course_key":   item.CourseKey,
			"key":          item.Key,
			"name":         item.Name,
			"unit_keys":    unitKeys,
		},
	}
	result, err := sa.db.customModule.UpdateOne(sa.context, filter, update, nil)
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
	result, err := sa.db.customModule.DeleteOne(sa.context, filter, nil)
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
func (sa *Adapter) GetCustomUnits(appID string, orgID string, id []string, name []string, key []string) ([]model.Unit, error) {
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
	var result []unit
	err := sa.db.customUnit.Find(sa.context, filter, &result, nil)
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
	err := sa.db.customUnit.FindOne(sa.context, filter, &result, nil)
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
	result.CourseKey = item.CourseKey
	result.ModuleKey = item.ModuleKey
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
	result := sa.customUnitConversionAPIToStorage(item)
	_, err := sa.db.customUnit.InsertOne(sa.context, result)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "", nil, err)
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
	result.CourseKey = item.CourseKey
	result.ModuleKey = item.ModuleKey
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
			"course_key":   item.CourseKey,
			"module_key":   item.ModuleKey,
			"key":          item.Key,
			"name":         item.Name,
			"content_keys": extractedKey,
			"schedule":     item.Schedule,
			"date_updated": time.Now(),
		},
	}
	result, err := sa.db.customUnit.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{"key": key}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

// DeleteCustomUnit deletes a unit
func (sa *Adapter) DeleteCustomUnit(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	result, err := sa.db.customUnit.DeleteOne(sa.context, filter, nil)
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
	result.CourseKey = item.CourseKey
	result.ModuleKey = item.ModuleKey
	result.UnitKey = item.UnitKey
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
		return errors.WrapErrorAction(logutils.ActionInsert, "", nil, err)
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
	content.CourseKey = item.CourseKey
	content.ModuleKey = item.ModuleKey
	content.UnitKey = item.UnitKey
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
			"course_key":     item.CourseKey,
			"module_key":     item.ModuleKey,
			"unit_key":       item.UnitKey,
			"key":            item.Key,
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

// FindUserCourses finds user courses by the given search parameters
func (sa *Adapter) FindUserCourses(id []string, appID string, orgID string, name []string, key []string, userID *string, timezoneOffsetPairs []model.TZOffsetPair, requirements map[string]interface{}) ([]model.UserCourse, error) {
	filter := bson.M{"app_id": appID, "org_id": orgID}

	//TODO: populate the missing fields in the filter

	// timezone offsets
	if len(timezoneOffsetPairs) > 0 {
		offsetFilters := make(bson.A, 0)
		for _, offsetPair := range timezoneOffsetPairs {
			offsetFilters = append(offsetFilters, bson.M{"$and": bson.A{bson.M{"$gte": offsetPair.Lower}, bson.M{"$lte": offsetPair.Upper}}})
		}
		filter["timezone_offset"] = bson.M{"$or": offsetFilters}
	}

	// notification requirements
	for reqKey, reqVal := range requirements {
		filter[reqKey] = reqVal
	}

	var userCourses []model.UserCourse
	err := sa.db.userCourse.Find(sa.context, filter, &userCourses, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUserCourse, nil, err)
	}

	return userCourses, nil
}

// FindCourseConfigs finds course configs by the given search parameters
func (sa *Adapter) FindCourseConfigs(notificationsActive *bool) ([]model.CourseConfig, error) {
	filter := bson.M{}

	if notificationsActive != nil {
		filter["active"] = *notificationsActive
	}

	var configs []model.CourseConfig
	err := sa.db.courseConfigs.Find(sa.context, filter, &configs, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourseConfig, nil, err)
	}

	return configs, nil
}
