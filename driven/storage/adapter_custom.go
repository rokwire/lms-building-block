package storage

import (
	"lms/core/model"
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
)

// FindCustomCourses finds courses by a set of parameters
func (sa *Adapter) FindCustomCourses(appID string, orgID string, id []string, name []string, key []string, moduleKeys []string) ([]model.Course, error) {
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
		singleConverted, err := sa.customCourseFromStorage(retrievedCourse)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// FindCustomCourse finds a course by id
func (sa *Adapter) FindCustomCourse(appID string, orgID string, key string) (*model.Course, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	var result course
	err := sa.db.customCourses.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.customCourseFromStorage(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertCustomCourse inserts a course
func (sa *Adapter) InsertCustomCourse(item model.Course) error {
	item.DateCreated = time.Now()
	course := sa.customCourseToStorage(item)

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

// DecrementUserCoursePauses decrements all matching user course pauses by 1
func (sa *Adapter) DecrementUserCoursePauses(appID string, orgID string, userIDs []string, key string) error {
	// should never decrement pauses for all user courses matching the first three filter fields at once
	if len(userIDs) == 0 {
		return errors.ErrorData(logutils.StatusMissing, "user ids", nil)
	}

	now := time.Now().UTC()
	filter := bson.M{"app_id": appID, "org_id": orgID, "course.key": key, "user_id": bson.M{"$in": userIDs}}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$inc": bson.M{
			"streak": 1, // using a pause counts toward the streak
			"pauses": -1,
		},
		"$push": bson.M{
			"pause_uses": now,
		},
		"$set": bson.M{
			"date_updated": now,
		},
	}
	res, err := sa.db.userCourses.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &errArgs, err)
	}
	if res.MatchedCount != res.ModifiedCount {
		errArgs["matched"] = res.MatchedCount
		errArgs["modified"] = res.ModifiedCount
		return errors.ErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &errArgs)
	}

	return nil
}

// ResetUserCourseStreaks resets all matching user course streaks to 0
func (sa *Adapter) ResetUserCourseStreaks(appID string, orgID string, userIDs []string, key string) error {
	// should never reset streaks for all user courses matching the first three filter fields at once
	if len(userIDs) == 0 {
		return errors.ErrorData(logutils.StatusMissing, "user ids", nil)
	}

	now := time.Now().UTC()
	filter := bson.M{"app_id": appID, "org_id": orgID, "course.key": key, "user_id": bson.M{"$in": userIDs}}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$push": bson.M{
			"streak_resets": now,
		},
		"$set": bson.M{
			"streak":       0,
			"date_updated": now,
		},
	}
	res, err := sa.db.userCourses.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &errArgs, err)
	}
	if res.MatchedCount != res.ModifiedCount {
		errArgs["matched"] = res.MatchedCount
		errArgs["modified"] = res.ModifiedCount
		return errors.ErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &errArgs)
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

// FindCustomModules finds courses by a set of parameters
func (sa *Adapter) FindCustomModules(appID string, orgID string, id []string, name []string, key []string, unitKeys []string) ([]model.Module, error) {
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
		singleConverted, err := sa.customModuleFromStorage(retrievedModule)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// FindCustomModule finds a module by id
func (sa *Adapter) FindCustomModule(appID string, orgID string, key string) (*model.Module, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	var result module
	err := sa.db.customModules.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.customModuleFromStorage(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertCustomModule inserts a module
func (sa *Adapter) InsertCustomModule(item model.Module) error {
	item.DateCreated = time.Now()
	module := sa.customModuleToStorage(item)
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
		module := sa.customModuleToStorage(item)
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
			"styles":       item.Styles,
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

// FindCustomUnits finds units by a set of parameters
func (sa *Adapter) FindCustomUnits(appID string, orgID string, id []string, name []string, key []string, contentKeys []string) ([]model.Unit, error) {
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
		singleConverted, err := sa.customUnitFromStorage(retrievedUnit)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// FindCustomUnit finds a unit by id
func (sa *Adapter) FindCustomUnit(appID string, orgID string, key string) (*model.Unit, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	var result unit
	err := sa.db.customUnits.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}

	convertedResult, err := sa.customUnitFromStorage(result)
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertCustomUnit inserts a unit
func (sa *Adapter) InsertCustomUnit(item model.Unit) error {
	unit := sa.customUnitToStorage(item)
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
		unit := sa.customUnitToStorage(item)
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
			"required":     len(item.Schedule),
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

// FindCustomContents finds contents by a set of parameters
func (sa *Adapter) FindCustomContents(appID string, orgID string, id []string, name []string, key []string) ([]model.Content, error) {
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
	var result []model.Content
	err := sa.db.customContents.Find(sa.context, filter, &result, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeContent, &errArgs, err)
	}

	return result, nil
}

// FindCustomContent finds a content by id
func (sa *Adapter) FindCustomContent(appID string, orgID string, key string) (*model.Content, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "key": key}
	var result model.Content
	err := sa.db.customContents.FindOne(sa.context, filter, &result, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeContent, &errArgs, err)
	}

	return &result, nil
}

// InsertCustomContent inserts a content
func (sa *Adapter) InsertCustomContent(item model.Content) error {
	item.DateCreated = time.Now()
	_, err := sa.db.customContents.InsertOne(sa.context, item)
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
		storeItems[i] = item
	}

	_, err := sa.db.customContents.InsertMany(sa.context, storeItems, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeContent, nil, err)
	}
	return nil
}

// UpdateCustomContent updates a content
func (sa *Adapter) UpdateCustomContent(key string, item model.Content) error {
	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "key": key}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$set": bson.M{
			"type":           item.Type,
			"details":        item.Details,
			"name":           item.Name,
			"reference":      item.Reference,
			"linked_content": item.LinkedContent,
			"styles":         item.Styles,
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

// FindCourseConfigs finds course configs by the given search parameters
func (sa *Adapter) FindCourseConfigs(appID *string, orgID *string, notificationsActive *bool) ([]model.CourseConfig, error) {
	filter := bson.M{}

	if appID != nil {
		filter["app_id"] = *appID
	}
	if orgID != nil {
		filter["org_id"] = *orgID
	}
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

// FindCourseConfig finds a single course config by course key
func (sa *Adapter) FindCourseConfig(appID string, orgID string, key string) (*model.CourseConfig, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "course_key": key}
	errArgs := logutils.FieldArgs(filter)

	var config model.CourseConfig
	err := sa.db.courseConfigs.FindOne(sa.context, filter, &config, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeCourseConfig, &errArgs, err)
	}

	return &config, nil
}

