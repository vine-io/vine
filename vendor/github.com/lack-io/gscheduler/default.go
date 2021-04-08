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

package gscheduler

import "sync"

var (
	defaultScheduler Scheduler
	oncedo           sync.Once
)

func GetJob(id string) (*Job, error) {
	return defaultScheduler.GetJob(id)
}

func GetJobs() ([]*Job, error) {
	return defaultScheduler.GetJobs()
}

func Count() uint64 {
	return defaultScheduler.Count()
}

func AddJob(job *Job) error {
	return defaultScheduler.AddJob(job)
}

func UpdateJob(job *Job) error {
	return defaultScheduler.UpdateJob(job)
}

func RemoveJob(job *Job) error {
	return defaultScheduler.RemoveJob(job)
}

func Start() {
	oncedo.Do(func() {
		defaultScheduler = NewScheduler()
	})
	defaultScheduler.Start()
}

func Stop() {
	defaultScheduler.Stop()
}

func StopGraceful() {
	defaultScheduler.StopGraceful()
}
func Reset() {
	defaultScheduler.Reset()
}
