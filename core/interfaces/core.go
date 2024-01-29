// Code generated by api-generator DO NOT EDIT.
package interfaces

import (
	"lms/core/model"

	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
)

// Default exposes default APIs to the driver adapters
type Default interface {
	// string

	GetVersion() (*string, error)
}

// Client exposes client APIs to the driver adapters
type Client interface {
	// model.ProviderCourse

	GetCourses(claims *tokenauth.Claims, courseType *string) ([]model.ProviderCourse, error)
	GetCourse(claims *tokenauth.Claims, id string) (*model.ProviderCourse, error)

	// model.AssignmentGroup

	GetAssignmentGroups(claims *tokenauth.Claims, id string, include *string) ([]model.AssignmentGroup, error)

	// model.User

	GetCourseUser(claims *tokenauth.Claims, id string, include *string) (*model.User, error)
	GetCurrentUser(claims *tokenauth.Claims) (*model.User, error)

	// model.UserCourse

	GetUserCourses(claims *tokenauth.Claims, id *string, name *string, key *string) ([]model.UserCourse, error)
	GetUserCourse(claims *tokenauth.Claims, key string) (*model.UserCourse, error)
	CreateUserCourse(claims *tokenauth.Claims, key string, item model.Timezone) (*model.UserCourse, error)
	DeleteUserCourse(claims *tokenauth.Claims, key string) error
	UpdateUserCourse(claims *tokenauth.Claims, key string, drop *bool) (*model.UserCourse, error)

	// model.UserUnit

	UpdateUserCourseUnitProgress(claims *tokenauth.Claims, courseKey string, unitKey string, item model.UserContentWithTimezone) (*model.UserUnit, error)
	GetUserCourseUnits(claims *tokenauth.Claims, key string) ([]model.UserUnit, error)

	// model.Course

	GetCustomCourses(claims *tokenauth.Claims) ([]model.Course, error)
	GetCustomCourse(claims *tokenauth.Claims, key string) (*model.Course, error)

	// model.CourseConfig

	GetCustomCourseConfig(claims *tokenauth.Claims, key string) (*model.CourseConfig, error)
}

// Admin exposes administrative APIs to the driver adapters
type Admin interface {
	// model.NudgesConfig

	GetNudgesConfig(claims *tokenauth.Claims) (*model.NudgesConfig, error)
	UpdateNudgesConfig(claims *tokenauth.Claims, item model.NudgesConfig) (*model.NudgesConfig, error)

	// model.Nudge

	GetNudges(claims *tokenauth.Claims) ([]model.Nudge, error)
	CreateNudge(claims *tokenauth.Claims, item model.Nudge) (*model.Nudge, error)
	UpdateNudge(claims *tokenauth.Claims, id string, item model.Nudge) (*model.Nudge, error)
	DeleteNudge(claims *tokenauth.Claims, id string) error

	// model.SentNudge

	FindSentNudges(claims *tokenauth.Claims, nudgeID *string, userID *string, netID *string, mode *string) ([]model.SentNudge, error)
	DeleteSentNudges(claims *tokenauth.Claims, ids *string) error
	ClearTestSentNudges(claims *tokenauth.Claims) error

	// model.NudgesProcess

	FindNudgesProcesses(claims *tokenauth.Claims, limit *int, offset *int) ([]model.NudgesProcess, error)

	// model.Course

	GetCustomCourses(claims *tokenauth.Claims, id *string, name *string, key *string, moduleKey *string) ([]model.Course, error)
	CreateCustomCourse(claims *tokenauth.Claims, item model.Course) (*model.Course, error)
	GetCustomCourse(claims *tokenauth.Claims, key string) (*model.Course, error)
	UpdateCustomCourse(claims *tokenauth.Claims, key string, item model.Course) (*model.Course, error)
	DeleteCustomCourse(claims *tokenauth.Claims, key string) error

	// model.Module

	GetCustomModules(claims *tokenauth.Claims, id *string, name *string, key *string, unitKey *string) ([]model.Module, error)
	CreateCustomModule(claims *tokenauth.Claims, item model.Module) (*model.Module, error)
	GetCustomModule(claims *tokenauth.Claims, key string) (*model.Module, error)
	UpdateCustomModule(claims *tokenauth.Claims, key string, item model.Module) (*model.Module, error)
	DeleteCustomModule(claims *tokenauth.Claims, key string) error

	// model.Unit

	GetCustomUnits(claims *tokenauth.Claims, id *string, name *string, key *string, contentKey *string) ([]model.Unit, error)
	CreateCustomUnit(claims *tokenauth.Claims, item model.Unit) (*model.Unit, error)
	GetCustomUnit(claims *tokenauth.Claims, key string) (*model.Unit, error)
	UpdateCustomUnit(claims *tokenauth.Claims, key string, item model.Unit) (*model.Unit, error)
	DeleteCustomUnit(claims *tokenauth.Claims, key string) error

	// model.Content

	GetCustomContents(claims *tokenauth.Claims, id *string, name *string, key *string) ([]model.Content, error)
	CreateCustomContent(claims *tokenauth.Claims, item model.Content) (*model.Content, error)
	GetCustomContent(claims *tokenauth.Claims, key string) (*model.Content, error)
	UpdateCustomContent(claims *tokenauth.Claims, key string, item model.Content) (*model.Content, error)
	DeleteCustomContent(claims *tokenauth.Claims, key string) error

	// model.CourseConfig

	GetCustomCourseConfigs(claims *tokenauth.Claims) ([]model.CourseConfig, error)
	CreateCustomCourseConfig(claims *tokenauth.Claims, item model.CourseConfig) (*model.CourseConfig, error)
	GetCustomCourseConfig(claims *tokenauth.Claims, key string) (*model.CourseConfig, error)
	UpdateCustomCourseConfig(claims *tokenauth.Claims, key string, item model.CourseConfig) (*model.CourseConfig, error)
	DeleteCustomCourseConfig(claims *tokenauth.Claims, key string) error
}