// InsertCourseConfig inserts a new course config
func (sa *Adapter) InsertCourseConfig(config model.CourseConfig) error {
	_, err := sa.db.courseConfigs.InsertOne(sa.context, config)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeCourseConfig, &logutils.FieldArgs{"id": config.ID, "app_id": config.AppID, "org_id": config.OrgID, "course_key": config.CourseKey}, err)
	}

	return nil
}

// UpdateCourseConfig updates an existing course config
func (sa *Adapter) UpdateCourseConfig(config model.CourseConfig) error {
	filter := bson.M{"org_id": config.OrgID, "app_id": config.AppID, "course_key": config.CourseKey}
	update := bson.M{
		"$set": bson.M{
			"initial_pauses":               config.InitialPauses,
			"max_pauses":                   config.MaxPauses,
			"pause_reward_streak":          config.PauseRewardStreak,
			"streaks_notifications_config": config.StreaksNotificationsConfig,
			"date_updated":                 time.Now().UTC(),
		},
	}
	errArgs := logutils.FieldArgs(filter)

	res, err := sa.db.courseConfigs.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeCourseConfig, &errArgs, err)
	}
	if res.ModifiedCount != 1 {
		errArgs["modified"] = res.ModifiedCount
		return errors.ErrorAction(logutils.ActionUpdate, model.TypeCourseConfig, &errArgs)
	}

	return nil
}

// DeleteCourseConfig deletes an existing course config by course key
func (sa *Adapter) DeleteCourseConfig(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "course_key": key}
	errArgs := logutils.FieldArgs(filter)

	result, err := sa.db.courseConfigs.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeCourseConfig, &errArgs, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "delete course config result", &errArgs, err)
	}

	if result.DeletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeCourseConfig, &errArgs, err)
	}
	return nil
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

// FindUserCourses finds user course by a set of parameters
func (sa *Adapter) FindUserCourses(id []string, appID string, orgID string, name []string, key []string, userID *string, timezoneOffsetPairs []model.TZOffsetPair) ([]model.UserCourse, error) {
	filter := bson.M{"app_id": appID, "org_id": orgID}
	if len(id) != 0 {
		filter["_id"] = bson.M{"$in": id}
	}

	if len(name) != 0 {
		filter["course.name"] = bson.M{"$in": name}
	}

	if len(key) != 0 {
		filter["course.key"] = bson.M{"$in": key}
	}

	if userID != nil {
		filter["user_id"] = userID
	}

	// timezone offsets
	if len(timezoneOffsetPairs) > 0 {
		offsetFilters := make(bson.A, 0)
		for _, offsetPair := range timezoneOffsetPairs {
			offsetFilters = append(offsetFilters,
				bson.M{
					"timezone_offset": bson.M{
						"$gte": offsetPair.Lower,
						"$lte": offsetPair.Upper,
					},
				},
			)
		}
		filter["$or"] = offsetFilters
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
		singleConverted, err := sa.userCourseFromStorage(retrievedResult)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleConverted)
	}

	return convertedResult, nil
}

// FindUserCourse finds a user course by id
func (sa *Adapter) FindUserCourse(appID string, orgID string, userID string, courseKey string) (*model.UserCourse, error) {
	filter := bson.M{"app_id": appID, "org_id": orgID, "user_id": userID, "course.key": courseKey}
	var result []userCourse
	err := sa.db.userCourses.Find(sa.context, filter, &result, nil)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}

	convertedResult, err := sa.userCourseFromStorage(result[0])
	if err != nil {
		return nil, err
	}

	return &convertedResult, nil
}

