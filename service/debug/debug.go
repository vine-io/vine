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

package debug

import (
	"github.com/lack-io/vine/internal/debug/log"
	memLog "github.com/lack-io/vine/internal/debug/log/memory"
	"github.com/lack-io/vine/internal/debug/profile"
	"github.com/lack-io/vine/internal/debug/stats"
	memStats "github.com/lack-io/vine/internal/debug/stats/memory"
	"github.com/lack-io/vine/internal/debug/trace"
	memTrace "github.com/lack-io/vine/internal/debug/trace/memory"
)

var (
	DefaultLog      log.Log         = memLog.NewLog()
	DefaultTracer   trace.Tracer    = memTrace.NewTracer()
	DefaultStats    stats.Stats     = memStats.NewStats()
	DefaultProfiler profile.Profile = nil
)
