// Code generated by api-generator DO NOT EDIT.
package web

import (
	"lms/core"
	"lms/core/model"
	"lms/utils"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// APIsHandler handles the rest APIs implementation
type APIsHandler struct {
	app *core.Application
}

// Default

func (a APIsHandler) defaultGetVersion(claims *tokenauth.Claims, params map[string]interface{}) (*string, error) {
	return a.app.Default.GetVersion()
}

// Client

func (a APIsHandler) clientGetCourses(claims *tokenauth.Claims, params map[string]interface{}) ([]model.ProviderCourse, error) {
	courseType, err := utils.GetValue[*string](params, "course_type", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("courseType"), err)
	}

	return a.app.Client.GetCourses(claims, courseType)
}

func (a APIsHandler) clientGetCourse(claims *tokenauth.Claims, params map[string]interface{}) (*model.ProviderCourse, error) {
	id, err := utils.GetValue[string](params, "id", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	return a.app.Client.GetCourse(claims, id)
}

func (a APIsHandler) clientGetAssignmentGroups(claims *tokenauth.Claims, params map[string]interface{}) ([]model.AssignmentGroup, error) {
	id, err := utils.GetValue[string](params, "id", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	include, err := utils.GetValue[*string](params, "include", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("include"), err)
	}

	return a.app.Client.GetAssignmentGroups(claims, id, include)
}

func (a APIsHandler) clientGetCourseUser(claims *tokenauth.Claims, params map[string]interface{}) (*model.User, error) {
	id, err := utils.GetValue[string](params, "id", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	include, err := utils.GetValue[*string](params, "include", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("include"), err)
	}

	return a.app.Client.GetCourseUser(claims, id, include)
}

func (a APIsHandler) clientGetCurrentUser(claims *tokenauth.Claims, params map[string]interface{}) (*model.User, error) {
	return a.app.Client.GetCurrentUser(claims)
}

func (a APIsHandler) clientGetUserCourses(claims *tokenauth.Claims, params map[string]interface{}) ([]model.UserCourse, error) {
	id, err := utils.GetValue[*string](params, "id", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	name, err := utils.GetValue[*string](params, "name", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("name"), err)
	}

	key, err := utils.GetValue[*string](params, "key", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Client.GetUserCourses(claims, id, name, key)
}

func (a APIsHandler) clientGetUserCourse(claims *tokenauth.Claims, params map[string]interface{}) (*model.UserCourse, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Client.GetUserCourse(claims, key)
}

func (a APIsHandler) clientCreateUserCourse(claims *tokenauth.Claims, params map[string]interface{}, item *model.Timezone) (*model.UserCourse, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Client.CreateUserCourse(claims, key, *item)
}

func (a APIsHandler) clientDeleteUserCourse(claims *tokenauth.Claims, params map[string]interface{}) error {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Client.DeleteUserCourse(claims, key)
}

func (a APIsHandler) clientDropUserCourse(claims *tokenauth.Claims, params map[string]interface{}, item *model.UserCourse) (*model.UserCourse, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Client.DropUserCourse(claims, key)
}

func (a APIsHandler) clientUpdateUserCourseUnitProgress(claims *tokenauth.Claims, params map[string]interface{}, item *model.UnitWithTimezone) (*model.UnitWithTimezone, error) {
	courseKey, err := utils.GetValue[string](params, "course_key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("courseKey"), err)
	}

	unitKey, err := utils.GetValue[string](params, "unit_key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("unitKey"), err)
	}

	return a.app.Client.UpdateUserCourseUnitProgress(claims, courseKey, unitKey, *item)
}

func (a APIsHandler) clientGetUserCourseUnits(claims *tokenauth.Claims, params map[string]interface{}) ([]model.UserUnit, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Client.GetUserCourseUnits(claims, key)
}

func (a APIsHandler) clientGetCustomCourses(claims *tokenauth.Claims, params map[string]interface{}) ([]model.Course, error) {
	return a.app.Client.GetCustomCourses(claims)
}

func (a APIsHandler) clientGetCustomCourseConfig(claims *tokenauth.Claims, params map[string]interface{}) (*model.CourseConfig, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Client.GetCustomCourseConfig(claims, key)
}

// Admin

func (a APIsHandler) adminGetNudgesConfig(claims *tokenauth.Claims, params map[string]interface{}) (*model.NudgesConfig, error) {
	return a.app.Admin.GetNudgesConfig(claims)
}

func (a APIsHandler) adminUpdateNudgesConfig(claims *tokenauth.Claims, params map[string]interface{}, item *model.NudgesConfig) (*model.NudgesConfig, error) {
	return a.app.Admin.UpdateNudgesConfig(claims, *item)
}

func (a APIsHandler) adminGetNudges(claims *tokenauth.Claims, params map[string]interface{}) ([]model.Nudge, error) {
	return a.app.Admin.GetNudges(claims)
}

func (a APIsHandler) adminCreateNudge(claims *tokenauth.Claims, params map[string]interface{}, item *model.Nudge) (*model.Nudge, error) {
	return a.app.Admin.CreateNudge(claims, *item)
}

func (a APIsHandler) adminUpdateNudge(claims *tokenauth.Claims, params map[string]interface{}, item *model.Nudge) (*model.Nudge, error) {
	id, err := utils.GetValue[string](params, "id", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	return a.app.Admin.UpdateNudge(claims, id, *item)
}

func (a APIsHandler) adminDeleteNudge(claims *tokenauth.Claims, params map[string]interface{}) error {
	id, err := utils.GetValue[string](params, "id", true)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	return a.app.Admin.DeleteNudge(claims, id)
}

func (a APIsHandler) adminFindSentNudges(claims *tokenauth.Claims, params map[string]interface{}) ([]model.SentNudge, error) {
	nudgeID, err := utils.GetValue[*string](params, "nudge-id", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("nudgeID"), err)
	}

	userID, err := utils.GetValue[*string](params, "user-id", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("userID"), err)
	}

	netID, err := utils.GetValue[*string](params, "net-id", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("netID"), err)
	}

	mode, err := utils.GetValue[*string](params, "mode", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("mode"), err)
	}

	return a.app.Admin.FindSentNudges(claims, nudgeID, userID, netID, mode)
}

