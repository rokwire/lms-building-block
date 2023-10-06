// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interfaces

import (
	"lms/core/model"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"
)

// Default exposes client APIs for the driver adapters
type Default interface {
	GetVersion() string
}

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	GetCourses(l *logs.Log, providerUserID string) ([]model.Course, error)
	GetCourse(l *logs.Log, providerUserID string, courseID int) (*model.Course, error)
	GetAssignmentGroups(l *logs.Log, providerUserID string, courseID int, includeAssignments bool, includeSubmission bool) ([]model.AssignmentGroup, error)
	GetCourseUser(l *logs.Log, providerUserID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error)
	GetCurrentUser(l *logs.Log, providerUserID string) (*model.User, error)
}

// Administration exposes APIs for the driver adapters
type Administration interface {
	GetNudgesConfig(l *logs.Log) (*model.NudgesConfig, error)
	UpdateNudgesConfig(l *logs.Log, active bool, groupName string, testGroupName string, mode string, processTime *int, blockSize *int) error

	GetNudges() ([]model.Nudge, error)
	CreateNudge(l *logs.Log, ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSource) error
	UpdateNudge(l *logs.Log, ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSources) error
	DeleteNudge(l *logs.Log, ID string) error

	FindSentNudges(l *logs.Log, nudgeID *string, userID *string, netID *string, mode *string) ([]model.SentNudge, error)
	DeleteSentNudges(l *logs.Log, ids []string) error
	ClearTestSentNudges(l *logs.Log) error

	FindNudgesProcesses(l *logs.Log, limit int, offset int) ([]model.NudgesProcess, error)
}

// Provider interface for LMS provider
type Provider interface {
	GetCourses(userID string) ([]model.Course, error)
	GetCourse(userID string, courseID int) (*model.Course, error)
	GetCourseUsers(courseID int) ([]model.User, error)
	GetAssignmentGroups(userID string, courseID int, includeAssignments bool, includeSubmission bool) ([]model.AssignmentGroup, error)
	GetCourseUser(userID string, courseID int, includeEnrolments bool, includeScores bool) (*model.User, error)
	GetCurrentUser(userID string) (*model.User, error)

	FindUsersByCanvasUserID(canvasUserIds []int) ([]ProviderUser, error)

	CacheCommonData(usersIDs map[string]string) error
	FindCachedData(usersIDs []string) ([]ProviderUser, error)
	CacheUserData(user ProviderUser) (*ProviderUser, error)
	CacheUserCoursesData(user ProviderUser, coursesIDs []int) (*ProviderUser, error)

	GetLastLogin(userID string) (*time.Time, error)
	GetMissedAssignments(userID string) ([]model.Assignment, error)
	GetCompletedAssignments(userID string) ([]model.Assignment, error)
	GetCalendarEvents(netID string, providerUserID int, courseID int, startAt time.Time, endAt time.Time) ([]model.CalendarEvent, error)
}

// ProviderUser cache entity
type ProviderUser struct {
	ID       string     `bson:"_id"`    //core BB account id
	NetID    string     `bson:"net_id"` //core BB external system id
	User     model.User `bson:"user"`
	SyncDate time.Time  `bson:"sync_date"`

	Courses *UserCourses `bson:"courses"`
}

// UserCourses cache entity
type UserCourses struct {
	Data     []UserCourse `bson:"data"`
	SyncDate time.Time    `bson:"sync_date"`
}

// UserCourse cache entity
type UserCourse struct {
	Data        model.Course       `bson:"data"`
	Assignments []CourseAssignment `bson:"assignments"`
	SyncDate    time.Time          `bson:"sync_date"`
}

// CourseAssignment cache entity
type CourseAssignment struct {
	Data       model.Assignment `bson:"data"`
	Submission *Submission      `bson:"submission"`
	SyncDate   time.Time        `bson:"sync_date"`
}

// Submission cache entity
type Submission struct {
	Data     *model.Submission `bson:"data"`
	SyncDate time.Time         `bson:"sync_date"`
}

// GroupsBB interface for the Groups building block communication
type GroupsBB interface {
	GetUsers(groupName string, offset int, limit int) ([]GroupsBBUser, error)
}

// GroupsBBUser entity
type GroupsBBUser struct {
	UserID string `json:"user_id"`
	NetID  string `json:"net_id"`
	Name   string `json:"name"`
}

// NotificationsBB interface for the Notifications building block communication
type NotificationsBB interface {
	SendNotifications(recipients []Recipient, text string, body string, data map[string]string) error
}

// Recipient entity
type Recipient struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

// BBs exposes Building Block APIs for the driver adapters
type BBs interface {
}

// TPS exposes third-party service APIs for the driver adapters
type TPS interface {
}

// System exposes system administrative APIs for the driver adapters
type System interface {
}
