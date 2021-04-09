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

// Package os runs processes locally
package os

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/lack-io/vine/service/runtime/local/process"
)

func (p *Process) Exec(exe *process.Executable) error {
	cmd := exec.Command(exe.Package.Path)
	return cmd.Run()
}

func (p *Process) Fork(exe *process.Executable) (*process.PID, error) {
	// create command
	cmd := exec.Command(exe.Package.Path, exe.Args...)
	// set env vars
	cmd.Env = append(cmd.Env, exe.Env...)

	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	er, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	// start the process
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &process.PID{
		ID:     fmt.Sprintf("%d", cmd.Process.Pid),
		Input:  in,
		Output: out,
		Error:  er,
	}, nil
}

func (p *Process) Kill(pid *process.PID) error {
	id, err := strconv.Atoi(pid.ID)
	if err != nil {
		return err
	}

	pr, err := os.FindProcess(id)
	if err != nil {
		return err
	}

	// now kill it
	err = pr.Kill()

	// return the kill error
	return err
}

func (p *Process) Wait(pid *process.PID) error {
	id, err := strconv.Atoi(pid.ID)
	if err != nil {
		return err
	}

	pr, err := os.FindProcess(id)
	if err != nil {
		return err
	}

	ps, err := pr.Wait()
	if err != nil {
		return err
	}

	if ps.Success() {
		return nil
	}

	return fmt.Errorf(ps.String())
}