// InsertUserCourse inserts a user course
func (sa *Adapter) InsertUserCourse(item model.UserCourse) error {
	userCourse := sa.userCourseToStorage(item)

	_, err := sa.db.userCourses.InsertOne(sa.context, userCourse)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserCourse, nil, err)
	}
	return nil
}

// UpdateUserCourse updates a user course
func (sa *Adapter) UpdateUserCourse(item model.UserCourse) error {
	filter := bson.M{"app_id": item.AppID, "org_id": item.OrgID, "course.key": item.Course.Key, "user_id": item.UserID}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$set": bson.M{
			"streak":          item.Streak,
			"pauses":          item.Pauses,
			"streak_restarts": item.StreakRestarts,
			"date_dropped":    item.DateDropped,
			"date_updated":    time.Now().UTC(),
		},
	}
	result, err := sa.db.userCourses.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserCourse, &errArgs, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeUserCourse, &errArgs, err)
	}
	return nil
}

// UpdateUserTimezone updates a user's timezone information in all its related userCourse storage struct
func (sa *Adapter) UpdateUserTimezone(appID string, orgID string, userID string, timezoneName string, timezoneOffset int) error {
	filter := bson.M{"app_id": appID, "org_id": orgID, "user_id": userID}

	update := bson.M{
		"$set": bson.M{
			"timezone_name":   timezoneName,
			"timezone_offset": timezoneOffset,
		},
	}
	result, err := sa.db.userCourses.UpdateMany(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{}, err)
	}
	return nil
}

// DeleteUserCourse deletes a user course
func (sa *Adapter) DeleteUserCourse(appID string, orgID string, userID string, courseKey string) error {
	filter := bson.M{"app_id": appID, "org_id": orgID, "user_id": userID, "course.key": courseKey}
	errArgs := logutils.FieldArgs(filter)
	result, err := sa.db.userCourses.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeUserCourse, &errArgs, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "delete user course result", &errArgs, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeUserCourse, &errArgs, err)
	}
	return nil
}

// DeleteUserCourses deletes all user courses for a course key
func (sa *Adapter) DeleteUserCourses(appID string, orgID string, key string) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "course.key": key}
	result, err := sa.db.userCourses.DeleteMany(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeCourse, &logutils.FieldArgs{"key": key}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"key": key}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeCourse, &logutils.FieldArgs{"key": key}, err)
	}
	return nil
}

// FindUserUnit finds a user unit
func (sa *Adapter) FindUserUnit(appID string, orgID string, userID string, courseKey string, unitKey *string) (*model.UserUnit, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "user_id": userID, "course_key": courseKey}
	if unitKey != nil {
		filter["unit.key"] = *unitKey
	}

	var results []userUnit
	err := sa.db.userUnits.Find(sa.context, filter, &results, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUserUnit, &errArgs, err)
	}
	if len(results) == 0 {
		return nil, nil
	}

	// no function needs to return UserUnit so not implementating this function yet
	convertedResult, err := sa.userUnitFromStorage(results[0])
	if err != nil {
		return nil, err
	}
	return &convertedResult, nil
}

// FindUserUnits finds user units by search parameters
func (sa *Adapter) FindUserUnits(appID string, orgID string, userIDs []string, courseKey string, current *bool) ([]model.UserUnit, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID, "course_key": courseKey}
	if len(userIDs) != 0 {
		filter["user_id"] = bson.M{"$in": userIDs}
	}
	if current != nil {
		filter["current"] = *current
	}

	var results []userUnit
	err := sa.db.userUnits.Find(sa.context, filter, &results, nil)
	if err != nil {
		errArgs := logutils.FieldArgs(filter)
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUserUnit, &errArgs, err)
	}

	userUnits := make([]model.UserUnit, len(results))
	for i, result := range results {
		convertedResult, err := sa.userUnitFromStorage(result)
		if err != nil {
			return nil, err
		}
		userUnits[i] = convertedResult
	}

	return userUnits, nil
}

// InsertUserUnit inserts a user unit
func (sa *Adapter) InsertUserUnit(item model.UserUnit) error {
	userUnit := sa.userUnitToStorage(item)

	_, err := sa.db.userUnits.InsertOne(sa.context, userUnit)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, model.TypeUserUnit, nil, err)
	}
	return nil
}

// UpdateUserUnit updates shcedules in a user unit
func (sa *Adapter) UpdateUserUnit(item model.UserUnit) error {
	filter := bson.M{"org_id": item.OrgID, "app_id": item.AppID, "user_id": item.UserID, "course_key": item.CourseKey, "unit.key": item.Unit.Key}
	errArgs := logutils.FieldArgs(filter)
	update := bson.M{
		"$set": bson.M{
			"unit.schedule":  item.Unit.Schedule,
			"completed":      item.Completed,
			"current":        item.Current,
			"last_completed": item.LastCompleted,
			"date_updated":   time.Now().UTC(),
		},
	}
	result, err := sa.db.userUnits.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeUserUnit, &errArgs, err)
	}
	if result.MatchedCount == 0 {
		return errors.ErrorData(logutils.StatusMissing, model.TypeUserUnit, &errArgs)
	}
	return nil
}
