// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	ver "github.com/hashicorp/go-version"
)

var (
	// populated by ldflags
	GitCommit string
	GitTag    string
	BuildDate string

	version    = "v1.0.0"
	prerelease = "" // blank if full release
)

func buildVersion() string {
	verStr := version
	if prerelease != "" {
		verStr = fmt.Sprintf("%s-%s", version, prerelease)
	}

	// check for git tag via ldflags
	if len(GitTag) > 0 {
		verStr = GitTag
	}

	// make sure we fail fast (panic) if bad version - this will get caught in CI tests
	ver.Must(ver.NewVersion(verStr))
	return verStr
}