func (a APIsHandler) adminDeleteSentNudges(claims *tokenauth.Claims, params map[string]interface{}) error {
	ids, err := utils.GetValue[*string](params, "ids", false)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("ids"), err)
	}

	return a.app.Admin.DeleteSentNudges(claims, ids)
}

func (a APIsHandler) adminClearTestSentNudges(claims *tokenauth.Claims, params map[string]interface{}) error {
	return a.app.Admin.ClearTestSentNudges(claims)
}

func (a APIsHandler) adminFindNudgesProcesses(claims *tokenauth.Claims, params map[string]interface{}) ([]model.NudgesProcess, error) {
	limit, err := utils.GetValue[*int](params, "limit", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("limit"), err)
	}

	offset, err := utils.GetValue[*int](params, "offset", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("offset"), err)
	}

	return a.app.Admin.FindNudgesProcesses(claims, limit, offset)
}

func (a APIsHandler) adminGetCustomCourses(claims *tokenauth.Claims, params map[string]interface{}) ([]model.Course, error) {
	id, err := utils.GetValue[*string](params, "id", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	name, err := utils.GetValue[*string](params, "name", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("name"), err)
	}

	key, err := utils.GetValue[*string](params, "key", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	moduleKey, err := utils.GetValue[*string](params, "module_key", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("moduleKey"), err)
	}

	return a.app.Admin.GetCustomCourses(claims, id, name, key, moduleKey)
}

func (a APIsHandler) adminCreateCustomCourse(claims *tokenauth.Claims, params map[string]interface{}, item *model.Course) (*model.Course, error) {
	return a.app.Admin.CreateCustomCourse(claims, *item)
}

func (a APIsHandler) adminGetCustomCourse(claims *tokenauth.Claims, params map[string]interface{}) (*model.Course, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.GetCustomCourse(claims, key)
}

func (a APIsHandler) adminUpdateCustomCourse(claims *tokenauth.Claims, params map[string]interface{}, item *model.Course) (*model.Course, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.UpdateCustomCourse(claims, key, *item)
}

func (a APIsHandler) adminDeleteCustomCourse(claims *tokenauth.Claims, params map[string]interface{}) error {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.DeleteCustomCourse(claims, key)
}

func (a APIsHandler) adminGetCustomModules(claims *tokenauth.Claims, params map[string]interface{}) ([]model.Module, error) {
	id, err := utils.GetValue[*string](params, "id", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	name, err := utils.GetValue[*string](params, "name", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("name"), err)
	}

	key, err := utils.GetValue[*string](params, "key", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	unitKey, err := utils.GetValue[*string](params, "unit_key", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("unitKey"), err)
	}

	return a.app.Admin.GetCustomModules(claims, id, name, key, unitKey)
}

func (a APIsHandler) adminCreateCustomModule(claims *tokenauth.Claims, params map[string]interface{}, item *model.Module) (*model.Module, error) {
	return a.app.Admin.CreateCustomModule(claims, *item)
}

