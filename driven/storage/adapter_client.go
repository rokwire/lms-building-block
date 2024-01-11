package storage

import (
	"lms/core/model"
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
)

// FindUserCourses finds user courses by the given search parameters
func (sa *Adapter) FindUserCourses(id []string, appID string, orgID string, name []string, key []string, userID *string, timezoneOffsetPairs []model.TZOffsetPair, requirements map[string]interface{}) ([]model.UserCourse, error) {
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

	// notification requirements
	for reqKey, reqVal := range requirements {
		if reqKey == "completed_tasks" {
			now := time.Now()
			y, m, d := now.Date()
			todayStart := time.Date(y, m, d, 0, 0, 0, now.Nanosecond(), time.UTC)
			if reqVal == true {
				filter["completed_tasks"] = bson.M{
					"$gte": todayStart,
				}
			} else if reqVal == false {
				noneCompletedFilter := make(bson.A, 0)
				noneCompletedFilter = append(noneCompletedFilter,
					bson.M{
						"completed_tasks": bson.M{
							"$lt": todayStart,
						},
					},
				)
				noneCompletedFilter = append(noneCompletedFilter,
					bson.M{
						"completed_tasks": bson.M{
							"$eq": nil,
						},
					},
				)
				filter["$or"] = noneCompletedFilter
			} else {
				// only accept boolean and nil
				return nil, errors.ErrorData(logutils.StatusInvalid, "notification requirement", &logutils.FieldArgs{"completed_tasks": reqVal})
			}
		} else {
			filter[reqKey] = reqVal
		}
	}

	var dbUserCourses []userCourse
	err := sa.db.userCourses.Find(sa.context, filter, &dbUserCourses, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeUserCourse, nil, err)
	}

	var convertedResult []model.UserCourse
	//convert storage format to desired return
	for _, duUserCourse := range dbUserCourses {
		singleResult, err := sa.userCourseConversionStorageToAPI(duUserCourse)
		if err != nil {
			return nil, err
		}
		convertedResult = append(convertedResult, singleResult)
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

// userCourseConversionHelper formats storage struct to appropirate struct for API request
func (sa *Adapter) userCourseConversionStorageToAPI(item userCourse) (model.UserCourse, error) {
	var result model.UserCourse
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.UserID = item.UserID
	result.TimezoneName = item.TimezoneName
	result.TimezoneOffset = item.TimezoneOffset
	result.Streaks = item.Streaks
	result.Pauses = item.Pauses
	result.CompletedTasks = item.CompletedTasks
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated

	convertedCourse, err := sa.customCourseConversionStorageToAPI(item.Course)
	if err != nil {
		return result, err
	}
	result.Course = convertedCourse

	return result, nil
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
		return errors.WrapErrorAction(logutils.ActionInsert, "", nil, err)
	}
	return nil
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
	userUnit.Unit = sa.customUnitConversionAPIToStorage(item.Unit)

	_, err := sa.db.userUnits.InsertOne(sa.context, userUnit)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionInsert, "", nil, err)
	}
	return nil
}

// UpdateUserUnit updates shcedules in a user unit
func (sa *Adapter) UpdateUserUnit(appID string, orgID string, userID string, userUnitID string, item model.Unit) error {
	filter := bson.M{"org_id": orgID, "app_id": appID, "user_id": userID, "_id": userUnitID}
	update := bson.M{
		"$set": bson.M{
			"unit.schedule": item.Schedule,
			"date_updated":  time.Now(),
		},
	}
	result, err := sa.db.userUnits.UpdateOne(sa.context, filter, update, nil)
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
	result, err := sa.db.userCourses.DeleteOne(sa.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, "", &logutils.FieldArgs{"courseKey": courseKey}, err)
	}
	if result == nil {
		return errors.WrapErrorData(logutils.StatusInvalid, "result", &logutils.FieldArgs{"courseKey": courseKey}, err)
	}
	deletedCount := result.DeletedCount
	if deletedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{"courseKey": courseKey}, err)
	}
	return nil
}

// UpdateUserCourseStreaks updates streaks, pauses, and completed_tasks fieled
func (sa *Adapter) UpdateUserCourseStreaks(appID string, orgID string, userID *string, userCourseID *string, courseKey string, streaks *int, pauses *int, userTime *time.Time) error {
	filter := bson.M{"app_id": appID, "org_id": orgID, "course.key": courseKey}
	if userID != nil {
		filter["user_id"] = userID
	}
	if userCourseID != nil {
		filter["_id"] = userCourseID
	}

	updateVals := bson.M{}
	if streaks != nil {
		updateVals["streaks"] = streaks
	}
	if pauses != nil {
		updateVals["pauses"] = pauses
	}
	if userTime != nil {
		updateVals["completed_tasks"] = userTime
	}

	update := bson.M{
		"$set": updateVals,
	}
	result, err := sa.db.userCourses.UpdateOne(sa.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, "", &logutils.FieldArgs{}, err)
	}
	if result.MatchedCount == 0 {
		return errors.WrapErrorData(logutils.StatusMissing, "", &logutils.FieldArgs{}, err)
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
