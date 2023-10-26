// Package Def provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package Def

import (
	"time"
)

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// Defines values for ContentType.
const (
	ContentTypeAssignment ContentType = "assignment"
	ContentTypeEvaluation ContentType = "evaluation"
	ContentTypeResource   ContentType = "resource"
	ContentTypeReward     ContentType = "reward"
)

// Defines values for NudgesConfigMode.
const (
	Normal NudgesConfigMode = "normal"
	Test   NudgesConfigMode = "test"
)

// Assignment defines model for Assignment.
type Assignment struct {
	CourseId *int    `json:"course_id,omitempty"`
	HtmlUrl  *string `json:"html_url,omitempty"`
	Id       *string `json:"id,omitempty"`
	Name     *string `json:"name,omitempty"`
	Position *int    `json:"position,omitempty"`
}

// AssignmentGroup defines model for AssignmentGroup.
type AssignmentGroup struct {
	Assigments *Assignment `json:"assigments,omitempty"`
	Id         *string     `json:"id,omitempty"`
}

// Content defines model for Content.
type Content struct {
	AppId         *string     `json:"app_id,omitempty"`
	CourseKey     string      `json:"course_key"`
	Details       *string     `json:"details,omitempty"`
	Id            *string     `json:"id,omitempty"`
	Key           string      `json:"key"`
	LinkedContent *[]string   `json:"linked_content"`
	ModuleKey     string      `json:"module_key"`
	Name          string      `json:"name"`
	OrgId         *string     `json:"org_id,omitempty"`
	Reference     Reference   `json:"reference"`
	Type          ContentType `json:"type"`
	UnitKey       string      `json:"unit_key"`
}

// ContentType defines model for Content.Type.
type ContentType string

// Course defines model for Course.
type Course struct {
	AccessRestrictedByDate *bool   `json:"access_restricted_by_date,omitempty"`
	Id                     *string `json:"id,omitempty"`
	Name                   *string `json:"name,omitempty"`
}

// CustomCourse defines model for CustomCourse.
type CustomCourse struct {
	AppId   *string  `json:"app_id,omitempty"`
	Id      *string  `json:"id,omitempty"`
	Key     string   `json:"key"`
	Modules []Module `json:"modules"`
	Name    string   `json:"name"`
	OrgId   *string  `json:"org_id,omitempty"`
}

// Enrollment defines model for Enrollment.
type Enrollment struct {
	Grade *Grade  `json:"grade,omitempty"`
	Id    *int    `json:"id,omitempty"`
	Type  *string `json:"type,omitempty"`
}

// Grade defines model for Grade.
type Grade struct {
	CurrentScore *float32 `json:"current_score,omitempty"`
}

// Module defines model for Module.
type Module struct {
	AppId     *string `json:"app_id,omitempty"`
	CourseKey string  `json:"course_key"`
	Id        *string `json:"id,omitempty"`
	Key       string  `json:"key"`
	Name      string  `json:"name"`
	OrgId     *string `json:"org_id,omitempty"`
	Units     []Unit  `json:"units"`
}

// Nudge defines model for Nudge.
type Nudge struct {
	Active *bool   `json:"active,omitempty"`
	Body   string  `json:"body"`
	Id     *string `json:"id,omitempty"`
	Name   string  `json:"name"`
	Params struct {
		AccountIds *[]int `json:"account_ids,omitempty"`
		CourseIds  *[]int `json:"course_ids,omitempty"`
	} `json:"params"`
	UsersSources *[]UsersSource `json:"users_sources,omitempty"`
}

// NudgesConfig defines model for NudgesConfig.
type NudgesConfig struct {
	Active        bool             `json:"active"`
	BlockSize     *int             `json:"block_size,omitempty"`
	GroupName     string           `json:"group_name"`
	Mode          NudgesConfigMode `json:"mode"`
	ProcessTime   *int             `json:"process_time,omitempty"`
	TestGroupName string           `json:"test_group_name"`
}

// NudgesConfigMode defines model for NudgesConfig.Mode.
type NudgesConfigMode string

// Reference defines model for Reference.
type Reference struct {
	Name         string `json:"name"`
	ReferenceKey string `json:"reference_key"`
	Type         string `json:"type"`
}

