package gscheduler

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

const _UNTOUCHED = time.Duration(math.MaxInt64)

var signal = struct{}{}

type Scheduler interface {
	GetJobs() ([]*Job, error)
	GetJob(id string) (*Job, error)
	AddJob(job *Job) error
	Count() uint64
	UpdateJob(job *Job) error
	RemoveJob(job *Job) error
	Start()
	Stop()
	StopGraceful()
	Reset()
}

type scheduler struct {
	store       Store
	count       uint64
	waitJobsNum uint64
	pauseChan   chan struct{}
	resumeChan  chan struct{}
	exitChan    chan struct{}
}

func NewScheduler() *scheduler {
	s := &scheduler{
		store:      newJobStore(),
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
		exitChan:   make(chan struct{}),
	}
	return s
}

func (s *scheduler) Start() {
	//开启守护协程
	go s.progress()
	s.resume()
}

// 开启任务调度器
func (s *scheduler) progress() {
	var (
		timeout time.Duration
		job     *Job
		err     error
		timer   = newSafeTimer(_UNTOUCHED)
	)
	defer timer.Stop()

Pause:
	<-s.resumeChan
	for {
		job, err = s.store.Min()
		if err != nil {
			timeout = _UNTOUCHED
		} else {
			timeout = job.nextTime.Sub(time.Now())
		}
		timer.SafeReset(timeout)

		select {
		case <-timer.C:
			timer.SCR()
			atomic.AddUint64(&s.count, 1)

			if job.isActive {
				job.start(true)
				if job.activeMax == 0 || job.activeMax > job.activeCount {
					s.removeJob(job)
					job.lastTime = time.Now()
					job.nextTime = job.cron.Next(job.lastTime)
					s.addJob(job)
				} else {
					s.removeJob(job)
				}
			} else {
				s.removeJob(job)
				job.nextTime = job.cron.Next(time.Now())
				s.addJob(job)
			}

		case <-s.pauseChan:
			goto Pause

		case <-s.exitChan:
			goto Exit
		}
	}
Exit:
}

func (s *scheduler) pause() {
	s.pauseChan <- signal
}

func (s *scheduler) resume() {
	s.resumeChan <- signal
}

func (s *scheduler) exit() {
	s.exitChan <- signal
}

func (s *scheduler) addJob(job *Job) error {
	if job.cron == nil {
		return fmt.Errorf("must set cron")
	}
	if job.fn == nil {
		job.fn = func() {}
	}
	job.status = Waiting
	atomic.AddUint64(&s.waitJobsNum, 1)
	job.nextTime = job.cron.Next(time.Now())
	return s.store.Put(job)
}

func (s *scheduler) removeJob(job *Job) error {
	err := s.store.Del(job)
	atomic.SwapUint64(&s.waitJobsNum, s.waitJobsNum-1)
	return err
}

func (s *scheduler) cleanJobs() {
	for {
		if item, _ := s.store.Min(); item != nil {
			_ = s.removeJob(item)
		} else {
			break
		}
	}
}

func (s *scheduler) immediate() {
	for {
		if item, _ := s.store.Min(); item != nil {
			atomic.AddUint64(&s.count, 1)
			item.start(false)
			_ = s.removeJob(item)
		} else {
			break
		}
	}
}

func (s *scheduler) AddJob(job *Job) error {
	s.pause()
	defer s.resume()
	return s.addJob(job)
}

func (s *scheduler) UpdateJob(job *Job) error {
	s.pause()
	defer s.resume()
	if err := s.removeJob(job); err != nil {
		return err
	}
	return s.addJob(job)
}

func (s *scheduler) RemoveJob(job *Job) error {
	s.pause()
	defer s.resume()
	return s.removeJob(job)
}

func (s *scheduler) GetJobs() ([]*Job, error) {
	return s.store.GetJobs()
}

func (s *scheduler) GetJob(id string) (*Job, error) {
	return s.store.GetById(id)
}

// Count 已经执行的任务数。对于重复任务，会计算多次
func (s *scheduler) Count() uint64 {
	return atomic.LoadUint64(&s.count)
}

// Reset 重置Clock的内部状态
func (s *scheduler) Reset() {
	s.exit()
	s.count = 0

	s.cleanJobs()
	s.Start()
	return
}

// Stop stop clock , and cancel all waiting jobs
func (s *scheduler) Stop() {
	s.exit()

	s.cleanJobs()
}

// StopGraceful stop clock ,and do once every waiting job including Once\Reapeat
// Note:对于任务队列中，即使安排执行多次或者不限次数的，也仅仅执行一次。
func (s *scheduler) StopGraceful() {
	s.exit()

	s.immediate()
}
