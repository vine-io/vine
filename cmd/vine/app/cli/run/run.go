// MIT License
//
// Copyright (c) 2021 Lack
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

package build

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/vine-io/vine/util/helper"
	"gopkg.in/fsnotify.v1"

	"github.com/vine-io/vine/cmd/vine/app/cli/util/tool"
	signalutil "github.com/vine-io/vine/util/signal"
)

type Runner struct {
	wg   sync.WaitGroup
	tmp  string
	cmd  *exec.Cmd
	args []string
}

func NewRunner(args ...string) *Runner {
	return &Runner{tmp: filepath.Join(os.TempDir(), uuid.New().String()), args: args}
}

func (r *Runner) Run() {
	r.wg.Add(1)
	r.init()
	go func() {
		defer r.wg.Done()
		if err := r.cmd.Start(); err != nil {
			fmt.Printf("vine project started failed: %v\n", err)
		}
		fmt.Printf("vine project running at %d\n", r.cmd.Process.Pid)
		r.cmd.Wait()
	}()
}

func (r *Runner) Kill() error {
	for {
		p := r.cmd.Process
		if p != nil {
			fmt.Printf("kill process: %d\n", p.Pid)
			if err := p.Kill(); err != nil {
				return err
			}
			return r.cmd.Wait()
		}
		time.Sleep(time.Second * 2)
	}
}

func (r *Runner) Wait() error {
	if err := r.Kill(); err != nil {
		return err
	}
	r.wg.Wait()
	return nil
}

func (r *Runner) Stop() error {
	defer os.Remove(r.tmp)
	return r.Wait()
}

func run(c *cobra.Command, cArgs []string) error {
	cfg, err := tool.New("vine.toml")
	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("invalid vine project: %v", err)
	}

	flags := c.PersistentFlags()
	name := helper.NewArgs(cArgs).First()
	watches, _ := flags.GetStringSlice("watch")
	auto, _ := flags.GetBool("auto-restart")
	interval, _ := flags.GetInt64("watch-interval")

	var mod *tool.Mod
	switch cfg.Package.Kind {
	case "cluster":
		if cfg.Mod != nil {
			for _, m := range *cfg.Mod {
				if m.Name == name {
					mod = &m
					break
				}
			}
		}
	case "single":
		if cfg.Pkg != nil && cfg.Pkg.Name == name {
			mod = cfg.Pkg
		}
	default:
		return fmt.Errorf("%s: unknown the kind of vine project", cfg.Package.Kind)
	}

	if mod == nil {
		return fmt.Errorf("not found service: " + name)
	}

	args := []string{mod.Main}
	if len(helper.NewArgs(args).Tail()) != 0 {
		args = append(args, helper.NewArgs(cArgs).Tail()...)
	}

	fmt.Printf("go run %s\n", strings.Join(args, " "))
	runner := NewRunner(args...)

	done := make(chan struct{})
	ech := make(chan fsnotify.Event, 1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("start watching: %v", err)
	}
	defer watcher.Close()

	watchAdd := func(w *fsnotify.Watcher, dirs ...string) {
		for _, d := range dirs {
			_ = filepath.Walk(d, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					_ = w.Add(path)
				}

				return nil
			})
		}
	}

	if auto {
		go func() {
			for {
				select {
				case <-done:
					return
				case e, ok := <-watcher.Events:
					if !ok {
						break
					}

					switch e.Op {
					case fsnotify.Write, fsnotify.Create:
						ech <- e
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						break
					}
					fmt.Printf("watching error: %v", err)
				}
			}
		}()

		switch cfg.Package.Kind {
		case "cluster":
			watchAdd(watcher, "cmd/"+name+"/", "pkg/"+name+"/", "proto/")
		case "single":
			watchAdd(watcher, "cmd/", "pkg/", "proto/")
		}

		if len(watches) > 0 {
			watchAdd(watcher, watches...)
		}
	}

	go func() {
		t := time.Now()
		runner.Run()
		for {
			select {
			case <-done:
				goto EXIT
			case e, ok := <-ech:
				if !ok {
					goto EXIT
				}

				if e.Op == fsnotify.Create && !strings.HasSuffix(e.Name, "~") {
					stat, _ := os.Stat(e.Name)
					if stat != nil && stat.IsDir() {
						watchAdd(watcher, e.Name)
					}
					continue
				}

				now := time.Now()
				if now.Sub(t).Seconds() < float64(interval) {
					continue
				}
				t = now

				if err = runner.Wait(); err != nil {
					fmt.Printf("kill go process faield: %v\n", err)
				}
				fmt.Printf("watching change, restart go binary: go %s\n", strings.Join(args, " "))
				runner.Run()
			}
		}
	EXIT:
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signalutil.Shutdown()...)
	<-ch
	close(ch)
	return runner.Stop()
}

func Commands() []*cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Start a vine project",
		RunE:  run,
	}
	runCmd.PersistentFlags().Bool("auto-restart", true, "auto restart project when code updating")
	runCmd.PersistentFlags().StringSlice("watch", nil, "specify directory in which for watching")
	runCmd.PersistentFlags().Int64("watch-interval", 3, "effective interval when event triggering")
	return []*cobra.Command{runCmd}
}
