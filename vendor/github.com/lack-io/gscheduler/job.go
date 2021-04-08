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

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/lack-io/gscheduler/cron"
	"github.com/lack-io/gscheduler/rbtree"
)

type Status string

const (
	Waiting Status = "waiting"
	Running Status = "running"
)

type Job struct {
	id          string
	name        string
	cron        cron.Crontab
	createTime  time.Time
	lastTime    time.Time
	nextTime    time.Time
	isActive    bool
	activeCount uint64
	activeMax   uint64
	status      Status
	err         error
	fn          func()
}

func (j *Job) Less(another rbtree.Item) bool {
	item, ok := another.(*Job)
	if !ok {
		return false
	}
	if !j.nextTime.Equal(item.nextTime) {
		return j.nextTime.Before(item.nextTime)
	}
	if j.lastTime.UnixNano() == 0 {
		j.lastTime = time.Now()
	}
	at, bt := j.cron.Next(j.lastTime), item.cron.Next(item.lastTime)
	if !at.Equal(bt) {
		return at.Before(bt)
	}
	if j.name != item.name {
		return j.name < item.name
	}
	return j.id < item.id
}

func (j Job) ID() string {
	return j.id
}

func (j Job) Name() string {
	return j.name
}

func (j Job) LastTime() time.Time {
	return j.lastTime
}

func (j Job) NextTime() time.Time {
	return j.nextTime
}

func (j *Job) SetCron(cron cron.Crontab) {
	j.cron = cron
}

func (j *Job) SetFn(fn func()) {
	j.fn = fn
}

func (j *Job) SetTimes(t uint64) {
	j.activeMax = t
}

func (j *Job) Start() {
	j.start(true)
}

func (j *Job) start(async bool) {
	j.status = Running
	j.activeCount++
	if async {
		go j.safeCall()
	} else {
		j.safeCall()
	}
}

func (j *Job) safeCall() {
	defer func() {
		if err := recover(); err != nil {
			j.err = fmt.Errorf("[%s] %s: %v", time.Now().Format("2006-01-02 15:04:05"), j.name, err)
		}
	}()
	j.fn()
}

type builder struct {
	j   *Job
	err error
}

// JobBuilder the builder of Job
//  examples:
//   job, err := JobBuilder().Name("cron-job").Spec("*/10 * * * * * *").Out()
//
//   job, err := JobBuilder().Name("delay-job").Delay(time.Now().Add(time.Hour*3)).Out()
//
//   job, err := JobBuilder().Name("duration-job").Duration(time.Second*10).Out()
//
//   job, err := JobBuilder().Name("once-job").Duration(time.Second*5).Times(1).Out()
func JobBuilder() *builder {
	return &builder{j: &Job{
		id:         uuid.New().String(),
		createTime: time.Now(),
		isActive:   true,
		status:     Waiting,
	}}
}

// FromJob get build of Job
func FromJob(job *Job) *builder {
	return &builder{
		j: job,
	}
}

// Name set the name of Job
func (b *builder) Name(name string) *builder {
	b.j.name = name
	return b
}

// Duration set the interval of Job
func (b *builder) Duration(d time.Duration) *builder {
	b.j.SetCron(cron.Every(d))
	return b
}

// Spec set the crontab expression of Job
//  */3 * * * * * * : every 3s
//  00 30 15 * * * * : 15:30:00 every day
func (b *builder) Spec(spec string) *builder {
	b.j.cron, b.err = cron.Parse(spec)
	return b
}

// Delay get a delay once Job
func (b *builder) Delay(t time.Time) *builder {
	b.j.activeMax = 1
	b.j.SetCron(cron.Every(t.Sub(time.Now())))
	return b
}

// Silent set isActive false
func (b *builder) Silent() *builder {
	b.j.isActive = false
	return b
}

// Times set the active count of Job
func (b *builder) Times(t uint64) *builder {
	b.j.SetTimes(t)
	return b
}

// Fn set the function of Job
func (b *builder) Fn(fn func()) *builder {
	b.j.SetFn(fn)
	return b
}

// Out get a Job
func (b *builder) Out() (*Job, error) {
	return b.j, b.err
}
