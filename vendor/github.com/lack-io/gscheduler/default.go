// MIT License
//
// Copyright (c) 2020 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
