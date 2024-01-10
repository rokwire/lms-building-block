// Code generated by api-generator DO NOT EDIT.
package web

import (
	"lms/core/model"
	Def "lms/driver/web/docs/gen"

	"github.com/gorilla/mux"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// apiDataType represents any stored data type that may be read/written by an API
type apiDataType interface {
	model.ProviderCourse | model.Content | model.Unit | model.Course | model.Module | model.UserCourse | model.AssignmentGroup | model.Nudge | model.SentNudge | string | model.NudgesProcess | model.NudgesConfig | model.User
}

// requestDataType represents any data type that may be sent in an API request body
type requestDataType interface {
	Def.AdminReqUpdateCourse | Def.NudgesConfig | Def.AdminReqCreateNudge | Def.AdminReqUpdateUnit | Def.AdminReqUpdateModule | Def.AdminReqUpdateNudge | apiDataType
}

func (a *Adapter) registerHandler(router *mux.Router, pathStr string, method string, tag string, coreFunc string, dataType string, authType interface{},
	requestBody interface{}, conversionFunc interface{}) error {
	authorization, err := a.getAuthHandler(tag, authType)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, "api auth handler", nil, err)
	}

	coreHandler, err := a.getCoreHandler(tag, coreFunc)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionGet, "api core handler", nil, err)
	}

	var convFunc interface{}
	if conversionFunc != nil {
		convFunc, err = a.getConversionFunc(conversionFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionGet, "request body conversion function", nil, err)
		}
	}

	switch dataType {
	case "model.ProviderCourse":
		handler := apiHandler[model.ProviderCourse, model.ProviderCourse]{authorization: authorization, messageDataType: model.TypeProviderCourse}
		err = setCoreHandler[model.ProviderCourse, model.ProviderCourse](&handler, coreHandler, method, tag, coreFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
		}

		router.HandleFunc(pathStr, handleRequest[model.ProviderCourse, model.ProviderCourse](&handler, a.paths, a.logger)).Methods(method)
	case "model.Content":
		handler := apiHandler[model.Content, model.Content]{authorization: authorization, messageDataType: model.TypeContent}
		err = setCoreHandler[model.Content, model.Content](&handler, coreHandler, method, tag, coreFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
		}

		router.HandleFunc(pathStr, handleRequest[model.Content, model.Content](&handler, a.paths, a.logger)).Methods(method)
	case "model.Unit":
		switch requestBody {
		case "#/components/schemas/_admin_req_update_unit":
			convert, ok := convFunc.(func(*tokenauth.Claims, *Def.AdminReqUpdateUnit) (*model.Unit, error))
			if !ok {
				return errors.ErrorData(logutils.StatusInvalid, "request body conversion function", &logutils.FieldArgs{"x-conversion-function": conversionFunc})
			}

			handler := apiHandler[model.Unit, Def.AdminReqUpdateUnit]{authorization: authorization, conversionFunc: convert, messageDataType: model.TypeUnit}
			err = setCoreHandler[model.Unit, Def.AdminReqUpdateUnit](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Unit, Def.AdminReqUpdateUnit](&handler, a.paths, a.logger)).Methods(method)
		default:
			handler := apiHandler[model.Unit, model.Unit]{authorization: authorization, messageDataType: model.TypeUnit}
			err = setCoreHandler[model.Unit, model.Unit](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Unit, model.Unit](&handler, a.paths, a.logger)).Methods(method)
		}
	case "model.Course":
		switch requestBody {
		case "#/components/schemas/_admin_req_update_course":
			convert, ok := convFunc.(func(*tokenauth.Claims, *Def.AdminReqUpdateCourse) (*model.Course, error))
			if !ok {
				return errors.ErrorData(logutils.StatusInvalid, "request body conversion function", &logutils.FieldArgs{"x-conversion-function": conversionFunc})
			}

			handler := apiHandler[model.Course, Def.AdminReqUpdateCourse]{authorization: authorization, conversionFunc: convert, messageDataType: model.TypeCourse}
			err = setCoreHandler[model.Course, Def.AdminReqUpdateCourse](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Course, Def.AdminReqUpdateCourse](&handler, a.paths, a.logger)).Methods(method)
		default:
			handler := apiHandler[model.Course, model.Course]{authorization: authorization, messageDataType: model.TypeCourse}
			err = setCoreHandler[model.Course, model.Course](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Course, model.Course](&handler, a.paths, a.logger)).Methods(method)
		}
	case "model.Module":
		switch requestBody {
		case "#/components/schemas/_admin_req_update_module":
			convert, ok := convFunc.(func(*tokenauth.Claims, *Def.AdminReqUpdateModule) (*model.Module, error))
			if !ok {
				return errors.ErrorData(logutils.StatusInvalid, "request body conversion function", &logutils.FieldArgs{"x-conversion-function": conversionFunc})
			}

			handler := apiHandler[model.Module, Def.AdminReqUpdateModule]{authorization: authorization, conversionFunc: convert, messageDataType: model.TypeModule}
			err = setCoreHandler[model.Module, Def.AdminReqUpdateModule](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Module, Def.AdminReqUpdateModule](&handler, a.paths, a.logger)).Methods(method)
		default:
			handler := apiHandler[model.Module, model.Module]{authorization: authorization, messageDataType: model.TypeModule}
			err = setCoreHandler[model.Module, model.Module](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Module, model.Module](&handler, a.paths, a.logger)).Methods(method)
		}
	case "model.UserCourse":
		handler := apiHandler[model.UserCourse, model.UserCourse]{authorization: authorization, messageDataType: model.TypeUserCourse}
		err = setCoreHandler[model.UserCourse, model.UserCourse](&handler, coreHandler, method, tag, coreFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
		}

		router.HandleFunc(pathStr, handleRequest[model.UserCourse, model.UserCourse](&handler, a.paths, a.logger)).Methods(method)
	case "model.AssignmentGroup":
		handler := apiHandler[model.AssignmentGroup, model.AssignmentGroup]{authorization: authorization, messageDataType: model.TypeAssignmentGroup}
		err = setCoreHandler[model.AssignmentGroup, model.AssignmentGroup](&handler, coreHandler, method, tag, coreFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
		}

		router.HandleFunc(pathStr, handleRequest[model.AssignmentGroup, model.AssignmentGroup](&handler, a.paths, a.logger)).Methods(method)
	case "model.Nudge":
		switch requestBody {
		case "#/components/schemas/_admin_req_create_nudge":
			convert, ok := convFunc.(func(*tokenauth.Claims, *Def.AdminReqCreateNudge) (*model.Nudge, error))
			if !ok {
				return errors.ErrorData(logutils.StatusInvalid, "request body conversion function", &logutils.FieldArgs{"x-conversion-function": conversionFunc})
			}

			handler := apiHandler[model.Nudge, Def.AdminReqCreateNudge]{authorization: authorization, conversionFunc: convert, messageDataType: model.TypeNudge}
			err = setCoreHandler[model.Nudge, Def.AdminReqCreateNudge](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Nudge, Def.AdminReqCreateNudge](&handler, a.paths, a.logger)).Methods(method)
		case "#/components/schemas/_admin_req_update_nudge":
			convert, ok := convFunc.(func(*tokenauth.Claims, *Def.AdminReqUpdateNudge) (*model.Nudge, error))
			if !ok {
				return errors.ErrorData(logutils.StatusInvalid, "request body conversion function", &logutils.FieldArgs{"x-conversion-function": conversionFunc})
			}

			handler := apiHandler[model.Nudge, Def.AdminReqUpdateNudge]{authorization: authorization, conversionFunc: convert, messageDataType: model.TypeNudge}
			err = setCoreHandler[model.Nudge, Def.AdminReqUpdateNudge](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Nudge, Def.AdminReqUpdateNudge](&handler, a.paths, a.logger)).Methods(method)
		default:
			handler := apiHandler[model.Nudge, model.Nudge]{authorization: authorization, messageDataType: model.TypeNudge}
			err = setCoreHandler[model.Nudge, model.Nudge](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.Nudge, model.Nudge](&handler, a.paths, a.logger)).Methods(method)
		}
	case "model.SentNudge":
		handler := apiHandler[model.SentNudge, model.SentNudge]{authorization: authorization, messageDataType: model.TypeSentNudge}
		err = setCoreHandler[model.SentNudge, model.SentNudge](&handler, coreHandler, method, tag, coreFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
		}

		router.HandleFunc(pathStr, handleRequest[model.SentNudge, model.SentNudge](&handler, a.paths, a.logger)).Methods(method)
	case "string":
		handler := apiHandler[string, string]{authorization: authorization, messageDataType: logutils.MessageDataType(dataType)}
		err = setCoreHandler[string, string](&handler, coreHandler, method, tag, coreFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
		}

		router.HandleFunc(pathStr, handleRequest[string, string](&handler, a.paths, a.logger)).Methods(method)
	case "model.NudgesProcess":
		handler := apiHandler[model.NudgesProcess, model.NudgesProcess]{authorization: authorization, messageDataType: model.TypeNudgesProcess}
		err = setCoreHandler[model.NudgesProcess, model.NudgesProcess](&handler, coreHandler, method, tag, coreFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
		}

		router.HandleFunc(pathStr, handleRequest[model.NudgesProcess, model.NudgesProcess](&handler, a.paths, a.logger)).Methods(method)
	case "model.NudgesConfig":
		switch requestBody {
		case "#/components/schemas/NudgesConfig":
			convert, ok := convFunc.(func(*tokenauth.Claims, *Def.NudgesConfig) (*model.NudgesConfig, error))
			if !ok {
				return errors.ErrorData(logutils.StatusInvalid, "request body conversion function", &logutils.FieldArgs{"x-conversion-function": conversionFunc})
			}

			handler := apiHandler[model.NudgesConfig, Def.NudgesConfig]{authorization: authorization, conversionFunc: convert, messageDataType: model.TypeNudgesConfig}
			err = setCoreHandler[model.NudgesConfig, Def.NudgesConfig](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.NudgesConfig, Def.NudgesConfig](&handler, a.paths, a.logger)).Methods(method)
		default:
			handler := apiHandler[model.NudgesConfig, model.NudgesConfig]{authorization: authorization, messageDataType: model.TypeNudgesConfig}
			err = setCoreHandler[model.NudgesConfig, model.NudgesConfig](&handler, coreHandler, method, tag, coreFunc)
			if err != nil {
				return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
			}

			router.HandleFunc(pathStr, handleRequest[model.NudgesConfig, model.NudgesConfig](&handler, a.paths, a.logger)).Methods(method)
		}
	case "model.User":
		handler := apiHandler[model.User, model.User]{authorization: authorization, messageDataType: model.TypeUser}
		err = setCoreHandler[model.User, model.User](&handler, coreHandler, method, tag, coreFunc)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionApply, "api core handler", nil, err)
		}

		router.HandleFunc(pathStr, handleRequest[model.User, model.User](&handler, a.paths, a.logger)).Methods(method)
	default:
		return errors.ErrorData(logutils.StatusInvalid, "data type reference", nil)
	}

	return nil
}