// ScheduleItem defines model for ScheduleItem.
type ScheduleItem struct {
	Contents []UserReference `json:"contents"`
	Duration int             `json:"duration"`
	Name     string          `json:"name"`
}

// Unit defines model for Unit.
type Unit struct {
	AppId     *string        `json:"app_id,omitempty"`
	Content   []Content      `json:"content"`
	CourseKey string         `json:"course_key"`
	Id        *string        `json:"id,omitempty"`
	Key       string         `json:"key"`
	ModuleKey string         `json:"module_key"`
	Name      string         `json:"name"`
	OrgId     *string        `json:"org_id,omitempty"`
	Schedule  []ScheduleItem `json:"schedule"`
}

// User defines model for User.
type User struct {
	Enrollments *Enrollment `json:"enrollments,omitempty"`
	Id          *int        `json:"id,omitempty"`
	Name        *string     `json:"name,omitempty"`
}

// UserCourse defines model for UserCourse.
type UserCourse struct {
	AppId  *string      `json:"app_id,omitempty"`
	Course CustomCourse `json:"course"`
	Id     *string      `json:"id,omitempty"`
	OrgId  *string      `json:"org_id,omitempty"`
	UserId *string      `json:"user_id,omitempty"`
}

// UserReference defines model for UserReference.
type UserReference struct {
	DateCompleted *time.Time              `json:"date_completed"`
	DateStarted   time.Time               `json:"date_started"`
	Name          string                  `json:"name"`
	ReferenceKey  string                  `json:"reference_key"`
	Type          string                  `json:"type"`
	UserData      *map[string]interface{} `json:"user_data"`
}

// UsersSource defines model for UsersSource.
type UsersSource struct {
	Params *map[string]interface{} `json:"params"`
	Type   string                  `json:"type"`
}

// AdminReqCreateNudge defines model for _admin_req_create_nudge.
type AdminReqCreateNudge struct {
	Active       bool                   `json:"active"`
	Body         string                 `json:"body"`
	DeepLink     string                 `json:"deep_link"`
	Id           string                 `json:"id"`
	Name         string                 `json:"name"`
	Params       map[string]interface{} `json:"params"`
	UsersSources *[]UsersSource         `json:"users_sources,omitempty"`
}

// AdminReqUpdateNudge defines model for _admin_req_update_nudge.
type AdminReqUpdateNudge struct {
	Active       bool                   `json:"active"`
	Body         string                 `json:"body"`
	DeepLink     string                 `json:"deep_link"`
	Name         string                 `json:"name"`
	Params       map[string]interface{} `json:"params"`
	UsersSources *[]UsersSource         `json:"users_sources,omitempty"`
}

// GetAdminContentParams defines parameters for GetAdminContent.
type GetAdminContentParams struct {
	// Id content ID
	Id *string `form:"id,omitempty" json:"id,omitempty"`

	// Name content name
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Key content key
	Key *string `form:"key,omitempty" json:"key,omitempty"`
}

// GetAdminCoursesParams defines parameters for GetAdminCourses.
type GetAdminCoursesParams struct {
	// Id course ID
	Id *string `form:"id,omitempty" json:"id,omitempty"`

	// Name course name
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Key course key
	Key *string `form:"key,omitempty" json:"key,omitempty"`

	// ModuleId comma separated list of module IDs
	ModuleId *string `form:"module_id,omitempty" json:"module_id,omitempty"`
}

// GetAdminModulesParams defines parameters for GetAdminModules.
type GetAdminModulesParams struct {
	// Id module ID
	Id *string `form:"id,omitempty" json:"id,omitempty"`

	// Name module name
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Key module key
	Key *string `form:"key,omitempty" json:"key,omitempty"`

	// UnitId comma separated list of unit IDs
	UnitId *string `form:"unit_id,omitempty" json:"unit_id,omitempty"`
}

