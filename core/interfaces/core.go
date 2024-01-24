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

	"github.com/rokwire/logging-library-go/v2/logs"
)

// Default exposes client APIs for the driver adapters
type Default interface {
	GetVersion() string
}

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string

	GetCourses(l *logs.Log, providerUserID string) ([]model.ProviderCourse, error)
	GetCourse(l *logs.Log, providerUserID string, courseID int) (*model.ProviderCourse, error)
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

// BBs exposes Building Block APIs for the driver adapters
type BBs interface {
}

// TPS exposes third-party service APIs for the driver adapters
type TPS interface {
}

// System exposes system administrative APIs for the driver adapters
type System interface {
}
