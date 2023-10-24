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
}
