// Package web provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.11.0 DO NOT EDIT.
package web

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

// Defines values for NudgesConfigMode.
const (
	Normal NudgesConfigMode = "normal"
	Test   NudgesConfigMode = "test"
)

// Assigment defines model for Assigment.
type Assigment struct {
	CourseId *int    `json:"course_id,omitempty"`
	HtmlUrl  *string `json:"html_url,omitempty"`
	Id       *string `json:"id,omitempty"`
	Name     *string `json:"name,omitempty"`
	Position *int    `json:"position,omitempty"`
}

// AssigmentGroup defines model for AssigmentGroup.
type AssigmentGroup struct {
	Assigments *Assigment `json:"assigments,omitempty"`
	Id         *string    `json:"id,omitempty"`
}

// Course defines model for Course.
type Course struct {
	AccessRestrictedByDate *bool   `json:"access_restricted_by_date,omitempty"`
	Id                     *string `json:"id,omitempty"`
	Name                   *string `json:"name,omitempty"`
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

// Nudge defines model for Nudge.
type Nudge struct {
	Active *bool                   `json:"active,omitempty"`
	Body   string                  `json:"body"`
	Id     *string                 `json:"id,omitempty"`
	Name   string                  `json:"name"`
	Params *map[string]interface{} `json:"params"`
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

// User defines model for User.
type User struct {
	Enrollments *Enrollment `json:"enrollments,omitempty"`
	Id          *int        `json:"id,omitempty"`
	Name        *string     `json:"name,omitempty"`
}

// AdminReqCreateNudge defines model for _admin_req_create_nudge.
type AdminReqCreateNudge struct {
	Active   bool                    `json:"active"`
	Body     string                  `json:"body"`
	DeepLink string                  `json:"deep_link"`
	Id       string                  `json:"id"`
	Name     string                  `json:"name"`
	Params   *map[string]interface{} `json:"params"`
}

// AdminReqUpdateNudge defines model for _admin_req_update_nudge.
type AdminReqUpdateNudge struct {
	Active   bool                    `json:"active"`
	Body     string                  `json:"body"`
	DeepLink string                  `json:"deep_link"`
	Name     string                  `json:"name"`
	Params   *map[string]interface{} `json:"params"`
}

// PostAdminNudgesJSONBody defines parameters for PostAdminNudges.
type PostAdminNudgesJSONBody = AdminReqCreateNudge

// PutAdminNudgesConfigJSONBody defines parameters for PutAdminNudgesConfig.
type PutAdminNudgesConfigJSONBody = NudgesConfig

// GetAdminNudgesProcessesParams defines parameters for GetAdminNudgesProcesses.
type GetAdminNudgesProcessesParams struct {
	// The maximum number  to return
	Limit *int `form:"limit,omitempty" json:"limit,omitempty"`

	// The index of the first nudges process to return
	Offset *int `form:"offset,omitempty" json:"offset,omitempty"`
}

// PutAdminNudgesIdJSONBody defines parameters for PutAdminNudgesId.
type PutAdminNudgesIdJSONBody = AdminReqUpdateNudge

// GetAdminNudgesBlocksParams defines parameters for GetAdminNudgesBlocks.
type GetAdminNudgesBlocksParams struct {
	// nudges process ID
	ProcessId *string `form:"process_id,omitempty" json:"process_id,omitempty"`
	Number    *int    `form:"number,omitempty" json:"number,omitempty"`
}

// DeleteAdminSentNudgesParams defines parameters for DeleteAdminSentNudges.
type DeleteAdminSentNudgesParams struct {
	// A comma-separated list of sent_nudge IDs
	Ids *string `form:"ids,omitempty" json:"ids,omitempty"`
}

// GetAdminSentNudgesParams defines parameters for GetAdminSentNudges.
type GetAdminSentNudgesParams struct {
	// nudge_id
	NudgeId *string `form:"nudge-id,omitempty" json:"nudge-id,omitempty"`

	// user_id
	UserId *string `form:"user-id,omitempty" json:"user-id,omitempty"`

	// net_id
	NetId *string `form:"net-id,omitempty" json:"net-id,omitempty"`

	// mode
	Mode *string `form:"mode,omitempty" json:"mode,omitempty"`
}

// GetApiCoursesIdAssignmentGroupsParams defines parameters for GetApiCoursesIdAssignmentGroups.
type GetApiCoursesIdAssignmentGroupsParams struct {
	// include = enrollments,scores
	Include string `form:"include" json:"include"`
}

// GetApiCoursesIdUsersParams defines parameters for GetApiCoursesIdUsers.
type GetApiCoursesIdUsersParams struct {
	// include = enrollments,scores
	Include string `form:"include" json:"include"`
}

// PostAdminNudgesJSONRequestBody defines body for PostAdminNudges for application/json ContentType.
type PostAdminNudgesJSONRequestBody = PostAdminNudgesJSONBody

// PutAdminNudgesConfigJSONRequestBody defines body for PutAdminNudgesConfig for application/json ContentType.
type PutAdminNudgesConfigJSONRequestBody = PutAdminNudgesConfigJSONBody

// PutAdminNudgesIdJSONRequestBody defines body for PutAdminNudgesId for application/json ContentType.
type PutAdminNudgesIdJSONRequestBody = PutAdminNudgesIdJSONBody