func (a *Adapter) getAuthHandler(tag string, ref interface{}) (tokenauth.Handler, error) {
	if ref == nil {
		return nil, nil
	}

	var handler tokenauth.Handlers
	switch tag {
	case "Client":
		handler = a.auth.client
	case "Admin":
		handler = a.auth.admin
	default:
		return nil, errors.ErrorData(logutils.StatusInvalid, "tag", &logutils.FieldArgs{"tag": tag})
	}

	switch ref {
	case "User":
		return handler.User, nil
	case "Standard":
		return handler.Standard, nil
	case "Authenticated":
		return handler.Authenticated, nil
	case "Permissions":
		return handler.Permissions, nil
	default:
		return nil, errors.ErrorData(logutils.StatusInvalid, "authentication type reference", &logutils.FieldArgs{"ref": ref})
	}
}

func (a *Adapter) getCoreHandler(tag string, ref string) (interface{}, error) {
	switch tag + ref {
	case "DefaultGetVersion":
		return a.apisHandler.defaultGetVersion, nil
	case "ClientGetCourses":
		return a.apisHandler.clientGetCourses, nil
	case "ClientGetCourse":
		return a.apisHandler.clientGetCourse, nil
	case "ClientGetAssignmentGroups":
		return a.apisHandler.clientGetAssignmentGroups, nil
	case "ClientGetCourseUser":
		return a.apisHandler.clientGetCourseUser, nil
	case "ClientGetCurrentUser":
		return a.apisHandler.clientGetCurrentUser, nil
	case "ClientGetUserCourses":
		return a.apisHandler.clientGetUserCourses, nil
	case "ClientGetUserCourse":
		return a.apisHandler.clientGetUserCourse, nil
	case "ClientCreateUserCourse":
		return a.apisHandler.clientCreateUserCourse, nil
	case "ClientDeleteUserCourse":
		return a.apisHandler.clientDeleteUserCourse, nil
	case "ClientUpdateUserCourseUnitProgress":
		return a.apisHandler.clientUpdateUserCourseUnitProgress, nil
	case "AdminGetNudgesConfig":
		return a.apisHandler.adminGetNudgesConfig, nil
	case "AdminUpdateNudgesConfig":
		return a.apisHandler.adminUpdateNudgesConfig, nil
	case "AdminGetNudges":
		return a.apisHandler.adminGetNudges, nil
	case "AdminCreateNudge":
		return a.apisHandler.adminCreateNudge, nil
	case "AdminUpdateNudge":
		return a.apisHandler.adminUpdateNudge, nil
	case "AdminDeleteNudge":
		return a.apisHandler.adminDeleteNudge, nil
	case "AdminFindSentNudges":
		return a.apisHandler.adminFindSentNudges, nil
	case "AdminDeleteSentNudges":
		return a.apisHandler.adminDeleteSentNudges, nil
	case "AdminClearTestSentNudges":
		return a.apisHandler.adminClearTestSentNudges, nil
	case "AdminFindNudgesProcesses":
		return a.apisHandler.adminFindNudgesProcesses, nil
	case "AdminGetCustomCourses":
		return a.apisHandler.adminGetCustomCourses, nil
	case "AdminCreateCustomCourse":
		return a.apisHandler.adminCreateCustomCourse, nil
	case "AdminGetCustomCourse":
		return a.apisHandler.adminGetCustomCourse, nil
	case "AdminUpdateCustomCourse":
		return a.apisHandler.adminUpdateCustomCourse, nil
	case "AdminDeleteCustomCourse":
		return a.apisHandler.adminDeleteCustomCourse, nil
	case "AdminGetCustomModules":
		return a.apisHandler.adminGetCustomModules, nil
	case "AdminCreateCustomModule":
		return a.apisHandler.adminCreateCustomModule, nil
	case "AdminGetCustomModule":
		return a.apisHandler.adminGetCustomModule, nil
	case "AdminUpdateCustomModule":
		return a.apisHandler.adminUpdateCustomModule, nil
	case "AdminDeleteCustomModule":
		return a.apisHandler.adminDeleteCustomModule, nil
	case "AdminGetCustomUnits":
		return a.apisHandler.adminGetCustomUnits, nil
	case "AdminCreateCustomUnit":
		return a.apisHandler.adminCreateCustomUnit, nil
	case "AdminGetCustomUnit":
		return a.apisHandler.adminGetCustomUnit, nil
	case "AdminUpdateCustomUnit":
		return a.apisHandler.adminUpdateCustomUnit, nil
	case "AdminDeleteCustomUnit":
		return a.apisHandler.adminDeleteCustomUnit, nil
	case "AdminGetCustomContents":
		return a.apisHandler.adminGetCustomContents, nil
	case "AdminCreateCustomContent":
		return a.apisHandler.adminCreateCustomContent, nil
	case "AdminGetCustomContent":
		return a.apisHandler.adminGetCustomContent, nil
	case "AdminUpdateCustomContent":
		return a.apisHandler.adminUpdateCustomContent, nil
	case "AdminDeleteCustomContent":
		return a.apisHandler.adminDeleteCustomContent, nil
	default:
		return nil, errors.ErrorData(logutils.StatusInvalid, "core function", logutils.StringArgs(tag+ref))
	}
}

func (a *Adapter) getConversionFunc(ref interface{}) (interface{}, error) {
	if ref == nil {
		return nil, nil
	}

	switch ref {
	case "nudgesConfigFromDef":
		return nudgesConfigFromDef, nil
	case "nudgeFromDefAdminReqCreate":
		return nudgeFromDefAdminReqCreate, nil
	case "nudgeFromDefAdminReqUpdate":
		return nudgeFromDefAdminReqUpdate, nil
	case "customCourseUpdateFromDef":
		return customCourseUpdateFromDef, nil
	case "customModuleUpdateFromDef":
		return customModuleUpdateFromDef, nil
	case "customUnitUpdateFromDef":
		return customUnitUpdateFromDef, nil
	default:
		return nil, errors.ErrorData(logutils.StatusInvalid, "conversion function reference", &logutils.FieldArgs{"ref": ref})
	}
}
