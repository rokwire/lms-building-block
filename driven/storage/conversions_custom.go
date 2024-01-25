package storage

import (
	"lms/core/model"
)

// customCourseToStorage formats API struct to storage struct
func (sa *Adapter) customCourseToStorage(item model.Course) course {
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

// customCourseFromStorage formats storage struct to appropriate struct for API request
func (sa *Adapter) customCourseFromStorage(item course) (model.Course, error) {
	var result model.Course
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.Key = item.Key
	result.Name = item.Name
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated
	if len(item.ModuleKeys) > 0 {
		modules, err := sa.FindCustomModules(item.AppID, item.OrgID, nil, nil, item.ModuleKeys, nil)
		if err != nil {
			return result, err
		}

		result.Modules = make([]model.Module, len(modules))
		for i, key := range item.ModuleKeys {
			for _, module := range modules {
				if module.Key == key {
					result.Modules[i] = module
					break
				}
			}
		}
	}
	return result, nil
}

// customModuleFromStorage formats storage struct to appropriate struct for API request
func (sa *Adapter) customModuleFromStorage(item module) (model.Module, error) {
	var result model.Module
	result.ID = item.ID
	result.AppID = item.AppID
	result.OrgID = item.OrgID
	result.Key = item.Key
	result.Name = item.Name
	result.Display = item.Display
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated

	if len(item.UnitKeys) > 0 {
		units, err := sa.FindCustomUnits(item.AppID, item.OrgID, nil, nil, item.UnitKeys, nil)
		if err != nil {
			return result, err
		}

		result.Units = make([]model.Unit, len(units))
		for i, key := range item.UnitKeys {
			for _, unit := range units {
				if unit.Key == key {
					result.Units[i] = unit
					break
				}
			}
		}
	}
	return result, nil
}

// customModuleToStorage formats API struct to storage struct
func (sa *Adapter) customModuleToStorage(item model.Module) module {
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
	module.Display = item.Display
	module.DateCreated = item.DateCreated
	module.DateUpdated = item.DateUpdated

	return module
}

// customUnitFromStorage formats storage struct to appropriate struct for API request
func (sa *Adapter) customUnitFromStorage(item unit) (model.Unit, error) {
	result := model.Unit{ID: item.ID, AppID: item.AppID, OrgID: item.OrgID, Key: item.Key, Name: item.Name, Schedule: item.Schedule,
		ScheduleStart: item.ScheduleStart, Required: item.Required, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}

	if len(item.ContentKeys) > 0 {
		contents, err := sa.FindCustomContents(item.AppID, item.OrgID, nil, nil, item.ContentKeys)
		if err != nil {
			return result, err
		}
		result.Contents = contents
	}
	return result, nil
}

func (sa *Adapter) customUnitToStorage(item model.Unit) unit {
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
	result.ScheduleStart = item.ScheduleStart
	result.Required = item.Required
	result.DateCreated = item.DateCreated
	result.DateUpdated = item.DateUpdated

	return result
}

// userCourseConversionHelper formats storage struct to appropriate struct for API request
func (sa *Adapter) userCourseFromStorage(item userCourse) (model.UserCourse, error) {
	timezone := model.Timezone{Name: item.TimezoneName, Offset: item.TimezoneOffset}
	result := model.UserCourse{ID: item.ID, AppID: item.AppID, OrgID: item.OrgID, UserID: item.UserID, Timezone: timezone,
		Streak: item.Streak, StreakResets: item.StreakResets, Pauses: item.Pauses, PauseUses: item.PauseUses,
		DateCreated: item.DateCreated, DateUpdated: item.DateUpdated, DateDropped: item.DateDropped}

	convertedCourse, err := sa.customCourseFromStorage(item.Course)
	if err != nil {
		return result, err
	}
	result.Course = convertedCourse

	return result, nil
}

func (sa *Adapter) userUnitFromStorage(item userUnit) (model.UserUnit, error) {
	result := model.UserUnit{ID: item.ID, AppID: item.AppID, OrgID: item.OrgID, UserID: item.UserID, CourseKey: item.CourseKey,
		Completed: item.Completed, Current: item.Current, LastCompleted: item.LastCompleted, DateCreated: item.DateCreated, DateUpdated: item.DateUpdated}

	unit, err := sa.customUnitFromStorage(item.Unit)
	if err != nil {
		return result, err
	}
	result.Unit = unit

	return result, nil
}
