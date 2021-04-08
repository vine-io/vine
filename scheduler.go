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

package vine

import "github.com/lack-io/gscheduler"

var defaultScheduler gscheduler.Scheduler

func init() {
	defaultScheduler = gscheduler.NewScheduler()
	defaultScheduler.Start()
}

func GetJob(id string) (*gscheduler.Job, error) {
	return defaultScheduler.GetJob(id)
}

func GetJobs() ([]*gscheduler.Job, error) {
	return defaultScheduler.GetJobs()
}

func AddJob(job *gscheduler.Job) error {
	return defaultScheduler.AddJob(job)
}

func UpdateJob(job *gscheduler.Job) error {
	return defaultScheduler.UpdateJob(job)
}

func RemoveJob(job *gscheduler.Job) error {
	return defaultScheduler.RemoveJob(job)
}
