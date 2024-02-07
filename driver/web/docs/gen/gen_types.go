// Package Def provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.2 DO NOT EDIT.
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
	NudgesConfigModeNormal NudgesConfigMode = "normal"
	NudgesConfigModeTest   NudgesConfigMode = "test"
)

// Defines values for StreaksNotificationsConfigNotificationsMode.
const (
	StreaksNotificationsConfigNotificationsModeNormal StreaksNotificationsConfigNotificationsMode = "normal"
	StreaksNotificationsConfigNotificationsModeTest   StreaksNotificationsConfigNotificationsMode = "test"
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
	Details       *string     `json:"details,omitempty"`
	Display       Display     `json:"display"`
	Id            *string     `json:"id,omitempty"`
	Key           string      `json:"key"`
	LinkedContent *[]string   `json:"linked_content"`
	Name          string      `json:"name"`
	OrgId         *string     `json:"org_id,omitempty"`
	Reference     Reference   `json:"reference"`
	Type          ContentType `json:"type"`
}

// ContentType defines model for Content.Type.
type ContentType string

// Course defines model for Course.
type Course struct {
	AppId   *string  `json:"app_id,omitempty"`
	Id      *string  `json:"id,omitempty"`
	Key     string   `json:"key"`
	Modules []Module `json:"modules"`
	Name    string   `json:"name"`
	OrgId   *string  `json:"org_id,omitempty"`
}

// CourseConfig defines model for CourseConfig.
type CourseConfig struct {
	AppId                      string                     `json:"app_id"`
	CourseKey                  string                     `json:"course_key"`
	Id                         *string                    `json:"id,omitempty"`
	InitialPauses              int                        `json:"initial_pauses"`
	MaxPauses                  int                        `json:"max_pauses"`
	OrgId                      string                     `json:"org_id"`
	PauseRewardStreak          int                        `json:"pause_reward_streak"`
	StreaksNotificationsConfig StreaksNotificationsConfig `json:"streaks_notifications_config"`
}

// Display defines model for Display.
type Display struct {
	AccentColor     *string `json:"accent_color,omitempty"`
	CompleteColor   *string `json:"complete_color,omitempty"`
	Icon            *string `json:"icon,omitempty"`
	IncompleteColor *string `json:"incomplete_color,omitempty"`
	PrimaryColor    *string `json:"primary_color,omitempty"`
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
	AppId   *string `json:"app_id,omitempty"`
	Display Display `json:"display"`
	Id      *string `json:"id,omitempty"`
	Key     string  `json:"key"`
	Name    string  `json:"name"`
	OrgId   *string `json:"org_id,omitempty"`
	Units   []Unit  `json:"units"`
}

