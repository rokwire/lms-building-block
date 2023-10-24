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
)

// Administration exposes APIs for the driver adapters
type Administration interface {
	GetNudgesConfig() (*model.NudgesConfig, error)
	UpdateNudgesConfig(active bool, groupName string, testGroupName string, mode string, processTime *int, blockSize *int) error

	GetNudges() ([]model.Nudge, error)
	CreateNudge(ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSource) error
	UpdateNudge(ID string, name string, body string, deepLink string, params model.NudgeParams, active bool, usersSourse []model.UsersSources) error
	DeleteNudge(ID string) error

	FindSentNudges(nudgeID *string, userID *string, netID *string, mode *string) ([]model.SentNudge, error)
	DeleteSentNudges(ids []string) error
	ClearTestSentNudges() error

	FindNudgesProcesses(limit int, offset int) ([]model.NudgesProcess, error)
}