func (a APIsHandler) adminGetCustomModule(claims *tokenauth.Claims, params map[string]interface{}) (*model.Module, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.GetCustomModule(claims, key)
}

func (a APIsHandler) adminUpdateCustomModule(claims *tokenauth.Claims, params map[string]interface{}, item *model.Module) (*model.Module, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.UpdateCustomModule(claims, key, *item)
}

func (a APIsHandler) adminDeleteCustomModule(claims *tokenauth.Claims, params map[string]interface{}) error {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.DeleteCustomModule(claims, key)
}

func (a APIsHandler) adminGetCustomUnits(claims *tokenauth.Claims, params map[string]interface{}) ([]model.Unit, error) {
	id, err := utils.GetValue[*string](params, "id", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	name, err := utils.GetValue[*string](params, "name", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("name"), err)
	}

	key, err := utils.GetValue[*string](params, "key", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	contentKey, err := utils.GetValue[*string](params, "content_key", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("contentKey"), err)
	}

	return a.app.Admin.GetCustomUnits(claims, id, name, key, contentKey)
}

func (a APIsHandler) adminCreateCustomUnit(claims *tokenauth.Claims, params map[string]interface{}, item *model.Unit) (*model.Unit, error) {
	return a.app.Admin.CreateCustomUnit(claims, *item)
}

func (a APIsHandler) adminGetCustomUnit(claims *tokenauth.Claims, params map[string]interface{}) (*model.Unit, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.GetCustomUnit(claims, key)
}

func (a APIsHandler) adminUpdateCustomUnit(claims *tokenauth.Claims, params map[string]interface{}, item *model.Unit) (*model.Unit, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.UpdateCustomUnit(claims, key, *item)
}

func (a APIsHandler) adminDeleteCustomUnit(claims *tokenauth.Claims, params map[string]interface{}) error {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.DeleteCustomUnit(claims, key)
}

func (a APIsHandler) adminGetCustomContents(claims *tokenauth.Claims, params map[string]interface{}) ([]model.Content, error) {
	id, err := utils.GetValue[*string](params, "id", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("id"), err)
	}

	name, err := utils.GetValue[*string](params, "name", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("name"), err)
	}

	key, err := utils.GetValue[*string](params, "key", false)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.GetCustomContents(claims, id, name, key)
}

func (a APIsHandler) adminCreateCustomContent(claims *tokenauth.Claims, params map[string]interface{}, item *model.Content) (*model.Content, error) {
	return a.app.Admin.CreateCustomContent(claims, *item)
}

func (a APIsHandler) adminGetCustomContent(claims *tokenauth.Claims, params map[string]interface{}) (*model.Content, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.GetCustomContent(claims, key)
}

func (a APIsHandler) adminUpdateCustomContent(claims *tokenauth.Claims, params map[string]interface{}, item *model.Content) (*model.Content, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.UpdateCustomContent(claims, key, *item)
}

func (a APIsHandler) adminDeleteCustomContent(claims *tokenauth.Claims, params map[string]interface{}) error {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.DeleteCustomContent(claims, key)
}

func (a APIsHandler) adminGetCustomCourseConfigs(claims *tokenauth.Claims, params map[string]interface{}) ([]model.CourseConfig, error) {
	return a.app.Admin.GetCustomCourseConfigs(claims)
}

func (a APIsHandler) adminCreateCustomCourseConfig(claims *tokenauth.Claims, params map[string]interface{}, item *model.CourseConfig) (*model.CourseConfig, error) {
	return a.app.Admin.CreateCustomCourseConfig(claims, *item)
}

func (a APIsHandler) adminGetCustomCourseConfig(claims *tokenauth.Claims, params map[string]interface{}) (*model.CourseConfig, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.GetCustomCourseConfig(claims, key)
}

func (a APIsHandler) adminUpdateCustomCourseConfig(claims *tokenauth.Claims, params map[string]interface{}, item *model.CourseConfig) (*model.CourseConfig, error) {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.UpdateCustomCourseConfig(claims, key, *item)
}

func (a APIsHandler) adminDeleteCustomCourseConfig(claims *tokenauth.Claims, params map[string]interface{}) error {
	key, err := utils.GetValue[string](params, "key", true)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, logutils.TypePathParam, logutils.StringArgs("key"), err)
	}

	return a.app.Admin.DeleteCustomCourseConfig(claims, key)
}

// NewAPIsHandler creates new API handler instance
func NewAPIsHandler(app *core.Application) APIsHandler {
	return APIsHandler{app: app}
}