// Notification defines model for Notification.
type Notification struct {
	Active       bool                   `json:"active"`
	Body         string                 `json:"body"`
	ProcessTime  int                    `json:"process_time"`
	Requirements map[string]interface{} `json:"requirements"`
	Subject      string                 `json:"subject"`
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

// ProviderCourse defines model for ProviderCourse.
type ProviderCourse struct {
	AccessRestrictedByDate *bool   `json:"access_restricted_by_date,omitempty"`
	Id                     *string `json:"id,omitempty"`
	Name                   *string `json:"name,omitempty"`
}

// Reference defines model for Reference.
type Reference struct {
	Name         string `json:"name"`
	ReferenceKey string `json:"reference_key"`
	Type         string `json:"type"`
}

// ScheduleItem defines model for ScheduleItem.
type ScheduleItem struct {
	DateCompleted *time.Time    `json:"date_completed"`
	DateStarted   *time.Time    `json:"date_started,omitempty"`
	Duration      int           `json:"duration"`
	Name          string        `json:"name"`
	UserContent   []UserContent `json:"user_content"`
}

// StreaksNotificationsConfig defines model for StreaksNotificationsConfig.
type StreaksNotificationsConfig struct {
	Notifications       []Notification                              `json:"notifications"`
	NotificationsActive bool                                        `json:"notifications_active"`
	NotificationsMode   StreaksNotificationsConfigNotificationsMode `json:"notifications_mode"`
	PreferEarly         bool                                        `json:"prefer_early"`
	StreaksProcessTime  int                                         `json:"streaks_process_time"`
	TimerDelayTolerance *int                                        `json:"timer_delay_tolerance,omitempty"`
	TimezoneName        string                                      `json:"timezone_name"`
	TimezoneOffset      *int                                        `json:"timezone_offset,omitempty"`
}

// StreaksNotificationsConfigNotificationsMode defines model for StreaksNotificationsConfig.NotificationsMode.
type StreaksNotificationsConfigNotificationsMode string

// Timezone defines model for Timezone.
type Timezone struct {
	TimezoneName   string `json:"timezone_name"`
	TimezoneOffset int    `json:"timezone_offset"`
}

// Unit defines model for Unit.
type Unit struct {
	AppId         *string        `json:"app_id,omitempty"`
	Content       []Content      `json:"content"`
	Id            *string        `json:"id,omitempty"`
	Key           string         `json:"key"`
	Name          string         `json:"name"`
	OrgId         *string        `json:"org_id,omitempty"`
	Schedule      []ScheduleItem `json:"schedule"`
	ScheduleStart *int           `json:"schedule_start,omitempty"`
}

// User defines model for User.
type User struct {
	Enrollments *Enrollment `json:"enrollments,omitempty"`
	Id          *int        `json:"id,omitempty"`
	Name        *string     `json:"name,omitempty"`
}

// UserContent defines model for UserContent.
type UserContent struct {
	ContentKey string                  `json:"content_key"`
	UserData   *map[string]interface{} `json:"user_data"`
}

// UserContentWithTimezone defines model for UserContentWithTimezone.
type UserContentWithTimezone struct {
	TimezoneName   string      `json:"timezone_name"`
	TimezoneOffset int         `json:"timezone_offset"`
	UserContent    UserContent `json:"user_content"`
}

// UserCourse defines model for UserCourse.
type UserCourse struct {
	AppId          *string     `json:"app_id,omitempty"`
	Course         Course      `json:"course"`
	DateDropped    *time.Time  `json:"date_dropped,omitempty"`
	Id             *string     `json:"id,omitempty"`
	OrgId          *string     `json:"org_id,omitempty"`
	PauseUses      []time.Time `json:"pause_uses"`
	Pauses         int         `json:"pauses"`
	Streak         int         `json:"streak"`
	StreakResets   []time.Time `json:"streak_resets"`
	StreakRestarts []time.Time `json:"streak_restarts"`
	TimezoneName   string      `json:"timezone_name"`
	TimezoneOffset int         `json:"timezone_offset"`
	UserId         *string     `json:"user_id,omitempty"`
}

// UserUnit defines model for UserUnit.
type UserUnit struct {
	AppId         *string    `json:"app_id,omitempty"`
	Completed     int        `json:"completed"`
	CourseKey     *string    `json:"course_key,omitempty"`
	Current       bool       `json:"current"`
	DateCreated   *time.Time `json:"date_created,omitempty"`
	DateUpdated   *time.Time `json:"date_updated,omitempty"`
	Id            *string    `json:"id,omitempty"`
	LastCompleted *time.Time `json:"last_completed,omitempty"`
	ModuleKey     *string    `json:"module_key,omitempty"`
	OrgId         *string    `json:"org_id,omitempty"`
	Unit          Unit       `json:"unit"`
	UserId        *string    `json:"user_id,omitempty"`
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

// AdminReqUpdateCourse defines model for _admin_req_update_course.
type AdminReqUpdateCourse struct {
	ModuleKeys []string `json:"module_keys"`
	Name       string   `json:"name"`
}

// AdminReqUpdateModule defines model for _admin_req_update_module.
type AdminReqUpdateModule struct {
	Name     string   `json:"name"`
	UnitKeys []string `json:"unit_keys"`
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

// AdminReqUpdateUnit defines model for _admin_req_update_unit.
type AdminReqUpdateUnit struct {
	ContentKeys []string       `json:"content_keys"`
	Name        string         `json:"name"`
	Schedule    []ScheduleItem `json:"schedule"`
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

	// ModuleKey comma separated list of module keys
	ModuleKey *string `form:"module_key,omitempty" json:"module_key,omitempty"`
}

// GetAdminModulesParams defines parameters for GetAdminModules.
type GetAdminModulesParams struct {
	// Id module ID
	Id *string `form:"id,omitempty" json:"id,omitempty"`

	// Name module name
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Key module key
	Key *string `form:"key,omitempty" json:"key,omitempty"`

	// UnitKey comma separated list of unit IDs
	UnitKey *string `form:"unit_key,omitempty" json:"unit_key,omitempty"`
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

	// ContentKey comma separated list of content keys
	ContentKey *string `form:"content_key,omitempty" json:"content_key,omitempty"`
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
	// Id userCourse ID
	Id *string `form:"id,omitempty" json:"id,omitempty"`

	// Name course name
	Name *string `form:"name,omitempty" json:"name,omitempty"`

	// Key course key
	Key *string `form:"key,omitempty" json:"key,omitempty"`
}

// PutApiUsersCoursesKeyParams defines parameters for PutApiUsersCoursesKey.
type PutApiUsersCoursesKeyParams struct {
	// Drop whether to drop the course
	Drop *bool `form:"drop,omitempty" json:"drop,omitempty"`
}

// PostAdminContentJSONRequestBody defines body for PostAdminContent for application/json ContentType.
type PostAdminContentJSONRequestBody = Content

// PutAdminContentKeyJSONRequestBody defines body for PutAdminContentKey for application/json ContentType.
type PutAdminContentKeyJSONRequestBody = Content

// PostAdminCourseConfigsJSONRequestBody defines body for PostAdminCourseConfigs for application/json ContentType.
type PostAdminCourseConfigsJSONRequestBody = CourseConfig

// PutAdminCourseConfigsKeyJSONRequestBody defines body for PutAdminCourseConfigsKey for application/json ContentType.
type PutAdminCourseConfigsKeyJSONRequestBody = CourseConfig

// PostAdminCoursesJSONRequestBody defines body for PostAdminCourses for application/json ContentType.
type PostAdminCoursesJSONRequestBody = Course

// PutAdminCoursesKeyJSONRequestBody defines body for PutAdminCoursesKey for application/json ContentType.
type PutAdminCoursesKeyJSONRequestBody = AdminReqUpdateCourse

// PostAdminModulesJSONRequestBody defines body for PostAdminModules for application/json ContentType.
type PostAdminModulesJSONRequestBody = Module

// PutAdminModulesKeyJSONRequestBody defines body for PutAdminModulesKey for application/json ContentType.
type PutAdminModulesKeyJSONRequestBody = AdminReqUpdateModule

// PostAdminNudgesJSONRequestBody defines body for PostAdminNudges for application/json ContentType.
type PostAdminNudgesJSONRequestBody = AdminReqCreateNudge

// PutAdminNudgesConfigJSONRequestBody defines body for PutAdminNudgesConfig for application/json ContentType.
type PutAdminNudgesConfigJSONRequestBody = NudgesConfig

// PutAdminNudgesIdJSONRequestBody defines body for PutAdminNudgesId for application/json ContentType.
type PutAdminNudgesIdJSONRequestBody = AdminReqUpdateNudge

// PostAdminUnitsJSONRequestBody defines body for PostAdminUnits for application/json ContentType.
type PostAdminUnitsJSONRequestBody = Unit

// PutAdminUnitsKeyJSONRequestBody defines body for PutAdminUnitsKey for application/json ContentType.
type PutAdminUnitsKeyJSONRequestBody = AdminReqUpdateUnit

// PutApiUsersCoursesCourseKeyModulesModuleKeyJSONRequestBody defines body for PutApiUsersCoursesCourseKeyModulesModuleKey for application/json ContentType.
type PutApiUsersCoursesCourseKeyModulesModuleKeyJSONRequestBody = UserContentWithTimezone

// PostApiUsersCoursesKeyJSONRequestBody defines body for PostApiUsersCoursesKey for application/json ContentType.
type PostApiUsersCoursesKeyJSONRequestBody = Timezone
