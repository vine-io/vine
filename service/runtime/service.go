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

package runtime

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/lack-io/vine/service/logger"
	"github.com/lack-io/vine/service/runtime/local/build"
	"github.com/lack-io/vine/service/runtime/local/process"
	proc "github.com/lack-io/vine/service/runtime/local/process/os"
)

type service struct {
	sync.RWMutex

	running bool
	closed  chan bool
	err     error
	updated time.Time

	retries    int
	maxRetries int

	// output for logs
	output io.Writer

	// service to manage
	*Service
	// process creator
	Process *proc.Process
	// Exec
	Exec *process.Executable
	// process pid
	PID *process.PID
}

func newService(s *Service, c CreateOptions) *service {
	var exec string
	var args []string

	// set command
	exec = strings.Join(c.Command, " ")
	args = c.Args

	return &service{
		Service: s,
		Process: new(proc.Process),
		Exec: &process.Executable{
			Package: &build.Package{
				Name: s.Name,
				Path: exec,
			},
			Env:  c.Env,
			Args: args,
			Dir:  s.Source,
		},
		closed:     make(chan bool),
		output:     c.Output,
		updated:    time.Now(),
		maxRetries: c.Retries,
	}
}

func (s *service) streamOutput() {
	go io.Copy(s.output, s.PID.Output)
	go io.Copy(s.output, s.PID.Error)
}

func (s *service) shouldStart() bool {
	if s.running {
		return false
	}
	return s.retries <= s.maxRetries
}

func (s *service) key() string {
	return fmt.Sprintf("%v:%v", s.Name, s.Version)
}

func (s *service) ShouldStart() bool {
	s.RLock()
	defer s.RUnlock()
	return s.shouldStart()
}

func (s *service) Running() bool {
	s.RLock()
	defer s.RUnlock()
	return s.running
}

// Start starts the service
func (s *service) Start() error {
	s.Lock()
	defer s.Unlock()

	if !s.shouldStart() {
		return nil
	}

	// reset
	s.err = nil
	s.closed = make(chan bool)
	s.retries = 0

	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}
	s.Status("starting", nil)

	// TODO: pull source & build binary
	log.Debugf("Runtime service %s forking new process", s.Service.Name)

	p, err := s.Process.Fork(s.Exec)
	if err != nil {
		s.Status("error", err)
		return err
	}
	// set the pid
	s.PID = p
	// set to running
	s.running = true
	// set status
	s.Status("running", nil)
	// set started
	s.Metadata["started"] = time.Now().Format(time.RFC3339)

	if s.output != nil {
		s.streamOutput()
	}

	// wait and watch
	go s.Wait()

	return nil
}

// Status updates the status of the service. Assumes it's called under a lock as it mutates state
func (s *service) Status(status string, err error) {
	s.Metadata["lastStatusUpdate"] = time.Now().Format(time.RFC3339)
	s.Metadata["status"] = status
	if err == nil {
		delete(s.Metadata, "error")
		return
	}
	s.Metadata["error"] = err.Error()

}

// Stop stops the service
func (s *service) Stop() error {
	s.Lock()
	defer s.Unlock()

	select {
	case <-s.closed:
		return nil
	default:
		close(s.closed)
		s.running = false
		s.retries = 0
		if s.PID == nil {
			return nil
		}

		// set status
		s.Status("stopping", nil)

		// kill the process
		err := s.Process.Kill(s.PID)
		if err == nil {
			// wait for it to exit
			s.Process.Wait(s.PID)
		}

		// set status
		s.Status("stopped", err)

		// return the kill error
		return err
	}
}

// Error returns the last error service has returned
func (s *service) Error() error {
	s.RLock()
	defer s.RUnlock()
	return s.err
}

// Wait waits for the service to finish running
func (s *service) Wait() {
	// wait for process to exit
	s.RLock()
	thisPID := s.PID
	s.RUnlock()
	err := s.Process.Wait(thisPID)

	s.Lock()
	defer s.Unlock()

	if s.PID.ID != thisPID.ID {
		// trying to update when it's already been switched out, ignore
		log.Warnf("Trying to update a process status but PID doesn't match. Old %s, New %s. Skipping update.", thisPID.ID, s.PID.ID)
		return
	}

	// save the error
	if err != nil {
		log.Errorf("Service %s terminated with error %s", s.Name, err)
		s.retries++
		s.Status("error", err)
		s.Metadata["retries"] = strconv.Itoa(s.retries)

		s.err = err
	} else {
		s.Status("done", nil)
	}

	// no longer running
	s.running = false
}