// GetAdminNudgesProcessesParams defines parameters for GetAdminNudgesProcesses.
type GetAdminNudgesProcessesParams struct {
	// Limit The maximum number  to return
	Limit *int `form:"limit,omitempty" json:"limit,omitempty"`

	// Offset The index of the first nudges process to return
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`
}

// DeleteAdminSentNudgesParams defines parameters for DeleteAdminSentNudges.
type DeleteAdminSentNudgesParams struct {
	// Ids A comma-separated list of sent_nudge IDs
	Ids *string `form:"ids,omitempty" json:"ids,omitempty"`
}

// GetAdminSentNudgesParams defines parameters for GetAdminSentNudges.
type GetAdminSentNudgesParams struct {
	// NudgeId nudge_id
	NudgeId *string `form:"nudge-id,omitempty" json:"nudge-id,omitempty"`

	// UserId user_id
	UserId *string `form:"user-id,omitempty" json:"user-id,omitempty"`

	// NetId net_id
	NetId *string `form:"net-id,omitempty" json:"net-id,omitempty"`

	// Mode mode
	Mode *string `form:"mode,omitempty" json:"mode,omitempty"`
}

// GetAdminUnitsParams defines parameters for GetAdminUnits.
type GetAdminUnitsParams struct {
	// Id unit ID
	Id *string `form:"id,omitempty" json:"id,omitempty"`

	// Name unit name
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Key unit key
	Key *string `form:"key,omitempty" json:"key,omitempty"`

	// ContentId comma separated list of content IDs
	ContentId *string `form:"content_id,omitempty" json:"content_id,omitempty"`
}

// GetApiCoursesParams defines parameters for GetApiCourses.
type GetApiCoursesParams struct {
	// CourseType course type
	CourseType *string `form:"course_type,omitempty" json:"course_type,omitempty"`
}

// GetApiCoursesIdAssignmentGroupsParams defines parameters for GetApiCoursesIdAssignmentGroups.
type GetApiCoursesIdAssignmentGroupsParams struct {
	// Include include = assignments,submission
	Include *string `form:"include,omitempty" json:"include,omitempty"`
}

// GetApiCoursesIdUsersParams defines parameters for GetApiCoursesIdUsers.
type GetApiCoursesIdUsersParams struct {
	// Include include = enrollments,scores
	Include *string `form:"include,omitempty" json:"include,omitempty"`
}

// GetApiUsersCoursesParams defines parameters for GetApiUsersCourses.
type GetApiUsersCoursesParams struct {
	// Id course ID
	Id *string `form:"id,omitempty" json:"id,omitempty"`

	// Name course name
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Key course key
	Key *string `form:"key,omitempty" json:"key,omitempty"`
}

// PostAdminContentJSONRequestBody defines body for PostAdminContent for application/json ContentType.
type PostAdminContentJSONRequestBody = Content

// PutAdminContentIdJSONRequestBody defines body for PutAdminContentId for application/json ContentType.
type PutAdminContentIdJSONRequestBody = Content

// PostAdminCoursesJSONRequestBody defines body for PostAdminCourses for application/json ContentType.
type PostAdminCoursesJSONRequestBody = CustomCourse

// PutAdminCoursesIdJSONRequestBody defines body for PutAdminCoursesId for application/json ContentType.
type PutAdminCoursesIdJSONRequestBody = CustomCourse

// PostAdminModulesJSONRequestBody defines body for PostAdminModules for application/json ContentType.
type PostAdminModulesJSONRequestBody = Module

// PutAdminModulesIdJSONRequestBody defines body for PutAdminModulesId for application/json ContentType.
type PutAdminModulesIdJSONRequestBody = Module

// PostAdminNudgesJSONRequestBody defines body for PostAdminNudges for application/json ContentType.
type PostAdminNudgesJSONRequestBody = AdminReqCreateNudge

// PutAdminNudgesConfigJSONRequestBody defines body for PutAdminNudgesConfig for application/json ContentType.
type PutAdminNudgesConfigJSONRequestBody = NudgesConfig

// PutAdminNudgesIdJSONRequestBody defines body for PutAdminNudgesId for application/json ContentType.
type PutAdminNudgesIdJSONRequestBody = AdminReqUpdateNudge

// PostAdminUnitsJSONRequestBody defines body for PostAdminUnits for application/json ContentType.
type PostAdminUnitsJSONRequestBody = Unit

// PutAdminUnitsIdJSONRequestBody defines body for PutAdminUnitsId for application/json ContentType.
type PutAdminUnitsIdJSONRequestBody = Unit

// PutApiUsersCoursesCourseIdUnitUnitIdJSONRequestBody defines body for PutApiUsersCoursesCourseIdUnitUnitId for application/json ContentType.
type PutApiUsersCoursesCourseIdUnitUnitIdJSONRequestBody = ScheduleItem
