package storage

import (
	"lms/core/model"

	"go.mongodb.org/mongo-driver/bson"
)

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
		//linked, err := sa.GetCustomModules(item.AppID, item.OrgID, nil, nil, item.ModuleKeys, nil)
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
	// if len(item.UnitKeys) > 0 {
	// 	var linked []unit
	// 	subFilter := bson.M{"org_id": item.OrgID, "app_id": item.AppID}
	// 	subFilter["key"] = bson.M{"$in": item.UnitKeys}
	// 	err := sa.db.customUnits.Find(sa.context, subFilter, &linked, nil)
	// 	if err != nil {
	// 		return result, err
	// 	}

	// 	for _, singleContent := range linked {
	// 		convertedContent, err := sa.customUnitConversionStorageToAPI(singleContent)
	// 		if err != nil {
	// 			return result, err
	// 		}
	// 		result.Units = append(result.Units, convertedContent)
	// 	}
	// }
	if len(item.UnitKeys) > 0 {
		var units []model.Unit
		units, err := sa.GetCustomUnits(item.AppID, item.OrgID, nil, nil, item.UnitKeys, nil)
		if err != nil {
			return result, err
		}
		result.Units = units
	}
	return result, nil
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
	// if len(item.ContentKeys) > 0 {
	// 	var linked []content
	// 	subFilter := bson.M{"org_id": item.OrgID, "app_id": item.AppID}
	// 	subFilter["key"] = bson.M{"$in": item.ContentKeys}
	// 	err := sa.db.customContents.Find(sa.context, subFilter, &linked, nil)
	// 	if err != nil {
	// 		return result, err
	// 	}
	// 	for _, singleContent := range linked {
	// 		convertedContent, err := sa.customContentConversionStorageToAPI(singleContent)
	// 		if err != nil {
	// 			return result, err
	// 		}
	// 		result.Contents = append(result.Contents, convertedContent)
	// 	}
	// }

	if len(item.ContentKeys) > 0 {
		var contents []model.Content
		contents, err := sa.GetCustomContents(item.AppID, item.OrgID, nil, nil, item.ContentKeys)
		if err != nil {
			return result, err
		}
		result.Contents = contents
	}
	return result, nil
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
	result.LinkedContent = item.LinkedContent

	// if len(item.LinkedContent) > 0 {
	// 	var linkedContents []content
	// 	subFilter := bson.M{"org_id": item.OrgID, "app_id": item.AppID}
	// 	subFilter["key"] = bson.M{"$in": item.LinkedContent}
	// 	err := sa.db.customContents.Find(sa.context, subFilter, &linkedContents, nil)
	// 	if err != nil {
	// 		return result, err
	// 	}
	// 	for _, singleContent := range linkedContents {
	// 		convertedContent, err := sa.customContentConversionStorageToAPI(singleContent)
	// 		if err != nil {
	// 			return result, err
	// 		}
	// 		result.LinkedContent = append(result.LinkedContent, convertedContent)
	// 	}
	// }
	return result, nil
}

// customContentConversionAPIToStorage formats API struct to stroage struct
func (sa *Adapter) customContentConversionAPIToStorage(item model.Content) content {
	//parse into the storage format and pass parameters
	// var extractedKey []string
	// for _, val := range item.LinkedContent {
	// 	extractedKey = append(extractedKey, val.Key)
	// }

	var content content
	content.ID = item.ID
	content.AppID = item.AppID
	content.OrgID = item.OrgID
	content.Key = item.Key
	content.Type = item.Type
	content.Name = item.Name
	content.Details = item.Details
	content.ContentReference = item.ContentReference
	content.LinkedContent = item.LinkedContent
	content.DateCreated = item.DateCreated
	content.DateUpdated = item.DateUpdated

	return content
}

// userCourseConversionStorageToAPI formats storage struct to appropirate struct for API request
func (sa *Adapter) userCourseConversionStorageToAPI(item userCourse) (model.UserCourse, error) {
	var result model.UserCourse
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.UserID = item.UserID
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated
	result.DateDropped = item.DateDropped

	convertedCourse, err := sa.customCourseConversionStorageToAPI(item.Course)
	if err != nil {
		return result, err
	}
	result.Course = convertedCourse

	return result, nil
}

// userUnitConversionStorageToAPI formats storage struct to appropirate struct for API request
func (sa *Adapter) userUnitConversionStorageToAPI(item userUnit) (model.UserUnit, error) {
	var result model.UserUnit
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.UserID = item.UserID
	result.CourseKey = item.CourseKey
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated

	convertedUnit, err := sa.customUnitConversionStorageToAPI(item.Unit)
	if err != nil {
		return result, err
	}
	result.Unit = convertedUnit

	return result, nil
}
